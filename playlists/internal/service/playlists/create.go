package playlists

import (
	"context"
	client "github.com/Benzogang-Tape/audio-hosting/playlists/internal/client/songs"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/service"
	"time"

	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/models"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/storage/postgres"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/logger"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/pgconv"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CreatePlaylistInput struct {
	Title    string
	TrackIDs []uuid.UUID
	AuthorID uuid.UUID
	CoverURL string
	IsAlbum  bool
}

func (s *ServicePlaylists) CreatePlaylist(ctx context.Context, in CreatePlaylistInput) (models.Playlist, error) {
	log := logger.GetLoggerFromCtx(ctx)
	log.Debug(ctx, "creating playlist", zap.String("layout", "service/playlists"))

	playlistParam := postgres.SavePlaylistParams{
		ID:        uuid.New(),
		Title:     in.Title,
		AuthorID:  in.AuthorID,
		TrackIds:  in.TrackIDs,
		CoverUrl:  pgconv.Text(in.CoverURL),
		CreatedAt: time.Now(),
		IsAlbum:   in.IsAlbum,
	}

	err := s.repo.SavePlaylist(ctx, playlistParam)
	if err != nil {
		log.Error(ctx, "failed to save playlist", zap.String("layout", "service/playlists"), zap.Error(err))
		return models.Playlist{}, service.ErrSavingPlaylist
	}

	playlistMetadata := models.PlaylistMetadata{ //nolint:exhaustruct
		ID:             playlistParam.ID.String(),
		Title:          playlistParam.Title,
		AuthorID:       in.AuthorID.String(),
		CoverURL:       in.CoverURL,
		CreatedAt:      playlistParam.CreatedAt,
		IsAlbum:        in.IsAlbum,
		IsMyCollection: false,
		IsPublic:       false,
	}

	songs, err := s.getSongs(ctx, playlistParam.TrackIds)
	if err != nil {
		log.Error(ctx, "failed to get songs", zap.String("layout", "service/playlists"), zap.Error(err))

		return models.Playlist{
			Metadata: playlistMetadata,
			Songs:    []client.Song{},
		}, service.ErrGetSongs
	}

	return models.Playlist{
		Metadata: playlistMetadata,
		Songs:    songs,
	}, nil
}

func convertUUIDtoString(in []uuid.UUID) []string {
	var out []string
	for _, id := range in {
		out = append(out, id.String())
	}

	return out
}
