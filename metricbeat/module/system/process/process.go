// +build darwin freebsd linux windows

package process

import (
	"fmt"
	"runtime"

	"packetbeat/libbeat/common"
	"packetbeat/libbeat/logp"
	"packetbeat/metricbeat/mb"
	"packetbeat/metricbeat/mb/parse"
	"packetbeat/metricbeat/module/system"

	"github.com/elastic/gosigar/cgroup"
	"github.com/pkg/errors"
)

var debugf = logp.MakeDebug("system-process")

func init() {
	if err := mb.Registry.AddMetricSet("system", "process", New, parse.EmptyHostParser); err != nil {
		panic(err)
	}
}

// MetricSet that fetches process metrics.
type MetricSet struct {
	mb.BaseMetricSet
	stats  *ProcStats
	cgroup *cgroup.Reader
}

// New creates and returns a new MetricSet.
func New(base mb.BaseMetricSet) (mb.MetricSet, error) {
	config := struct {
		Procs        []string `config:"processes"`
		Cgroups      *bool    `config:"process.cgroups.enabled"`
		EnvWhitelist []string `config:"process.env.whitelist"`
		CPUTicks     bool     `config:"cpu_ticks"`
	}{
		Procs: []string{".*"}, // collect all processes by default
	}
	if err := base.Module().UnpackConfig(&config); err != nil {
		return nil, err
	}

	m := &MetricSet{
		BaseMetricSet: base,
		stats: &ProcStats{
			Procs:        config.Procs,
			EnvWhitelist: config.EnvWhitelist,
			CpuTicks:     config.CPUTicks,
		},
	}
	err := m.stats.InitProcStats()
	if err != nil {
		return nil, err
	}

	if runtime.GOOS == "linux" {
		systemModule, ok := base.Module().(*system.Module)
		if !ok {
			return nil, fmt.Errorf("unexpected module type")
		}

		if config.Cgroups == nil || *config.Cgroups {
			debugf("process cgroup data collection is enabled, using hostfs='%v'", systemModule.HostFS)
			m.cgroup, err = cgroup.NewReader(systemModule.HostFS, true)
			if err != nil {
				if err == cgroup.ErrCgroupsMissing {
					logp.Warn("cgroup data collection will be disabled: %v", err)
				} else {
					return nil, errors.Wrap(err, "error initializing cgroup reader")
				}
			}
		}
	}

	return m, nil
}

// Fetch fetches metrics for all processes. It iterates over each PID and
// collects process metadata, CPU metrics, and memory metrics.
func (m *MetricSet) Fetch() ([]common.MapStr, error) {
	procs, err := m.stats.GetProcStats()
	if err != nil {
		return nil, errors.Wrap(err, "process stats")
	}

	if m.cgroup != nil {
		for _, proc := range procs {
			pid, ok := proc["pid"].(int)
			if !ok {
				debugf("error converting pid to int for proc %+v", proc)
				continue
			}
			stats, err := m.cgroup.GetStatsForProcess(pid)
			if err != nil {
				debugf("error getting cgroups stats for pid=%d, %v", pid, err)
				continue
			}

			if statsMap := cgroupStatsToMap(stats); statsMap != nil {
				proc["cgroup"] = statsMap
			}
		}
	}

	return procs, err
}
