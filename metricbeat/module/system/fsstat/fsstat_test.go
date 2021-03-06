// +build darwin freebsd linux openbsd windows

package fsstat

import (
	"testing"

	"time"

	mbtest "packetbeat/metricbeat/mb/testing"
)

func TestData(t *testing.T) {
	f := mbtest.NewEventFetcher(t, getConfig())

	// Do a first fetch to have percentages
	f.Fetch()
	time.Sleep(1 * time.Second)

	err := mbtest.WriteEvent(f, t)
	if err != nil {
		t.Fatal("write", err)
	}
}

func getConfig() map[string]interface{} {
	return map[string]interface{}{
		"module":     "system",
		"metricsets": []string{"fsstat"},
	}
}
