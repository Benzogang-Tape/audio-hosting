package songs

import (
	"context"
	"errors"
	"time"

	"github.com/Benzogang-Tape/audio-hosting/songs/internal/clients/users"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/storage/postgres"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/erix"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/logger"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/pgconv"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/repoerrs"

	"dev.gaijin.team/go/golib/e"
	"github.com/AlekSi/pointer"
	"github.com/google/uuid"
)

var (
	ErrSongNotFound    = erix.NewStatus("song not found", erix.CodeNotFound)
	ErrMultipleFilters = erix.NewStatus("multiple filters not allowed", erix.CodeBadRequest)
	ErrNoFilters       = erix.NewStatus("no filters provided", erix.CodeBadRequest)
)

type GetSongInput struct {
	Id uuid.UUID
}

type GetSongOutput struct {
	Id          uuid.UUID
	Singer      users.Artist
	Artists     []users.Artist
	Name        string
	SongUrl     string
	ImageUrl    *string
	Duration    time.Duration
	WeightBytes int32
	UploadedAt  time.Time
	ReleasedAt  *time.Time
}

func (s *Service) GetSong(ctx context.Context, input GetSongInput) (GetSongOutput, error) {
	var (
		null GetSongOutput
		log  = logger.FromContext(ctx)
	)

	log.Debug().
		Str("song_id", input.Id.String()).
		Msg("getting song")

	song, err := s.songRepo.Song(ctx, input.Id)

	switch {
	case errors.Is(err, repoerrs.ErrEmptyResult):
		return null, ErrSongNotFound.Wrap(err)

	case err != nil:
		return null, e.NewFrom("getting song", err)
	}

	artists, err := s.artists(ctx, song.Song.SingerFk, song.ArtistsIds)
	if err != nil {
		log.Warn().Err(err).Msg("error getting artists")
		return null, err
	}

	return GetSongOutput{
		Id:          song.Song.SongID,
		Singer:      artists.Singer(),
		Artists:     artists.Artists(),
		Name:        song.Song.Name,
		SongUrl:     song.Song.S3ObjectName.String,
		ImageUrl:    pgconv.FromText(song.Song.ImageUrl),
		Duration:    *pgconv.FromInterval(song.Song.Duration),
		WeightBytes: *pgconv.FromInt4(song.Song.WeightBytes),
		UploadedAt:  song.Song.UploadedAt,
		ReleasedAt:  pgconv.FromTimestamptz(song.Song.ReleasedAt),
	}, nil
}

type GetSongsInput struct {
	ArtistId    *uuid.UUID
	MatchArtist *string
	MatchName   *string
	Ids         []uuid.UUID
	// pagination
	Page     int32
	PageSize int32
}

type Song struct {
	Id          uuid.UUID
	Singer      users.Artist
	Artists     []users.Artist
	Name        string
	SongUrl     string
	ImageUrl    *string
	Duration    time.Duration
	WeightBytes int32
	UploadedAt  time.Time
	ReleasedAt  time.Time
}

type GetSongsOutput struct {
	Songs []Song
	// pagination
	LastPage int32
}

func (s *Service) GetSongs(ctx context.Context, input GetSongsInput) (GetSongsOutput, error) {
	var (
		null = GetSongsOutput{Songs: []Song{}, LastPage: 0}
		log  = logger.FromContext(ctx)
	)

	log.Debug().
		Interface("input", input).
		Msg("getting songs")

	if err := checkFiltersCount(input); err != nil {
		return null, err
	}

	var (
		rows paginatedRows[postgres.ReleasedSongsRow]
		err  error
	)

	switch {
	case len(input.Ids) > 0:
		rows, err = s.getSongsWithIds(ctx, input.Ids)

	case input.MatchName != nil && len(*input.MatchName) > 0:
		rows, err = s.getSongsMatchName(ctx, *input.MatchName, input.Page, input.PageSize)

	case input.ArtistId != nil:
		rows, err = s.getSongsWithArtistId(ctx, *input.ArtistId, input.Page, input.PageSize)

	case input.MatchArtist != nil:
		rows, err = s.getSongsMatchArtist(ctx, *input.MatchArtist, input.Page, input.PageSize)

	}

	if err != nil {
		return null, err
	}

	songsCh := artistsOrderedFanOut(ctx, rows.Rows, s,
		func(row postgres.ReleasedSongsRow, a artists) Song {
			return Song{
				Id:          row.Song.SongID,
				Singer:      a.Singer(),
				Artists:     a.Artists(),
				Name:        row.Song.Name,
				SongUrl:     s.rawService.SongUrl(row.Song.S3ObjectName.String),
				ImageUrl:    pgconv.FromText(row.Song.ImageUrl),
				Duration:    pointer.Get(pgconv.FromInterval(row.Song.Duration)),
				WeightBytes: row.Song.WeightBytes.Int32,
				UploadedAt:  row.Song.UploadedAt,
				ReleasedAt:  row.Song.ReleasedAt.Time,
			}
		},
	)

	outSongs := orderedFanIn(songsCh, len(rows.Rows))

	log.Debug().
		Int("count", len(outSongs)).
		Int("init_count", len(rows.Rows)).
		Int("lost", len(rows.Rows)-len(outSongs)).
		Msg("got songs")

	return GetSongsOutput{
		Songs:    outSongs,
		LastPage: rows.LastPage,
	}, nil
}

func checkFiltersCount(input GetSongsInput) error {
	filtersSum := bit(input.ArtistId != nil) +
		bit(input.MatchArtist != nil) +
		bit(input.MatchName != nil && len(*input.MatchName) > 0) +
		bit(len(input.Ids) > 0)
	if filtersSum == 0 {
		return ErrNoFilters
	} else if filtersSum > 1 {
		return ErrMultipleFilters
	}

	return nil
}

type paginatedRows[T any] struct {
	Rows []T
	// pagination
	LastPage int32
}

func (s *Service) getSongsWithArtistId(ctx context.Context,
	artistId uuid.UUID,
	page, pageSize int32,
) (paginatedRows[postgres.ReleasedSongsRow], error) {
	var (
		null = paginatedRows[postgres.ReleasedSongsRow]{Rows: []postgres.ReleasedSongsRow{}, LastPage: 0}
		log  = logger.FromContext(ctx)
	)

	params := postgres.ReleasedSongsParams{ //nolint:exhaustruct
		BySinger:   false,
		WithArtist: true,
		SingersIds: []uuid.UUID{artistId},
		Limitv:     pageSize,
		Offsetv:    (page - 1) * pageSize,
	}

	log.Debug().
		Str("artist_id", artistId.String()).
		Msg("getting songs by artist id")

	rows, err := s.releasedSongsFromRepo(ctx, params)
	if err != nil {
		return null, err
	}

	songsCount, err := s.songRepo.CountSongsWithArtistsIds(ctx, []uuid.UUID{artistId})
	if err != nil {
		return null, e.NewFrom("getting songs count", err)
	}

	return paginatedRows[postgres.ReleasedSongsRow]{
		Rows:     rows,
		LastPage: (songsCount-1)/pageSize + 1,
	}, nil
}

func (s *Service) getSongsMatchArtist(ctx context.Context,
	name string,
	page, pageSize int32,
) (paginatedRows[postgres.ReleasedSongsRow], error) {
	var (
		null = paginatedRows[postgres.ReleasedSongsRow]{Rows: []postgres.ReleasedSongsRow{}, LastPage: 0}
		log  = logger.FromContext(ctx)
	)

	log.Debug().
		Str("artist_name", name).
		Msg("getting songs by artist name")

	artists, err := s.artistsMatchingName(ctx, name)

	switch {
	case errors.Is(err, ErrArtistsNotFound):
		return null, nil

	case err != nil:
		return null, err
	}

	artistsIds := make([]uuid.UUID, len(artists))
	for i, a := range artists {
		artistsIds[i] = a.Id
	}

	log.Debug().
		Int("count", len(artists)).
		Array("ids", logger.Stringers[uuid.UUID](artistsIds)).
		Msg("got artists by name")

	params := postgres.ReleasedSongsParams{ //nolint:exhaustruct
		WithArtist: true,
		SingersIds: artistsIds,
		Limitv:     pageSize,
		Offsetv:    (page - 1) * pageSize,
	}

	log.Debug().Msg("getting songs by artist ids")

	rows, err := s.releasedSongsFromRepo(ctx, params)
	if err != nil {
		return null, err
	}

	songsCount, err := s.songRepo.CountSongsWithArtistsIds(ctx, artistsIds)
	if err != nil {
		return null, e.NewFrom("getting songs count", err)
	}

	return paginatedRows[postgres.ReleasedSongsRow]{
		Rows:     rows,
		LastPage: (songsCount-1)/pageSize + 1,
	}, nil
}

func (s *Service) getSongsMatchName(ctx context.Context,
	name string,
	page, pageSize int32,
) (paginatedRows[postgres.ReleasedSongsRow], error) {
	var (
		null = paginatedRows[postgres.ReleasedSongsRow]{Rows: []postgres.ReleasedSongsRow{}, LastPage: 0}
		log  = logger.FromContext(ctx)
	)

	params := postgres.ReleasedSongsParams{ //nolint:exhaustruct
		ByName:    true,
		MatchName: name,
		Limitv:    pageSize,
		Offsetv:   (page - 1) * pageSize,
	}

	log.Debug().
		Str("name", name).
		Msg("getting songs by name")

	rows, err := s.releasedSongsFromRepo(ctx, params)
	if err != nil {
		return null, err
	}

	songsCount, err := s.songRepo.CountSongsMatchName(ctx, name)
	if err != nil {
		return null, e.NewFrom("getting songs count", err)
	}

	return paginatedRows[postgres.ReleasedSongsRow]{
		Rows:     rows,
		LastPage: (songsCount-1)/pageSize + 1,
	}, nil
}

func (s *Service) getSongsWithIds(ctx context.Context,
	ids []uuid.UUID,
) (paginatedRows[postgres.ReleasedSongsRow], error) {
	var (
		null = paginatedRows[postgres.ReleasedSongsRow]{Rows: []postgres.ReleasedSongsRow{}, LastPage: 0}
		log  = logger.FromContext(ctx)
	)

	params := postgres.ReleasedSongsParams{ //nolint:exhaustruct
		ByIds:   true,
		Ids:     ids,
		Limitv:  int32(len(ids)), //nolint:gosec
		Offsetv: 0,
	}

	log.Debug().
		Array("ids", logger.Stringers[uuid.UUID](ids)).
		Msg("getting songs by ids")

	rows, err := s.releasedSongsFromRepo(ctx, params)
	if err != nil {
		return null, err
	}

	return paginatedRows[postgres.ReleasedSongsRow]{
		Rows:     rows,
		LastPage: 1,
	}, nil
}

type GetMySongsInput struct {
	UserId uuid.UUID // must be an artist
	ByIds  bool
	Ids    []uuid.UUID
	// pagination
	Page     int32
	PageSize int32
}

type MySong struct {
	Id          uuid.UUID
	Singer      users.Artist
	Artists     []users.Artist
	Name        string
	SongUrl     *string
	ImageUrl    *string
	Duration    *time.Duration
	WeightBytes *int32
	UploadedAt  time.Time
	ReleasedAt  *time.Time
}
type GetMySongsOutput struct {
	Songs    []MySong
	LastPage int32
}

func (s *Service) GetMySongs(ctx context.Context, input GetMySongsInput) (GetMySongsOutput, error) {
	var (
		null = GetMySongsOutput{Songs: []MySong{}, LastPage: 0}
		log  = logger.FromContext(ctx)
	)

	log.Debug().Msg("getting songs")

	params := postgres.MySongsParams{ //nolint:exhaustruct
		SingerID: input.UserId,
		ByIds:    input.ByIds,
	}
	if input.ByIds {
		params.Ids = input.Ids
		params.Limitv = int32(len(input.Ids)) //nolint:gosec
		params.Offsetv = 0
	} else {
		params.Limitv = input.PageSize
		params.Offsetv = (input.Page - 1) * input.PageSize
	}

	songs, err := s.songRepo.MySongs(ctx, params)

	switch {
	case errors.Is(err, repoerrs.ErrEmptyResult):
		log.Debug().Stringer("user_id", input.UserId).Msg("no songs found")
		return null, nil

	case err != nil:
		return null, e.NewFrom("getting songs", err)
	}

	log.Debug().Int("count", len(songs)).Msg("got songs")
	log.Debug().Msg("getting songs count")

	songsCount := int32(len(input.Ids)) //nolint:gosec

	if !input.ByIds {
		songsCount, err = s.songRepo.CountMySongs(ctx, input.UserId)
		if err != nil {
			return null, e.NewFrom("getting songs count", err)
		}
	}

	songsCh := artistsOrderedFanOut(ctx, songs, s,
		func(row postgres.MySongsRow, a artists) MySong {
			songUrl := pgconv.FromText(row.Song.S3ObjectName)
			if songUrl != nil {
				songUrlAboba := s.rawService.SongUrl(*songUrl)
				songUrl = &songUrlAboba
			}

			return MySong{
				Id:          row.Song.SongID,
				Singer:      a.Singer(),
				Artists:     a.Artists(),
				Name:        row.Song.Name,
				SongUrl:     songUrl,
				ImageUrl:    pgconv.FromText(row.Song.ImageUrl),
				Duration:    pgconv.FromInterval(row.Song.Duration),
				WeightBytes: pgconv.FromInt4(row.Song.WeightBytes),
				UploadedAt:  row.Song.UploadedAt,
				ReleasedAt:  pgconv.FromTimestamptz(row.Song.ReleasedAt),
			}
		},
	)

	outSongs := orderedFanIn(songsCh, len(songs))

	log.Debug().
		Int("count", len(outSongs)).
		Int("init_count", len(songs)).
		Int("lost", len(songs)-len(outSongs)).
		Msg("got my songs")

	return GetMySongsOutput{
		Songs:    outSongs,
		LastPage: (songsCount-1)/input.PageSize + 1,
	}, nil
}
