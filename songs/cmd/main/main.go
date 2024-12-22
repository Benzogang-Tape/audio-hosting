package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/Benzogang-Tape/audio-hosting/songs/internal/app"

	"dev.gaijin.team/go/golib/must"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	must.NoErr(app.New().Run(ctx))
}
