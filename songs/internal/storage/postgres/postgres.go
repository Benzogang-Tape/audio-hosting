package postgres

import (
	"context"
	"embed"

	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/pg"

	_ "github.com/golang-migrate/migrate/v4/database/postgres" // it is needed for migrations to work

	"dev.gaijin.team/go/golib/e"
)

type PgStorage struct {
	db pg.Database
	*Queries
}

//go:embed migrations/*.sql
var migrations embed.FS

func Connect(ctx context.Context, conn string) (*PgStorage, error) {
	db, err := pg.Connect(ctx, conn, migrations)
	if err != nil {
		return nil, e.NewFrom("connecting to pg database", err)
	}

	return &PgStorage{
		db:      db,
		Queries: New(db),
	}, nil
}

func (s *PgStorage) Close() error {
	return s.db.Close() //nolint:wrapcheck
}
