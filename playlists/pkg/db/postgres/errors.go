package postgres

import (
	"database/sql"
	"errors"

	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/repoerrors"

	"dev.gaijin.team/go/golib/e"
	"github.com/jackc/pgx/v5/pgconn"
)

// copied from https://github.com/jackc/pgerrcode/blob/master/errcode.go
const (
	notNullViolation    = "23502"
	foreignKeyViolation = "23503"
	uniqueViolation     = "23505"
	checkViolation      = "23514"
)

func wrapError(err error) error {
	if err == nil {
		return nil
	}

	if err.Error() == sql.ErrNoRows.Error() {
		return repoerrors.ErrEmptyResult.Wrap(err)
	}

	pgErr := new(pgconn.PgError)

	if !errors.As(err, &pgErr) {
		return e.NewFrom("pg.wrapError unexpected error", err)
	}

	switch pgErr.Code {
	case uniqueViolation:
		return repoerrors.ErrUnique.Wrap(err)

	case foreignKeyViolation:
		return repoerrors.ErrFK.Wrap(err)

	case notNullViolation:
		return repoerrors.ErrNull.Wrap(err)

	case checkViolation:
		return repoerrors.ErrChecked.Wrap(err)
	}

	return e.NewFrom("pg.wrapError unexpected error code", err)
}
