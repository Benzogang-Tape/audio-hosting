package songs

import (
	"context"

	"github.com/Benzogang-Tape/audio-hosting/songs/internal/clients/users"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/storage/broker"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/storage/postgres"

	"github.com/google/uuid"
)

type Service struct {
	c             Config
	songRepo      SongRepo
	userRepo      UserRepo
	rawService    RawService
	messageBroker Broker
}

type SongRepo interface {
	SaveSong(context.Context, postgres.SaveSongParams) error
	Song(context.Context, uuid.UUID) (postgres.SongRow, error)
	ReleasedSongs(context.Context, postgres.ReleasedSongsParams) ([]postgres.ReleasedSongsRow, error)
	MySongs(context.Context, postgres.MySongsParams) ([]postgres.MySongsRow, error)
	CountMySongs(context.Context, uuid.UUID) (int32, error)
	CountSongsWithArtistsIds(context.Context, []uuid.UUID) (int32, error)
	CountSongsMatchName(context.Context, string) (int32, error)
	PatchSongs(context.Context, postgres.PatchSongsParams) error
}

type UserRepo interface {
	ArtistsByIds(context.Context, []uuid.UUID) ([]users.Artist, error)
	ArtistsMatchingName(context.Context, string) ([]users.Artist, error)
}

type Broker interface {
	SendReleasedMessages(context.Context, []broker.SongReleasedMessage) error
}

type RawService interface {
	SongUrl(rawSongId string) string
}

type Dependencies struct {
	SongRepo   SongRepo
	UserRepo   UserRepo
	RawService RawService
	Broker     Broker
}

type Config struct {
	Dependencies
}

func New(deps Dependencies) *Service {
	return NewWithConfig(Config{
		Dependencies: deps,
	})
}

func NewWithConfig(conf Config) *Service {
	return &Service{
		songRepo:      conf.SongRepo,
		userRepo:      conf.UserRepo,
		rawService:    conf.RawService,
		messageBroker: conf.Broker,
		c:             conf,
	}
}
