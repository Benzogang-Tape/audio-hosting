package grpcserver

import (
	"context"
	"net/http"

	"github.com/Benzogang-Tape/audio-hosting/songs/api/protogen/api"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/grpc-server/grpcgw"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/grpc-server/uniceptors"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/services/songs"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/transport"

	"dev.gaijin.team/go/golib/e"
	gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type songsServer struct {
	log zerolog.Logger
	api.UnimplementedSongsServiceServer

	service     Service
	tokenParser uniceptors.TokenParser
}

type Service interface {
	CreateSong(ctx context.Context, in songs.CreateSongInput) (songs.CreateSongOutput, error)
	GetMySongs(ctx context.Context, input songs.GetMySongsInput) (songs.GetMySongsOutput, error)
	GetSong(ctx context.Context, input songs.GetSongInput) (songs.GetSongOutput, error)
	GetSongs(ctx context.Context, input songs.GetSongsInput) (songs.GetSongsOutput, error)
	ReleaseSongs(ctx context.Context, in songs.ReleaseSongsInput) (songs.ReleaseSongsOutput, error)
}

type Dependencies struct {
	Service     Service
	RawService  grpcgw.RawService
	TokenParser uniceptors.TokenParser
}

func Register(log zerolog.Logger, server *grpc.Server, gatewayMux *gateway.ServeMux, deps Dependencies) {
	srv := &songsServer{
		log:                             log,
		UnimplementedSongsServiceServer: api.UnimplementedSongsServiceServer{},
		service:                         deps.Service,
		tokenParser:                     deps.TokenParser,
	}

	api.RegisterSongsServiceServer(server, srv)

	err := registerHandlers(gatewayMux, log, deps)
	if err != nil {
		log.Error().Err(err).Msg("register gateway handlers")
	}

	// it never returns error
	_ = api.RegisterSongsServiceHandlerServer(context.Background(), gatewayMux, srv)
}

func (*songsServer) Health(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func registerHandlers(mux *gateway.ServeMux, log zerolog.Logger, deps Dependencies) error {
	h := grpcgw.RawHandlers{
		Service: deps.RawService,
	}

	mws := func(hand grpcgw.HandlerErrFunc) gateway.HandlerFunc {
		return grpcgw.ContextWithLogger(log,
			grpcgw.RecoveryMw(
				grpcgw.LoggingMw(hand)))
	}
	authMws := func(hand grpcgw.HandlerErrFunc) gateway.HandlerFunc {
		return mws(grpcgw.AuthMw(true, deps.TokenParser, hand))
	}

	err := mux.HandlePath(http.MethodPost, "/songs/api/v1/song/{song_id}/raw", authMws(h.UploadRawSongHandler()))
	if err != nil {
		return e.NewFrom("register post song/raw", err)
	}

	err = mux.HandlePath(http.MethodGet, "/songs/api/v1/song/raw/{id}", mws(h.GetRawSongHandler()))
	if err != nil {
		return e.NewFrom("register get song/raw", err)
	}

	err = mux.HandlePath(http.MethodPost, "/songs/api/v1/song/{song_id}/image/raw",
		authMws(h.UploadRawSongImageHandler()))
	if err != nil {
		return e.NewFrom("register post song/image/raw", err)
	}

	err = mux.HandlePath(http.MethodGet, "/songs/api/v1/song/image/raw/{id}", mws(h.GetRawSongImageHandler()))
	if err != nil {
		return e.NewFrom("register get song/image/raw", err)
	}

	return nil
}

func applyUnis[T transport.ValidatorAll, T2 any](ctx context.Context,
	log zerolog.Logger,
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
