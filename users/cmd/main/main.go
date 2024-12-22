package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Benzogang-Tape/audio-hosting/users/internal/app"
)

func main() {
	ctx := context.Background()

	app, err := app.NewApp(ctx)
	if err != nil {
		panic(err)
	}

	go func() {
		err = app.Run(ctx)
		if err != nil {
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	err = app.GracefulShutdown(ctx)
	if err != nil {
		panic(err)
	}
}
