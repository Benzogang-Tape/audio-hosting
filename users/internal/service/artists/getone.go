package artists

import (
	"context"
	"fmt"

	"github.com/Benzogang-Tape/audio-hosting/users/internal/model"
	"github.com/google/uuid"
)

type DTOGetArtistInput struct {
	ArtistID uuid.UUID
}

type DTOGetArtistOutput struct {
	Artist model.Artist
}

func (se *Service) GetArtist(
	ctx context.Context,
	input DTOGetArtistInput,
) (DTOGetArtistOutput, error) {
	artist, err := se.artistsRepository.GetArtist(ctx, input.ArtistID)
	if err != nil {
		return DTOGetArtistOutput{}, fmt.Errorf(
			"artists.Service.GetArtist - get artist: %w",
			err,
		)
	}

	return DTOGetArtistOutput{
		Artist: artist,
	}, nil
}
