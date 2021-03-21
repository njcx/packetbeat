package console

import (
	"packetbeat/libbeat/outputs"
)

type Config struct {
	Codec outputs.CodecConfig `config:"codec"`

	// old pretty settings to use if no codec is configured
	Pretty bool `config:"pretty"`
}
