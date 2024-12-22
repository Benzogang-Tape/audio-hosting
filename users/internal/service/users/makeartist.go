package users

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type DTOMakeArtistInput struct {
	UserID uuid.UUID
}

func (se *Service) MakeArtist(ctx context.Context, input DTOMakeArtistInput) error {
	err := se.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		err := se.usersRepository.MakeArtist(ctx, input.UserID)
		if err != nil {
			return fmt.Errorf("users.Service.MakeArtist - make artist: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("users.Service.MakeArtist - within transaction: %w", err)
	}

	return nil
}
