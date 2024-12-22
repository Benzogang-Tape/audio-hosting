package app

import (
	"net/http"

	"github.com/Benzogang-Tape/audio-hosting/songs/internal/config"
	grpcserver "github.com/Benzogang-Tape/audio-hosting/songs/internal/grpc-server"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/storage"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/audio/auth"
	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/transport"

	gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func newServers(log zerolog.Logger, conf config.Config, db *storage.Storage) (*grpc.Server, *http.Server, error) {
	var (
		creds credentials.TransportCredentials
		err   error
	)

	if conf.Servers.Grpc.UseTls {
		log.Info().
			Str("cert_path", conf.Servers.Tls.CertPath).
			Str("key_path", conf.Servers.Tls.KeyPath).
			Msg("grpc uses TLS certificate")

		creds, err = credentials.NewServerTLSFromFile(conf.Servers.Tls.CertPath, conf.Servers.Tls.KeyPath)
		if err != nil {
			creds = insecure.NewCredentials()

			log.Error().Err(err).Msg("loading creds for grpc, using insecure")
		}
	} else {
		log.Info().Msg("use insecure grpc")

		creds = insecure.NewCredentials()
	}

	srv := grpc.NewServer(
		grpc.ConnectionTimeout(conf.Servers.Grpc.Timeout),
		grpc.Creds(creds),
	)

	mux := gateway.NewServeMux(transport.MuxWithAuthAndTraceHeaders())

	service, err := newService(db)
	if err != nil {
		return nil, nil, err
	}

	tokenParser, err := auth.NewParser(conf.Features.Auth.PublicKey)
	if err != nil {
		return nil, nil, err //nolint:wrapcheck
	}

	grpcserver.Register(log, srv, mux, grpcserver.Dependencies{
		Service:     service,
		RawService:  service,
		TokenParser: tokenParser,
	})

	log.Info().Msg("registered grpcserver")

	return srv, &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: conf.Servers.Http.Timeout,
		ReadTimeout:       conf.Servers.Http.Timeout,
		WriteTimeout:      conf.Servers.Http.Timeout,
	}, nil
}
