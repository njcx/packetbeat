package module

import (
	"packetbeat/libbeat/cfgfile"
	"packetbeat/libbeat/common"
	"packetbeat/libbeat/publisher"
	"packetbeat/metricbeat/mb"
)

// Factory is used to register and reload modules
type Factory struct {
	client func() publisher.Client
}

// NewFactory creates new Reloader instance for the given config
func NewFactory(p publisher.Publisher) *Factory {
	return &Factory{
		client: p.Connect,
	}
}

func (r *Factory) Create(c *common.Config) (cfgfile.Runner, error) {
	w, err := NewWrapper(c, mb.Registry)
	if err != nil {
		return nil, err
	}

	mr := NewRunner(r.client, w)
	return mr, nil
}
