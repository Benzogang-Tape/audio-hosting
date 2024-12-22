package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/Benzogang-Tape/audio-hosting/songs/internal/config"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/storage"

	"dev.gaijin.team/go/golib/must"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

// Application represents the top-level application structure.
type Application struct {
	cfg        config.Config
	log        zerolog.Logger
	grpcServer *grpc.Server
	gateway    *http.Server
	db         *storage.Storage
}

// New creates a new Application instance with loaded configuration.
func New() *Application {
	cfg := must.OK(config.Load("songs.yaml", "/etc/app/songs.yaml"))
	config.Set(cfg)

	return NewWithConfig(cfg)
}

// NewWithConfig creates a new Application instance with the provided configuration.
func NewWithConfig(cfg config.Config) *Application {
	logger := newLogger(cfg)

	const dbConnectionTimeout = time.Second * 10

	ctx, cancel := context.WithTimeout(context.Background(), dbConnectionTimeout)
	defer cancel()

	db, err := storage.New(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("connecting to database")
	}

	logger.Info().Msg("connected to database")

	srv, gw, err := newServers(logger, cfg, db)
	if err != nil {
		logger.Fatal().Err(err).Msg("creating servers")
	}

	return &Application{
		cfg:        cfg,
		log:        logger,
		grpcServer: srv,
		gateway:    gw,
		db:         db,
	}
}

// Run starts the application and listens for shutdown signals.
func (a *Application) Run(ctx context.Context) error {
	defer a.shutdown(ctx)

	a.log.Info().Msg("starting application")

	grpcLis, err := net.Listen("tcp",
		fmt.Sprintf(":%d", a.cfg.Servers.Grpc.Port))
	if err != nil {
		a.log.Error().Err(err).Msg("starting grpc")
		return fmt.Errorf("net.Listen for grpc: %w", err)
	}

	go func() {
		a.log.Info().Int("port", a.cfg.Servers.Grpc.Port).Msg("started grpc")

		if err := a.grpcServer.Serve(grpcLis); err != nil {
			a.log.Error().Err(err).Msg("grpc serve failed")
		}
	}()

	httpLis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.cfg.Servers.Http.Port))
	if err != nil {
		a.log.Error().Err(err).Msg("starting http")
		return fmt.Errorf("net.Listen for http: %w", err)
	}

	go func() {
		a.log.Info().Int("port", a.cfg.Servers.Http.Port).Msg("started http")

		if err := a.gateway.Serve(httpLis); err != nil {
			a.log.Error().Err(err).Msg("http serve failed")
		}
	}()

	a.log.Info().Msg("started application")

	<-ctx.Done()

	return nil
}

// shutdown gracefully stops the application components.
func (a *Application) shutdown(ctx context.Context) {
	a.log.Info().Msg("stopping application")

	err := a.gateway.Shutdown(ctx)
	a.log.Info().Err(err).Msg("stopped http")

	a.grpcServer.GracefulStop()
	a.log.Info().Msg("stopped grpc")

	err = a.db.Close()
	a.log.Info().Err(err).Msg("disconnected from database")

	a.log.Info().Msg("stopped application")
}
