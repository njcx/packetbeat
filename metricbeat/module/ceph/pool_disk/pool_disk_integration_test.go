package pool_disk

import (
	"fmt"
	mbtest "packetbeat/metricbeat/mb/testing"
	"os"
	"testing"
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
		"module":     "ceph",
		"metricsets": []string{"pool_disk"},
		"hosts":      getTestCephHost(),
	}
}

const (
	cephDefaultHost = "127.0.0.1"
	cephDefaultPort = "5000"
)

func getTestCephHost() string {
	return fmt.Sprintf("%v:%v",
		getenv("CEPH_HOST", cephDefaultHost),
		getenv("CEPH_PORT", cephDefaultPort),
	)
}

func getenv(name, defaultValue string) string {
	return strDefault(os.Getenv(name), defaultValue)
}

func strDefault(a, defaults string) string {
	if len(a) == 0 {
		return defaults
	}
	return a
}
