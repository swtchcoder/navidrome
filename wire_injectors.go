//+build wireinject

package main

import (
	"github.com/cloudsonic/sonic-server/api"
	"github.com/cloudsonic/sonic-server/domain"
	"github.com/cloudsonic/sonic-server/engine"
	"github.com/cloudsonic/sonic-server/itunesbridge"
	"github.com/cloudsonic/sonic-server/persistence"
	"github.com/cloudsonic/sonic-server/scanner_legacy"
	"github.com/cloudsonic/sonic-server/server"
	"github.com/google/wire"
)

type Provider struct {
	AlbumRepository       domain.AlbumRepository
	ArtistRepository      domain.ArtistRepository
	CheckSumRepository    domain.CheckSumRepository
	ArtistIndexRepository domain.ArtistIndexRepository
	MediaFileRepository   domain.MediaFileRepository
	MediaFolderRepository domain.MediaFolderRepository
	NowPlayingRepository  domain.NowPlayingRepository
	PlaylistRepository    domain.PlaylistRepository
	PropertyRepository    domain.PropertyRepository
}

var allProviders = wire.NewSet(
	itunesbridge.NewItunesControl,
	engine.Set,
	scanner_legacy.Set,
	api.NewRouter,
	wire.FieldsOf(new(*Provider), "AlbumRepository", "ArtistRepository", "CheckSumRepository",
		"ArtistIndexRepository", "MediaFileRepository", "MediaFolderRepository", "NowPlayingRepository",
		"PlaylistRepository", "PropertyRepository"),
	createPersistenceProvider,
)

func CreateApp(musicFolder string) *server.Server {
	panic(wire.Build(
		server.New,
		allProviders,
	))
}

func CreateSubsonicAPIRouter() *api.Router {
	panic(wire.Build(allProviders))
}

// When implementing a different persistence layer, duplicate this function (in separated files) and use build tags
// to conditionally select which function to use
func createPersistenceProvider() *Provider {
	panic(wire.Build(
		persistence.Set,
		wire.Struct(new(Provider), "*"),
	))
}
