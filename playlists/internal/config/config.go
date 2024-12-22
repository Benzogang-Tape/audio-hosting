package config

import (
	"github.com/Benzogang-Tape/audio-hosting/playlists/internal/client/songs"
	s3 "github.com/Benzogang-Tape/audio-hosting/playlists/pkg/db/minio"
	"time"

	pg "github.com/Benzogang-Tape/audio-hosting/playlists/pkg/db/postgres"
	"github.com/Benzogang-Tape/audio-hosting/playlists/pkg/db/redis"
)

type Config struct {
	ENV         string      `env:"ENV" env-default:"dev" yaml:"env"`
	Servers     Servers     `yaml:"servers"`
	Connections Connections `yaml:"connections"`
	Secrets     Secrets     `yaml:"secrets"`
}

type Servers struct {
	GRPC GRPCConfig `yaml:"grpc"`
	HTTP HTTPConfig `yaml:"http"`
}

type GRPCConfig struct {
	Host    string        `env:"GRPC_HOST" env-default:"" yaml:"host" `
	Port    int           `env:"GRPC_SERVER_PORT" env-default:"50051" yaml:"port"`
	Timeout time.Duration `env:"GRPC_TIMEOUT" env-default:"5s" yaml:"timeout"`
}

type HTTPConfig struct {
	Host    string        `env:"HTTP_HOST" env-default:"" yaml:"host" `
	Port    int           `env:"HTTP_PORT" env-default:"8080" yaml:"port"`
	Timeout time.Duration `env:"HTTP_TIMEOUT" env-default:"5s" yaml:"timeout"`
}

type Connections struct {
	PGConfig    pg.Config    `yaml:"postgres"`
	RedisConfig redis.Config `yaml:"redis"`
	SongsConn   songs.Config `yaml:"songs"`
	S3Config    s3.Config    `yaml:"s3"`
}

type Secrets struct {
	Public string `env:"PUBLIC_KEY" env-default:"" yaml:"public"`
}
