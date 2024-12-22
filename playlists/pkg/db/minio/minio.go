package minio

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Endpoint     string `env:"MINIO_ENDPOINT" env-default:"minio:9000" yaml:"endpoint"`
	AccessKey    string `env:"MINIO_ACCESS_KEY" env-default:"" yaml:"accessKey"`
	SecretKey    string `env:"MINIO_SECRET_KEY" env-default:"" yaml:"secretKey"`
	UseSsl       bool   `env:"MINIO_USE_SSL" env-default:"false" yaml:"useSsl"`
	CoversBucket string `env:"MINIO_COVERS_BUCKET" env-default:"covers" yaml:"coversBucket"`
}

func Connect(conf Config) (*minio.Client, error) {
	return minio.New( //nolint:wrapcheck
		conf.Endpoint,
		&minio.Options{ //nolint:exhaustruct
			Creds:  credentials.NewStaticV2(conf.AccessKey, conf.SecretKey, ""),
			Secure: conf.UseSsl,
		},
	)
}
