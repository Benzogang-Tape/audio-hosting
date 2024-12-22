package grpcserver

import (
	"context"

	"github.com/Benzogang-Tape/audio-hosting/songs/api/protogen/api"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/grpc-server/uniceptors"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/services/songs"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *songsServer) GetSongs(ctx context.Context, req *api.GetSongsRequest) (*api.GetSongsResponse, error) {
	return applyUnis[*api.GetSongsRequest, *api.GetSongsResponse](
		ctx, s.log, req, "GetSongs")(s.getSongsImpl)
}

func (s *songsServer) getSongsImpl(ctx context.Context, req *api.GetSongsRequest) (*api.GetSongsResponse, error) {
	var (
		artistId       *uuid.UUID
		page, pageSize int32 = 1, 10
	)

	if req.GetArtistId() != "" {
		id := uuid.MustParse(req.GetArtistId())
		artistId = &id
	}

	if req.GetPage() > 0 {
		page = req.GetPage()
	}

	if req.GetPageSize() > 0 {
		pageSize = req.GetPageSize()
	}

	result, err := s.service.GetSongs(ctx, songs.GetSongsInput{
		ArtistId:    artistId,
		MatchArtist: req.MatchArtist, //nolint:protogetter
		MatchName:   req.MatchName,   //nolint:protogetter
		Ids:         mapUuids(req.GetIds()),
		Page:        page,
		PageSize:    pageSize,
	})
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	outSongs := make([]*api.Song, len(result.Songs))

	for i, song := range result.Songs {
		outSongs[i] = &api.Song{
			Id:          song.Id.String(),
			Singer:      mapArtist(song.Singer),
			Artists:     mapArtists(song.Artists),
			Name:        song.Name,
			SongUrl:     song.SongUrl,
			ImageUrl:    song.ImageUrl,
			Duration:    durationpb.New(song.Duration),
			WeightBytes: song.WeightBytes,
			UploadedAt:  timestamppb.New(song.UploadedAt),
			ReleasedAt:  timestamppb.New(song.ReleasedAt),
		}
	}

	return &api.GetSongsResponse{
		Songs: outSongs,
		Pagination: &api.PaginationResponse{
			LastPage: result.LastPage,
		},
	}, nil
}

func (s *songsServer) GetMySongs(ctx context.Context, req *api.GetMySongsRequest) (*api.GetMySongsResponse, error) {
	return applyUnis(
		ctx, s.log, req, "GetMySongs",
		uniceptors.Auth[*api.GetMySongsRequest, *api.GetMySongsResponse](true, s.tokenParser))(s.getMySongsImpl)
}

func (s *songsServer) getMySongsImpl(ctx context.Context, req *api.GetMySongsRequest,
) (*api.GetMySongsResponse, error) {
	var (
		page, pageSize int32 = 1, 10
	)

	token := uniceptors.TokenFromCtx(ctx)

	if req.GetPage() > 0 {
		page = req.GetPage()
	}

	if req.GetPageSize() > 0 {
		pageSize = req.GetPageSize()
	}

	input := songs.GetMySongsInput{ //nolint:exhaustruct
		UserId:   token.Subject,
		Page:     page,
		PageSize: pageSize,
	}

	if len(req.GetIds()) > 0 {
		input.ByIds = true
		input.Ids = mapUuids(req.GetIds())
		input.Page = 1
		input.PageSize = int32(len(req.GetIds())) //nolint:gosec
	}

	result, err := s.service.GetMySongs(ctx, input)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	outSongs := make([]*api.MySong, len(result.Songs))

	for i, song := range result.Songs {
		var dur *durationpb.Duration
		if song.Duration != nil {
			dur = durationpb.New(*song.Duration)
		}

		uploadedAt := timestamppb.New(song.UploadedAt)

		var releasedAt *timestamppb.Timestamp
		if song.ReleasedAt != nil {
			releasedAt = timestamppb.New(*song.ReleasedAt)
		}

		outSongs[i] = &api.MySong{
			Id:          song.Id.String(),
			Singer:      mapArtist(song.Singer),
			Artists:     mapArtists(song.Artists),
			Name:        song.Name,
			SongUrl:     song.SongUrl,
			ImageUrl:    song.ImageUrl,
			Duration:    dur,
			WeightBytes: song.WeightBytes,
			UploadedAt:  uploadedAt,
			ReleasedAt:  releasedAt,
		}
	}

	return &api.GetMySongsResponse{
		Songs: outSongs,
		Pagination: &api.PaginationResponse{
			LastPage: result.LastPage,
		},
	}, nil
}

func mapUuids(ids []string) []uuid.UUID {
	out := make([]uuid.UUID, len(ids))
	for i, id := range ids {
		out[i] = uuid.MustParse(id)
	}

	return out
}
