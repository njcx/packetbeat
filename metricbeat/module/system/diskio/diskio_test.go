// +build !integration
// +build freebsd linux windows

package diskio

import (
	"testing"

	"time"

	mbtest "packetbeat/metricbeat/mb/testing"
)

func TestData(t *testing.T) {
	f := mbtest.NewEventsFetcher(t, getConfig())

	// Do a first fetch to have percentages
	f.Fetch()
	time.Sleep(1 * time.Second)

	err := mbtest.WriteEvents(f, t)
	if err != nil {
		t.Fatal("write", err)
	}
}

func getConfig() map[string]interface{} {
	return map[string]interface{}{
		"module":     "system",
		"metricsets": []string{"diskio"},
	}
}
