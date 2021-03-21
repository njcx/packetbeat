package processors

import (
	"errors"

	"packetbeat/libbeat/common"
	"packetbeat/libbeat/logp"
	p "packetbeat/libbeat/plugin"
)

type processorPlugin struct {
	name   string
	constr Constructor
}

var pluginKey = "libbeat.processor"

func Plugin(name string, c Constructor) map[string][]interface{} {
	return p.MakePlugin(pluginKey, processorPlugin{name, c})
}

func init() {
	p.MustRegisterLoader(pluginKey, func(ifc interface{}) error {
		p, ok := ifc.(processorPlugin)
		if !ok {
			return errors.New("plugin does not match processor plugin type")
		}

		return registry.Register(p.name, p.constr)
	})
}

type Processor interface {
	Run(event common.MapStr) (common.MapStr, error)
	String() string
}

type Constructor func(config common.Config) (Processor, error)

var registry = NewNamespace()

func RegisterPlugin(name string, constructor Constructor) {
	logp.Debug("processors", "Register plugin %s", name)

	err := registry.Register(name, constructor)
	if err != nil {
		panic(err)
	}
}
