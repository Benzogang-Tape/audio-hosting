package handlers

import (
	"context"
	"dev.gaijin.team/go/golib/e"
	"errors"
	"github.com/Benzogang-Tape/audio-hosting/playlists/api/protogen"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/models"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/service"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/service/playlists"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/transport/grpc/uniceptors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *PlaylistsService) CopyPlaylist(
	ctx context.Context,
	req *protogen.CopyPlaylistRequest,
) (*protogen.CopyPlaylistResponse, error) {
	return applyUnis(
		ctx, s.logger, req, "CopyPlaylist",
		uniceptors.Auth[*protogen.CopyPlaylistRequest, *protogen.CopyPlaylistResponse](false, s.tokenParser),
	)(s.copyPlaylistImpl)
}

func (s *PlaylistsService) copyPlaylistImpl(
	ctx context.Context,
	req *protogen.CopyPlaylistRequest,
) (*protogen.CopyPlaylistResponse, error) {
	playlistID := req.GetPlaylistId()

	userToken, ok := uniceptors.TokenFromCtx(ctx)
	if !ok {
		return nil, ErrTokenNotFound
	}

	id, err := s.service.CopyPlaylist(ctx, playlists.CopyPlaylistInput{
		PlaylistID: playlistID,
		UserID:     userToken.Subject.String(),
	})
	if err != nil {
		return nil, err
	}

	return &protogen.CopyPlaylistResponse{
		PlaylistId: id,
	}, nil
}

func (s *PlaylistsService) LikePlaylist(
	ctx context.Context,
	req *protogen.LikeDislikePlaylistRequest,
) (*protogen.LikeDislikePlaylistResponse, error) {
	return applyUnis(
		ctx, s.logger, req, "LikePlaylist",
		uniceptors.Auth[*protogen.LikeDislikePlaylistRequest, *protogen.LikeDislikePlaylistResponse](
			false,
			s.tokenParser,
		),
	)(s.likePlaylistImpl)
}

func (s *PlaylistsService) likePlaylistImpl(
	ctx context.Context,
	req *protogen.LikeDislikePlaylistRequest,
) (*protogen.LikeDislikePlaylistResponse, error) {
	playlistID := req.GetPlaylistId()

	userToken, ok := uniceptors.TokenFromCtx(ctx)
	if !ok {
		return nil, e.New("transport LikePlaylist: token not found in ctx")
	}

	err := s.service.LikePlaylist(ctx, playlists.LikeDislikePlaylistInput{
		PlaylistID: playlistID,
		UserID:     userToken.Subject.String(),
	})

	if err != nil {
		return &protogen.LikeDislikePlaylistResponse{Success: false}, err
	}

	return &protogen.LikeDislikePlaylistResponse{Success: true}, nil
}

func (s *PlaylistsService) DislikePlaylist(
	ctx context.Context,
	req *protogen.LikeDislikePlaylistRequest,
) (*protogen.LikeDislikePlaylistResponse, error) {
	return applyUnis(
		ctx, s.logger, req, "DislikePlaylist",
		uniceptors.Auth[*protogen.LikeDislikePlaylistRequest, *protogen.LikeDislikePlaylistResponse](
			false,
			s.tokenParser,
		),
	)(s.dislikePlaylistImpl)
}

func (s *PlaylistsService) dislikePlaylistImpl(
	ctx context.Context,
	req *protogen.LikeDislikePlaylistRequest,
) (*protogen.LikeDislikePlaylistResponse, error) {
	playlistID := req.GetPlaylistId()

	userToken, ok := uniceptors.TokenFromCtx(ctx)
	if !ok {
		return nil, e.New("transport DislikePlaylist: token not found in ctx")
	}

	err := s.service.DislikePlaylist(ctx, playlists.LikeDislikePlaylistInput{
		PlaylistID: playlistID,
		UserID:     userToken.Subject.String(),
	})

	if err != nil {
		return &protogen.LikeDislikePlaylistResponse{Success: false}, err
	}

	return &protogen.LikeDislikePlaylistResponse{Success: true}, nil
}

func (s *PlaylistsService) GetMyPlaylists(
	ctx context.Context,
	req *protogen.GetMyPlaylistsRequest,
) (*protogen.GetMyPlaylistsResponse, error) {
	return applyUnis(
		ctx, s.logger, req, "GetMyPlaylists",
		uniceptors.Auth[*protogen.GetMyPlaylistsRequest, *protogen.GetMyPlaylistsResponse](false, s.tokenParser),
	)(s.getMyPlaylistsImpl)
}

func (s *PlaylistsService) getMyPlaylistsImpl(
	ctx context.Context,
	_ *protogen.GetMyPlaylistsRequest,
) (*protogen.GetMyPlaylistsResponse, error) {
	userToken, ok := uniceptors.TokenFromCtx(ctx)
	if !ok {
		return nil, ErrTokenNotFound
	}

	pl, err := s.service.GetMyPlaylists(ctx, userToken.Subject.String())

	switch {
	case errors.Is(err, service.ErrPlaylistNotFound):
		return &protogen.GetMyPlaylistsResponse{Playlists: []*protogen.PlaylistMetadata{}}, nil
	case err != nil:
		return nil, err
	}

	resp := make([]*protogen.PlaylistMetadata, len(pl))
	for i, p := range pl {
		resp[i] = &protogen.PlaylistMetadata{
			Id:             p.ID,
			Title:          p.Title,
			AuthorId:       p.AuthorID,
			CoverUrl:       p.CoverURL,
			CreatedAt:      timestamppb.New(p.CreatedAt),
			UpdatedAt:      timestamppb.New(p.UpdatedAt),
			ReleasedAt:     timestamppb.New(p.ReleasedAt),
			IsAlbum:        p.IsAlbum,
			IsPublic:       p.IsPublic,
			IsMyCollection: p.IsMyCollection,
		}
	}

	return &protogen.GetMyPlaylistsResponse{Playlists: resp}, nil
}

func (s *PlaylistsService) GetMyCollection(
	ctx context.Context,
	req *protogen.GetMyCollectionRequest,
) (*protogen.GetMyCollectionResponse, error) {
	return applyUnis(
		ctx, s.logger, req, "GetMyCollection",
		uniceptors.Auth[*protogen.GetMyCollectionRequest, *protogen.GetMyCollectionResponse](false, s.tokenParser),
	)(s.getMyCollectionImpl)
}

func (s *PlaylistsService) getMyCollectionImpl(
	ctx context.Context,
	_ *protogen.GetMyCollectionRequest,
) (*protogen.GetMyCollectionResponse, error) {
	userToken, ok := uniceptors.TokenFromCtx(ctx)
	if !ok {
		return nil, ErrTokenNotFound
	}

	playlist, err := s.service.GetMyCollection(ctx, userToken.Subject.String())

	switch {
	case errors.Is(err, service.ErrPlaylistNotFound):
		return &protogen.GetMyCollectionResponse{
			Playlist: &protogen.Playlist{
				Metadata: convPlaylistMetadata(models.PlaylistMetadata{}),
				Songs:    nil,
			},
		}, nil
	case err != nil:
		return nil, err
	}

	return &protogen.GetMyCollectionResponse{
		Playlist: &protogen.Playlist{
			Metadata: convPlaylistMetadata(playlist.Metadata),
			Songs:    convSongs(playlist.Songs),
		},
	}, nil
}

func (s *PlaylistsService) LikeTrack(
	ctx context.Context,
	req *protogen.LikeDislikeTrackRequest,
) (*protogen.LikeDislikeTrackResponse, error) {
	return applyUnis(
		ctx, s.logger, req, "LikeTrack",
		uniceptors.Auth[*protogen.LikeDislikeTrackRequest, *protogen.LikeDislikeTrackResponse](false, s.tokenParser),
	)(s.likeTrackImpl)
}

func (s *PlaylistsService) likeTrackImpl(
	ctx context.Context,
	req *protogen.LikeDislikeTrackRequest,
) (*protogen.LikeDislikeTrackResponse, error) {
	userToken, ok := uniceptors.TokenFromCtx(ctx)
	if !ok {
		return nil, ErrTokenNotFound
	}

	trackID := req.GetTrackId()

	err := s.service.LikeTrack(ctx, playlists.LikeDislikeTrackInput{
		TrackID: trackID,
		UserID:  userToken.Subject.String(),
	})
	if err != nil {
		return &protogen.LikeDislikeTrackResponse{Success: false}, err
	}

	return &protogen.LikeDislikeTrackResponse{Success: true}, nil
}

func (s *PlaylistsService) DislikeTrack(
	ctx context.Context,
	req *protogen.LikeDislikeTrackRequest,
) (*protogen.LikeDislikeTrackResponse, error) {
	return applyUnis(
		ctx, s.logger, req, "DislikeTrack",
		uniceptors.Auth[*protogen.LikeDislikeTrackRequest, *protogen.LikeDislikeTrackResponse](false, s.tokenParser),
	)(s.dislikeTrackImpl)
}

func (s *PlaylistsService) dislikeTrackImpl(
	ctx context.Context,
	req *protogen.LikeDislikeTrackRequest,
) (*protogen.LikeDislikeTrackResponse, error) {
	userToken, ok := uniceptors.TokenFromCtx(ctx)
	if !ok {
		return nil, ErrTokenNotFound
	}

	trackID := req.GetTrackId()

	err := s.service.DislikeTrack(ctx, playlists.LikeDislikeTrackInput{
		TrackID: trackID,
		UserID:  userToken.Subject.String(),
	})
	if err != nil {
		return &protogen.LikeDislikeTrackResponse{Success: false}, err
	}

	return &protogen.LikeDislikeTrackResponse{Success: true}, nil
}
