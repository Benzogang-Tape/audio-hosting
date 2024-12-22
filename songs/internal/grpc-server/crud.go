package grpcserver

import (
	"context"

	"github.com/Benzogang-Tape/audio-hosting/songs/api/protogen/api"
	"github.com/Benzogang-Tape/audio-hosting/songs/api/protogen/api/clients/users"
	usersclient "github.com/Benzogang-Tape/audio-hosting/songs/internal/clients/users"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/grpc-server/uniceptors"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/services/songs"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *songsServer) CreateSong(ctx context.Context, req *api.CreateSongRequest) (*api.CreateSongResponse, error) {
	return applyUnis(
		ctx, s.log, req, "CreateSong",
		uniceptors.Auth[*api.CreateSongRequest, *api.CreateSongResponse](true, s.tokenParser))(s.createSongImpl)
}

func (s *songsServer) createSongImpl(ctx context.Context, req *api.CreateSongRequest) (*api.CreateSongResponse, error) {
	token := uniceptors.TokenFromCtx(ctx)

	artistsIds := make([]uuid.UUID, 0, len(req.GetFeatArtistsIds()))
	for _, artist := range req.GetFeatArtistsIds() {
		artistsIds = append(artistsIds, uuid.MustParse(artist))
	}

	out, err := s.service.CreateSong(ctx, songs.CreateSongInput{
		Name:        req.GetName(),
		ImageUrl:    req.ImageUrl, //nolint:protogetter
		SingerId:    token.Subject,
		FeatArtists: artistsIds,
	})
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	return &api.CreateSongResponse{
		Id:         out.Id.String(),
		Singer:     mapArtist(out.Singer),
		Name:       out.Name,
		Artists:    mapArtists(out.Artists),
		UploadedAt: timestamppb.New(out.UploadedAt),
		ImageUrl:   out.ImageUrl,
	}, nil
}

func (s *songsServer) GetSong(ctx context.Context, req *api.GetSongRequest) (*api.GetSongResponse, error) {
	return applyUnis(
		ctx, s.log, req, "GetSong",
		uniceptors.Auth[*api.GetSongRequest, *api.GetSongResponse](false, s.tokenParser))(s.getSongImpl)
}

func (s *songsServer) getSongImpl(ctx context.Context, req *api.GetSongRequest) (*api.GetSongResponse, error) {
	out, err := s.service.GetSong(ctx, songs.GetSongInput{
		Id: uuid.MustParse(req.GetId()),
	})
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	uploadedAt := timestamppb.New(out.UploadedAt)

	var releasedAt *timestamppb.Timestamp
	if out.ReleasedAt != nil {
		releasedAt = timestamppb.New(*out.ReleasedAt)
	}

	return &api.GetSongResponse{
		Song: &api.Song{
			Id:          out.Id.String(),
			Singer:      mapArtist(out.Singer),
			Artists:     mapArtists(out.Artists),
			Name:        out.Name,
			SongUrl:     out.SongUrl,
			ImageUrl:    out.ImageUrl,
			Duration:    durationpb.New(out.Duration),
			WeightBytes: out.WeightBytes,
			UploadedAt:  uploadedAt,
			ReleasedAt:  releasedAt,
		}}, nil

}

func (s *songsServer) UpdateSong(ctx context.Context, req *api.UpdateSongRequest) (*api.UpdateSongResponse, error) {
	return applyUnis(
		ctx, s.log, req, "UpdateSong",
		uniceptors.Auth[*api.UpdateSongRequest, *api.UpdateSongResponse](true, s.tokenParser))(s.updateSongImpl)
}

func (*songsServer) updateSongImpl(_ context.Context, _ *api.UpdateSongRequest) (*api.UpdateSongResponse, error) {
	return nil, status.Error(codes.Unimplemented, "will be implemented later") //nolint:wrapcheck
}

func (s *songsServer) DeleteSongs(ctx context.Context, req *api.DeleteSongsRequest) (*api.DeleteSongsResponse, error) {
	return applyUnis(
		ctx, s.log, req, "DeleteSongs",
		uniceptors.Auth[*api.DeleteSongsRequest, *api.DeleteSongsResponse](true, s.tokenParser))(s.deleteSongsImpl)
}

func (*songsServer) deleteSongsImpl(_ context.Context, _ *api.DeleteSongsRequest) (*api.DeleteSongsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "will be implemented later") //nolint:wrapcheck
}

func mapArtists(artists []usersclient.Artist) []*users.Artist {
	out := make([]*users.Artist, len(artists))
	for i, artist := range artists {
		out[i] = mapArtist(artist)
	}

	return out
}

func mapArtist(artist usersclient.Artist) *users.Artist {
	return &users.Artist{
		Id:        artist.Id.String(),
		Name:      artist.Name,
		Label:     artist.Label,
		AvatarUrl: artist.AvatarUrl,
	}
}
