package artists

import (
	"context"

	"github.com/Benzogang-Tape/audio-hosting/users/internal/model"
	"github.com/Benzogang-Tape/audio-hosting/users/internal/repo/options"
	"github.com/google/uuid"
)

type ArtistsRepository interface {
	GetArtists(ctx context.Context, options options.Options) ([]model.Artist, error)
	GetArtist(ctx context.Context, id uuid.UUID) (model.Artist, error)
	UpdateArtist(ctx context.Context, artist model.Artist) error
	DeleteArtist(ctx context.Context, id uuid.UUID) error
	GetArtistsCount(ctx context.Context, options options.Options) (int, error)
}

type Service struct {
	artistsRepository ArtistsRepository
}

func New(artistRepository ArtistsRepository) *Service {
	return &Service{
		artistsRepository: artistRepository,
	}
}
