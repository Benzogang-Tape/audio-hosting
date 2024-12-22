package pqrepo

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"

	"github.com/Benzogang-Tape/audio-hosting/users/internal/model"
	"github.com/Benzogang-Tape/audio-hosting/users/internal/repo/options"
	"github.com/Benzogang-Tape/audio-hosting/users/internal/repo/pqerrs"
	"github.com/Benzogang-Tape/audio-hosting/users/pkg/database/postgres"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type Artists struct {
	DB *postgres.DB
}

func (ar *Artists) GetArtists(
	ctx context.Context,
	options options.Options,
) ([]model.Artist, error) {
	conn := ar.DB.GetConnection(ctx)

	query := squirrel.Select("label, users.id, name, email, avatar_url, password_hash").
		From("artists").
		Join("users ON artists.user_id = users.id")

	if options.Filter.Enable {
		or := squirrel.Or{}
		for _, filter := range options.Filter.Fields {
			or = append(or, squirrel.Expr(filter.ToQueryWithParameter(), filter.FormattedValue()))
		}

		query = query.Where(or)
	}

	if options.Sort.Enable {
		query = query.OrderBy(
			options.Sort.QuottedField() + " " + options.Sort.QuottedOrder(),
		)
	} else {
		query = query.OrderBy("users.id", "name")
	}

	if options.Pagination.Enable {
		if options.Pagination.Limit > 0 {
			query = query.Limit(uint64(options.Pagination.Limit))
		}

		query = query.Offset(uint64(options.Pagination.Offset))
	}

	sql, args, err := query.PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf(
			"pqrepo.Artists.GetArtists - failed to generate SQL: %w",
			err,
		)
	}

	fmt.Println(sql, args)

	rows, err := conn.Query(ctx, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pqerrs.ErrNotFound
		}

		return nil, fmt.Errorf(
			"pqrepo.Artists.GetArtists - failed to execute SQL: %w",
			err,
		)
	}
	defer rows.Close()

	artists := make([]model.Artist, 0)

	for rows.Next() {
		var artist model.Artist
		if err = rows.Scan(
			&artist.Label,
			&artist.ID,
			&artist.Name,
			&artist.Email,
			&artist.AvatarURL,
			&artist.PasswordHash,
		); err != nil {
			return nil, fmt.Errorf(
				"pqrepo.Artists.GetArtists - failed to scan row: %w",
				err,
			)
		}

		artists = append(artists, artist)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"pqrepo.Artists.GetArtists - failed to iterate rows: %w",
			err,
		)
	}

	return artists, nil
}

func (ar *Artists) GetArtist(ctx context.Context, id uuid.UUID) (model.Artist, error) {
	conn := ar.DB.GetConnection(ctx)

	sql, args, err := squirrel.Select("label, users.id, name, email, avatar_url, password_hash").
		From("artists").
		Where(squirrel.Eq{"artists.user_id": id}).
		Join("users ON artists.user_id = users.id").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return model.Artist{}, fmt.Errorf(
			"pqrepo.Artists.GetArtist - failed to generate SQL: %w",
			err,
		)
	}

	var artist model.Artist
	if err = conn.QueryRow(ctx, sql, args...).Scan(
		&artist.Label,
		&artist.ID,
		&artist.Name,
		&artist.Email,
		&artist.AvatarURL,
		&artist.PasswordHash,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Artist{}, pqerrs.ErrNotFound
		}

		return model.Artist{}, fmt.Errorf(
			"pqrepo.Artists.GetArtist - failed to execute SQL: %w",
			err,
		)
	}

	return artist, nil
}

func (ar *Artists) UpdateArtist(ctx context.Context, artist model.Artist) error {
	conn := ar.DB.GetConnection(ctx)

	subQuery := squirrel.Update("users").
		Set("name", artist.Name).
		Set("email", artist.Email).
		Set("avatar_url", artist.AvatarURL).
		Set("password_hash", artist.PasswordHash).
		Where(squirrel.Eq{"id": artist.ID}).
		Suffix("RETURNING id")

	with := subQuery.Prefix("WITH updated_user AS (").Suffix(")")

	subSQL, subSQLArgs, err := with.ToSql()
	if err != nil {
		return fmt.Errorf("pqrepo.Artists.UpdateArtist - failed to generate sub SQL: %w", err)
	}

	query := squirrel.Update("artists").
		Set("label", artist.Label).
		Prefix(subSQL).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("pqrepo.Artists.UpdateArtist - failed to generate SQL: %w", err)
	}

	fmt.Println(sql, args)

	_, err = conn.Exec(ctx, sql, append(subSQLArgs, args...)...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pqerrs.ErrUniqueViolationCode {
				return pqerrs.ErrUniqueViolation
			}
		}

		return fmt.Errorf("pqrepo.Artists.UpdateArtist - failed to execute SQL: %w", err)
	}

	return nil
}

//nolint:dupl
func (ar *Artists) DeleteArtist(ctx context.Context, id uuid.UUID) error {
	conn := ar.DB.GetConnection(ctx)

	subQuery := squirrel.Delete("users").
		Where(squirrel.Eq{"id": id}).
		Suffix("RETURNING id").
		PlaceholderFormat(squirrel.Dollar)

	with := subQuery.Prefix("WITH deleted_user AS (").Suffix(")")

	subSQL, subSQLArgs, err := with.ToSql()
	if err != nil {
		return fmt.Errorf("pqrepo.Artists.DeleteArtist - failed to generate sub SQL: %w", err)
	}

	query := squirrel.Delete("artists").
		Where(squirrel.Expr("user_id IN (SELECT id FROM deleted_user)")).
		Prefix(subSQL).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("pqrepo.Artists.DeleteArtist - failed to generate SQL: %w", err)
	}

	_, err = conn.Exec(ctx, sql, append(subSQLArgs, args...)...)
	if err != nil {
		return fmt.Errorf("pqrepo.Artists.DeleteArtist - failed to execute SQL: %w", err)
	}

	return nil
}

func (ar *Artists) GetArtistsCount(ctx context.Context, options options.Options) (int, error) {
	conn := ar.DB.GetConnection(ctx)

	query := squirrel.Select("count(*)").
		From("artists").
		Join("users ON artists.user_id = users.id").
		PlaceholderFormat(squirrel.Dollar)

	if options.Filter.Enable {
		or := squirrel.Or{}
		for _, filter := range options.Filter.Fields {
			or = append(or, squirrel.Expr(filter.ToQueryWithParameter(), filter.FormattedValue()))
		}

		query = query.Where(or)

	}

	query = query.PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return 0, fmt.Errorf("pqrepo.Artists.GetArtistsCount - failed to generate SQL: %w", err)
	}

	var count int
	if err = conn.QueryRow(ctx, sql, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("pqrepo.Artists.GetArtistsCount - failed to execute SQL: %w", err)
	}

	return count, nil
}

func NewArtists(db *postgres.DB) *Artists {
	return &Artists{
		DB: db,
	}
}
