// +build integration

package cluster

import (
	"testing"

	mbtest "packetbeat/metricbeat/mb/testing"
	"packetbeat/metricbeat/module/couchbase"
)

func TestData(t *testing.T) {
	f := mbtest.NewEventFetcher(t, getConfig())

	err := mbtest.WriteEvent(f, t)
	if err != nil {
		t.Fatal("write", err)
	}
}

func getConfig() map[string]interface{} {
	return map[string]interface{}{
		"module":     "couchbase",
		"metricsets": []string{"cluster"},
		"hosts":      []string{couchbase.GetEnvDSN()},
	}
}
