package raw

import (
	"context"
	"errors"
	"io"
	"path/filepath"

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

type UploadRawSongImageInput struct {
	ArtistId    uuid.UUID
	SongId      uuid.UUID
	Extension   string
	WeightBytes int32
	Content     io.Reader
}

type UploadRawSongImageOutput struct {
	ImageUrl string
}

var (
	ErrInvalidImageExtension = erix.NewStatus("invalid extension, only jpg, png, jpeg supported", erix.CodeBadRequest)
)

func (s *ServiceRaw) UploadRawSongImage(ctx context.Context,
	input UploadRawSongImageInput,
) (UploadRawSongImageOutput, error) {
	var (
		null UploadRawSongImageOutput
		log  = logger.FromContext(ctx)
	)

	if input.Extension != "jpg" && input.Extension != "png" && input.Extension != "jpeg" {
		return null, ErrInvalidImageExtension
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

	case err != nil:
		return null, e.NewFrom("getting song", err, fields.F("song_id", input.SongId))
	}

	objectId := getObjectId(input.ArtistId, songRow.Song.Name, input.Extension)

	log.Debug().Str("object_id", objectId).Msg("calculated object id")

	txRepo, err := s.repo.Begin(ctx)
	if err != nil {
		return null, e.NewFrom("begin transaction", err)
	}
	defer txRepo.Rollback(ctx) //nolint:errcheck

	song := postgres.PatchSongParams{ //nolint:exhaustruct
		ID:       input.SongId,
		ImageUrl: pgconv.Text(s.ImageUrl(objectId)),
	}

	patchedSong, err := txRepo.PatchSong(ctx, song)
	if err != nil {
		return null, e.NewFrom("patching song", err, fields.F("song_id", input.SongId))
	}

	log.Debug().Object("songs_diff", songsDiff(songRow.Song, patchedSong)).Msg("patched song")
	log.Debug().Msg("putting song object")

	err = s.storage.PutImageObject(ctx, s3minio.ImageObject{
		Id:          objectId,
		Extension:   input.Extension,
		WeightBytes: input.WeightBytes,
		Content:     input.Content,
	})
	if err != nil {
		return null, e.NewFrom("putting image object", err, fields.F("song_id", input.SongId))
	}

	err = txRepo.Commit(ctx)
	if err != nil {
		return null, e.NewFrom("commit transaction", err)
	}

	return UploadRawSongImageOutput{
		ImageUrl: s.ImageUrl(objectId),
	}, nil
}

type GetRawSongImageOutput struct {
	Extension string
	Content   io.Reader
}

func (s *ServiceRaw) GetRawSongImage(ctx context.Context, songId string) (GetRawSongImageOutput, error) {
	reader, err := s.storage.GetImageObject(ctx, songId)
	if err != nil {
		return GetRawSongImageOutput{}, e.NewFrom("getting song object", err, fields.F("song_id", songId))
	}

	return GetRawSongImageOutput{
		Extension: filepath.Ext(songId)[1:],
		Content:   reader,
	}, nil
}
