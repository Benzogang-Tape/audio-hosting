package app

import (
	"context"

	"github.com/Benzogang-Tape/audio-hosting/users/internal/service/artists"
	"github.com/Benzogang-Tape/audio-hosting/users/internal/service/listeners"
	"github.com/Benzogang-Tape/audio-hosting/users/internal/service/users"
)

type (
	UsersService interface {
		Login(context.Context, users.DTOLoginInput) (users.DTOLoginOutput, error)
		Register(context.Context, users.DTORegistedInput) (users.DTORegistedOutput, error)
		Refresh(context.Context, users.DTORefreshInput) (users.DTORefreshOutput, error)
		Logout(context.Context, users.DTOLogoutInput) error
		Follow(context.Context, users.DTOFollowInput) error
		Unfollow(context.Context, users.DTOUnfollowInput) error
		GetFollowed(context.Context, users.DTOGetFollowedInput) (users.DTOGetFollowedOutput, error)
		GetFollowers(
			context.Context,
			users.DTOGetFollowersInput,
		) (users.DTOGetFollowersOutput, error)
		MakeArtist(context.Context, users.DTOMakeArtistInput) error
	}
	ListenersService interface {
		GetListener(
			context.Context,
			listeners.DTOGetListenerInput,
		) (listeners.DTOGetListenerOutput, error)
		GetListeners(
			context.Context,
			listeners.DTOGetListenersInput,
		) (listeners.DTOGetListenersOutput, error)
		UpdateListener(context.Context, listeners.DTOUpdateListenerInput) error
		DeleteListener(context.Context, listeners.DTODeleteListenerInput) error
	}
	ArtistsService interface {
		GetArtist(context.Context, artists.DTOGetArtistInput) (artists.DTOGetArtistOutput, error)
		GetArtists(context.Context, artists.DTOGetArtistsInput) (artists.DTOGetArtistsOutput, error)
		UpdateArtist(context.Context, artists.DTOUpdateArtistInput) error
		DeleteArtist(context.Context, artists.DTODeleteArtistInput) error
	}
)

func (p *Provider) UsersService(ctx context.Context) UsersService {
	if p.usersService == nil {
		p.usersService = users.New(
			p.UsersRepository(ctx),
			p.ArtistsRepository(ctx),
			p.ListenersRepository(ctx),
			p.RefreshSessionsRepository(ctx),
			p.Hasher(),
			p.Signer(),
			p.DB(ctx),
			p.Cfg().Auth.AccessTTL,
			p.Cfg().Auth.RefreshTTL,
			p.Cfg().Auth.RefreshSessionsLimit,
		)
	}

	return p.usersService
}

func (p *Provider) ArtistsService(ctx context.Context) ArtistsService {
	if p.artistsService == nil {
		p.artistsService = artists.New(
			p.ArtistsRepository(ctx),
		)
	}

	return p.artistsService
}

func (p *Provider) ListenersService(ctx context.Context) ListenersService {
	if p.listenersService == nil {
		p.listenersService = listeners.New(
			p.ListenersRepository(ctx),
		)
	}

	return p.listenersService
}
