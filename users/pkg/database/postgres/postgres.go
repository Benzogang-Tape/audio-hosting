package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host     string `env:"POSTGRES_HOST"     env-required:"true"`
	Port     string `env:"POSTGRES_PORT"     env-required:"true"`
	User     string `env:"POSTGRES_USER"     env-required:"true"`
	Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
	DB       string `env:"POSTGRES_DB"       env-required:"true"`
	SSLMode  string `env:"POSTGRES_SSL"      env-required:"true"`
}

type DB struct {
	Pool *pgxpool.Pool
}

type txKey struct{}

type connector interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, optionsAndArgs ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, optionsAndArgs ...interface{}) pgx.Row
}

func New(ctx context.Context, cfg Config) (*DB, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DB,
		cfg.SSLMode,
	)

	connPool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	connection, err := connPool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't acquire connection: %w", err)
	}
	defer connection.Release()

	err = connection.Conn().Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't ping database: %w", err)
	}

	return &DB{
		Pool: connPool,
	}, nil
}

func (db *DB) Close() {
	db.Pool.Close()
}

func (db *DB) WithinTransaction(
	ctx context.Context,
	tFunc func(ctx context.Context) error,
) error {
	conn, err := db.Pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("can't acquire connection: %w", err)
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("can't begin transaction: %w", err)
	}

	err = tFunc(injectTx(ctx, tx))
	if err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("transaction failed: %w", err)
	}

	tx.Commit(ctx)

	return nil
}

func (db *DB) GetConnection(ctx context.Context) connector {
	conn := extractTx(ctx)
	if conn == nil {
		conn = db.Pool
	}

	return conn
}

func injectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func extractTx(ctx context.Context) connector {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}

	return nil
}
