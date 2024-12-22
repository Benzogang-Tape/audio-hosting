package config

import (
	"time"

	"github.com/Benzogang-Tape/audio-hosting/users/pkg/database/postgres"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string `env:"ENV" env-default:"dev"`
	GRPC     GRPCConfig
	HTTP     HTTPConfig
	Postgres postgres.Config
	Auth     Auth
}

type Auth struct {
	AccessTTL            time.Duration `env:"ACCESS_TTL"             env-default:"15m"`
	RefreshTTL           time.Duration `env:"REFRESH_TTL"            env-default:"24h"`
	RefreshSessionsLimit int           `env:"REFRESH_SESSIONS_LIMIT" env-default:"5"`

	PrivateKey string `env:"AUTH_PRIVATE_KEY" env-required:"true"`
	PublicKey  string `env:"AUTH_PUBLIC_KEY"  env-required:"true"`
}

type GRPCConfig struct {
	Host string `env:"GRPC_HOST" env-default:"localhost"`
	Port int    `env:"GRPC_PORT" env-default:"9090"`
}

type HTTPConfig struct {
	Host string `env:"HTTP_HOST" env-default:"localhost"`
	Port int    `env:"HTTP_PORT" env-default:"8080"`
}

func New() (*Config, error) {
	var cfg Config

	err := cleanenv.ReadConfig("./configs/.env", &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
