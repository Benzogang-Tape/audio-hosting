package postgres

import (
	"embed"

	pg "github.com/Benzogang-Tape/audio-hosting/playlists/pkg/db/postgres"

	"dev.gaijin.team/go/golib/e"
)

type PGStorage struct {
	db *pg.DB
	*Queries
}

//go:embed migrations/*.sql
var migrations embed.FS

func Connect(conf pg.Config) (*PGStorage, error) {
	db, err := pg.New(conf, migrations)
	if err != nil {
		return nil, e.NewFrom("creating db", err)
	}

	return &PGStorage{
		db:      db,
		Queries: New(db),
	}, nil
}

func (s *PGStorage) Close() {
	s.db.Pool.Close()
}
