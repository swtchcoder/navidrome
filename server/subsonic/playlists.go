package subsonic

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/deluan/navidrome/log"
	"github.com/deluan/navidrome/model"
	"github.com/deluan/navidrome/server/subsonic/engine"
	"github.com/deluan/navidrome/server/subsonic/responses"
	"github.com/deluan/navidrome/utils"
)

type PlaylistsController struct {
	pls engine.Playlists
}

func NewPlaylistsController(pls engine.Playlists) *PlaylistsController {
	return &PlaylistsController{pls: pls}
}

func (c *PlaylistsController) GetPlaylists(w http.ResponseWriter, r *http.Request) (*responses.Subsonic, error) {
	allPls, err := c.pls.GetAll(r.Context())
	if err != nil {
		log.Error(r, err)
		return nil, newError(responses.ErrorGeneric, "Internal error")
	}
	playlists := make([]responses.Playlist, len(allPls))
	for i, p := range allPls {
		playlists[i].Id = p.ID
		playlists[i].Name = p.Name
		playlists[i].Comment = p.Comment
		playlists[i].SongCount = p.SongCount
		playlists[i].Duration = int(p.Duration)
		playlists[i].Owner = p.Owner
		playlists[i].Public = p.Public
		playlists[i].Created = p.CreatedAt
		playlists[i].Changed = p.UpdatedAt
	}
	response := newResponse()
	response.Playlists = &responses.Playlists{Playlist: playlists}
	return response, nil
}

func (c *PlaylistsController) GetPlaylist(w http.ResponseWriter, r *http.Request) (*responses.Subsonic, error) {
	id, err := requiredParamString(r, "id", "id parameter required")
	if err != nil {
		return nil, err
	}
	pinfo, err := c.pls.Get(r.Context(), id)
	switch {
	case err == model.ErrNotFound:
		log.Error(r, err.Error(), "id", id)
		return nil, newError(responses.ErrorDataNotFound, "Directory not found")
	case err != nil:
		log.Error(r, err)
		return nil, newError(responses.ErrorGeneric, "Internal Error")
	}

	response := newResponse()
	response.Playlist = c.buildPlaylistWithSongs(r.Context(), pinfo)
	return response, nil
}

func (c *PlaylistsController) CreatePlaylist(w http.ResponseWriter, r *http.Request) (*responses.Subsonic, error) {
	songIds := utils.ParamStrings(r, "songId")
	playlistId := utils.ParamString(r, "playlistId")
	name := utils.ParamString(r, "name")
	if playlistId == "" && name == "" {
		return nil, errors.New("Required parameter name is missing")
	}
	err := c.pls.Create(r.Context(), playlistId, name, songIds)
	if err != nil {
		log.Error(r, err)
		return nil, newError(responses.ErrorGeneric, "Internal Error")
	}
	return newResponse(), nil
}

func (c *PlaylistsController) DeletePlaylist(w http.ResponseWriter, r *http.Request) (*responses.Subsonic, error) {
	id, err := requiredParamString(r, "id", "Required parameter id is missing")
	if err != nil {
		return nil, err
	}
	err = c.pls.Delete(r.Context(), id)
	if err == model.ErrNotAuthorized {
		return nil, newError(responses.ErrorAuthorizationFail)
	}
	if err != nil {
		log.Error(r, err)
		return nil, newError(responses.ErrorGeneric, "Internal Error")
	}
	return newResponse(), nil
}

func (c *PlaylistsController) UpdatePlaylist(w http.ResponseWriter, r *http.Request) (*responses.Subsonic, error) {
	playlistId, err := requiredParamString(r, "playlistId", "Required parameter playlistId is missing")
	if err != nil {
		return nil, err
	}
	songsToAdd := utils.ParamStrings(r, "songIdToAdd")
	songIndexesToRemove := utils.ParamInts(r, "songIndexToRemove")

	var pname *string
	if len(r.URL.Query()["name"]) > 0 {
		s := r.URL.Query()["name"][0]
		pname = &s
	}

	log.Debug(r, "Updating playlist", "id", playlistId)
	if pname != nil {
		log.Trace(r, fmt.Sprintf("-- New Name: '%s'", *pname))
	}
	log.Trace(r, fmt.Sprintf("-- Adding: '%v'", songsToAdd))
	log.Trace(r, fmt.Sprintf("-- Removing: '%v'", songIndexesToRemove))

	err = c.pls.Update(r.Context(), playlistId, pname, songsToAdd, songIndexesToRemove)
	if err == model.ErrNotAuthorized {
		return nil, newError(responses.ErrorAuthorizationFail)
	}
	if err != nil {
		log.Error(r, err)
		return nil, newError(responses.ErrorGeneric, "Internal Error")
	}
	return newResponse(), nil
}

func (c *PlaylistsController) buildPlaylistWithSongs(ctx context.Context, d *engine.PlaylistInfo) *responses.PlaylistWithSongs {
	pls := &responses.PlaylistWithSongs{
		Playlist: *c.buildPlaylist(d),
	}
	pls.Entry = toChildren(ctx, d.Entries)
	return pls
}

func (c *PlaylistsController) buildPlaylist(d *engine.PlaylistInfo) *responses.Playlist {
	pls := &responses.Playlist{}
	pls.Id = d.Id
	pls.Name = d.Name
	pls.Comment = d.Comment
	pls.SongCount = d.SongCount
	pls.Owner = d.Owner
	pls.Duration = d.Duration
	pls.Public = d.Public
	pls.Created = d.Created
	pls.Changed = d.Changed
	return pls
}
