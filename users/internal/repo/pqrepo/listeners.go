package pqrepo

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"

	"github.com/Benzogang-Tape/audio-hosting/users/internal/model"
	"github.com/Benzogang-Tape/audio-hosting/users/internal/repo/pqerrs"
	"github.com/Benzogang-Tape/audio-hosting/users/pkg/database/postgres"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type Listeners struct {
	DB *postgres.DB
}

func (li *Listeners) GetListener(ctx context.Context, id uuid.UUID) (model.Listener, error) {
	conn := li.DB.GetConnection(ctx)

	sql, args, err := squirrel.Select("users.id, name, email, avatar_url, password_hash").
		From("listeners").
		Where(squirrel.Eq{"listeners.user_id": id}).
		Join("users ON listeners.user_id = users.id").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return model.Listener{}, fmt.Errorf(
			"pqrepo.Listeners.GetListener - failed to generate SQL: %w",
			err,
		)
	}

	var listener model.Listener
	if err = conn.QueryRow(ctx, sql, args...).Scan(
		&listener.ID,
		&listener.Name,
		&listener.Email,
		&listener.AvatarURL,
		&listener.PasswordHash,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Listener{}, pqerrs.ErrNotFound
		}

		return model.Listener{}, fmt.Errorf(
			"pqrepo.Listeners.GetListener - failed to execute SQL: %w",
			err,
		)
	}

	return listener, nil
}

func (li *Listeners) UpdateListener(ctx context.Context, listener model.Listener) error {
	conn := li.DB.GetConnection(ctx)

	sql, args, err := squirrel.Update("users").
		Set("name", listener.Name).
		Set("email", listener.Email).
		Set("avatar_url", listener.AvatarURL).
		Set("password_hash", listener.PasswordHash).
		Where(squirrel.Eq{"id": listener.ID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf(
			"pqrepo.Listeners.UpdateListener - failed to generate SQL: %w",
			err,
		)
	}

	_, err = conn.Exec(ctx, sql, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pqerrs.ErrUniqueViolationCode {
				return pqerrs.ErrUniqueViolation
			}
		}

		return fmt.Errorf(
			"pqrepo.Listeners.UpdateListener - failed to execute SQL: %w",
			err,
		)
	}

	return nil
}

//nolint:dupl
func (li *Listeners) DeleteListener(ctx context.Context, id uuid.UUID) error {
	conn := li.DB.GetConnection(ctx)

	subquery := squirrel.Delete("users").
		Where(squirrel.Eq{"id": id}).
		Suffix("RETURNING id").
		PlaceholderFormat(squirrel.Dollar)

	with := subquery.Prefix("WITH deleted_user AS (").Suffix(")")

	subSQL, subSQLArgs, err := with.ToSql()
	if err != nil {
		return fmt.Errorf("pqrepo.Listeners.DeleteListener - failed to generate sub SQL: %w", err)
	}

	query := squirrel.Delete("listeners").
		Where(squirrel.Expr("user_id IN (SELECT id FROM deleted_user)")).
		Prefix(subSQL).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("pqrepo.Listeners.DeleteListener - failed to generate SQL: %w", err)
	}

	_, err = conn.Exec(ctx, sql, append(subSQLArgs, args...)...)
	if err != nil {
		return fmt.Errorf("pqrepo.Listeners.DeleteListener - failed to execute SQL: %w", err)
	}

	return nil
}

func (li *Listeners) GetListeners(ctx context.Context) ([]model.Listener, error) {
	conn := li.DB.GetConnection(ctx)

	sql, args, err := squirrel.Select("users.id, name, email, avatar_url, password_hash").
		From("listeners").
		Join("users ON listeners.user_id = users.id").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf(
			"pqrepo.Listeners.GetListeners - failed to generate SQL: %w",
			err,
		)
	}

	rows, err := conn.Query(ctx, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pqerrs.ErrNotFound
		}

		return nil, fmt.Errorf(
			"pqrepo.Listeners.GetListeners - failed to execute SQL: %w",
			err,
		)
	}
	defer rows.Close()

	listeners := make([]model.Listener, 0)

	for rows.Next() {
		var listener model.Listener
		if err = rows.Scan(
			&listener.ID,
			&listener.Name,
			&listener.Email,
			&listener.AvatarURL,
			&listener.PasswordHash,
		); err != nil {
			return nil, fmt.Errorf(
				"pqrepo.Listeners.GetListeners - failed to scan row: %w",
				err,
			)
		}

		listeners = append(listeners, listener)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf(
			"pqrepo.Listeners.GetListeners - failed to iterate rows: %w",
			err,
		)
	}

	return listeners, nil
}

func (li *Listeners) CreateListener(ctx context.Context, listener model.Listener) error {
	conn := li.DB.GetConnection(ctx)

	subQuery := squirrel.Insert("users").
		Columns("id", "name", "email", "avatar_url", "password_hash").
		Values(listener.ID, listener.Name, listener.Email, listener.AvatarURL, listener.PasswordHash).
		Suffix("RETURNING id")

	with := subQuery.Prefix("WITH new_user AS (").Suffix(")")

	subSQL, subSQLArgs, err := with.ToSql()
	if err != nil {
		return fmt.Errorf("pqrepo.Listeners.CreateListener - failed to generate sub SQL: %w", err)
	}

	query := squirrel.Insert("listeners").
		Columns("user_id").
		Select(squirrel.Select("id").From("new_user")).
		Prefix(subSQL).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("pqrepo.Listeners.CreateListener - failed to generate SQL: %w", err)
	}

	_, err = conn.Exec(ctx, sql, append(subSQLArgs, args...)...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pqerrs.ErrUniqueViolationCode {
				return pqerrs.ErrUniqueViolation
			}
		}

		return fmt.Errorf("pqrepo.Listeners.CreateListener - failed to execute SQL: %w", err)
	}

	return nil
}

func NewListeners(db *postgres.DB) *Listeners {
	return &Listeners{
		DB: db,
	}
}
