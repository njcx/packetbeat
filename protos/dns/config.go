package dns

import (
	"packetbeat/config"
	"packetbeat/protos"
)

type dnsConfig struct {
	config.ProtocolCommon `config:",inline"`
	IncludeAuthorities    bool `config:"include_authorities"`
	IncludeAdditionals    bool `config:"include_additionals"`
}

var (
	defaultConfig = dnsConfig{
		ProtocolCommon: config.ProtocolCommon{
			TransactionTimeout: protos.DefaultTransactionExpiration,
		},
	}
)
