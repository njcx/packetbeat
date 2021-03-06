// +build !integration

package module_test

import (
	"testing"

	"packetbeat/libbeat/common"
	"packetbeat/libbeat/publisher"
	pubtest "packetbeat/libbeat/publisher/testing"
	"packetbeat/metricbeat/mb"
	"packetbeat/metricbeat/mb/module"

	"github.com/stretchr/testify/assert"
)

func TestRunner(t *testing.T) {
	pubClient, factory := newPubClientFactory()

	config, err := common.NewConfigFrom(map[string]interface{}{
		"module":     moduleName,
		"metricsets": []string{metricSetName},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create a new Wrapper based on the configuration.
	m, err := module.NewWrapper(config, mb.Registry)
	if err != nil {
		t.Fatal(err)
	}

	// Create the Runner facade.
	runner := module.NewRunner(factory, m)

	// Start the module and have it publish to a new publisher.Client.
	runner.Start()

	assert.NotNil(t, <-pubClient.Channel)

	// Stop the module. This blocks until all MetricSets in the Module have
	// stopped and the publisher.Client is closed.
	runner.Stop()
}

// newPubClientFactory returns a new ChanClient and a function that returns
// the same Client when invoked. This simulates the return value of
// Publisher.Connect.
func newPubClientFactory() (*pubtest.ChanClient, func() publisher.Client) {
	client := pubtest.NewChanClient(10)
	return client, func() publisher.Client { return client }
}
