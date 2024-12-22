package handlers

import (
	"context"
	"github.com/Benzogang-Tape/audio-hosting/playlists/api/protogen"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/lib/auth"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/models"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/service/covers"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/service/playlists"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/transport/grpc/uniceptors"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/logger"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/transport"
)

type Service interface {
	// crud
	CreatePlaylist(
		ctx context.Context,
		input playlists.CreatePlaylistInput,
	) (models.Playlist, error)
	GetPlaylist(
		ctx context.Context,
		playlistID string,
	) (models.Playlist, error)
	DeletePlaylist(
		ctx context.Context,
		in playlists.DeletePlaylistInput,
	) error
	UpdatePlaylist(
		ctx context.Context,
		in playlists.UpdatePlaylistInput,
	) (models.PlaylistMetadata, error)
	GetPlaylists(
		ctx context.Context,
		in playlists.GetPlaylistsInput,
	) (playlists.GetPlaylistsOutput, error)

	// user's actions
	LikePlaylist(
		ctx context.Context,
		input playlists.LikeDislikePlaylistInput,
	) error
	DislikePlaylist(
		ctx context.Context,
		input playlists.LikeDislikePlaylistInput,
	) error
	CopyPlaylist(
		ctx context.Context,
		input playlists.CopyPlaylistInput,
	) (string, error)
	GetMyPlaylists(
		ctx context.Context,
		userID string,
	) ([]models.PlaylistMetadata, error)
	GetMyCollection(
		ctx context.Context,
		userID string,
	) (models.Playlist, error)
	LikeTrack(
		ctx context.Context,
		in playlists.LikeDislikeTrackInput,
	) error
	DislikeTrack(
		ctx context.Context,
		in playlists.LikeDislikeTrackInput,
	) error

	// Album related actions
	ReleaseAlbum(
		ctx context.Context,
		in playlists.ReleaseAlbumInput,
	) error

	// Cover related actions
	UploadRawCover(
		ctx context.Context,
		input covers.UploadRawCoverInput,
	) (covers.UploadRawCoverOutput, error)
	GetRawCover(
		ctx context.Context,
		playlistID string,
	) (covers.GetRawCoverOutput, error)
}

type PlaylistsService struct {
	protogen.UnimplementedPlaylistsServiceServer

	service Service

	logger      logger.Logger
	tokenParser uniceptors.TokenParser
}

func NewPlaylistsService(service Service, pub string, log logger.Logger) (*PlaylistsService, error) {
	tokenParser, err := auth.NewParser(pub)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	return &PlaylistsService{ //nolint:exhaustruct
		service:     service,
		tokenParser: tokenParser,
		logger:      log,
	}, nil
}

func applyUnis[T transport.ValidatorAll, T2 any](ctx context.Context,
	log logger.Logger,
	req T,
	method string,
	unis ...transport.Uniceptor[T, T2]) transport.HandInvoker[T, T2] {
	const defaultUnisCount = 4

	allUnis := make([]transport.Uniceptor[T, T2], 0, len(unis)+defaultUnisCount)
	allUnis = append(allUnis,
		transport.ContextWithLogger[T, T2](log),
		transport.Recovery[T, T2](method),
		transport.Validation[T, T2](true),
		transport.Logging[T, T2](method))
	allUnis = append(allUnis, unis...)

	return transport.Apply(ctx, req, allUnis...)
}
