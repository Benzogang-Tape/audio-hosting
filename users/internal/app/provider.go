package app

import (
	"context"
	"log/slog"

	"github.com/Benzogang-Tape/audio-hosting/users/api/protogen"
	usersHandlers "github.com/Benzogang-Tape/audio-hosting/users/internal/adapters/grpc/handlers/users"
	"github.com/Benzogang-Tape/audio-hosting/users/internal/config"
	"github.com/Benzogang-Tape/audio-hosting/users/pkg/auth"
	"github.com/Benzogang-Tape/audio-hosting/users/pkg/closer"
	"github.com/Benzogang-Tape/audio-hosting/users/pkg/database/postgres"
	"github.com/Benzogang-Tape/audio-hosting/users/pkg/hasher"
	"github.com/Benzogang-Tape/audio-hosting/users/pkg/logger"
)

type Hasher interface {
	HashPassword(password string) (string, error)
	CheckPasswordHash(password, hash string) error
}

type Provider struct {
	logger *slog.Logger
	config *config.Config
	closer *closer.Closer

	db     *postgres.DB
	signer auth.Signer
	parser auth.Parser
	hasher Hasher

	usersRepository           UsersRepository
	artistsRepository         ArtistsRepository
	listenersRepository       ListenersRepository
	refreshSessionsRepository RefreshSessionsRepository

	usersService     UsersService
	artistsService   ArtistsService
	listenersService ListenersService

	usersHandler protogen.UsersServiceServer
}

func NewProvider() *Provider {
	return &Provider{}
}

func (p *Provider) Logger() *slog.Logger {
	if p.logger == nil {
		p.logger = logger.InitLogger(p.Cfg().Env)
	}

	return p.logger
}

func (p *Provider) Closer() *closer.Closer {
	if p.closer == nil {
		p.closer = closer.New(p.Logger())
	}

	return p.closer
}

func (p *Provider) Cfg() *config.Config {
	if p.config == nil {
		cfg, err := config.New()
		if err != nil {
			panic(err)
		}

		p.config = cfg
	}

	return p.config
}

func (p *Provider) Hasher() Hasher {
	if p.hasher == nil {
		p.hasher = hasher.NewBcryptHasher()
	}

	return p.hasher
}

func (p *Provider) Parser() auth.Parser {
	if p.parser == nil {
		parser, err := auth.NewParser(p.Cfg().Auth.PublicKey)
		if err != nil {
			panic(err)
		}

		p.parser = parser
	}

	return p.parser
}

func (p *Provider) Signer() auth.Signer {
	if p.signer == nil {
		signer, err := auth.NewSigner(p.Cfg().Auth.PrivateKey, p.Cfg().Auth.AccessTTL)
		if err != nil {
			panic(err)
		}

		p.signer = signer
	}

	return p.signer
}

func (p *Provider) DB(ctx context.Context) *postgres.DB {
	if p.db == nil {
		db, err := postgres.New(ctx, p.Cfg().Postgres)
		if err != nil {
			panic(err)
		}

		p.Closer().Add(func() error {
			db.Close()

			p.Logger().Info("database connection closed")
			return nil
		})

		p.db = db
	}

	return p.db
}

func (p *Provider) UsersHandler(ctx context.Context) protogen.UsersServiceServer {
	if p.usersHandler == nil {
		p.usersHandler = usersHandlers.NewHandler(
			p.UsersService(ctx),
			p.ListenersService(ctx),
			p.ArtistsService(ctx),
		)
	}

	return p.usersHandler
}
