package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/Benzogang-Tape/audio-hosting/users/internal/repo/pqerrs"
	"github.com/Benzogang-Tape/audio-hosting/users/internal/service"
)

type DTOLoginInput struct {
	Email    string
	Password string
}

type DTOLoginOutput struct {
	RefreshToken string
	AccessToken  string
}

func (se *Service) Login(ctx context.Context, input DTOLoginInput) (DTOLoginOutput, error) {
	user, err := se.usersRepository.GetUserByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, pqerrs.ErrNotFound) {
			return DTOLoginOutput{}, service.ErrUserNotFound
		}

		return DTOLoginOutput{}, fmt.Errorf(
			"users.Service.Login - get user by email: %w",
			err,
		)
	}

	if err := se.hasher.CheckPasswordHash(input.Password, user.PasswordHash); err != nil {
		return DTOLoginOutput{}, service.ErrWrongPassword
	}

	isArtist := false
	_, err = se.artistsRepository.GetArtist(ctx, user.ID)

	switch {
	case err != nil && !errors.Is(err, pqerrs.ErrNotFound):
		return DTOLoginOutput{}, fmt.Errorf(
			"users.Service.Login - get artist: %w",
			err,
		)

	case errors.Is(err, pqerrs.ErrNotFound):
		isArtist = false

	case err == nil:
		isArtist = true

	}

	var accessToken, refreshToken string
	err = se.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		accessToken, refreshToken, err = se.createTokensPair(ctx, user.ID, isArtist)
		if err != nil {
			return fmt.Errorf(
				"users.Service.Login - create tokens pair: %w",
				err,
			)
		}

		return nil
	})
	if err != nil {
		return DTOLoginOutput{}, fmt.Errorf(
			"users.Service.Login - within transaction: %w",
			err,
		)
	}

	return DTOLoginOutput{
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
	}, nil
}
