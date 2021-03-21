package redis

import (
	"packetbeat/config"
	"packetbeat/protos"
)

type redisConfig struct {
	config.ProtocolCommon `config:",inline"`
}

var (
	defaultConfig = redisConfig{
		ProtocolCommon: config.ProtocolCommon{
			TransactionTimeout: protos.DefaultTransactionExpiration,
		},
	}
)
