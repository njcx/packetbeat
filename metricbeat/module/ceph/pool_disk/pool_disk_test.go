package pool_disk

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"packetbeat/libbeat/common"
	mbtest "packetbeat/metricbeat/mb/testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchEventContents(t *testing.T) {
	absPath, err := filepath.Abs("../_meta/testdata/")

	response, err := ioutil.ReadFile(absPath + "/df_sample_response.json")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "appication/json;")
		w.Write([]byte(response))
	}))
	defer server.Close()

	config := map[string]interface{}{
		"module":     "ceph",
		"metricsets": []string{"pool_disk"},
		"hosts":      []string{server.URL},
	}

	f := mbtest.NewEventsFetcher(t, config)
	events, err := f.Fetch()
	event := events[0]
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	t.Logf("%s/%s event: %+v", f.Module().Name(), f.Name(), event.StringToPrint())

	assert.EqualValues(t, "rbd", event["name"])
	assert.EqualValues(t, 0, event["id"])

	stats := event["stats"].(common.MapStr)

	used := stats["used"].(common.MapStr)
	assert.EqualValues(t, 0, used["bytes"])
	assert.EqualValues(t, 0, used["kb"])

	available := stats["available"].(common.MapStr)
	assert.EqualValues(t, 5003444224, available["bytes"])

}
