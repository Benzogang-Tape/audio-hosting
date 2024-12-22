package playlists

import (
	"context"
	"dev.gaijin.team/go/golib/e"
	"errors"
	"fmt"
	client "github.com/Benzogang-Tape/audio-hosting/playlists/internal/client/songs"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/models"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/service"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/storage/postgres"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/erix"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/logger"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/pgconv"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type LikeDislikePlaylistInput struct {
	PlaylistID string
	UserID     string
}

func (s *ServicePlaylists) LikePlaylist(ctx context.Context, in LikeDislikePlaylistInput) error {
	log := logger.GetLoggerFromCtx(ctx)

	log.Debug(
		ctx, "liking playlist",
		zap.String("layout", "service/playlists"),
		zap.String("playlist_id", in.PlaylistID),
		zap.String("user_id", in.UserID),
	)

	id, err := s.repo.LikePlaylist(ctx, postgres.LikePlaylistParams{
		PlaylistID: uuid.MustParse(in.PlaylistID),
		UserID:     uuid.MustParse(in.UserID),
	})

	log.Debug(
		ctx, "liked playlist",
		zap.String("layout", "service/playlists"),
		zap.Any("playlist_id", id))

	if errors.Is(err, pgx.ErrNoRows) {
		return service.ErrNoPlaylistToLike
	}

	if err != nil {
		log.Error(
			ctx, "failed to like playlist",
			zap.String("layout", "service/playlists"),
			zap.String("playlist_id", in.PlaylistID),
			zap.String("user_id", in.UserID),
			zap.Error(err),
		)

		return erix.NewStatus("failed to like playlist", erix.CodeInternalServerError)
	}

	return nil
}

func (s *ServicePlaylists) DislikePlaylist(ctx context.Context, in LikeDislikePlaylistInput) error {
	log := logger.GetLoggerFromCtx(ctx)

	log.Debug(
		ctx, "disliking playlist",
		zap.String("layout", "service/playlists"),
		zap.String("playlist_id", in.PlaylistID),
		zap.String("user_id", in.UserID),
	)

	// it never returns error, because DELETE operation always completes succesfully
	// even if there is no such row to delete
	_ = s.repo.DislikePlaylist(ctx, postgres.DislikePlaylistParams{
		PlaylistID: uuid.MustParse(in.PlaylistID),
		UserID:     uuid.MustParse(in.UserID),
	})

	return nil
}

type CopyPlaylistInput struct {
	PlaylistID string
	UserID     string
}

func (s *ServicePlaylists) CopyPlaylist(ctx context.Context, in CopyPlaylistInput) (string, error) {
	log := logger.GetLoggerFromCtx(ctx)

	log.Debug(
		ctx, "copying playlist",
		zap.String("layout", "service/playlists"),
		zap.String("playlist_id", in.PlaylistID),
		zap.String("user_id", in.UserID),
	)

	id := uuid.New()

	_, err := s.repo.CopyPlaylist(ctx, postgres.CopyPlaylistParams{
		UserID:        uuid.MustParse(in.UserID),
		PlaylistID:    uuid.MustParse(in.PlaylistID),
		NewPlaylistID: id,
	})

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return "", service.ErrPlaylistNotFound

	case err != nil:
		log.Error(
			ctx, "failed to copy playlist",
			zap.String("layout", "service/playlists"),
			zap.String("playlist_id", in.PlaylistID),
			zap.String("user_id", in.UserID),
			zap.Error(err))

		return "", erix.NewStatus("failed to copy playlist", erix.CodeInternalServerError)
	}

	return id.String(), nil
}

func (s *ServicePlaylists) GetMyPlaylists(ctx context.Context, userID string) ([]models.PlaylistMetadata, error) {
	log := logger.GetLoggerFromCtx(ctx)

	rows, err := s.repo.UserPlaylists(ctx, uuid.MustParse(userID))

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		log.Debug(ctx, "playlist not found", zap.String("layout", "service/playlists"))
		return nil, service.ErrPlaylistNotFound

	case err != nil:
		log.Error(
			ctx, "failed to get playlist",
			zap.String("layout", "service/playlists"),
			zap.String("user_id", userID),
			zap.Error(err))

		return nil, erix.NewStatus("failed to get playlist", erix.CodeInternalServerError)
	}

	resp := make([]models.PlaylistMetadata, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, models.PlaylistMetadata{ //nolint:exhaustruct
			ID:             row.ID.String(),
			Title:          row.Title,
			AuthorID:       row.AuthorID.String(),
			CoverURL:       *pgconv.FromText(row.CoverUrl),
			CreatedAt:      row.CreatedAt,
			IsAlbum:        row.IsAlbum,
			IsMyCollection: false,
			IsPublic:       row.IsPublic,
		})
	}

	return resp, nil
}

func (s *ServicePlaylists) GetMyCollection(
	ctx context.Context,
	userID string,
) (models.Playlist, error) {
	playlistMetadata := models.PlaylistMetadata{ //nolint:exhaustruct
		ID:             fmt.Sprint("my_collection_", userID),
		Title:          "My Collection",
		AuthorID:       userID,
		CoverURL:       "", // TODO: default My Collection cover
		CreatedAt:      time.Time{},
		UpdatedAt:      time.Time{},
		IsAlbum:        false,
		IsMyCollection: true,
		IsPublic:       false,
	}

	log := logger.GetLoggerFromCtx(ctx)

	log.Debug(
		ctx, "getting my collection",
		zap.String("layout", "service/playlists"),
		zap.String("user_id", userID),
	)

	rows, err := s.repo.MyCollection(ctx, uuid.MustParse(userID))

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return models.Playlist{
			Metadata: playlistMetadata,
			Songs:    []client.Song{},
		}, nil

	case err != nil:
		log.Error(
			ctx, "failed to get playlist",
			zap.String("layout", "service/playlists"),
			zap.String("user_id", userID))

		return models.Playlist{}, e.NewFrom("getting liked tracks", err)
	}

	likedTracks, err := s.getSongs(ctx,
		func() []uuid.UUID {
			ids := make([]uuid.UUID, 0, len(rows))
			for _, row := range rows {
				ids = append(ids, row.TrackID)
			}

			return ids
		}())

	if err != nil {
		log.Error(
			ctx, "failed to get playlist",
			zap.String("layout", "service/playlists"),
			zap.String("user_id", userID),
			zap.Error(err))

		return models.Playlist{
			Metadata: playlistMetadata,
			Songs:    []client.Song{},
		}, service.ErrGetSongs
	}

	return models.Playlist{
		Metadata: playlistMetadata,
		Songs:    likedTracks,
	}, nil
}

type LikeDislikeTrackInput struct {
	TrackID string
	UserID  string
}

func (s *ServicePlaylists) LikeTrack(ctx context.Context, in LikeDislikeTrackInput) error {
	_, err := s.songs.GetSong(ctx, in.TrackID)

	switch {
	case err != nil && status.Code(err) == codes.NotFound:
		return service.ErrSongDoesNotExist

	case err != nil:
		return e.NewFrom("getting song", err)
	}

	_, err = s.repo.LikeTrack(ctx, postgres.LikeTrackParams{
		TrackID: uuid.MustParse(in.TrackID),
		UserID:  uuid.MustParse(in.UserID),
	})
	if err != nil {
		return e.NewFrom("liking track", err)
	}

	return nil
}

func (s *ServicePlaylists) DislikeTrack(ctx context.Context, in LikeDislikeTrackInput) error {
	err := s.repo.DislikeTrack(ctx, postgres.DislikeTrackParams{
		TrackID: uuid.MustParse(in.TrackID),
		UserID:  uuid.MustParse(in.UserID),
	})
	if err != nil {
		return e.NewFrom("disliking track", err)
	}

	return nil
}
