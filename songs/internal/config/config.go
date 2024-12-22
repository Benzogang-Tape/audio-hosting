package config

import (
	"time"

	"github.com/rs/zerolog"
)

type Config struct {
	Environment string      `env:"ENVIRONMENT" env-default:"dev" yaml:"environment"`
	Connections Connections `yaml:"connections"`
	Servers     Servers     `yaml:"servers"`
	Logging     Logging     `yaml:"logging"`
	Features    Features    `yaml:"features"`
}

type Servers struct {
	Tls  Tls    `yaml:"common.tls"`
	Host string `env:"HOST" env-default:"localhost:8080" yaml:"common.host"`
	Grpc Grpc   `yaml:"grpc"`
	Http Http   `yaml:"http"`
}

type Tls struct {
	CertPath string `env:"TLS_CERT_PATH" env-default:"/etc/sso/tls.crt" yaml:"certPath"`
	KeyPath  string `env:"TLS_KEY_PATH"  env-default:"/etc/sso/tls.key" yaml:"keyPath"`
}

type Http struct {
	Port    int           `env:"HTTP_PORT"    env-default:"8080"  yaml:"port"`
	UseTls  bool          `env:"HTTP_USE_TLS" env-default:"false" yaml:"useTls"`
	Timeout time.Duration `env:"HTTP_TIMEOUT" env-default:"5s" yaml:"timeout"`
}

type Grpc struct {
	Port    int           `env:"GRPC_PORT"    env-default:"5050"  yaml:"port"`
	UseTls  bool          `env:"GRPC_USE_TLS" env-default:"false" yaml:"useTls"`
	Timeout time.Duration `env:"GRPC_TIMEOUT" env-default:"5s"    yaml:"timeout"`
}

type Connections struct {
	Postgres     Postgres     `yaml:"postgres"`
	Redis        Redis        `yaml:"redis"`
	Kafka        Kafka        `yaml:"kafka"`
	S3           S3           `yaml:"s3"`
	UsersService UsersService `yaml:"usersService"`
}

type Postgres struct {
	Host     string `env:"POSTGRES_HOST" e.g:"postgres" yaml:"host"`
	Port     int    `env:"POSTGRES_PORT" env-default:"5432" yaml:"port"`
	User     string `env:"POSTGRES_USER" e.g:"songs_user" yaml:"user"`
	Password string `env:"POSTGRES_PASS" e.g:"hard_password1234" yaml:"password"`
	Database string `env:"POSTGRES_DB" e.g:"songs" yaml:"database"`
	SslMode  string `env:"POSTGRES_SSL" env-default:"disable" yaml:"sslmode"`
}

type Redis struct {
	Host     string `env:"REDIS_HOST" e.g:"redis" yaml:"host"`
	Port     int    `env:"REDIS_PORT" env-default:"6379" yaml:"port"`
	User     string `env:"REDIS_USER" e.g:"default" yaml:"user"`
	Password string `env:"REDIS_PASS" e.g:"hard_password1234" yaml:"password"`
	Db       int    `env:"REDIS_DB" env-default:"0"  yaml:"db"`
}

type Kafka struct {
	Topic   string   `env:"KAFKA_TOPIC" env-default:"released-songs" yaml:"songReleasedTopic"`
	Brokers []string `env:"KAFKA_BROKERS" env-default:"kafka:9092" yaml:"brokers"`
}

type S3 struct {
	Endpoint     string `env:"S3_ENDPOINT" env-default:"minio:9000" yaml:"endpoint"`
	AccessKey    string `env:"S3_ACCESS_KEY" e.g:"Q3AM3UQ867SPQQA43P2F" yaml:"accessKey"`
	SecretKey    string `env:"S3_SECRET_KEY" e.g:"zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG" yaml:"secretKey"`
	UseSsl       bool   `env:"S3_USE_SSL" env-default:"false" yaml:"useSsl"`
	SongsBucket  string `env:"S3_SONGS_BUCKET" e.g:"songs" yaml:"songsBucket"`
	ImagesBucket string `env:"S3_IMAGES_BUCKET" e.g:"songs_images" yaml:"imagesBucket"`
}

type UsersService struct {
	UseFake bool   `env:"USERS_SERVICE_USE_FAKE" env-default:"false" yaml:"useFake"`
	Target  string `env:"USERS_SERVICE_TARGET" e.g:"users:9090" yaml:"target"`
}

type Logging struct {
	Level zerolog.Level `env:"LOG_LEVEL" env-default:"info" yaml:"level"`
}

type Features struct {
	Auth struct { //nolint:revive
		PublicKey string `env:"JWT_PUBLIC_KEY" e.g:"Gvbo6JyyS410wg87Gq0N9kphc67A5Lb1VS1wupzwYTU=" yaml:"publicKey"`
	} `yaml:"auth"`
	Cache struct { //nolint:revive
		SongsTtl   time.Duration `env:"CACHE_SONGS_TTL" env-default:"5m" yaml:"songsTtl"`
		MySongsTtl time.Duration `env:"CACHE_MY_SONGS_TTL" env-default:"5m" yaml:"mySongsTtl"`
	} `yaml:"cache"`
}
