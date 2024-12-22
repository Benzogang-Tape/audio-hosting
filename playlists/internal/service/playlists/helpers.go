package playlists

import (
	"context"
	"dev.gaijin.team/go/golib/e"
	"dev.gaijin.team/go/golib/fields"
	"github.com/AlekSi/pointer"
	client "github.com/Benzogang-Tape/audio-hosting/playlists/internal/client/songs"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/models"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/service"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/storage/postgres"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/logger"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/pgconv"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *ServicePlaylists) getSongs(ctx context.Context, trackIDs []uuid.UUID) ([]client.Song, error) {
	if len(trackIDs) == 0 {
		return []client.Song{}, nil
	}

	log := logger.GetLoggerFromCtx(ctx)

	songs, err := s.songs.GetSongs(ctx, convertUUIDtoString(trackIDs))
	if err != nil {
		log.Error(ctx, "failed to get songs", zap.String("layout", "service/playlists"), zap.Error(err))
		return nil, e.From(err, fields.F("GetSongs; getting songs:", err))
	}

	return songs, nil
}

func (*ServicePlaylists) convToPlaylistMetadata(row postgres.Playlist) models.PlaylistMetadata {
	return models.PlaylistMetadata{
		ID:             row.ID.String(),
		Title:          row.Title,
		AuthorID:       row.AuthorID.String(),
		CoverURL:       pointer.Get(pgconv.FromText(row.CoverUrl)),
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      pointer.Get(pgconv.FromTimestamptz(row.UpdatedAt)),
		ReleasedAt:     pointer.Get(pgconv.FromTimestamptz(row.ReleasedAt)),
		IsAlbum:        row.IsAlbum,
		IsMyCollection: false,
		IsPublic:       row.IsPublic,
	}
}

func convertToUUID(ids []string) []uuid.UUID {
	m := make([]uuid.UUID, len(ids))

	for i, id := range ids {
		m[i] = uuid.MustParse(id)
	}

	return m
}

func checkFiltersCount(input GetPlaylistsInput) error {
	filtersSum := bit(input.ArtistID != nil) +
		bit(input.MatchTitle != nil && len(*input.MatchTitle) > 0) +
		bit(len(input.PlaylistIDs) > 0)
	if filtersSum == 0 {
		return service.ErrNoFilters
	} else if filtersSum > 1 {
		return service.ErrMultipleFilters
	}

	return nil
}

func bit(b bool) int {
	if b {
		return 1
	}

	return 0
}
