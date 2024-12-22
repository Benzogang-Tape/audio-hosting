package songs

import (
	"context"
	"errors"
	"time"

	"github.com/Benzogang-Tape/audio-hosting/songs/internal/clients/users"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/storage/postgres"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/erix"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/logger"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/pgconv"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/repoerrs"

	"dev.gaijin.team/go/golib/e"
	"dev.gaijin.team/go/golib/fields"
	"github.com/google/uuid"
)

var (
	ErrSongExists      = erix.NewStatus("artist already has song with this name", erix.CodeConflict)
	ErrArtistsNotFound = erix.NewStatus("artists not found", erix.CodeNotFound)
)

type CreateSongInput struct {
	Name        string
	SingerId    uuid.UUID
	ImageUrl    *string
	FeatArtists []uuid.UUID
}

type CreateSongOutput struct {
	Id         uuid.UUID
	Singer     users.Artist
	Artists    []users.Artist
	Name       string
	UploadedAt time.Time
	ImageUrl   *string
}

func (s *Service) CreateSong(ctx context.Context, in CreateSongInput) (CreateSongOutput, error) {
	var (
		null CreateSongOutput
		log  = logger.FromContext(ctx)
	)

	log.Debug().Msg("saving song")

	songParams := postgres.SaveSongParams{
		SongID:       uuid.New(),
		SingerFk:     in.SingerId,
		Name:         in.Name,
		UploadedAt:   time.Now(),
		ArtistsIds:   in.FeatArtists,
		S3ObjectName: pgconv.NullText(),
		ImageUrl:     pgconv.TextPtr(in.ImageUrl),
		Duration:     pgconv.NullInterval(),
		WeightBytes:  pgconv.NullInt4(),
		ReleasedAt:   pgconv.NullTimestamptz(),
	}

	err := s.songRepo.SaveSong(ctx, songParams)

	switch {
	case errors.Is(err, repoerrs.ErrUnique):
		return null, ErrSongExists.Wrap(err, fields.F("name", in.Name), fields.F("singer_id", in.SingerId))

	case err != nil:
		return null, e.NewFrom("saving song", err)
	}

	log.Debug().Msg("getting artists by id")

	artists, err := s.artists(ctx, in.SingerId, in.FeatArtists)
	if err != nil {
		return null, err
	}

	return CreateSongOutput{
		Id:         songParams.SongID,
		Singer:     artists.Singer(),
		Artists:    artists.Artists(),
		Name:       songParams.Name,
		UploadedAt: songParams.UploadedAt,
		ImageUrl:   in.ImageUrl,
	}, nil
}
