package app

import (
	"context"
	"io"

	"github.com/Benzogang-Tape/audio-hosting/songs/internal/clients/users"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/config"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/services/raw"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/services/songs"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/storage"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/storage/postgres"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/audiodecoder"
)

type service struct {
	*songs.Service
	*raw.ServiceRaw
	closer io.Closer
}

func newService(db *storage.Storage) (*service, error) {
	rawService := raw.New(raw.Dependencies{
		ObjectStorage: db,
		SongRepo:      rawSongRepo{db},
		SoundDecoder:  audiodecoder.Decoder{},
	})

	var usersClient interface {
		songs.UserRepo
		io.Closer
	}

	if config.Get().Connections.UsersService.UseFake {
		usersClient = users.NewFake()
	} else {
		client, err := users.New()
		if err != nil {
			return nil, err //nolint:wrapcheck
		}

		usersClient = client
	}

	songsService := songs.New(songs.Dependencies{
		SongRepo:   db,
		UserRepo:   usersClient,
		RawService: rawService,
		Broker:     db,
	})

	return &service{
		Service:    songsService,
		ServiceRaw: rawService,
		closer:     usersClient,
	}, nil
}

type rawSongRepo struct {
	*storage.Storage
}

func (r rawSongRepo) Begin(ctx context.Context) (raw.SongRepo, error) {
	tx, err := r.Storage.Begin(ctx)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	return &rawSongRepoTx{tx}, nil
}

type rawSongRepoTx struct {
	postgres.PgTx
}

func (r rawSongRepoTx) Begin(context.Context) (raw.SongRepo, error) {
	return r, nil
}
