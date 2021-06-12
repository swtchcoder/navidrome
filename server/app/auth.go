package app

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/deluan/rest"
	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/navidrome/navidrome/conf"
	"github.com/navidrome/navidrome/consts"
	"github.com/navidrome/navidrome/core/auth"
	"github.com/navidrome/navidrome/log"
	"github.com/navidrome/navidrome/model"
	"github.com/navidrome/navidrome/model/request"
	"github.com/navidrome/navidrome/utils/gravatar"
)

var (
	ErrFirstTime = errors.New("no users created")
)

func Login(ds model.DataStore) func(w http.ResponseWriter, r *http.Request) {
	auth.Init(ds)

	return func(w http.ResponseWriter, r *http.Request) {
		username, password, err := getCredentialsFromBody(r)
		if err != nil {
			log.Error(r, "Parsing request body", err)
			_ = rest.RespondWithError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}

		handleLogin(ds, username, password, w, r)
	}
}

func handleLoginFromHeaders(ds model.DataStore, r *http.Request) *map[string]interface{} {
	if !validateIPAgainstList(r.RemoteAddr, conf.Server.ReverseProxyWhitelist) {
		log.Warn("Ip is not whitelisted for reverse proxy login", "ip", r.RemoteAddr)
		return nil
	}

	username := r.Header.Get(conf.Server.ReverseProxyUserHeader)

	userRepo := ds.User(r.Context())
	user, err := userRepo.FindByUsername(username)
	if user == nil || err != nil {
		log.Warn("User passed in header not found", "user", username)
		return nil
	}

	err = userRepo.UpdateLastLoginAt(user.ID)
	if err != nil {
		log.Error("Could not update LastLoginAt", "user", username, err)
		return nil
	}

	tokenString, err := auth.CreateToken(user)
	if err != nil {
		log.Error("Could not create token", "user", username, err)
		return nil
	}

	payload := buildPayload(user, tokenString)

	bytes := make([]byte, 3)
	_, err = rand.Read(bytes)
	if err != nil {
		log.Error("Could not create subsonic salt", "user", username, err)
		return nil
	}
	salt := hex.EncodeToString(bytes)
	payload["subsonicSalt"] = salt

	h := md5.New()
	_, err = io.WriteString(h, user.Password+salt)
	if err != nil {
		log.Error("Could not create subsonic token", "user", username, err)
		return nil
	}
	payload["subsonicToken"] = hex.EncodeToString(h.Sum(nil))

	return &payload
}

func handleLogin(ds model.DataStore, username string, password string, w http.ResponseWriter, r *http.Request) {
	user, err := validateLogin(ds.User(r.Context()), username, password)
	if err != nil {
		_ = rest.RespondWithError(w, http.StatusInternalServerError, "Unknown error authentication user. Please try again")
		return
	}
	if user == nil {
		log.Warn(r, "Unsuccessful login", "username", username, "request", r.Header)
		_ = rest.RespondWithError(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	tokenString, err := auth.CreateToken(user)
	if err != nil {
		_ = rest.RespondWithError(w, http.StatusInternalServerError, "Unknown error authenticating user. Please try again")
		return
	}
	payload := buildPayload(user, tokenString)
	_ = rest.RespondWithJSON(w, http.StatusOK, payload)
}

func buildPayload(user *model.User, tokenString string) map[string]interface{} {
	payload := map[string]interface{}{
		"message":  "User '" + user.UserName + "' authenticated successfully",
		"token":    tokenString,
		"id":       user.ID,
		"name":     user.Name,
		"username": user.UserName,
		"isAdmin":  user.IsAdmin,
	}
	if conf.Server.EnableGravatar && user.Email != "" {
		payload["avatar"] = gravatar.Url(user.Email, 50)
	}
	return payload
}

func validateIPAgainstList(ip string, comaSeparatedList string) bool {
	if comaSeparatedList == "" || ip == "" {
		return false
	}

	if net.ParseIP(ip) == nil {
		ip, _, _ = net.SplitHostPort(ip)
	}

	if ip == "" {
		return false
	}

	cidrs := strings.Split(comaSeparatedList, ",")
	testedIP, _, err := net.ParseCIDR(fmt.Sprintf("%s/32", ip))

	if err != nil {
		return false
	}

	for _, cidr := range cidrs {
		_, ipnet, err := net.ParseCIDR(cidr)
		if err == nil && ipnet.Contains(testedIP) {
			return true
		}
	}

	return false
}

func getCredentialsFromBody(r *http.Request) (username string, password string, err error) {
	data := make(map[string]string)
	decoder := json.NewDecoder(r.Body)
	if err = decoder.Decode(&data); err != nil {
		log.Error(r, "parsing request body", err)
		err = errors.New("invalid request payload")
		return
	}
	username = data["username"]
	password = data["password"]
	return username, password, nil
}

func CreateAdmin(ds model.DataStore) func(w http.ResponseWriter, r *http.Request) {
	auth.Init(ds)

	return func(w http.ResponseWriter, r *http.Request) {
		username, password, err := getCredentialsFromBody(r)
		if err != nil {
			log.Error(r, "parsing request body", err)
			_ = rest.RespondWithError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		c, err := ds.User(r.Context()).CountAll()
		if err != nil {
			_ = rest.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if c > 0 {
			_ = rest.RespondWithError(w, http.StatusForbidden, "Cannot create another first admin")
			return
		}
		err = createDefaultUser(r.Context(), ds, username, password)
		if err != nil {
			_ = rest.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		handleLogin(ds, username, password, w, r)
	}
}

func createDefaultUser(ctx context.Context, ds model.DataStore, username, password string) error {
	log.Warn("Creating initial user", "user", username)
	now := time.Now()
	initialUser := model.User{
		ID:          uuid.NewString(),
		UserName:    username,
		Name:        strings.Title(username),
		Email:       "",
		NewPassword: password,
		IsAdmin:     true,
		LastLoginAt: &now,
	}
	err := ds.User(ctx).Put(&initialUser)
	if err != nil {
		log.Error("Could not create initial user", "user", initialUser, err)
	}
	return nil
}

func validateLogin(userRepo model.UserRepository, userName, password string) (*model.User, error) {
	u, err := userRepo.FindByUsername(userName)
	if err == model.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if u.Password != password {
		return nil, nil
	}
	err = userRepo.UpdateLastLoginAt(u.ID)
	if err != nil {
		log.Error("Could not update LastLoginAt", "user", userName)
	}
	return u, nil
}

func contextWithUser(ctx context.Context, ds model.DataStore, token jwt.Token) context.Context {
	userName := token.Subject()
	user, _ := ds.User(ctx).FindByUsername(userName)
	return request.WithUser(ctx, *user)
}

func getToken(ds model.DataStore, ctx context.Context) (jwt.Token, error) {
	token, claims, err := jwtauth.FromContext(ctx)

	valid := err == nil && token != nil
	valid = valid && claims["sub"] != nil
	if valid {
		return token, nil
	}

	c, err := ds.User(ctx).CountAll()
	firstTime := c == 0 && err == nil
	if firstTime {
		return nil, ErrFirstTime
	}
	return nil, errors.New("invalid authentication")
}

// This method maps the custom authorization header to the default 'Authorization', used by the jwtauth library
func mapAuthHeader() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bearer := r.Header.Get(consts.UIAuthorizationHeader)
			r.Header.Set("Authorization", bearer)
			next.ServeHTTP(w, r)
		})
	}
}

func verifier() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return jwtauth.Verify(auth.TokenAuth, jwtauth.TokenFromHeader, jwtauth.TokenFromCookie, jwtauth.TokenFromQuery)(next)
	}
}

func authenticator(ds model.DataStore) func(next http.Handler) http.Handler {
	auth.Init(ds)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := getToken(ds, r.Context())
			if err == ErrFirstTime {
				_ = rest.RespondWithJSON(w, http.StatusUnauthorized, map[string]string{"message": ErrFirstTime.Error()})
				return
			}
			if err != nil {
				_ = rest.RespondWithError(w, http.StatusUnauthorized, "Not authenticated")
				return
			}

			newCtx := contextWithUser(r.Context(), ds, token)
			newTokenString, err := auth.TouchToken(token)
			if err != nil {
				log.Error(r, "signing new token", err)
				_ = rest.RespondWithError(w, http.StatusUnauthorized, "Not authenticated")
				return
			}

			w.Header().Set(consts.UIAuthorizationHeader, newTokenString)
			next.ServeHTTP(w, r.WithContext(newCtx))
		})
	}
}
