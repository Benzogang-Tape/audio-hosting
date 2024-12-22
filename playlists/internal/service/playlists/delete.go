package playlists

import (
	"context"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/storage/postgres"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DeletePlaylistInput struct {
	PlaylistIDs []string
	UserID      string
}

func (s *ServicePlaylists) DeletePlaylist(ctx context.Context, in DeletePlaylistInput) error {
	log := logger.GetLoggerFromCtx(ctx)

	err := s.repo.DeletePlaylists(ctx, postgres.DeletePlaylistsParams{
		Ids:    convertToUUID(in.PlaylistIDs),
		UserID: uuid.MustParse(in.UserID),
	})

	if err != nil {
		log.Error(
			ctx, "failed to delete playlist",
			zap.String("layout", "service/playlists"),
			zap.String("user_id", in.UserID),
			zap.Error(err))

		return err //nolint:wrapcheck
	}

	return nil
}
