package cluster_disk

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
	assert.NoError(t, err)

	response, err := ioutil.ReadFile(absPath + "/df_sample_response.json")
	assert.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "appication/json;")
		w.Write([]byte(response))
	}))
	defer server.Close()

	config := map[string]interface{}{
		"module":     "ceph",
		"metricsets": []string{"cluster_disk"},
		"hosts":      []string{server.URL},
	}

	f := mbtest.NewEventFetcher(t, config)
	event, err := f.Fetch()

	t.Logf("%s/%s event: %+v", f.Module().Name(), f.Name(), event.StringToPrint())

	used := event["used"].(common.MapStr)
	assert.EqualValues(t, 1428520960, used["bytes"])

	total := event["total"].(common.MapStr)
	assert.EqualValues(t, 6431965184, total["bytes"])

	available := event["available"].(common.MapStr)
	assert.EqualValues(t, 5003444224, available["bytes"])
}
