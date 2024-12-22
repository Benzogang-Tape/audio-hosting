package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/Benzogang-Tape/audio-hosting/users/internal/model"
	"github.com/Benzogang-Tape/audio-hosting/users/internal/repo/pqerrs"
	"github.com/Benzogang-Tape/audio-hosting/users/internal/service"

	"github.com/google/uuid"
)

type DTORegistedInput struct {
	Email    string
	Name     string
	Password string
}

type DTORegistedOutput struct {
	AccessToken  string
	RefreshToken string
}

func (se *Service) Register(
	ctx context.Context,
	input DTORegistedInput,
) (DTORegistedOutput, error) {
	passwordHash, err := se.hasher.HashPassword(input.Password)
	if err != nil {
		return DTORegistedOutput{}, fmt.Errorf(
			"users.Service.Register - hash password: %w",
			err,
		)
	}

	user := model.Listener{
		User: model.User{
			ID:           uuid.New(),
			Email:        input.Email,
			Name:         input.Name,
			PasswordHash: passwordHash,
			AvatarURL:    "", // TODO: add avatar gen
		},
	}

	var accessToken, refreshToken string
	err = se.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		if err = se.listenersRepository.CreateListener(ctx, user); err != nil {
			if errors.Is(err, pqerrs.ErrUniqueViolation) {
				return service.ErrUserAlreadyExists
			}

			return fmt.Errorf(
				"users.Service.Register - create listener: %w",
				err,
			)
		}

		accessToken, refreshToken, err = se.createTokensPair(ctx, user.ID, false)
		if err != nil {
			return fmt.Errorf(
				"users.Service.Register - create tokens pair: %w",
				err,
			)
		}

		return nil
	})
	if err != nil {
		return DTORegistedOutput{}, fmt.Errorf("users.Service.Register - transaction: %w", err)
	}

	return DTORegistedOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
