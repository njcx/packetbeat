package node

import (
	"packetbeat/libbeat/common"
	"packetbeat/libbeat/logp"
	"packetbeat/metricbeat/helper"
	"packetbeat/metricbeat/mb"
	"packetbeat/metricbeat/mb/parse"
)

const (
	defaultScheme = "http"
	defaultPath   = "/pools/default"
)

var (
	hostParser = parse.URLHostParserBuilder{
		DefaultScheme: defaultScheme,
		DefaultPath:   defaultPath,
	}.Build()
)

// init registers the MetricSet with the central registry.
// The New method will be called after the setup of the module and before starting to fetch data
func init() {
	if err := mb.Registry.AddMetricSet("couchbase", "node", New, hostParser); err != nil {
		panic(err)
	}
}

// MetricSet type defines all fields of the MetricSet
type MetricSet struct {
	mb.BaseMetricSet
	http *helper.HTTP
}

// New create a new instance of the MetricSet
func New(base mb.BaseMetricSet) (mb.MetricSet, error) {
	logp.Warn("BETA: The couchbase node metricset is beta")

	return &MetricSet{
		BaseMetricSet: base,
		http:          helper.NewHTTP(base),
	}, nil
}

// Fetch methods implements the data gathering and data conversion to the right format
// It returns the event which is then forward to the output. In case of an error, a
// descriptive error must be returned.
func (m *MetricSet) Fetch() ([]common.MapStr, error) {
	content, err := m.http.FetchContent()
	if err != nil {
		return nil, err
	}

	return eventsMapping(content), nil
}
