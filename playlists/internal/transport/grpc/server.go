package grpcserver

import (
	"context"
	"fmt"
	client "github.com/Benzogang-Tape/audio-hosting/playlists/internal/client/songs"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/storage"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/transport/grpc/gateway"
	"net"
	"net/http"

	"github.com/Benzogang-Tape/audio-hosting/playlists/api/protogen"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/config"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/transport/grpc/handlers"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/logger"

	"dev.gaijin.team/go/golib/e"
	"dev.gaijin.team/go/golib/fields"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type Server struct {
	grpcServer  *grpc.Server
	restServer  *http.Server
	clientSongs *client.Client
	listener    net.Listener
}

// New returns a new gRPC server with a unary interceptor that logs the
// request and response, and also sets up a gRPC-gateway server to expose
// the gRPC service as a RESTful API.
//
// The gRPC server and REST server are both configured with the given
// configuration. The gRPC server has a connection timeout, and the
// REST server has read, write, and read header timeouts.
//
// The gRPC server is registered with the given service, and the REST
// server is also registered with the same service.
//
// The returned Server has a Close method that can be used to shut down
// both the gRPC and REST servers.
func New(
	ctx context.Context,
	cfg config.Servers,
	publicKey string,
	storage *storage.Storage,
	clientCfg client.Config,
) (*Server, error) {
	clSongs, err := client.New(clientCfg)
	if err != nil {
		return nil, e.NewFrom("creating songs client", err)
	}

	service := NewService(storage, cfg.HTTP, clSongs)

	lis, err := net.Listen("tcp", fmt.Sprintf("%v:%d", cfg.GRPC.Host, cfg.GRPC.Port))
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	opts := []grpc.ServerOption{
		grpc.ConnectionTimeout(cfg.GRPC.Timeout),
	}

	playlistService, err := handlers.NewPlaylistsService(service, publicKey, logger.GetLoggerFromCtx(ctx))
	if err != nil {
		return nil, e.From(err, fields.F("creating service", "playlists"))
	}

	grpcServer := grpc.NewServer(opts...)
	protogen.RegisterPlaylistsServiceServer(grpcServer, playlistService)

	restSrv := runtime.NewServeMux(
		runtime.SetQueryParameterParser(&QueryParser{}),
	)
	if err = protogen.RegisterPlaylistsServiceHandlerServer(ctx, restSrv, playlistService); err != nil { //nolint:revive
		return nil, err //nolint:wrapcheck
	}

	err = gateway.RegisterHandlers(restSrv, logger.GetLoggerFromCtx(ctx), service, publicKey)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	restServer := &http.Server{
		Addr:              fmt.Sprintf("%v:%d", cfg.HTTP.Host, cfg.HTTP.Port),
		Handler:           restSrv,
		ReadHeaderTimeout: cfg.HTTP.Timeout,
		ReadTimeout:       cfg.HTTP.Timeout,
		WriteTimeout:      cfg.HTTP.Timeout,
	}

	return &Server{
		grpcServer:  grpcServer,
		restServer:  restServer,
		clientSongs: clSongs,
		listener:    lis,
	}, nil
}

// Run starts both the gRPC and REST servers concurrently.
// It listens for incoming connections and serves requests using the provided context.
// The method returns an error if either server fails to start or encounters an issue during execution.
func (s *Server) Run(ctx context.Context) error {
	eg := errgroup.Group{}

	eg.Go(func() error {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "starting gRPC server",
			zap.Int("port", s.listener.Addr().(*net.TCPAddr).Port)) //nolint:forcetypeassert
		return s.grpcServer.Serve(s.listener)
	})

	eg.Go(func() error {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "starting REST server", zap.String("addr", s.restServer.Addr))
		return s.restServer.ListenAndServe()
	})

	return eg.Wait() //nolint:wrapcheck
}

// Stop gracefully stops both the gRPC and REST servers.
//
// It calls `http.Server.Shutdown` to stop the REST server and
// `grpc.Server.GracefulStop` to stop the gRPC server. If the REST server
// was not able to shut down, the error from `http.Server.Shutdown` is
// returned.
func (s *Server) Stop(ctx context.Context) error {
	l := logger.GetLoggerFromCtx(ctx)

	l.Info(ctx, "stopping gateway server")

	err := s.restServer.Shutdown(ctx)
	if err != nil {
		l.Error(ctx, "failed to shutdown gateway server", zap.Error(err))
	}

	l.Info(ctx, "stopping gRPC server")
	s.grpcServer.GracefulStop()
	l.Info(ctx, "gRPC server stopped")

	l.Info(ctx, "closing songs conn")

	err = s.clientSongs.Close()

	l.Info(ctx, "songs conn closed") //nolint:wsl

	return err //nolint:wrapcheck
}
