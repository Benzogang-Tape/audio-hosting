package redis

import (
	"context"
	"errors"
	"time"

	"github.com/Benzogang-Tape/audio-hosting/songs/pkg/repoerrs"

	"dev.gaijin.team/go/golib/e"
	"dev.gaijin.team/go/golib/fields"
	"github.com/redis/go-redis/v9"
)

const nsSep = ":"

type RedStorage struct {
	db *redis.Client
}

type Config struct {
	Addr     string
	Username string
	Password string
	Db       int
}

func Connect(ctx context.Context, cfg Config) (*RedStorage, error) {
	rdb := redis.NewClient(&redis.Options{ //nolint:exhaustruct
		Addr:     cfg.Addr,
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.Db,
	})

	err := rdb.Ping(ctx).Err()
	if err != nil {
		return nil, e.NewFrom("redis ping", err)
	}

	return &RedStorage{
		db: rdb,
	}, nil
}

func (rs *RedStorage) Close() error {
	return rs.db.Close() //nolint:wrapcheck
}

type RedNs struct {
	*RedStorage
	ns string
}

func (rs *RedStorage) With(ns string) RedNs {
	return RedNs{
		RedStorage: rs,
		ns:         ns,
	}
}

// Set stores a value in the sub-storage with an optional expiration.
func (r RedNs) SetBytes(ctx context.Context, key string, val []byte, exp time.Duration) error {
	err := r.db.Set(ctx, r.namespaced(key), val, exp).Err()
	if err != nil {
		return e.NewFrom("setting key", err, fields.F("key", key))
	}

	return nil
}

// Get retrieves a value from the sub-storage.
func (r RedNs) GetBytes(ctx context.Context, key string) ([]byte, error) {
	bytes, err := r.db.Get(ctx, r.namespaced(key)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, repoerrs.ErrEmptyResult
	}

	if err != nil {
		return nil, e.NewFrom("getting key", err, fields.F("key", key))
	}

	return bytes, nil
}

// Del removes a value from the sub-storage.
func (r RedNs) Del(ctx context.Context, key string) error {
	err := r.db.Del(ctx, r.namespaced(key)).Err()
	if err != nil {
		return e.NewFrom("deleting key", err, fields.F("key", key))
	}

	return nil
}

func (r RedNs) namespaced(s string) string {
	return r.ns + nsSep + s
}
