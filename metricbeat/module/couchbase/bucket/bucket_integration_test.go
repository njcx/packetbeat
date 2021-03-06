// +build integration

package bucket

import (
	"testing"

	mbtest "packetbeat/metricbeat/mb/testing"
	"packetbeat/metricbeat/module/couchbase"
)

func TestData(t *testing.T) {
	f := mbtest.NewEventsFetcher(t, getConfig())

	err := mbtest.WriteEvents(f, t)
	if err != nil {
		t.Fatal("write", err)
	}
}

func getConfig() map[string]interface{} {
	return map[string]interface{}{
		"module":     "couchbase",
		"metricsets": []string{"bucket"},
		"hosts":      []string{couchbase.GetEnvDSN()},
	}
}
