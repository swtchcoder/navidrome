package server

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/deluan/navidrome/conf"
	"github.com/deluan/navidrome/consts"
	"github.com/deluan/navidrome/log"
	"github.com/deluan/navidrome/model"
	"github.com/google/uuid"
)

func initialSetup(ds model.DataStore) {
	_ = ds.WithTx(func(tx model.DataStore) error {
		properties := ds.Property(context.TODO())
		_, err := properties.Get(consts.InitialSetupFlagKey)
		if err == nil {
			return nil
		}
		log.Warn("Running initial setup")
		if err = createJWTSecret(ds); err != nil {
			return err
		}

		if conf.Server.DevAutoCreateAdminPassword != "" {
			if err = createInitialAdminUser(ds); err != nil {
				return err
			}
		}

		err = properties.Put(consts.InitialSetupFlagKey, time.Now().String())
		return err
	})
}

func createInitialAdminUser(ds model.DataStore) error {
	users := ds.User(context.TODO())
	c, err := users.CountAll()
	if err != nil {
		panic(fmt.Sprintf("Could not access User table: %s", err))
	}
	if c == 0 {
		id, _ := uuid.NewRandom()
		random, _ := uuid.NewRandom()
		initialPassword := random.String()
		if conf.Server.DevAutoCreateAdminPassword != "" {
			initialPassword = conf.Server.DevAutoCreateAdminPassword
		}
		log.Warn("Creating initial admin user. This should only be used for development purposes!!", "user", consts.DevInitialUserName, "password", initialPassword)
		initialUser := model.User{
			ID:       id.String(),
			UserName: consts.DevInitialUserName,
			Name:     consts.DevInitialName,
			Email:    "",
			Password: initialPassword,
			IsAdmin:  true,
		}
		err := users.Put(&initialUser)
		if err != nil {
			log.Error("Could not create initial admin user", "user", initialUser, err)
		}
	}
	return err
}

func createJWTSecret(ds model.DataStore) error {
	properties := ds.Property(context.TODO())
	_, err := properties.Get(consts.JWTSecretKey)
	if err == nil {
		return nil
	}
	jwtSecret, _ := uuid.NewRandom()
	log.Warn("Creating JWT secret, used for encrypting UI sessions")
	err = properties.Put(consts.JWTSecretKey, jwtSecret.String())
	if err != nil {
		log.Error("Could not save JWT secret in DB", err)
	}
	return err
}

func checkFfmpegInstallation() {
	path, err := exec.LookPath("ffmpeg")
	if err == nil {
		log.Info("Found ffmpeg", "path", path)
		return
	}
	log.Warn("Unable to find ffmpeg. Transcoding will fail if used", err)
	if conf.Server.Scanner.Extractor == "ffmpeg" {
		log.Warn("ffmpeg cannot be used for metadata extraction. Falling back to taglib")
		conf.Server.Scanner.Extractor = "taglib"
	}
}

func checkExternalCredentials() {
	if conf.Server.LastFM.ApiKey == "" || conf.Server.LastFM.Secret == "" {
		log.Info("Last.FM integration not available: missing ApiKey/Secret")
	}

	if conf.Server.Spotify.ID == "" || conf.Server.Spotify.Secret == "" {
		log.Info("Spotify integration is not enabled: artist images will not be available")
	}
}
