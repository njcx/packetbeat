package processors

import (
	"errors"
	"testing"

	"packetbeat/libbeat/common"
	"github.com/stretchr/testify/assert"
)

type testFilterRule struct {
	str func() string
	run func(common.MapStr) (common.MapStr, error)
}

func TestNamespace(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"test"},
		{"test.test"},
		{"abc.def.test"},
	}

	for i, test := range tests {
		t.Logf("run (%v): %v", i, test.name)

		ns := NewNamespace()
		err := ns.Register(test.name, newTestFilterRule)
		fatalError(t, err)

		cfg, _ := common.NewConfigFrom(map[string]interface{}{
			test.name: nil,
		})

		filter, err := ns.Plugin()(*cfg)

		assert.NoError(t, err)
		assert.NotNil(t, filter)
	}
}

func TestNamespaceRegisterFail(t *testing.T) {
	ns := NewNamespace()
	err := ns.Register("test", newTestFilterRule)
	fatalError(t, err)

	err = ns.Register("test", newTestFilterRule)
	assert.Error(t, err)
}

func TestNamespaceError(t *testing.T) {
	tests := []struct {
		title   string
		factory Constructor
		config  interface{}
	}{
		{
			"no module configured",
			newTestFilterRule,
			map[string]interface{}{},
		},
		{
			"unknown module configured",
			newTestFilterRule,
			map[string]interface{}{
				"notTest": nil,
			},
		},
		{
			"too many modules",
			newTestFilterRule,
			map[string]interface{}{
				"a":    nil,
				"b":    nil,
				"test": nil,
			},
		},
		{
			"filter init fail",
			func(_ common.Config) (Processor, error) {
				return nil, errors.New("test")
			},
			map[string]interface{}{
				"test": nil,
			},
		},
	}

	for i, test := range tests {
		t.Logf("run (%v): %v", i, test.title)

		ns := NewNamespace()
		err := ns.Register("test", test.factory)
		fatalError(t, err)

		config, err := common.NewConfigFrom(test.config)
		fatalError(t, err)

		_, err = ns.Plugin()(*config)
		assert.Error(t, err)
	}
}

func newTestFilterRule(_ common.Config) (Processor, error) {
	return &testFilterRule{}, nil
}

func (r *testFilterRule) String() string {
	if r.str == nil {
		return "test"
	}
	return r.str()
}

func (r *testFilterRule) Run(evt common.MapStr) (common.MapStr, error) {
	if r.run == nil {
		return evt, nil
	}
	return r.Run(evt)
}

func fatalError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}
