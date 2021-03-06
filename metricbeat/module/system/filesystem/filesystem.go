// +build darwin freebsd linux openbsd windows

package filesystem

import (
	"packetbeat/libbeat/common"
	"packetbeat/libbeat/logp"
	"packetbeat/metricbeat/mb"
	"packetbeat/metricbeat/mb/parse"

	"github.com/pkg/errors"
)

var debugf = logp.MakeDebug("system.filesystem")

func init() {
	if err := mb.Registry.AddMetricSet("system", "filesystem", New, parse.EmptyHostParser); err != nil {
		panic(err)
	}
}

// MetricSet for fetching filesystem metrics.
type MetricSet struct {
	mb.BaseMetricSet
	config Config
}

// New creates and returns a new instance of MetricSet.
func New(base mb.BaseMetricSet) (mb.MetricSet, error) {
	var config Config
	if err := base.Module().UnpackConfig(&config); err != nil {
		return nil, err
	}

	return &MetricSet{
		BaseMetricSet: base,
		config:        config,
	}, nil
}

// Fetch fetches filesystem metrics for all mounted filesystems and returns
// an event for each mount point.
func (m *MetricSet) Fetch() ([]common.MapStr, error) {
	fss, err := GetFileSystemList()
	if err != nil {
		return nil, errors.Wrap(err, "filesystem list")
	}

	if len(m.config.IgnoreTypes) > 0 {
		fss = Filter(fss, BuildTypeFilter(m.config.IgnoreTypes...))
	}

	filesSystems := make([]common.MapStr, 0, len(fss))
	for _, fs := range fss {
		fsStat, err := GetFileSystemStat(fs)
		if err != nil {
			debugf("error getting filesystem stats for '%s': %v", fs.DirName, err)
			continue
		}
		AddFileSystemUsedPercentage(fsStat)
		filesSystems = append(filesSystems, GetFilesystemEvent(fsStat))
	}

	return filesSystems, nil
}
