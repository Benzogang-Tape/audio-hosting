package handlers

import (
	"context"
	"errors"
	"github.com/Benzogang-Tape/audio-hosting/playlists/api/protogen"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/service"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/service/playlists"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/transport/grpc/uniceptors"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/erix"

	"github.com/google/uuid"
)

// for CreatePlaylist
const isAlbumKey = "is_album"

var (
	ErrTokenNotFound = erix.NewStatus("token not found in context", erix.CodeInternalServerError)
)

func (s *PlaylistsService) CreatePlaylist(
	ctx context.Context,
	req *protogen.CreatePlaylistRequest,
) (*protogen.CreatePlaylistResponse, error) {
	ctx = context.WithValue(ctx, isAlbumKey, false)

	return applyUnis(
		ctx, s.logger, req, "CreatePlaylist",
		uniceptors.Auth[*protogen.CreatePlaylistRequest, *protogen.CreatePlaylistResponse](false, s.tokenParser),
	)(s.createPlaylistImpl)
}

func (s *PlaylistsService) createPlaylistImpl(
	ctx context.Context,
	req *protogen.CreatePlaylistRequest,
) (*protogen.CreatePlaylistResponse, error) {
	title := req.GetTitle()

	var trackIDs []uuid.UUID //nolint:prealloc
	for _, id := range req.GetTrackIds() {
		trackIDs = append(trackIDs, uuid.MustParse(id))
	}

	coverURL := req.GetCoverUrl()

	userToken, ok := uniceptors.TokenFromCtx(ctx)
	if !ok {
		return nil, ErrTokenNotFound
	}

	isAlbum := ctx.Value(isAlbumKey).(bool)

	playlist, err := s.service.CreatePlaylist(ctx, playlists.CreatePlaylistInput{
		Title:    title,
		TrackIDs: trackIDs,
		AuthorID: userToken.Subject,
		CoverURL: coverURL,
		IsAlbum:  isAlbum,
	})

	switch {
	case errors.Is(err, service.ErrGetSongs):
		return &protogen.CreatePlaylistResponse{
			Playlist: &protogen.Playlist{
				Metadata: convPlaylistMetadata(playlist.Metadata),
				Songs:    nil,
			},
		}, nil

	case err != nil:
		return nil, err
	}

	return &protogen.CreatePlaylistResponse{
		Playlist: &protogen.Playlist{
			Metadata: convPlaylistMetadata(playlist.Metadata),
			Songs:    convSongs(playlist.Songs),
		},
	}, nil
}

func (s *PlaylistsService) GetPlaylist(
	ctx context.Context,
	req *protogen.GetPlaylistRequest,
) (*protogen.GetPlaylistResponse, error) {
	return applyUnis[*protogen.GetPlaylistRequest, *protogen.GetPlaylistResponse](
		ctx, s.logger, req, "GetPlaylist",
		uniceptors.ParseToken[*protogen.GetPlaylistRequest, *protogen.GetPlaylistResponse](s.tokenParser),
	)(s.getPlaylistImpl)
}

func (s *PlaylistsService) getPlaylistImpl(
	ctx context.Context,
	req *protogen.GetPlaylistRequest,
) (*protogen.GetPlaylistResponse, error) {
	playlistID := req.GetPlaylistId()

	playlist, err := s.service.GetPlaylist(ctx, playlistID)

	switch {
	case errors.Is(err, service.ErrGetSongs):
		return &protogen.GetPlaylistResponse{
			Playlist: &protogen.Playlist{
				Metadata: convPlaylistMetadata(playlist.Metadata),
				Songs:    nil,
			},
		}, nil

	case err != nil:
		return nil, err
	}

	return &protogen.GetPlaylistResponse{
		Playlist: &protogen.Playlist{
			Metadata: convPlaylistMetadata(playlist.Metadata),
			Songs:    convSongs(playlist.Songs),
		},
	}, nil
}

func (s *PlaylistsService) UpdatePlaylist(
	ctx context.Context,
	req *protogen.UpdatePlaylistRequest,
) (*protogen.UpdatePlaylistResponse, error) {
	return applyUnis(
		ctx, s.logger, req, "UpdatePlaylist",
		uniceptors.Auth[*protogen.UpdatePlaylistRequest, *protogen.UpdatePlaylistResponse](false, s.tokenParser),
	)(s.updatePlaylistImpl)
}

func (s *PlaylistsService) updatePlaylistImpl(
	ctx context.Context,
	req *protogen.UpdatePlaylistRequest,
) (*protogen.UpdatePlaylistResponse, error) {
	userToken, ok := uniceptors.TokenFromCtx(ctx)
	if !ok {
		return nil, ErrTokenNotFound
	}

	playlistID := req.GetPlaylistId()

	params := playlists.UpdatePlaylistInput{
		PlaylistID: playlistID,
		UserID:     userToken.Subject.String(),
		Title:      req.Title,
		CoverURL:   req.CoverUrl,
		IsPublic:   req.IsPublic,
		TrackIDs:   req.GetTrackIds(),
	}

	playlistMetadata, err := s.service.UpdatePlaylist(ctx, params)
	if err != nil {
		return nil, err
	}

	return &protogen.UpdatePlaylistResponse{
		Playlist: convPlaylistMetadata(playlistMetadata),
	}, nil
}

func (s *PlaylistsService) DeletePlaylist(
	ctx context.Context,
	req *protogen.DeletePlaylistRequest,
) (*protogen.DeletePlaylistResponse, error) {
	return applyUnis(
		ctx, s.logger, req, "DeletePlaylist",
		uniceptors.Auth[*protogen.DeletePlaylistRequest, *protogen.DeletePlaylistResponse](false, s.tokenParser),
	)(s.deletePlaylistImpl)
}

func (s *PlaylistsService) deletePlaylistImpl(
	ctx context.Context,
	req *protogen.DeletePlaylistRequest,
) (*protogen.DeletePlaylistResponse, error) {
	var response *protogen.DeletePlaylistResponse

	playlistIDs := req.GetPlaylistId()

	userToken, ok := uniceptors.TokenFromCtx(ctx)
	if !ok {
		return nil, ErrTokenNotFound
	}

	err := s.service.DeletePlaylist(ctx, playlists.DeletePlaylistInput{
		PlaylistIDs: playlistIDs,
		UserID:      userToken.Subject.String(),
	})
	if err != nil {
		return response, err
	}

	return response, nil
}

func (s *PlaylistsService) GetPlaylists(
	ctx context.Context,
	req *protogen.GetPlaylistsRequest,
) (*protogen.GetPlaylistsResponse, error) {
	return applyUnis[*protogen.GetPlaylistsRequest, *protogen.GetPlaylistsResponse](
		ctx, s.logger, req, "GetPlaylists",
	)(s.getPlaylistsImpl)
}

func (s *PlaylistsService) getPlaylistsImpl(
	ctx context.Context,
	req *protogen.GetPlaylistsRequest,
) (*protogen.GetPlaylistsResponse, error) {
	params := playlists.GetPlaylistsInput{
		Page:        req.GetPagination().GetPage(),
		Limit:       req.GetPagination().GetLimit(),
		ArtistID:    req.GetFilter().ArtistId,   //nolint:protogetter
		MatchTitle:  req.GetFilter().MatchTitle, //nolint:protogetter
		PlaylistIDs: req.GetIds(),
	}

	resp, err := s.service.GetPlaylists(ctx, params)
	if err != nil {
		return nil, err
	}

	playlistsMetadata := make([]*protogen.PlaylistMetadata, 0, len(resp.Playlists))
	for _, p := range resp.Playlists {
		playlistsMetadata = append(playlistsMetadata, convPlaylistMetadata(p))
	}

	return &protogen.GetPlaylistsResponse{
		Playlists: playlistsMetadata,
		Pagination: &protogen.PaginationResponse{
			Total:    int64(len(playlistsMetadata)),
			HasNext:  resp.HasNext,
			LastPage: int64(resp.LastPage),
		},
	}, nil
}
