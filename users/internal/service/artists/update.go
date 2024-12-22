package artists

import (
	"context"
	"fmt"

	"github.com/Benzogang-Tape/audio-hosting/users/internal/model"
	"github.com/google/uuid"
)

type DTOUpdateArtistInput struct {
	ArtistID  uuid.UUID
	Name      string
	Email     string
	Label     string
	AvatarURL string
}

func (se *Service) UpdateArtist(ctx context.Context, input DTOUpdateArtistInput) error {
	artist := model.Artist{
		User: model.User{
			ID:        input.ArtistID,
			Name:      input.Name,
			Email:     input.Email,
			AvatarURL: input.AvatarURL,
		},
		Label: input.Label,
	}

	err := se.artistsRepository.UpdateArtist(ctx, artist)
	if err != nil {
		return fmt.Errorf(
			"artists.Service.UpdateArtist - update artist: %w",
			err,
		)
	}

	return nil
}
