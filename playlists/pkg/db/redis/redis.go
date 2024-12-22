package redis

import (
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Host string `env:"REDIS_HOST" env-default:"localhost" yaml:"host"`
	Port int    `env:"REDIS_PORT" env-default:"6379" yaml:"port"`
}

func New(config Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", config.Host, config.Port),
	})
}
