package songs

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Benzogang-Tape/audio-hosting/songs/internal/clients/users"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/storage/broker"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/storage/postgres"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/logger"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/repoerrs"

	"dev.gaijin.team/go/golib/e"
	"github.com/google/uuid"
)

type artists []users.Artist

func (a artists) Singer() users.Artist {
	return a[len(a)-1]
}

func (a artists) Artists() []users.Artist {
	return a[:len(a)-1]
}

func (s *Service) artists(ctx context.Context, singer uuid.UUID, artists []uuid.UUID) (artists, error) {
	if len(artists) == 1 && artists[0] == uuid.Nil {
		artists = make([]uuid.UUID, 0, 1)
	}

	usersArtists, err := s.userRepo.ArtistsByIds(ctx, append(artists, singer))

	switch {
	case errors.Is(err, repoerrs.ErrEmptyResult):
		return nil, ErrArtistsNotFound.Wrap(err)

	case err != nil:
		return nil, e.NewFrom("getting artists by id", err)
	}

	return usersArtists, nil
}

func (s *Service) artistsMatchingName(ctx context.Context, name string) ([]users.Artist, error) {
	artists, err := s.userRepo.ArtistsMatchingName(ctx, name)

	switch {
	case errors.Is(err, repoerrs.ErrEmptyResult):
		return nil, ErrArtistsNotFound.Wrap(err)

	case err != nil:
		return nil, e.NewFrom("getting artists matching name", err)
	}

	return artists, nil
}

func (s *Service) releasedSongsFromRepo(ctx context.Context,
	params postgres.ReleasedSongsParams,
) ([]postgres.ReleasedSongsRow, error) {
	songs, err := s.songRepo.ReleasedSongs(ctx, params)

	switch {
	case errors.Is(err, repoerrs.ErrEmptyResult):
		return nil, nil

	case err != nil:
		return nil, e.NewFrom("getting songs", err)
	}

	return songs, nil
}

type ordered[T any] struct {
	value T
	valid bool
	order int
}

type songRow interface {
	GetSingerFk() uuid.UUID
	GetArtistsIds() []uuid.UUID
}

type ctorFn[T songRow, T2 any] func(row T, a artists) T2

func artistsOrderedFanOut[T songRow, T2 any](ctx context.Context,
	rows []T,
	s *Service,
	ctor ctorFn[T, T2],
) <-chan ordered[T2] {
	log := logger.FromContext(ctx)
	ch := make(chan ordered[T2])

	go func() {
		wg := sync.WaitGroup{}
		wg.Add(len(rows))

		for i := range rows {
			go func() {
				defer wg.Done()

				artists, err := s.artists(ctx, rows[i].GetSingerFk(), rows[i].GetArtistsIds())
				if err != nil {
					log.Warn().Err(err).Int("order", i).
						Array("ids", logger.Stringers[uuid.UUID](
							append(rows[i].GetArtistsIds(), rows[i].GetSingerFk()))).
						Msg("error getting artists")

					ch <- ordered[T2]{ //nolint:exhaustruct
						valid: false,
						order: i,
					}

					return
				}

				ch <- ordered[T2]{
					value: ctor(rows[i], artists),
					valid: true,
					order: i,
				}
			}()
		}

		wg.Wait()
		close(ch)
	}()

	return ch
}

func orderedFanIn[T any](ch <-chan ordered[T], expectedCount int) []T {
	queue := make(map[int]ordered[T], expectedCount)
	out := make([]T, 0, expectedCount)
	expecting := 0

	for songElem := range ch {
		if songElem.order != expecting {
			queue[songElem.order] = songElem
			continue
		}

		if songElem.valid {
			out = append(out, songElem.value)
		}

		expecting++

		for v, ok := queue[expecting]; ok; v, ok = queue[expecting] {
			if v.valid {
				out = append(out, v.value)
			}

			expecting++
		}
	}

	return out
}

func bit(b bool) int {
	if b {
		return 1
	}

	return 0
}

func (s *Service) sendReleasedMessages(ctx context.Context, songs []postgres.MySongsRow, releaseTime time.Time) error {
	messages := make([]broker.SongReleasedMessage, len(songs))
	for i := range songs {
		messages[i] = broker.SongReleasedMessage{
			SongId:     songs[i].Song.SongID,
			ArtistId:   songs[i].Song.SingerFk,
			Name:       songs[i].Song.Name,
			ReleasedAt: releaseTime,
		}
	}

	err := s.messageBroker.SendReleasedMessages(ctx, messages)
	if err != nil {
		return e.NewFrom("sending messages", err)
	}

	return nil
}
