package pgsql

import (
	"packetbeat/config"
	"packetbeat/protos"
)

type pgsqlConfig struct {
	config.ProtocolCommon `config:",inline"`
	MaxRowLength          int `config:"max_row_length"`
	MaxRows               int `config:"max_rows"`
}

var (
	defaultConfig = pgsqlConfig{
		ProtocolCommon: config.ProtocolCommon{
			TransactionTimeout: protos.DefaultTransactionExpiration,
		},
		MaxRowLength: 1024,
		MaxRows:      10,
	}
)
