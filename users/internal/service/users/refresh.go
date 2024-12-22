package users

import (
	"context"
	"fmt"
	"time"

	"github.com/Benzogang-Tape/audio-hosting/users/internal/service"
)

type DTORefreshInput struct {
	RefreshToken string
}

type DTORefreshOutput struct {
	AccessToken  string
	RefreshToken string
}

func (se *Service) Refresh(ctx context.Context, input DTORefreshInput) (DTORefreshOutput, error) {
	session, err := se.refreshSessionsRepository.GetRefreshSession(ctx, input.RefreshToken)
	if err != nil {
		return DTORefreshOutput{}, fmt.Errorf(
			"users.Service.Refresh - get refresh session: %w",
			err,
		)
	}

	var accessToken, refreshToken string
	err = se.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		err = se.refreshSessionsRepository.DeleteRefreshSession(ctx, input.RefreshToken)
		if err != nil {
			return fmt.Errorf(
				"users.Service.Refresh - delete refresh session: %w",
				err,
			)
		}

		if session.ExpiresAt.Before(time.Now()) {
			return service.ErrSessionExpired
		}

		accessToken, refreshToken, err = se.createTokensPair(ctx, session.UserID, false)
		if err != nil {
			return fmt.Errorf(
				"users.Service.Refresh - create tokens pair: %w",
				err,
			)
		}

		return nil
	})
	if err != nil {
		return DTORefreshOutput{}, fmt.Errorf(
			"users.Service.Refresh - within transaction: %w",
			err,
		)
	}

	return DTORefreshOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
