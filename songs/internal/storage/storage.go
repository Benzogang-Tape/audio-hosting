package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/Benzogang-Tape/audio-hosting/songs/internal/config"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/storage/broker"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/storage/postgres"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/storage/redis"
	"github.com/Benzogang-Tape/audio-hosting/songs/internal/storage/s3minio"
)

type Storage struct {
	*postgres.PgStorage
	*redis.RedStorage
	*s3minio.S3Storage
	*broker.KafkaProducer

	c Config
}

type PostgresConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
	SslMode  string
}

type RedisConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Db       int
}

type MinioConfig struct {
	Endpoint     string
	AccessKey    string
	SecretKey    string
	UseSsl       bool
	SongsBucket  string
	ImagesBucket string
}

type KafkaConfig struct {
	Brokers []string
	Topic   string
}

type Config struct {
	MySongsTtl time.Duration
	SongsTtl   time.Duration
}

func New(ctx context.Context) (*Storage, error) {
	cfg := config.Get()
	pgconf := cfg.Connections.Postgres
	rdconf := cfg.Connections.Redis
	s3conf := cfg.Connections.S3

	return NewWithConfig(ctx, PostgresConfig{
		Host:     pgconf.Host,
		Port:     pgconf.Port,
		Username: pgconf.User,
		Password: pgconf.Password,
		Database: pgconf.Database,
		SslMode:  pgconf.SslMode,
	}, RedisConfig{
		Host:     rdconf.Host,
		Port:     rdconf.Port,
		Username: rdconf.User,
		Password: rdconf.Password,
		Db:       rdconf.Db,
	}, MinioConfig{
		Endpoint:     s3conf.Endpoint,
		AccessKey:    s3conf.AccessKey,
		SecretKey:    s3conf.SecretKey,
		UseSsl:       s3conf.UseSsl,
		SongsBucket:  s3conf.SongsBucket,
		ImagesBucket: s3conf.ImagesBucket,
	}, KafkaConfig{
		Brokers: cfg.Connections.Kafka.Brokers,
		Topic:   cfg.Connections.Kafka.Topic,
	},
		Config{
			MySongsTtl: cfg.Features.Cache.MySongsTtl,
			SongsTtl:   cfg.Features.Cache.SongsTtl,
		})
}

func NewWithConfig(ctx context.Context,
	pgconf PostgresConfig,
	rdconf RedisConfig,
	s3conf MinioConfig,
	kconf KafkaConfig,
	cfg Config,
) (*Storage, error) {
	pgconn := fmt.Sprintf( //nolint:nosprintfhostport
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		pgconf.Username, pgconf.Password, pgconf.Host, pgconf.Port, pgconf.Database, pgconf.SslMode)

	postgresDatabase, err := postgres.Connect(ctx, pgconn)
	if err != nil {
		return nil, fmt.Errorf("postgres connect: %w", err)
	}

	redisDatabase, err := redis.Connect(ctx, redis.Config{
		Addr:     fmt.Sprintf("%s:%d", rdconf.Host, rdconf.Port),
		Username: rdconf.Username,
		Password: rdconf.Password,
		Db:       rdconf.Db,
	})
	if err != nil {
		return nil, fmt.Errorf("redis connect: %w", err)
	}

	s3Database, err := s3minio.Connect(ctx, s3minio.Config(s3conf))
	if err != nil {
		return nil, fmt.Errorf("s3 connect: %w", err)
	}

	kBroker, err := broker.Connect(kconf.Topic, kconf.Brokers)
	if err != nil {
		return nil, fmt.Errorf("kafka connect: %w", err)
	}

	return &Storage{
		PgStorage:     postgresDatabase,
		RedStorage:    redisDatabase,
		S3Storage:     s3Database,
		KafkaProducer: kBroker,
		c:             cfg,
	}, nil
}

func (s *Storage) Close() error {
	err := tripleError{} //nolint:exhaustruct

	err.err1 = s.RedStorage.Close()
	err.err2 = s.PgStorage.Close()
	err.err3 = s.KafkaProducer.Close()

	if err.IsError() {
		return fmt.Errorf("disconnecting from storages: %w", &err)
	}

	return nil
}

type tripleError struct {
	err1 error
	err2 error
	err3 error
}

func (e *tripleError) Error() string {
	return fmt.Sprintf("%v, %v, %v", e.err1, e.err2, e.err3)
}

func (e *tripleError) IsError() bool {
	return e.err1 != nil || e.err2 != nil || e.err3 != nil
}
