package migrations

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/Benzogang-Tape/audio-hosting/users/pkg/database/postgres"
	_ "github.com/lib/pq"
)

//go:embed *.sql
var embedMigrations embed.FS

func Run(cfg postgres.Config) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DB,
		cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		panic(err)
	}

	if err := goose.Up(db, "."); err != nil {
		panic(err)
	}
}
