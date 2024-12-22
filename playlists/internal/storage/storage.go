package storage

import (
	"context"
	"dev.gaijin.team/go/golib/e"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/storage/minio"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/storage/postgres"
	s3 "github.com/Benzogang-Tape/audio-hosting/playlists/pkg/db/minio"
	pg "github.com/Benzogang-Tape/audio-hosting/playlists/pkg/db/postgres"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/db/redis"
)

type Storage struct {
	*postgres.PGStorage
	*minio.S3Storage
	// TODO: do redis storage
}

func New(pgConfig pg.Config, _ redis.Config, s3Config s3.Config) (*Storage, error) {
	db, err := postgres.Connect(pgConfig)
	if err != nil {
		return nil, e.NewFrom("connecting to postgres", err)
	}

	minioClient, err := minio.New(context.Background(), s3Config)
	if err != nil {
		return nil, e.NewFrom("connecting to minio", err)
	}

	return &Storage{
		PGStorage: db,
		S3Storage: minioClient,
	}, nil
}

func (s *Storage) Close() error {
	s.PGStorage.Close()
	return nil
}
