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

type Users struct {
	DB *postgres.DB
}

func (us *Users) GetUserByEmail(ctx context.Context, email string) (model.User, error) {
	conn := us.DB.GetConnection(ctx)

	sql, args, err := squirrel.Select("id, name, email, avatar_url, password_hash").
		From("users").
		Where(squirrel.Eq{"email": email}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return model.User{}, fmt.Errorf(
			"pqrepo.Users.GetUser - failed to generate SQL: %w",
			err,
		)
	}

	var user model.User
	if err = conn.QueryRow(ctx, sql, args...).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.AvatarURL,
		&user.PasswordHash,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, pqerrs.ErrNotFound
		}

		return model.User{}, fmt.Errorf(
			"pqrepo.Users.GetUser - failed to execute SQL: %w",
			err,
		)
	}

	return user, nil
}

//nolint:dupl
func (us *Users) GetFollowers(ctx context.Context, userID uuid.UUID) ([]model.User, error) {
	conn := us.DB.GetConnection(ctx)

	sql, args, err := squirrel.Select("follower_id, name, email, avatar_url").
		From("users_users").
		Where(squirrel.Eq{"followed_id": userID}).
		Join("users ON users_users.follower_id = users.id").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf(
			"pqrepo.Users.GetFollowers - failed to generate SQL: %w",
			err,
		)
	}

	rows, err := conn.Query(ctx, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pqerrs.ErrNotFound
		}

		return nil, fmt.Errorf(
			"pqrepo.Users.GetFollowers - failed to execute SQL: %w",
			err,
		)
	}
	defer rows.Close()

	followers := make([]model.User, 0)

	for rows.Next() {
		var follower model.User
		if err = rows.Scan(
			&follower.ID,
			&follower.Name,
			&follower.Email,
			&follower.AvatarURL,
		); err != nil {
			return nil, fmt.Errorf(
				"pqrepo.Users.GetFollowers - failed to scan row: %w",
				err,
			)
		}

		followers = append(followers, follower)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf(
			"pqrepo.Users.GetFollowers - failed to iterate rows: %w",
			err,
		)
	}

	return followers, nil
}

//nolint:dupl
func (us *Users) GetFollowed(ctx context.Context, userID uuid.UUID) ([]model.User, error) {
	conn := us.DB.GetConnection(ctx)

	sql, args, err := squirrel.Select("followed_id, name, email, avatar_url").
		From("users_users").
		Where(squirrel.Eq{"follower_id": userID}).
		Join("users ON users_users.followed_id = users.id").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf(
			"pqrepo.Users.GetFollowed - failed to generate SQL: %w",
			err,
		)
	}

	rows, err := conn.Query(ctx, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pqerrs.ErrNotFound
		}

		return nil, fmt.Errorf(
			"pqrepo.Users.GetFollowed - failed to execute SQL: %w",
			err,
		)
	}
	defer rows.Close()

	followed := make([]model.User, 0)

	for rows.Next() {
		var follower model.User
		if err = rows.Scan(
			&follower.ID,
			&follower.Name,
			&follower.Email,
			&follower.AvatarURL,
		); err != nil {
			return nil, fmt.Errorf(
				"pqrepo.Users.GetFollowed - failed to scan row: %w",
				err,
			)
		}

		followed = append(followed, follower)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf(
			"pqrepo.Users.GetFollowed - failed to iterate rows: %w",
			err,
		)
	}

	return followed, nil
}

func (us *Users) Follow(ctx context.Context, followerID uuid.UUID, followedID uuid.UUID) error {
	conn := us.DB.GetConnection(ctx)

	sql, args, err := squirrel.Insert("users_users").
		Columns("follower_id", "followed_id").
		Values(followerID, followedID).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		var pqErr *pgconn.PgError
		if errors.As(err, &pqErr) {
			if pqErr.Code == pqerrs.ErrFKViolationCode {
				return pqerrs.ErrFKViolation
			}
		}

		return fmt.Errorf(
			"pqrepo.Users.Follow - failed to generate SQL: %w",
			err,
		)
	}

	_, err = conn.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf(
			"pqrepo.Users.Follow - failed to execute SQL: %w",
			err,
		)
	}

	return nil
}

func (us *Users) Unfollow(ctx context.Context, followerID uuid.UUID, followedID uuid.UUID) error {
	conn := us.DB.GetConnection(ctx)

	sql, args, err := squirrel.Delete("users_users").
		Where(squirrel.Eq{"follower_id": followerID}).
		Where(squirrel.Eq{"followed_id": followedID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf(
			"pqrepo.Users.Unfollow - failed to generate SQL: %w",
			err,
		)
	}

	_, err = conn.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf(
			"pqrepo.Users.Unfollow - failed to execute SQL: %w",
			err,
		)
	}

	return nil
}

func (us *Users) UpdateNotificationsSettings(
	ctx context.Context,
	settings model.NotificationSettings,
) error {
	conn := us.DB.GetConnection(ctx)

	sql, args, err := squirrel.Update("notifications_settings").
		Set("email_notifications", settings.EmailNotifications).
		Where(squirrel.Eq{"user_id": settings.UserID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf(
			"pqrepo.Users.UpdateNotificationsSettings - failed to generate SQL: %w",
			err,
		)
	}

	_, err = conn.Exec(ctx, sql, args...)
	if err != nil {
		var pqErr *pgconn.PgError
		if errors.As(err, &pqErr) {
			if pqErr.Code == pqerrs.ErrFKViolationCode {
				return pqerrs.ErrFKViolation
			}
		}

		return fmt.Errorf(
			"pqrepo.Users.UpdateNotificationsSettings - failed to execute SQL: %w",
			err,
		)
	}

	return nil
}

func (us *Users) GetNotificationsSettings(
	ctx context.Context,
	userID uuid.UUID,
) (model.NotificationSettings, error) {
	conn := us.DB.GetConnection(ctx)

	sql, args, err := squirrel.Select("user_id, email_notifications").
		From("notifications_settings").
		Where(squirrel.Eq{"user_id": userID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return model.NotificationSettings{}, fmt.Errorf(
			"pqrepo.Users.GetNotificationsSettings - failed to generate SQL: %w",
			err,
		)
	}

	var settings model.NotificationSettings
	if err = conn.QueryRow(ctx, sql, args...).Scan(
		&settings.UserID,
		&settings.EmailNotifications,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.NotificationSettings{}, pqerrs.ErrNotFound
		}

		return model.NotificationSettings{}, fmt.Errorf(
			"pqrepo.Users.GetNotificationsSettings - failed to execute SQL: %w",
			err,
		)
	}

	return settings, nil
}

func (us *Users) MakeArtist(ctx context.Context, userID uuid.UUID) error {
	conn := us.DB.GetConnection(ctx)

	deleteSql, deleteArgs, err := squirrel.Delete("listeners").
		Where(squirrel.Eq{"user_id": userID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf(
			"pqrepo.Users.MakeArtist - failed to generate delete SQL: %w",
			err,
		)
	}

	sql, args, err := squirrel.Insert("artists").
		Columns("user_id").
		Values(userID).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf(
			"pqrepo.Users.MakeArtist - failed to generate insert SQL: %w",
			err,
		)
	}

	_, err = conn.Exec(ctx, deleteSql, deleteArgs...)
	if err != nil {
		return fmt.Errorf(
			"pqrepo.Users.MakeArtist - failed to execute delete SQL: %w",
			err,
		)
	}

	_, err = conn.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf(
			"pqrepo.Users.MakeArtist - failed to execute insert SQL: %w",
			err,
		)
	}

	return nil
}

func NewUsers(db *postgres.DB) *Users {
	return &Users{
		DB: db,
	}
}
