package memory

import (
	"packetbeat/libbeat/common"
	"packetbeat/libbeat/logp"
	"packetbeat/metricbeat/mb"
	"packetbeat/metricbeat/module/docker"

	dc "github.com/fsouza/go-dockerclient"
)

func init() {
	if err := mb.Registry.AddMetricSet("docker", "memory", New, docker.HostParser); err != nil {
		panic(err)
	}
}

type MetricSet struct {
	mb.BaseMetricSet
	memoryService *MemoryService
	dockerClient  *dc.Client
}

// New creates a new instance of the docker memory MetricSet.
func New(base mb.BaseMetricSet) (mb.MetricSet, error) {
	logp.Warn("BETA: The docker memory metricset is beta")

	config := docker.Config{}
	if err := base.Module().UnpackConfig(&config); err != nil {
		return nil, err
	}

	client, err := docker.NewDockerClient(base.HostData().URI, config)
	if err != nil {
		return nil, err
	}

	return &MetricSet{
		BaseMetricSet: base,
		memoryService: &MemoryService{},
		dockerClient:  client,
	}, nil
}

// Fetch creates a list of memory events for each container.
func (m *MetricSet) Fetch() ([]common.MapStr, error) {
	stats, err := docker.FetchStats(m.dockerClient, m.Module().Config().Timeout)
	if err != nil {
		return nil, err
	}

	memoryStats := m.memoryService.getMemoryStatsList(stats)
	return eventsMapping(memoryStats), nil
}
