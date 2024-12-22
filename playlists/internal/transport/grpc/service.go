package grpcserver

import (
	"context"
	client "github.com/Benzogang-Tape/audio-hosting/playlists/internal/client/songs"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/config"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/service/covers"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/service/playlists"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/storage"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/storage/postgres"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/transport/grpc/handlers"
	"strconv"
)

type Service struct {
	*playlists.ServicePlaylists
	*covers.ServiceCovers
}

type Repo interface {
	covers.Repository
	covers.ObjectRepository

	playlists.PlaylistsRepo
}

func NewService(strg *storage.Storage, cfg config.HTTPConfig, clientSongs *client.Client) handlers.Service {
	coversService := covers.New(coversRepo{strg}, strg, covers.Config{
		HostUsesTLS: false,
		Host:        cfg.Host + ":" + strconv.Itoa(cfg.Port),
	})

	playlistsService := playlists.New(playlistsRepo{strg}, clientSongs)

	return Service{
		ServicePlaylists: playlistsService,
		ServiceCovers:    coversService,
	}
}

type coversRepo struct {
	*storage.Storage
}

func (r coversRepo) BeginCovers(ctx context.Context) (covers.Repository, error) {
	tx, err := r.Storage.Begin(ctx)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	return &coversRepoTx{tx}, nil
}

type coversRepoTx struct {
	postgres.PgTx
}

func (r coversRepoTx) BeginCovers(context.Context) (covers.Repository, error) {
	return r, nil
}

type playlistsRepo struct {
	*storage.Storage
}

func (r playlistsRepo) Begin(ctx context.Context) (playlists.PlaylistsRepo, error) {
	tx, err := r.Storage.Begin(ctx)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	return &playlistsRepoTx{tx}, nil
}

type playlistsRepoTx struct {
	postgres.PgTx
}

func (r playlistsRepoTx) Begin(context.Context) (playlists.PlaylistsRepo, error) {
	return r, nil
}
