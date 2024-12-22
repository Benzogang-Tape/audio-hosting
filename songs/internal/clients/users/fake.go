package users

import (
	"context"
	"math/rand/v2"
	"sync"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
)

type Fake struct {
	memo sync.Map
}

func NewFake() *Fake {
	return &Fake{
		memo: sync.Map{},
	}
}

func (f *Fake) ArtistsByIds(_ context.Context, ids []uuid.UUID) ([]Artist, error) {
	artists := make([]Artist, len(ids))
	for i := range ids {
		artists[i] = f.randomArtist(ids[i])
	}

	return artists, nil
}

func (*Fake) ArtistsMatchingName(_ context.Context, name string) ([]Artist, error) {
	artists := make([]Artist, rand.Int()%10+1) //nolint:gosec
	for i := range artists {
		artists[i] = Artist{
			Id:        uuid.New(),
			Name:      name + " " + gofakeit.LastName(),
			Label:     gofakeit.Company(),
			AvatarUrl: gofakeit.URL() + gofakeit.LoremIpsumWord() + ".png",
		}
	}

	return artists, nil
}

func (*Fake) Close() error {
	return nil
}

func (f *Fake) randomArtist(id uuid.UUID) Artist {
	if v, ok := f.memo.Load(id); ok {
		if artist, ok := v.(Artist); ok {
			return artist
		}
	}

	artist := Artist{
		Id:        id,
		Name:      gofakeit.Name(),
		Label:     gofakeit.Company(),
		AvatarUrl: gofakeit.URL() + gofakeit.LoremIpsumWord() + ".png",
	}

	f.memo.Store(id, artist)

	return artist
}
