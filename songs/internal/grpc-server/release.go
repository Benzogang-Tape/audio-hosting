package grpcserver

import (
	"context"

	"github.com/Benzogang-Tape/audio-hosting/songs/api/protogen/api"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/grpc-server/uniceptors"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/services/songs"
)

func (s *songsServer) ReleaseSongs(ctx context.Context, req *api.ReleaseSongsRequest,
) (*api.ReleaseSongsResponse, error) {
	return applyUnis(
		ctx, s.log, req, "ReleaseSongs",
		uniceptors.Auth[*api.ReleaseSongsRequest, *api.ReleaseSongsResponse](true, s.tokenParser))(s.releaseSongsImpl)
}

func (s *songsServer) releaseSongsImpl(ctx context.Context, req *api.ReleaseSongsRequest,
) (*api.ReleaseSongsResponse, error) {
	token := uniceptors.TokenFromCtx(ctx)

	_, err := s.service.ReleaseSongs(ctx, songs.ReleaseSongsInput{
		UserId:   token.Subject,
		SongsIds: mapUuids(req.GetIds()),
		Notify:   req.GetNotify(),
	})
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	return &api.ReleaseSongsResponse{}, nil
}
