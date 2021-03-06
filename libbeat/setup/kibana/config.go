package kibana

import (
	"time"

	"packetbeat/libbeat/outputs"
)

type kibanaConfig struct {
	Protocol string             `config:"protocol"`
	Host     string             `config:"host"`
	Path     string             `config:"path"`
	Username string             `config:"username"`
	Password string             `config:"password"`
	TLS      *outputs.TLSConfig `config:"ssl"`
	Timeout  time.Duration      `config:"timeout"`
}

var (
	defaultKibanaConfig = kibanaConfig{
		Protocol: "http",
		Host:     "",
		Path:     "",
		Username: "",
		Password: "",
		Timeout:  90 * time.Second,
		TLS:      nil,
	}
)
