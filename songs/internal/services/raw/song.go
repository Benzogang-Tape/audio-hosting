package raw

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/Benzogang-Tape/audio-hosting/songs/internal/storage/postgres"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/storage/s3minio"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/erix"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/logger"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/pgconv"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/repoerrs"

	"dev.gaijin.team/go/golib/e"
	"dev.gaijin.team/go/golib/fields"
	"github.com/google/uuid"
)

var (
	ErrInvalidExtension    = erix.NewStatus("invalid extension, only mp3 supported", erix.CodeBadRequest)
	ErrSongNotExists       = erix.NewStatus("song not exists", erix.CodeNotFound)
	ErrSongAlreadyReleased = erix.NewStatus("not able to upload song, it is released", erix.CodePreconditionFailed)
	ErrFileNotFound        = erix.NewStatus("file not found", erix.CodeNotFound)
)

type UploadRawSongInput struct {
	ArtistId    uuid.UUID
	SongId      uuid.UUID
	Extension   string
	WeightBytes int32
	Content     io.Reader
}

type UploadRawSongOutput struct {
	SongUrl string
}

func (s *ServiceRaw) UploadRawSong(ctx context.Context, input UploadRawSongInput) (UploadRawSongOutput, error) {
	var (
		null UploadRawSongOutput
		log  = logger.FromContext(ctx)
	)

	if input.Extension != "mp3" {
		return null, ErrInvalidExtension
	}

	log.Debug().
		Stringer("song_id", input.SongId).Stringer("artist_id", input.ArtistId).
		Msg("getting info about song")

	songRow, err := s.repo.MySong(ctx, postgres.MySongParams{
		SingerID: input.ArtistId,
		SongID:   input.SongId,
	})

	switch {
	case errors.Is(err, repoerrs.ErrEmptyResult):
		return null, ErrSongNotExists

	case songRow.Song.ReleasedAt.Valid:
		return null, ErrSongAlreadyReleased

	case err != nil:
		return null, e.NewFrom("getting song", err, fields.F("song_id", input.SongId))
	}

	log.Debug().Msg("teeing content into buffer")

	contentBuf := &bytes.Buffer{}
	content := io.TeeReader(input.Content, contentBuf)

	dur, err := s.decoder.GetMp3Duration(ctx, content)
	if err != nil {
		return null, e.NewFrom("getting mp3 duration", err)
	}

	log.Debug().Dur("song_duration", dur).Msg("got mp3 duration")

	objectId := getObjectId(input.ArtistId, songRow.Song.Name, input.Extension)

	log.Debug().Str("object_id", objectId).Msg("calculated object id")

	txRepo, err := s.repo.Begin(ctx)
	if err != nil {
		return null, e.NewFrom("begin transaction", err)
	}
	defer txRepo.Rollback(ctx) //nolint:errcheck

	song := postgres.PatchSongParams{ //nolint:exhaustruct
		ID:           input.SongId,
		S3ObjectName: pgconv.Text(objectId),
		Duration:     pgconv.Interval(dur),
		WeightBytes:  pgconv.Int4(input.WeightBytes),
	}

	patchedSong, err := txRepo.PatchSong(ctx, song)
	if err != nil {
		return null, e.NewFrom("patching song", err, fields.F("song_id", input.SongId))
	}

	log.Debug().Object("songs_diff", songsDiff(songRow.Song, patchedSong)).Msg("patched song")

	log.Debug().Msg("putting song object")

	err = s.storage.PutSongObject(ctx, s3minio.SongObject{
		Id:          objectId,
		Duration:    dur,
		Extension:   input.Extension,
		WeightBytes: input.WeightBytes,
		Content:     contentBuf,
	})
	if err != nil {
		return null, e.NewFrom("putting song object", err, fields.F("song_id", input.SongId))
	}

	err = txRepo.Commit(ctx)
	if err != nil {
		return null, e.NewFrom("commit transaction", err)
	}

	return UploadRawSongOutput{
		SongUrl: s.SongUrl(objectId),
	}, nil
}

func (s *ServiceRaw) GetRawSong(ctx context.Context, songId string) (io.Reader, error) {
	reader, err := s.storage.GetSongObject(ctx, songId)
	if err != nil {
		return nil, e.NewFrom("getting song object", err, fields.F("song_id", songId))
	}

	return reader, nil
}
