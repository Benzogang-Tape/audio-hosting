package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_relatePath(t *testing.T) {
	tt := []struct {
		name   string
		path   string
		target string
		want   string
	}{
		{
			name:   "from example",
			path:   "./certs/tls.key",
			target: "../configs/conf.yaml",
			want:   "../configs/certs/tls.key",
		},
		{
			name:   "absolute path",
			path:   "/etc/app/tls.crt",
			target: "../configs/conf.yaml",
			want:   "/etc/app/tls.crt",
		},
		{
			name:   "empty path",
			path:   "",
			target: "../configs/conf.yaml",
			want:   "",
		},
		{
			name:   "from parent",
			path:   "certs/tls.key",
			target: "../configs/conf.yaml",
			want:   "../configs/certs/tls.key",
		},
		{
			name:   "absolute target",
			path:   "../../ssl/tls.key",
			target: "/etc/app/configs/conf.yaml",
			want:   "/etc/ssl/tls.key",
		},

		{
			name:   "same dir",
			path:   "tls.key",
			target: "./configs/conf.yaml",
			want:   "configs/tls.key",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got := relatePath(tc.path, tc.target)

			assert.Equal(t, tc.want, got)
		})
	}
}
