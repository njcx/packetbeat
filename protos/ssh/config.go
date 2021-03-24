package ssh

import (
	"packetbeat/config"
	"packetbeat/protos"
)

type sshConfig struct {
	config.ProtocolCommon `config:",inline"`
}

var (
	defaultConfig = sshConfig{
		ProtocolCommon: config.ProtocolCommon{
			TransactionTimeout: protos.DefaultTransactionExpiration,
		},
	}
)
