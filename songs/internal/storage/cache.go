package storage

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Benzogang-Tape/audio-hosting/songs/internal/storage/postgres"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/logger"

	"dev.gaijin.team/go/golib/e"
	"dev.gaijin.team/go/golib/fields"
	"github.com/google/uuid"
)

func (s *Storage) ReleasedSongs(ctx context.Context, params postgres.ReleasedSongsParams,
) ([]postgres.ReleasedSongsRow, error) {
	if !params.ByIds {
		return s.PgStorage.ReleasedSongs(ctx, params) //nolint:wrapcheck
	}

	const initRestIdsCap = 2

	cache := s.RedStorage.With("released")
	result := make([]postgres.ReleasedSongsRow, 0, len(params.Ids))
	restIds := make([]uuid.UUID, 0, initRestIdsCap)
	log := logger.FromContext(ctx)

	for _, id := range params.Ids {
		songBytes, err := cache.GetBytes(ctx, id.String())
		if err != nil {
			log.Trace().Err(err).Stringer("song_id", id).Msg("one of the songs not found in cache")
			restIds = append(restIds, id)

			continue
		}

		var song postgres.ReleasedSongsRow

		err = json.Unmarshal(songBytes, &song) //nolint:musttag
		if err != nil {
			log.Warn().Err(err).Stringer("song_id", id).Msg("one of songs from cache is not JSON")
			restIds = append(restIds, id)

			continue
		}

		result = append(result, song)
	}

	if len(restIds) == 0 {
		log.Debug().Msg("all songs found in cache")
		return result, nil
	}

	log.Debug().Array("ids", logger.Stringers[uuid.UUID](restIds)).Msg("some songs not found in cache")

	params.Ids = restIds

	restRows, err := s.PgStorage.ReleasedSongs(ctx, params)
	if err != nil {
		return nil, e.NewFrom("getting songs from db", err)
	}

	go func() { //nolint:contextcheck
		cacheCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		for _, row := range restRows {
			songBytes, err := json.Marshal(row) //nolint:musttag
			if err != nil {
				log.Warn().Err(err).Stringer("song_id", row.Song.SongID).
					Msg("one of songs from db could not be marshalled to JSON")

				continue
			}

			err = cache.SetBytes(cacheCtx, row.Song.SongID.String(), songBytes, s.c.SongsTtl)
			if err != nil {
				log.Warn().Err(err).Stringer("song_id", row.Song.SongID).
					Msg("one of songs from db could not be saved to cache")
			}
		}
	}()

	return append(result, restRows...), nil
}

func (s *Storage) MySongs(ctx context.Context, params postgres.MySongsParams) ([]postgres.MySongsRow, error) {
	if !params.ByIds {
		return s.PgStorage.MySongs(ctx, params) //nolint:wrapcheck
	}

	cache := s.RedStorage.With("my")
	result := make([]postgres.MySongsRow, 0, len(params.Ids))
	restIds := make([]uuid.UUID, 0, len(params.Ids))
	log := logger.FromContext(ctx)

	for _, id := range params.Ids {
		songBytes, err := cache.GetBytes(ctx, id.String())
		if err != nil {
			log.Trace().Err(err).Stringer("song_id", id).Msg("one of the songs not found in cache")
			restIds = append(restIds, id)

			continue
		}

		var song postgres.MySongsRow

		err = json.Unmarshal(songBytes, &song) //nolint:musttag
		if err != nil {
			log.Warn().Err(err).Stringer("song_id", id).Msg("one of songs from cache is not JSON")
			restIds = append(restIds, id)

			continue
		}

		result = append(result, song)
	}

	if len(restIds) == 0 {
		log.Debug().Msg("all songs found in cache")
		return result, nil
	}

	log.Debug().Array("ids", logger.Stringers[uuid.UUID](restIds)).Msg("some songs not found in cache")

	params.Ids = restIds

	restRows, err := s.PgStorage.MySongs(ctx, params)
	if err != nil {
		return nil, e.NewFrom("getting songs from db", err)
	}

	go func() { //nolint:contextcheck
		cacheCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		for _, row := range restRows {
			songBytes, err := json.Marshal(row) //nolint:musttag
			if err != nil {
				log.Warn().Err(err).Stringer("song_id", row.Song.SongID).
					Msg("one of songs from db could not be marshalled to JSON")

				continue
			}

			err = cache.SetBytes(cacheCtx, row.Song.SongID.String(), songBytes, s.c.MySongsTtl)
			if err != nil {
				log.Warn().Err(err).Stringer("song_id", row.Song.SongID).
					Msg("one of songs from db could not be saved to cache")
			}
		}
	}()

	return append(result, restRows...), nil
}

func (s *Storage) Song(ctx context.Context, id uuid.UUID) (postgres.SongRow, error) {
	cache := s.RedStorage.With("released")
	log := logger.FromContext(ctx)

	songBytes, err := cache.GetBytes(ctx, id.String())
	if err == nil {
		var song postgres.SongRow

		err = json.Unmarshal(songBytes, &song) //nolint:musttag
		if err == nil {
			log.Trace().Stringer("song_id", id).Msg("song found in cache")

			return song, nil
		}

		log.Warn().Err(err).Stringer("song_id", id).Msg("song from cache is not JSON")
	}

	log.Debug().Stringer("song_id", id).Msg("song not found in cache")

	song, err := s.PgStorage.Song(ctx, id)
	if err != nil {
		return postgres.SongRow{}, e.NewFrom("getting song from db", err, fields.F("song_id", id))
	}

	go func() { //nolint:contextcheck
		cacheCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		songBytes, err := json.Marshal(song) //nolint:musttag
		if err != nil {
			log.Warn().Err(err).Stringer("song_id", id).Msg("song from db could not be marshalled to JSON")

			return
		}

		err = cache.SetBytes(cacheCtx, id.String(), songBytes, s.c.SongsTtl)
		if err != nil {
			log.Warn().Err(err).Stringer("song_id", id).Msg("song from db could not be saved to cache")
		}
	}()

	return song, nil
}

func (s *Storage) PatchSong(ctx context.Context, params postgres.PatchSongParams) (postgres.Song, error) {
	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		log := logger.FromContext(ctx)

		log.Trace().Stringer("song_id", params.ID).Msg("song patched, invalidating cache")

		cache1 := s.RedStorage.With("released")
		cache2 := s.RedStorage.With("my")

		err := cache1.Del(cacheCtx, params.ID.String()) //nolint:contextcheck
		if err != nil {
			log.Warn().Err(err).Str("ns", "released").Stringer("song_id", params.ID).
				Msg("error deleting song from cache")
		}

		err = cache2.Del(cacheCtx, params.ID.String()) //nolint:contextcheck
		if err != nil {
			log.Warn().Err(err).Str("ns", "my").Stringer("song_id", params.ID).
				Msg("error deleting song from cache")
		}
	}()

	return s.PgStorage.PatchSong(ctx, params) //nolint:wrapcheck
}

func (s *Storage) PatchSongs(ctx context.Context, params postgres.PatchSongsParams) error {
	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		log := logger.FromContext(ctx)

		log.Trace().Array("song_ids", logger.Stringers[uuid.UUID](params.Ids)).
			Msg("songs patched, invalidating cache")

		cache1 := s.RedStorage.With("released")
		cache2 := s.RedStorage.With("my")

		for _, id := range params.Ids {
			err := cache1.Del(cacheCtx, id.String()) //nolint:contextcheck
			if err != nil {
				log.Warn().Err(err).Str("ns", "released").Stringer("song_id", id).
					Msg("error deleting song from cache")
			}

			err = cache2.Del(cacheCtx, id.String()) //nolint:contextcheck
			if err != nil {
				log.Warn().Err(err).Str("ns", "my").Stringer("song_id", id).
					Msg("error deleting song from cache")
			}
		}
	}()

	return s.PgStorage.PatchSongs(ctx, params) //nolint:wrapcheck
}

func (s *Storage) UpdateSong(ctx context.Context, params postgres.UpdateSongParams) (postgres.Song, error) {
	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		log := logger.FromContext(ctx)

		log.Trace().Stringer("song_id", params.SongID).Msg("song updated, invalidating cache")

		// I guess, it is not able to edit a released song
		cache := s.RedStorage.With("my")

		err := cache.Del(cacheCtx, params.SongID.String()) //nolint:contextcheck
		if err != nil {
			log.Warn().Err(err).Stringer("song_id", params.SongID).Msg("error deleting song from cache")
		}
	}()

	return s.PgStorage.UpdateSong(ctx, params) //nolint:wrapcheck
}

func (s *Storage) DeleteSong(ctx context.Context, ids []uuid.UUID) error {
	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		log := logger.FromContext(ctx)

		log.Trace().Array("ids", logger.Stringers[uuid.UUID](ids)).Msg("songs deleted, invalidating cache")

		cache1 := s.RedStorage.With("my")
		cache2 := s.RedStorage.With("released")

		for _, id := range ids {
			err := cache1.Del(cacheCtx, id.String()) //nolint:contextcheck
			if err != nil {
				log.Warn().Err(err).Stringer("song_id", id).Str("ns", "my").
					Msg("error deleting song from cache")
			}

			err = cache2.Del(cacheCtx, id.String()) //nolint:contextcheck
			if err != nil {
				log.Warn().Err(err).Stringer("song_id", id).Str("ns", "released").
					Msg("error deleting song from cache")
			}
		}
	}()

	return s.PgStorage.DeleteSongs(ctx, ids) //nolint:wrapcheck

}
