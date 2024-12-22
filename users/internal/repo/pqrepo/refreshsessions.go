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

type RefreshSessions struct {
	DB *postgres.DB
}

func (re *RefreshSessions) CreateRefreshSession(
	ctx context.Context,
	session model.RefreshSession,
) error {
	conn := re.DB.GetConnection(ctx)

	sql, args, err := squirrel.Insert("refresh_sessions").
		Columns("token", "expires_at", "user_id").
		Values(session.RefreshToken, session.ExpiresAt, session.UserID).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf(
			"pqrepo.RefreshSessions.CreateRefreshSession - failed to generate SQL: %w",
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
			"pqrepo.RefreshSessions.CreateRefreshSession - failed to execute SQL: %w",
			err,
		)
	}

	return nil
}

func (re *RefreshSessions) GetRefreshSession(
	ctx context.Context,
	token string,
) (model.RefreshSession, error) {
	conn := re.DB.GetConnection(ctx)

	sql, args, err := squirrel.Select("token, expires_at, user_id").
		From("refresh_sessions").
		Where(squirrel.Eq{"token": token}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return model.RefreshSession{}, fmt.Errorf(
			"pqrepo.RefreshSessions.GetRefreshSession - failed to generate SQL: %w",
			err,
		)
	}

	var session model.RefreshSession
	if err = conn.QueryRow(ctx, sql, args...).Scan(
		&session.RefreshToken,
		&session.ExpiresAt,
		&session.UserID,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.RefreshSession{}, pqerrs.ErrNotFound
		}

		return model.RefreshSession{}, fmt.Errorf(
			"pqrepo.RefreshSessions.GetRefreshSession - failed to execute SQL: %w",
			err,
		)
	}

	return session, nil
}

func (re *RefreshSessions) GetUserRefreshSession(
	ctx context.Context,
	userID uuid.UUID,
) ([]model.RefreshSession, error) {
	conn := re.DB.GetConnection(ctx)

	sql, args, err := squirrel.Select("token, expires_at, user_id").
		From("refresh_sessions").
		Where(squirrel.Eq{"user_id": userID}).
		OrderBy("expires_at DESC").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf(
			"pqrepo.RefreshSessions.GetUserRefreshSession - failed to generate SQL: %w",
			err,
		)
	}

	var sessions []model.RefreshSession
	rows, err := conn.Query(ctx, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sessions, pqerrs.ErrNotFound
		}

		return sessions, fmt.Errorf(
			"pqrepo.RefreshSessions.GetUserRefreshSession - failed to execute SQL: %w",
			err,
		)
	}
	defer rows.Close()

	for rows.Next() {
		var session model.RefreshSession
		if err = rows.Scan(
			&session.RefreshToken,
			&session.ExpiresAt,
			&session.UserID,
		); err != nil {
			return sessions, fmt.Errorf(
				"pqrepo.RefreshSessions.GetUserRefreshSession - failed to scan row: %w",
				err,
			)
		}

		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return sessions, fmt.Errorf(
			"pqrepo.RefreshSessions.GetUserRefreshSession - failed to iterate rows: %w",
			err,
		)
	}

	return sessions, nil
}

func (re *RefreshSessions) DeleteRefreshSession(ctx context.Context, token string) error {
	conn := re.DB.GetConnection(ctx)

	sql, args, err := squirrel.Delete("refresh_sessions").
		Where(squirrel.Eq{"token": token}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf(
			"pqrepo.RefreshSessions.DeleteRefreshSession - failed to generate SQL: %w",
			err,
		)
	}

	_, err = conn.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf(
			"pqrepo.RefreshSessions.DeleteRefreshSession - failed to execute SQL: %w",
			err,
		)
	}

	return nil
}

func NewRefreshSessions(db *postgres.DB) *RefreshSessions {
	return &RefreshSessions{
		DB: db,
	}
}
