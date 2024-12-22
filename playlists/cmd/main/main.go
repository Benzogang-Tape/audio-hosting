package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	application "github.com/Benzogang-Tape/audio-hosting/playlists/internal/app"
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/config"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/logger"
)

const serviceName = "playlists"

func main() {
	cfg, err := config.Load("./configs/service.yaml", "/etc/app/service.yaml", "service.yaml")
	if cfg == nil || err != nil {
		panic(fmt.Sprintf("failed to read config: %s", err))
	}

	ctx := context.Background()
	mainLogger := logger.New(serviceName, cfg.ENV)
	ctx = context.WithValue(ctx, logger.LoggerKey, mainLogger)

	app, err := application.New(ctx, cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to create app: %v", err))
	}

	graceCh := make(chan os.Signal, 1)
	signal.Notify(graceCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := app.Run(ctx); err != nil {
			mainLogger.Error(ctx, err.Error())
		}
	}()

	<-graceCh

	if err := app.Stop(ctx); err != nil {
		mainLogger.Error(ctx, err.Error())
	}
}
