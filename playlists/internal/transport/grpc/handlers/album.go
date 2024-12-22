package handlers

import (
	"context"
	"github.com/Benzogang-Tape/audio-hosting/playlists/api/protogen"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/service/playlists"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/transport/grpc/uniceptors"
)

func (s *PlaylistsService) CreateAlbum(
	ctx context.Context,
	req *protogen.CreatePlaylistRequest,
) (*protogen.CreatePlaylistResponse, error) {
	ctx = context.WithValue(ctx, isAlbumKey, true)

	return applyUnis(
		ctx, s.logger, req, "CreateAlbum",
		uniceptors.Auth[*protogen.CreatePlaylistRequest, *protogen.CreatePlaylistResponse](true, s.tokenParser),
	)(s.createPlaylistImpl)
}

func (s *PlaylistsService) ReleaseAlbum(
	ctx context.Context,
	req *protogen.ReleaseAlbumRequest,
) (*protogen.ReleaseAlbumResponse, error) {
	return applyUnis(
		ctx, s.logger, req, "ReleaseAlbum",
		uniceptors.Auth[*protogen.ReleaseAlbumRequest, *protogen.ReleaseAlbumResponse](true, s.tokenParser),
	)(s.releaseAlbumImpl)
}

func (s *PlaylistsService) releaseAlbumImpl(
	ctx context.Context,
	req *protogen.ReleaseAlbumRequest,
) (*protogen.ReleaseAlbumResponse, error) {
	albumId := req.GetAlbumId()
	suppressNotifications := req.GetSuppressNotifications()

	userToken, ok := uniceptors.TokenFromCtx(ctx)
	if !ok {
		return nil, ErrTokenNotFound
	}

	err := s.service.ReleaseAlbum(ctx, playlists.ReleaseAlbumInput{
		AlbumID:               albumId,
		UserID:                userToken.Subject.String(),
		SuppressNotifications: suppressNotifications,
	})
	if err != nil {
		return &protogen.ReleaseAlbumResponse{
			Success: false,
		}, err
	}

	return &protogen.ReleaseAlbumResponse{
		Success: true,
	}, nil
}
