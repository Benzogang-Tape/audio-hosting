package pg

import (
	"errors"
	"github.com/jackc/pgx/v5"

	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/repoerrs"

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

	if err.Error() == pgx.ErrNoRows.Error() {
		return repoerrs.ErrEmptyResult.Wrap(err)
	}

	pgErr := new(pgconn.PgError)

	if !errors.As(err, &pgErr) {
		return e.NewFrom("pg.wrapError unexpected error", err)
	}

	switch pgErr.Code {
	case uniqueViolation:
		return repoerrs.ErrUnique.Wrap(err)

	case foreignKeyViolation:
		return repoerrs.ErrFK.Wrap(err)

	case notNullViolation:
		return repoerrs.ErrNull.Wrap(err)

	case checkViolation:
		return repoerrs.ErrChecked.Wrap(err)
	}

	return e.NewFrom("pg.wrapError unexpected error code", err)
}
