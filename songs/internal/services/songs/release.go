package songs

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Benzogang-Tape/audio-hosting/songs/internal/storage/postgres"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/erix"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/logger"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/pgconv"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/repoerrs"

	"dev.gaijin.team/go/golib/e"
	"github.com/google/uuid"
)

type ReleaseSongsInput struct {
	UserId   uuid.UUID
	SongsIds []uuid.UUID
	Notify   bool
}

type ReleaseSongsOutput struct {
}

func (s *Service) ReleaseSongs(ctx context.Context, in ReleaseSongsInput) (ReleaseSongsOutput, error) {
	var (
		null = ReleaseSongsOutput{}
		log  = logger.FromContext(ctx)
	)

	log.Debug().Array("songs_ids", logger.Stringers[uuid.UUID](in.SongsIds)).Msg("releasing songs")

	songs, err := s.songRepo.MySongs(ctx, postgres.MySongsParams{
		SingerID: in.UserId,
		ByIds:    true,
		Ids:      in.SongsIds,
		Limitv:   int32(len(in.SongsIds)), //nolint:gosec
		Offsetv:  0,
	})

	switch {
	// repoerrs didn't work properly, so I've added another check
	case errors.Is(err, repoerrs.ErrEmptyResult) || (len(songs) == 0 && err == nil):
		return null, ErrSongNotFound

	case err != nil:
		return null, e.NewFrom("getting my songs", err)
	}

	err = validateReleasingSongs(songs)
	if err != nil {
		return null, err
	}

	releaseTime := time.Now()
	log.Debug().Time("release_time", releaseTime).Msg("patching songs")

	err = s.songRepo.PatchSongs(ctx, postgres.PatchSongsParams{ //nolint:exhaustruct
		Ids:        in.SongsIds,
		ReleasedAt: pgconv.Timestamptz(releaseTime),
	})
	if err != nil {
		return null, e.NewFrom("patching songs", err)
	}

	if !in.Notify {
		log.Debug().Msg("not sending messages")
		return null, nil
	}

	// TODO: add outbox
	err = s.sendReleasedMessages(ctx, songs, releaseTime)
	if err != nil {
		return null, e.NewFrom("sending messages", err)
	}

	return null, nil
}

func validateReleasingSongs(songs []postgres.MySongsRow) error {
	var errs []error

	for _, song := range songs {
		if song.Song.ReleasedAt.Valid {
			errs = append(errs, newReleaseErr("already released", song))
		}

		if !song.Song.S3ObjectName.Valid {
			errs = append(errs, newReleaseErr("has not been loaded", song))
		}
	}

	if len(errs) > 0 {
		return erix.NewStatus(multiErr(errs).Error(), erix.CodePreconditionFailed)
	}

	return nil
}

type releaseSongErr struct {
	Id   uuid.UUID
	Name string
	Err  error
}

func newReleaseErr(reason string, song postgres.MySongsRow) releaseSongErr {
	return releaseSongErr{
		Id:   song.Song.SongID,
		Name: song.Song.Name,
		Err:  e.New(reason),
	}
}

func (r releaseSongErr) Error() string {
	return r.Name + " (" + r.Id.String() + ") " + r.Err.Error()
}

func (r releaseSongErr) Unwrap() error {
	return r.Err
}

type multiErr []error

func (m multiErr) Error() string {
	b := strings.Builder{}
	for _, err := range m {
		_, _ = b.WriteString(err.Error())
		_, _ = b.WriteString(" | ")
	}

	return b.String()
}

func (m multiErr) Unwrap() []error {
	return m
}
