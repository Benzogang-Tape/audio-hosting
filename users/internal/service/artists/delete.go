package artists

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type DTODeleteArtistInput struct {
	ArtsitID uuid.UUID
}

func (se *Service) DeleteArtist(ctx context.Context, input DTODeleteArtistInput) error {
	err := se.artistsRepository.DeleteArtist(ctx, input.ArtsitID)
	if err != nil {
		return fmt.Errorf(
			"artists.Service.DeleteArtist - delete artist: %w",
			err,
		)
	}

	return nil
}
