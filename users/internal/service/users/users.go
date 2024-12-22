package users

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Benzogang-Tape/audio-hosting/users/internal/model"
	"github.com/Benzogang-Tape/audio-hosting/users/internal/repo/pqerrs"
	"github.com/Benzogang-Tape/audio-hosting/users/pkg/auth"

	"github.com/google/uuid"
)

type artistsRepository interface {
	GetArtist(ctx context.Context, id uuid.UUID) (model.Artist, error)
}

type listenersRepository interface {
	CreateListener(ctx context.Context, listener model.Listener) error
}

type refreshSessionsRepository interface {
	CreateRefreshSession(ctx context.Context, session model.RefreshSession) error
	GetUserRefreshSession(
		ctx context.Context,
		userID uuid.UUID,
	) ([]model.RefreshSession, error)
	DeleteRefreshSession(ctx context.Context, token string) error
	GetRefreshSession(
		ctx context.Context,
		token string,
	) (model.RefreshSession, error)
}

type usersRepository interface {
	GetUserByEmail(ctx context.Context, email string) (model.User, error)
	MakeArtist(ctx context.Context, userID uuid.UUID) error
}

type hasher interface {
	HashPassword(password string) (string, error)
	CheckPasswordHash(password, hash string) error
}

type transactor interface {
	WithinTransaction(
		ctx context.Context,
		tFunc func(ctx context.Context) error,
	) error
}

type Service struct {
	usersRepository           usersRepository
	artistsRepository         artistsRepository
	listenersRepository       listenersRepository
	refreshSessionsRepository refreshSessionsRepository

	hasher     hasher
	signer     auth.Signer
	transactor transactor

	accessTokenTTL       time.Duration
	refreshTokenTTL      time.Duration
	refreshSessionsLimit int
}

//nolint:revive
func New(
	usersRepository usersRepository,
	artistsRepository artistsRepository,
	listenersRepository listenersRepository,
	refreshSessionsRepository refreshSessionsRepository,
	hasher hasher,
	signer auth.Signer,
	transactor transactor,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
	refreshSessionsLimit int,
) *Service {
	return &Service{
		usersRepository:           usersRepository,
		artistsRepository:         artistsRepository,
		listenersRepository:       listenersRepository,
		refreshSessionsRepository: refreshSessionsRepository,
		hasher:                    hasher,
		signer:                    signer,
		transactor:                transactor,
		accessTokenTTL:            accessTokenTTL,
		refreshTokenTTL:           refreshTokenTTL,
		refreshSessionsLimit:      refreshSessionsLimit,
	}
}

func (se *Service) createTokensPair(
	ctx context.Context,
	userID uuid.UUID,
	isArtist bool,
) (access string, refresh string, err error) {
	accessToken, err := se.signer.Sign(auth.Token{
		Subject:  userID,
		IsArtist: isArtist,
		Exp:      time.Now().Add(se.accessTokenTTL).Unix(),
	})
	if err != nil {
		return "", "", fmt.Errorf(
			"users.Service.createTokensPair - sign token: %w",
			err,
		)
	}

	refreshSession, err := se.createRefreshSessionInLimit(ctx, userID)
	if err != nil {
		return "", "", fmt.Errorf(
			"users.Service.createTokensPair - create refresh session: %w",
			err,
		)
	}

	return accessToken, refreshSession.RefreshToken, nil
}

func (se *Service) createRefreshSessionInLimit(
	ctx context.Context,
	userID uuid.UUID,
) (model.RefreshSession, error) {
	var refreshSession model.RefreshSession

	sessions, err := se.refreshSessionsRepository.GetUserRefreshSession(ctx, userID)
	if err != nil && !errors.Is(err, pqerrs.ErrNotFound) {
		return model.RefreshSession{}, fmt.Errorf(
			"users.Service.createRefreshSessionInLimit - get refresh sessions: %w",
			err,
		)
	}

	if len(sessions) > se.refreshSessionsLimit {
		err := se.refreshSessionsRepository.DeleteRefreshSession(ctx, sessions[0].RefreshToken)
		if err != nil {
			return model.RefreshSession{}, fmt.Errorf(
				"users.Service.createRefreshSessionInLimit - delete refresh session: %w",
				err,
			)
		}
	}

	refreshSession = model.RefreshSession{
		RefreshToken: uuid.New().String(),
		ExpiresAt:    time.Now().Add(se.refreshTokenTTL),
		UserID:       userID,
	}

	if err := se.refreshSessionsRepository.CreateRefreshSession(ctx, refreshSession); err != nil {
		return model.RefreshSession{}, fmt.Errorf(
			"users.Service.createRefreshSessionInLimit - create refresh session: %w",
			err,
		)
	}

	return refreshSession, nil
}
