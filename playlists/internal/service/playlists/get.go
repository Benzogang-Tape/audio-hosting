package playlists

import (
	"context"
	"errors"
	client "github.com/Benzogang-Tape/audio-hosting/playlists/internal/client/songs"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/service"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/storage/postgres"
	"github.com/jackc/pgx/v5"
	"time"

	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/models"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/transport/grpc/uniceptors"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/logger"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/pgconv"

	"dev.gaijin.team/go/golib/e"
	"dev.gaijin.team/go/golib/fields"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *ServicePlaylists) GetPlaylist(ctx context.Context, playlistID string) (models.Playlist, error) {
	log := logger.GetLoggerFromCtx(ctx)

	log.Debug(ctx, "getting playlist", zap.String("layout", "service/playlists"))

	id := uuid.MustParse(playlistID)

	playlistRow, err := s.repo.Playlist(ctx, id)

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		log.Debug(ctx, "playlist not found", zap.String("layout", "service/playlists"))
		return models.Playlist{}, service.ErrPlaylistNotFound

	case err != nil:
		log.Error(ctx, "failed to get playlist", zap.String("layout", "service/playlists"), zap.Error(err))
		return models.Playlist{}, e.From(err, fields.F("GetPlaylist; getting playlist:", err))
	}

	pl := playlistRow.Playlist

	userToken, ok := uniceptors.TokenFromCtx(ctx)
	if (!ok && !pl.IsPublic) || (userToken.Subject.String() != pl.AuthorID.String() && !pl.IsPublic) {
		return models.Playlist{}, service.ErrPlaylistNotFound
	}

	var updatedAt, releasedAt time.Time

	tmp := pgconv.FromTimestamptz(pl.UpdatedAt)
	if tmp != nil {
		updatedAt = *tmp
	}

	tmp = pgconv.FromTimestamptz(pl.ReleasedAt)
	if tmp != nil {
		releasedAt = *tmp
	}

	playlistMetadata := models.PlaylistMetadata{
		ID:             pl.ID.String(),
		Title:          pl.Title,
		AuthorID:       pl.AuthorID.String(),
		CoverURL:       *pgconv.FromText(pl.CoverUrl),
		CreatedAt:      pl.CreatedAt,
		UpdatedAt:      updatedAt,
		ReleasedAt:     releasedAt,
		IsAlbum:        pl.IsAlbum,
		IsPublic:       pl.IsPublic,
		IsMyCollection: false,
	}

	songs, err := s.getSongs(ctx, pl.TrackIds)
	if err != nil {
		log.Error(ctx, "failed to get songs", zap.String("layout", "service/playlists"), zap.Error(err))

		return models.Playlist{
			Metadata: playlistMetadata,
			Songs:    []client.Song{},
		}, service.ErrGetSongs
	}

	log.Debug(
		ctx, "songs count",
		zap.String("layout", "service/playlists"),
		zap.Int("was", len(pl.TrackIds)),
		zap.Int("got", len(songs)),
		zap.Int("loss", len(pl.TrackIds)-len(songs)),
	)

	return models.Playlist{
		Metadata: playlistMetadata,
		Songs:    songs,
	}, nil
}

type GetPlaylistsInput struct {
	// Pagination
	Page  int32
	Limit int32
	// Filters
	ArtistID    *string
	MatchTitle  *string
	PlaylistIDs []string
}

type GetPlaylistsOutput struct {
	Playlists []models.PlaylistMetadata
	HasNext   bool
	LastPage  int32
}

func (s *ServicePlaylists) GetPlaylists(ctx context.Context, in GetPlaylistsInput) (GetPlaylistsOutput, error) {
	err := checkFiltersCount(in)
	if err != nil {
		return GetPlaylistsOutput{}, err
	}

	var rows paginatedRows[postgres.PublicPlaylistsRow]

	switch {
	case in.ArtistID != nil:
		rows, err = s.getPlaylistsByArtistID(ctx, uuid.MustParse(*in.ArtistID), in.Page, in.Limit)

	case in.MatchTitle != nil:
		rows, err = s.getPlaylistsByTitle(ctx, *in.MatchTitle, in.Page, in.Limit)

	case len(in.PlaylistIDs) > 0:
		rows, err = s.getPlaylistsByIDs(ctx, convertToUUID(in.PlaylistIDs), in.Limit)
	}

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return GetPlaylistsOutput{
			Playlists: []models.PlaylistMetadata{},
			HasNext:   false,
			LastPage:  0,
		}, nil

	case err != nil:
		return GetPlaylistsOutput{}, e.From(err, fields.F("GetPlaylists; getting playlists:", err))
	}

	playlistsMetadata := make([]models.PlaylistMetadata, 0, len(rows.Rows))
	for _, row := range rows.Rows {
		playlistsMetadata = append(playlistsMetadata, s.convToPlaylistMetadata(row.Playlist))
	}

	return GetPlaylistsOutput{
		Playlists: playlistsMetadata,
		HasNext:   in.Page < rows.LastPage,
		LastPage:  rows.LastPage,
	}, nil
}

type paginatedRows[T any] struct {
	Rows     []T
	LastPage int32
}

func (s *ServicePlaylists) getPlaylistsByIDs(
	ctx context.Context,
	ids []uuid.UUID,
	limit int32,
) (paginatedRows[postgres.PublicPlaylistsRow], error) {
	params := postgres.PublicPlaylistsParams{ //nolint:exhaustruct
		ByArtistID: false,
		ByTitle:    false,
		ByIds:      true,
		Ids:        ids,
		Offsetv:    0,
		Limitv:     int32(len(ids)), //nolint:gosec
	}

	rows, err := s.repo.PublicPlaylists(ctx, params)
	if err != nil {
		return paginatedRows[postgres.PublicPlaylistsRow]{}, err
	}

	return paginatedRows[postgres.PublicPlaylistsRow]{
		Rows:     rows,
		LastPage: int32(len(rows))/limit + 1, //nolint:gosec
	}, nil
}

func (s *ServicePlaylists) getPlaylistsByArtistID(
	ctx context.Context,
	artistIDs uuid.UUID,
	page int32,
	limit int32,
) (paginatedRows[postgres.PublicPlaylistsRow], error) {
	params := postgres.PublicPlaylistsParams{ //nolint:exhaustruct
		ByArtistID: true,
		ArtistID:   artistIDs,
		ByTitle:    false,
		ByIds:      false,
		Offsetv:    (page - 1) * limit,
		Limitv:     limit,
	}

	rows, err := s.repo.PublicPlaylists(ctx, params)
	if err != nil {
		return paginatedRows[postgres.PublicPlaylistsRow]{}, err
	}

	return paginatedRows[postgres.PublicPlaylistsRow]{
		Rows:     rows,
		LastPage: int32(len(rows))/limit + 1, //nolint:gosec
	}, nil
}

func (s *ServicePlaylists) getPlaylistsByTitle(
	ctx context.Context,
	title string,
	page int32,
	limit int32,
) (paginatedRows[postgres.PublicPlaylistsRow], error) {
	params := postgres.PublicPlaylistsParams{ //nolint:exhaustruct
		ByArtistID: false,
		ByTitle:    true,
		MatchName:  title,
		ByIds:      false,
		Offsetv:    (page - 1) * limit,
		Limitv:     limit,
	}

	rows, err := s.repo.PublicPlaylists(ctx, params)
	if err != nil {
		return paginatedRows[postgres.PublicPlaylistsRow]{}, err
	}

	return paginatedRows[postgres.PublicPlaylistsRow]{
		Rows:     rows,
		LastPage: int32(len(rows))/limit + 1, //nolint:gosec
	}, nil
}
