package app

import (
	"context"

	"github.com/Benzogang-Tape/audio-hosting/users/internal/model"
	"github.com/Benzogang-Tape/audio-hosting/users/internal/repo/options"
	"github.com/Benzogang-Tape/audio-hosting/users/internal/repo/pqrepo"

	"github.com/google/uuid"
)

type (
	UsersRepository interface {
		GetFollowers(ctx context.Context, userID uuid.UUID) ([]model.User, error)
		GetFollowed(ctx context.Context, userID uuid.UUID) ([]model.User, error)
		Follow(ctx context.Context, followerID uuid.UUID, followedID uuid.UUID) error
		Unfollow(ctx context.Context, followerID uuid.UUID, followedID uuid.UUID) error
		UpdateNotificationsSettings(ctx context.Context, settings model.NotificationSettings) error
		GetNotificationsSettings(
			ctx context.Context,
			userID uuid.UUID,
		) (model.NotificationSettings, error)
		GetUserByEmail(ctx context.Context, email string) (model.User, error)
		MakeArtist(ctx context.Context, userID uuid.UUID) error
	}
	ListenersRepository interface {
		GetListener(ctx context.Context, id uuid.UUID) (model.Listener, error)
		UpdateListener(ctx context.Context, listener model.Listener) error
		DeleteListener(ctx context.Context, id uuid.UUID) error
		GetListeners(ctx context.Context) ([]model.Listener, error)
		CreateListener(ctx context.Context, listener model.Listener) error
	}
	ArtistsRepository interface {
		GetArtistsCount(ctx context.Context, options options.Options) (int, error)
		GetArtists(ctx context.Context, options options.Options) ([]model.Artist, error)
		GetArtist(ctx context.Context, id uuid.UUID) (model.Artist, error)
		UpdateArtist(ctx context.Context, artist model.Artist) error
		DeleteArtist(ctx context.Context, id uuid.UUID) error
	}
	RefreshSessionsRepository interface {
		CreateRefreshSession(ctx context.Context, session model.RefreshSession) error
		GetRefreshSession(ctx context.Context, token string) (model.RefreshSession, error)
		DeleteRefreshSession(ctx context.Context, token string) error
		GetUserRefreshSession(
			ctx context.Context,
			userID uuid.UUID,
		) ([]model.RefreshSession, error)
	}
)

func (p *Provider) ListenersRepository(ctx context.Context) ListenersRepository {
	if p.listenersRepository == nil {
		p.listenersRepository = pqrepo.NewListeners(p.DB(ctx))
	}

	return p.listenersRepository
}

func (p *Provider) ArtistsRepository(ctx context.Context) ArtistsRepository {
	if p.artistsRepository == nil {
		p.artistsRepository = pqrepo.NewArtists(p.DB(ctx))
	}

	return p.artistsRepository
}

func (p *Provider) RefreshSessionsRepository(ctx context.Context) RefreshSessionsRepository {
	if p.refreshSessionsRepository == nil {
		p.refreshSessionsRepository = pqrepo.NewRefreshSessions(p.DB(ctx))
	}

	return p.refreshSessionsRepository
}

func (p *Provider) UsersRepository(ctx context.Context) UsersRepository {
	if p.usersRepository == nil {
		p.usersRepository = pqrepo.NewUsers(p.DB(ctx))
	}

	return p.usersRepository
}
