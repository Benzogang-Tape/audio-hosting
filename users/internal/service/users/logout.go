package users

import (
	"context"
	"fmt"
)

type DTOLogoutInput struct {
	RefreshToken string
}

func (se *Service) Logout(ctx context.Context, input DTOLogoutInput) error {
	err := se.refreshSessionsRepository.DeleteRefreshSession(ctx, input.RefreshToken)
	if err != nil {
		return fmt.Errorf(
			"users.Service.Logout - delete refresh session: %w",
			err,
		)
	}

	return nil
}
