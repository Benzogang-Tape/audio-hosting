package playlists

import (
	"context"
	"dev.gaijin.team/go/golib/e"
	"errors"
	"github.com/AlekSi/pointer"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/models"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/service"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/storage/postgres"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/logger"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/pgconv"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	"time"
)

type UpdatePlaylistInput struct {
	PlaylistID string
	UserID     string
	Title      *string
	CoverURL   *string
	IsPublic   *bool
	TrackIDs   []string
}

func (s *ServicePlaylists) UpdatePlaylist(
	ctx context.Context,
	in UpdatePlaylistInput,
) (models.PlaylistMetadata, error) {
	var playlistMetadata models.PlaylistMetadata

	log := logger.GetLoggerFromCtx(ctx)

	log.Debug(
		ctx, "updating playlist",
		zap.String("layout", "service/playlists"),
		zap.String("playlist_id", in.PlaylistID))

	params := postgres.PatchPlaylistParams{ //nolint:exhaustruct
		ID:        uuid.MustParse(in.PlaylistID),
		UserID:    uuid.MustParse(in.UserID),
		UpdatedAt: time.Now(),
	}

	if in.Title != nil {
		params.Title = pgconv.Text(*in.Title)
	}

	if in.CoverURL != nil {
		params.CoverUrl = pgconv.Text(*in.CoverURL)
	}

	if in.IsPublic != nil {
		params.IsPublic = pgconv.Bool(*in.IsPublic)
	}

	playlist, err := s.repo.PatchPlaylist(ctx, params)

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return playlistMetadata, service.ErrUpdatePlaylist

	case err != nil:
		log.Error(
			ctx, "failed to update playlist",
			zap.String("layout", "service/playlists"),
			zap.String("playlist_id", in.PlaylistID),
			zap.Error(err))

		return playlistMetadata, e.NewFrom("failed to update playlist", err)
	}

	playlistMetadata = models.PlaylistMetadata{
		ID:             playlist.ID.String(),
		Title:          playlist.Title,
		AuthorID:       playlist.AuthorID.String(),
		CoverURL:       pointer.GetString(pgconv.FromText(playlist.CoverUrl)),
		CreatedAt:      playlist.CreatedAt,
		UpdatedAt:      pointer.GetTime(pgconv.FromTimestamptz(playlist.UpdatedAt)),
		ReleasedAt:     pointer.GetTime(pgconv.FromTimestamptz(playlist.ReleasedAt)),
		IsAlbum:        playlist.IsAlbum,
		IsMyCollection: false,
		IsPublic:       playlist.IsPublic,
	}

	return playlistMetadata, nil
}

type ReleaseAlbumInput struct {
	AlbumID               string
	UserID                string
	SuppressNotifications bool
}

func (s *ServicePlaylists) ReleaseAlbum(ctx context.Context, in ReleaseAlbumInput) error {
	log := logger.GetLoggerFromCtx(ctx)

	log.Debug(
		ctx, "releasing album",
		zap.String("layout", "service/playlists"),
		zap.String("album_id", in.AlbumID))

	params := postgres.PatchPlaylistParams{ //nolint:exhaustruct
		ReleasedAt: pgconv.Timestamptz(time.Now()),
		ID:         uuid.MustParse(in.AlbumID),
		UserID:     uuid.MustParse(in.UserID),
		IsPublic:   pgconv.Bool(true),
	}

	tx, err := s.repo.Begin(ctx)
	if err != nil {
		return e.NewFrom("beginning transaction", err)
	}
	defer tx.Rollback(ctx)

	playlist, err := tx.PatchPlaylist(ctx, params)

	switch {
	case errors.Is(err, pgx.ErrNoRows) || playlist.AuthorID != uuid.MustParse(in.UserID) || !playlist.IsAlbum:
		return service.ErrAlbumNotFound

	case err != nil:
		log.Error(
			ctx, "failed to release album",
			zap.String("layout", "service/playlists"),
			zap.String("album_id", in.AlbumID),
			zap.Error(err))

		return e.NewFrom("failed to release album", err)
	}

	err = s.songs.ReleaseSongs(ctx, convertUUIDtoString(playlist.TrackIds))
	if err != nil {
		log.Error(
			ctx, "failed to release album",
			zap.String("layout", "service/playlists"),
			zap.String("album_id", in.AlbumID),
			zap.Error(err))

		return e.NewFrom("failed to release album", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return e.NewFrom("committing transaction", err)
	}

	// TODO: notifications (kafka producer)

	return nil
}
