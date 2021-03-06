package config

import (
	"time"

	"packetbeat/libbeat/common"
	"packetbeat/libbeat/common/droppriv"
	"packetbeat/procs"
)

type Config struct {
	Interfaces     InterfacesConfig          `config:"interfaces"`
	Flows          *Flows                    `config:"flows"`
	Protocols      map[string]*common.Config `config:"protocols"`
	Procs          procs.ProcsConfig         `config:"procs"`
	IgnoreOutgoing bool                      `config:"ignore_outgoing"`
	RunOptions     droppriv.RunOptions
}

type InterfacesConfig struct {
	Device       string `config:"device"`
	Type         string `config:"type"`
	File         string `config:"file"`
	WithVlans    bool   `config:"with_vlans"`
	BpfFilter    string `config:"bpf_filter"`
	Snaplen      int    `config:"snaplen"`
	BufferSizeMb int    `config:"buffer_size_mb"`
	TopSpeed     bool
	Dumpfile     string
	OneAtATime   bool
	Loop         int
}

type Flows struct {
	Enabled *bool  `config:"enabled"`
	Timeout string `config:"timeout"`
	Period  string `config:"period"`
}

type ProtocolCommon struct {
	Ports              []int         `config:"ports"`
	SendRequest        bool          `config:"send_request"`
	SendResponse       bool          `config:"send_response"`
	TransactionTimeout time.Duration `config:"transaction_timeout"`
}

func (f *Flows) IsEnabled() bool {
	return f != nil && (f.Enabled == nil || *f.Enabled)
}
