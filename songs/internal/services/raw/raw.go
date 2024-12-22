package raw

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Benzogang-Tape/audio-hosting/songs/internal/config"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/storage/postgres"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/storage/s3minio"
)

type ServiceRaw struct {
	c       Config
	storage ObjectStorage
	repo    SongRepo
	decoder SoundDecoder

	songUrlTpl  string
	imageUrlTpl string
}

type ObjectStorage interface {
	PutSongObject(context.Context, s3minio.SongObject) error
	GetSongObject(ctx context.Context, id string) (io.Reader, error)
	PutImageObject(ctx context.Context, image s3minio.ImageObject) error
	GetImageObject(ctx context.Context, id string) (io.Reader, error)
}

type SongRepo interface {
	MySong(context.Context, postgres.MySongParams) (postgres.MySongRow, error)
	PatchSong(context.Context, postgres.PatchSongParams) (postgres.Song, error)
	Begin(context.Context) (SongRepo, error)
	Commit(context.Context) error
	Rollback(context.Context) error
}

type SoundDecoder interface {
	GetMp3Duration(context.Context, io.Reader) (time.Duration, error)
}

type Dependencies struct {
	ObjectStorage ObjectStorage
	SongRepo      SongRepo
	SoundDecoder  SoundDecoder
}

type Config struct {
	Dependencies
	HostUsesTls bool
	Host        string
}

func New(deps Dependencies) *ServiceRaw {
	conf := config.Get().Servers

	return NewWithConfig(Config{
		Dependencies: deps,
		HostUsesTls:  conf.Http.UseTls,
		Host:         conf.Host,
	})
}

func NewWithConfig(conf Config) *ServiceRaw {
	schema := "http"
	if conf.HostUsesTls {
		schema = "https"
	}

	return &ServiceRaw{
		c:           conf,
		storage:     conf.ObjectStorage,
		repo:        conf.SongRepo,
		decoder:     conf.SoundDecoder,
		songUrlTpl:  fmt.Sprintf("%s://%s/songs/api/v1/song/raw/", schema, conf.Host),
		imageUrlTpl: fmt.Sprintf("%s://%s/songs/api/v1/song/image/raw/", schema, conf.Host),
	}
}
