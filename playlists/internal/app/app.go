package app

import (
	"context"

	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/config"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/storage"
	grpcserver "github.com/Benzogang-Tape/audio-hosting/playlists/internal/transport/grpc"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/logger"

	"go.uber.org/zap"
)

type App struct {
	server *grpcserver.Server
	db     *storage.Storage
}

// New creates a new App instance.
func New(ctx context.Context, cfg *config.Config) (*App, error) {
	log := logger.GetLoggerFromCtx(ctx)

	db, err := storage.New(cfg.Connections.PGConfig, cfg.Connections.RedisConfig, cfg.Connections.S3Config) //nolint:contextcheck
	if err != nil {
		return nil, err
	}

	server, err := grpcserver.New(ctx, cfg.Servers, cfg.Secrets.Public, db, cfg.Connections.SongsConn)
	if err != nil {
		log.Error(ctx, "failed to create server", zap.Error(err))
		return nil, err
	}

	return &App{
		server: server,
		db:     db,
	}, nil
}

// Run starts the GRPC server and begins listening for incoming requests.
func (a *App) Run(ctx context.Context) error {
	return a.server.Run(ctx)
}

// Stop closes the database connection and stops the GRPC server.
func (a *App) Stop(ctx context.Context) error {
	a.db.Close()
	return a.server.Stop(ctx)
}
