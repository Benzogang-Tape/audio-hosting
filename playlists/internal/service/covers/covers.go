package covers

import (
	"context"
	"dev.gaijin.team/go/golib/e"
	"dev.gaijin.team/go/golib/fields"
	"errors"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/service"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/storage/minio"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/storage/postgres"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/logger"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/pgconv"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	"io"
	"path/filepath"
	"strings"
	"time"
)

type UploadRawCoverInput struct {
	UserId      uuid.UUID
	PlaylistId  uuid.UUID
	Extension   string
	WeightBytes int32
	Content     io.Reader
}

type UploadRawCoverOutput struct {
	CoverUrl string
}

func (s *ServiceCovers) UploadRawCover(
	ctx context.Context,
	input UploadRawCoverInput,
) (UploadRawCoverOutput, error) {
	var (
		null UploadRawCoverOutput
		log  = logger.GetLoggerFromCtx(ctx)
	)

	if input.Extension != ".jpg" && input.Extension != ".png" && input.Extension != ".jpeg" {
		return null, service.ErrInvalidCoverExtension
	}

	log.Debug(
		ctx, "uploading cover",
		zap.String("layout", "service/covers"),
		zap.String("user_id", input.UserId.String()),
		zap.String("playlist_id", input.PlaylistId.String()),
	)

	playlistRow, err := s.repo.Playlist(ctx, input.PlaylistId)

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return null, service.ErrPlaylistNotFound

	case err != nil:
		return null, e.NewFrom("getting playlist", err, fields.F("playlist_id", input.PlaylistId))
	}

	objectID := getObjectID(input.UserId, playlistRow.Playlist.ID, input.Extension)

	log.Debug(
		ctx, "calculated object id",
		zap.String("layout", "service/covers"),
		zap.String("object_id", objectID),
	)

	txRepo, err := s.repo.BeginCovers(ctx)
	if err != nil {
		return null, e.NewFrom("begin transaction", err)
	}
	defer txRepo.Rollback(ctx) //nolint:errcheck

	playlist := postgres.PatchPlaylistParams{ //nolint:exhaustruct
		ID:        input.PlaylistId,
		UserID:    input.UserId,
		CoverUrl:  pgconv.Text(s.CoverURL(playlistRow.Playlist.ID.String(), objectID)),
		UpdatedAt: time.Now(),
	}

	_, err = txRepo.PatchPlaylist(ctx, playlist)
	if err != nil {

		return null, e.NewFrom(
			"patching playlist",
			err,
			fields.F("playlist_id", input.PlaylistId),
			fields.F("user_id", input.UserId),
		)
	}

	log.Debug(
		ctx, "putting song object",
		zap.String("layout", "service/covers"),
		zap.String("playlist_id", input.PlaylistId.String()),
	)

	err = s.objRepo.PutCoverObject(ctx, minio.CoverObject{
		ID:          objectID,
		Extension:   input.Extension,
		WeightBytes: input.WeightBytes,
		Content:     input.Content,
	})
	if err != nil {
		return null, e.NewFrom("putting cover object", err, fields.F("playlist_id", input.PlaylistId))
	}

	err = txRepo.Commit(ctx)
	if err != nil {
		return null, e.NewFrom("commit transaction", err)
	}

	return UploadRawCoverOutput{
		CoverUrl: s.CoverURL(playlistRow.Playlist.ID.String(), objectID),
	}, nil
}

type GetRawCoverOutput struct {
	Extension string
	Content   io.Reader
}

func (s *ServiceCovers) GetRawCover(ctx context.Context, coverID string) (GetRawCoverOutput, error) {
	reader, err := s.objRepo.GetCoverObject(ctx, coverID)
	if err != nil {
		return GetRawCoverOutput{}, e.NewFrom("getting cover object", err, fields.F("cover_id", coverID))
	}

	return GetRawCoverOutput{
		Extension: strings.TrimPrefix(filepath.Ext(coverID), "."),
		Content:   reader,
	}, nil
}
