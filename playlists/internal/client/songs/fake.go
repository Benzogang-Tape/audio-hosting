package songs

import (
	"context"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"sync"
	"time"
)

type FakeClient struct {
	memo sync.Map
}

func NewFakeClient() *FakeClient {
	return &FakeClient{
		memo: sync.Map{},
	}
}

func (c *FakeClient) GetSong(_ context.Context, id string) (Song, error) {
	return c.randomSong(id), nil
}

func (c *FakeClient) GetSongs(_ context.Context, ids []string) ([]Song, error) {
	songs := make([]Song, 0, len(ids))
	for _, id := range ids {
		songs = append(songs, c.randomSong(id))
	}

	return songs, nil
}

func (c *FakeClient) randomSong(id string) Song {
	if v, ok := c.memo.Load(id); ok {
		return v.(Song)
	}

	song := Song{
		ID:          id,
		Singer:      c.randomArtist(uuid.New()),
		Artists:     nil,
		Name:        gofakeit.BookTitle(),
		SongURL:     gofakeit.URL(),
		ImageURL:    gofakeit.URL(),
		Duration:    time.Duration(gofakeit.Int()),
		WeightBytes: gofakeit.Int32(),
		ReleasedAt:  gofakeit.Date(),
	}
	c.memo.Store(id, song)

	return song
}

func (c *FakeClient) randomArtist(id uuid.UUID) Artist {
	if v, ok := c.memo.Load(id); ok {
		if artist, ok := v.(Artist); ok {
			return artist
		}
	}

	artist := Artist{
		ID:        id.String(),
		Name:      gofakeit.Name(),
		Label:     gofakeit.Company(),
		AvatarURL: gofakeit.URL() + gofakeit.LoremIpsumWord() + ".png",
	}

	c.memo.Store(id, artist)

	return artist
}
