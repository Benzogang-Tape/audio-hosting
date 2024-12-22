package app

import (
	"os"

	"github.com/Benzogang-Tape/audio-hosting/songs/internal/config"

	"github.com/rs/zerolog"
)

// newLogger creates a new configured zerolog logger instance.
//
// Currently, it sets up a logger that writes to standard output
// and includes caller information.
// In the future, this function will be extended to integrate
// with OpenTelemetry for distributed logging.
func newLogger(conf config.Config) zerolog.Logger {
	var callerHook zerolog.HookFunc = func(e *zerolog.Event, _ zerolog.Level, _ string) {
		const skipFrames = 4

		e.Timestamp().Caller(skipFrames)
	}

	return zerolog.New(os.Stdout).Hook(callerHook).Level(conf.Logging.Level)
}
