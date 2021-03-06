// +build integration

package status

import (
	"testing"

	"packetbeat/libbeat/common"
	mbtest "packetbeat/metricbeat/mb/testing"
	"packetbeat/metricbeat/module/mysql"

	"github.com/stretchr/testify/assert"
)

func TestFetch(t *testing.T) {
	f := mbtest.NewEventFetcher(t, getConfig(false))
	event, err := f.Fetch()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	t.Logf("%s/%s event: %+v", f.Module().Name(), f.Name(), event)

	// Check event fields
	connections := event["connections"].(int64)
	open := event["open"].(common.MapStr)
	openTables := open["tables"].(int64)
	openFiles := open["files"].(int64)
	openStreams := open["streams"].(int64)

	assert.True(t, connections > 0)
	assert.True(t, openTables > 0)
	assert.True(t, openFiles >= 0)
	assert.True(t, openStreams == 0)
}

func TestFetchRaw(t *testing.T) {
	f := mbtest.NewEventFetcher(t, getConfig(true))
	event, err := f.Fetch()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	t.Logf("%s/%s event: %+v", f.Module().Name(), f.Name(), event)

	// Check event fields
	cachedThreads := event["threads"].(common.MapStr)["cached"].(int64)
	assert.True(t, cachedThreads >= 0)

	rawData := event["raw"].(common.MapStr)

	// Make sure field was removed from raw fields as in schema
	_, exists := rawData["Threads_cached"]
	assert.False(t, exists)

	// Check a raw field if it is available
	_, exists = rawData["Slow_launch_threads"]
	assert.True(t, exists)
}

func TestData(t *testing.T) {
	f := mbtest.NewEventFetcher(t, getConfig(false))

	err := mbtest.WriteEvent(f, t)
	if err != nil {
		t.Fatal("write", err)
	}
}

func getConfig(raw bool) map[string]interface{} {
	return map[string]interface{}{
		"module":     "mysql",
		"metricsets": []string{"status"},
		"hosts":      []string{mysql.GetMySQLEnvDSN()},
		"raw":        raw,
	}
}
