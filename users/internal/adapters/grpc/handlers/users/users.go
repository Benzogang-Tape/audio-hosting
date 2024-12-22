package users

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Benzogang-Tape/audio-hosting/users/api/protogen"
	"github.com/Benzogang-Tape/audio-hosting/users/internal/adapters/grpc/handlers"
	"github.com/Benzogang-Tape/audio-hosting/users/internal/repo/options"
	"github.com/Benzogang-Tape/audio-hosting/users/internal/service"
	"github.com/Benzogang-Tape/audio-hosting/users/internal/service/artists"
	"github.com/Benzogang-Tape/audio-hosting/users/internal/service/listeners"
	"github.com/Benzogang-Tape/audio-hosting/users/internal/service/users"
	"github.com/Benzogang-Tape/audio-hosting/users/pkg/logger"
	"github.com/google/uuid"
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

type UsersHandler struct {
	usersService     UsersService
	listenersService ListenersService
	artistsService   ArtistsService

	protogen.UnimplementedUsersServiceServer
}

func NewHandler(
	usersService UsersService,
	listenersService ListenersService,
	artistsService ArtistsService,
) *UsersHandler {
	return &UsersHandler{
		usersService:     usersService,
		listenersService: listenersService,
		artistsService:   artistsService,
	}
}

// auth related things.
func (us *UsersHandler) Login(
	ctx context.Context,
	req *protogen.LoginRequest,
) (*protogen.LoginResponse, error) {
	output, err := us.usersService.Login(ctx, users.DTOLoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound) || errors.Is(err, service.ErrWrongPassword):
			logger.GetLoggerFromCtx(ctx).Info("wrong creds", slog.String("reason", err.Error()))

			return nil, errors.New("wrong creds")
		}

		logger.GetLoggerFromCtx(ctx).
			Error("users.UsersHandler.Login - login", slog.String("error", err.Error()))

		return nil, errors.New("internal server error")
	}

	logger.GetLoggerFromCtx(ctx).Info("user logged in", slog.String("email", req.Email))

	return &protogen.LoginResponse{
		Tokens: &protogen.Tokens{
			AccessToken:  output.AccessToken,
			RefreshToken: output.RefreshToken,
		},
	}, nil
}

func (us *UsersHandler) Register(
	ctx context.Context,
	req *protogen.RegisterRequest,
) (*protogen.RegisterResponse, error) {
	output, err := us.usersService.Register(ctx, users.DTORegistedInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserAlreadyExists):
			logger.GetLoggerFromCtx(ctx).
				Info("user already exists", slog.String("email", req.Email))

			return nil, errors.New("user already exists")
		}

		logger.GetLoggerFromCtx(ctx).
			Error("users.UsersHandler.Register - register", slog.String("error", err.Error()))

		return nil, errors.New("internal server error")
	}

	logger.GetLoggerFromCtx(ctx).Info("user registered", slog.String("email", req.Email))

	return &protogen.RegisterResponse{
		Tokens: &protogen.Tokens{
			AccessToken:  output.AccessToken,
			RefreshToken: output.RefreshToken,
		},
	}, nil
}

func (us *UsersHandler) Refresh(
	ctx context.Context,
	req *protogen.RefreshRequest,
) (*protogen.RefreshResponse, error) {
	output, err := us.usersService.Refresh(ctx, users.DTORefreshInput{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSessionExpired):
			logger.GetLoggerFromCtx(ctx).
				Info("session expired", slog.String("token", req.RefreshToken))

			return nil, errors.New("session expired")
		}

		logger.GetLoggerFromCtx(ctx).
			Error("users.UsersHandler.Refresh - refresh", slog.String("error", err.Error()))

		return nil, errors.New("internal server error")
	}

	logger.GetLoggerFromCtx(ctx).Info("user refreshed tokens")

	return &protogen.RefreshResponse{
		Tokens: &protogen.Tokens{
			AccessToken:  output.AccessToken,
			RefreshToken: output.RefreshToken,
		},
	}, nil
}

func (us *UsersHandler) Logout(
	ctx context.Context,
	req *protogen.LogoutRequest,
) (*protogen.LogoutResponse, error) {
	err := us.usersService.Logout(ctx, users.DTOLogoutInput{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		logger.GetLoggerFromCtx(ctx).
			Error("users.UsersHandler.Logout - logout", slog.String("error", err.Error()))

		return nil, err
	}

	logger.GetLoggerFromCtx(ctx).Info("user logged out")

	return &protogen.LogoutResponse{}, nil
}

// one user related things.
func (us *UsersHandler) GetMe(
	_ context.Context,
	_ *protogen.GetMeRequest,
) (*protogen.GetMeResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (us *UsersHandler) UpdateMe(
	_ context.Context,
	_ *protogen.UpdateMeRequest,
) (*protogen.UpdateMeResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (us *UsersHandler) ChangePassword(
	_ context.Context,
	_ *protogen.ChangePasswordRequest,
) (*protogen.ChangePasswordResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (us *UsersHandler) Follow(
	_ context.Context,
	_ *protogen.FollowRequest,
) (*protogen.FollowResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (us *UsersHandler) Unfollow(
	_ context.Context,
	_ *protogen.UnfollowRequest,
) (*protogen.UnfollowResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (us *UsersHandler) GetFollowers(
	_ context.Context,
	_ *protogen.GetFollowersRequest,
) (*protogen.GetFollowersResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (us *UsersHandler) GetFollowed(
	_ context.Context,
	_ *protogen.GetFollowedRequest,
) (*protogen.GetFollowedResponse, error) {
	panic("not implemented") // TODO: Implement
}

// artist related things.
func (us *UsersHandler) GetArtists(
	ctx context.Context,
	req *protogen.GetArtistsRequest,
) (*protogen.GetArtistsResponse, error) {
	pagination, err := handlers.ParsePagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	sort, err := handlers.ParseSort(req.Sort, handlers.ArtistEntityName)
	if err != nil {
		return nil, err
	}

	filter, err := handlers.ParseFilters(req.Filter, handlers.ArtistEntityName)
	if err != nil {
		return nil, err
	}

	artists, err := us.artistsService.GetArtists(ctx, artists.DTOGetArtistsInput{
		Options: options.Options{
			Pagination: pagination,
			Sort:       sort,
			Filter:     filter,
		},
	})
	if err != nil {
		return nil, err
	}

	outputArtists := make([]*protogen.Artist, 0, len(artists.Artists))
	for _, artist := range artists.Artists {
		outputArtists = append(outputArtists, &protogen.Artist{
			Id:        artist.ID.String(),
			Label:     artist.Label,
			Name:      artist.Name,
			AvatarUrl: artist.AvatarURL,
		})
	}

	return &protogen.GetArtistsResponse{
		Artists: outputArtists,
		Pagination: &protogen.PaginationResponse{
			Total:    int64(artists.Pagination.Total),
			HasNext:  artists.Pagination.HasNext,
			LastPage: int64(artists.Pagination.LastPage),
		},
	}, nil
}

func (us *UsersHandler) GetArtist(
	ctx context.Context,
	input *protogen.GetArtistRequest,
) (*protogen.GetArtistResponse, error) {
	id, err := uuid.Parse(input.Id)
	if err != nil {
		return nil, err
	}

	artist, err := us.artistsService.GetArtist(ctx, artists.DTOGetArtistInput{
		ArtistID: id,
	})
	if err != nil {
		return nil, err
	}

	return &protogen.GetArtistResponse{
		Artist: &protogen.Artist{
			Id:        artist.Artist.ID.String(),
			Label:     artist.Artist.Label,
			Name:      artist.Artist.Name,
			AvatarUrl: artist.Artist.AvatarURL,
		},
	}, nil
}

func (us *UsersHandler) MakeArtist(
	ctx context.Context,
	req *protogen.MakeArtistRequest,
) (*protogen.MakeArtistResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	err = us.usersService.MakeArtist(ctx, users.DTOMakeArtistInput{
		UserID: id,
	})
	if err != nil {
		return nil, err
	}

	return &protogen.MakeArtistResponse{}, nil
}

func (us *UsersHandler) UpdateArtist(
	ctx context.Context,
	input *protogen.UpdateArtistRequest,
) (*protogen.UpdateArtistResponse, error) {
	id, err := uuid.Parse(input.Id)
	if err != nil {
		return nil, err
	}

	err = us.artistsService.UpdateArtist(ctx, artists.DTOUpdateArtistInput{
		ArtistID:  id,
		Name:      input.Name,
		Email:     input.Email,
		AvatarURL: input.AvatarUrl,
	})
	if err != nil {
		return nil, err
	}

	return &protogen.UpdateArtistResponse{}, nil
}

func (us *UsersHandler) DeleteArtist(
	ctx context.Context,
	input *protogen.DeleteArtistRequest,
) (*protogen.DeleteArtistResponse, error) {
	id, err := uuid.Parse(input.Id)
	if err != nil {
		return nil, err
	}

	err = us.artistsService.DeleteArtist(ctx, artists.DTODeleteArtistInput{
		ArtsitID: id,
	})
	if err != nil {
		return nil, err
	}

	return &protogen.DeleteArtistResponse{}, nil
}

// listener related things.

func (us *UsersHandler) GetListener(
	_ context.Context,
	_ *protogen.GetListenerRequest,
) (*protogen.GetListenerResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (us *UsersHandler) UpdateListener(
	_ context.Context,
	_ *protogen.UpdateListenerRequest,
) (*protogen.UpdateListenerResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (us *UsersHandler) DeleteListener(
	_ context.Context,
	_ *protogen.DeleteListenerRequest,
) (*protogen.DeleteListenerResponse, error) {
	panic("not implemented") // TODO: Implement
}
