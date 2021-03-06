package healthcheck

import (
	"strings"

	dc "github.com/fsouza/go-dockerclient"

	"packetbeat/libbeat/common"
	"packetbeat/libbeat/logp"
	"packetbeat/metricbeat/mb"
	"packetbeat/metricbeat/module/docker"
)

func eventsMapping(containers []dc.APIContainers, m *MetricSet) []common.MapStr {
	var events []common.MapStr
	for _, container := range containers {
		event := eventMapping(&container, m)
		if event != nil {
			events = append(events, event)
		}
	}
	return events
}

func eventMapping(cont *dc.APIContainers, m *MetricSet) common.MapStr {
	if !hasHealthCheck(cont.Status) {
		return nil
	}

	container, err := m.dockerClient.InspectContainer(cont.ID)
	if err != nil {
		logp.Err("Error inpsecting container %v: %v", cont.ID, err)
		return nil
	}
	lastEvent := len(container.State.Health.Log) - 1

	// Checks if a healthcheck already happened
	if lastEvent < 0 {
		return nil
	}

	return common.MapStr{
		mb.ModuleData: common.MapStr{
			"container": docker.NewContainer(cont).ToMapStr(),
		},
		"status":        container.State.Health.Status,
		"failingstreak": container.State.Health.FailingStreak,
		"event": common.MapStr{
			"start_date": common.Time(container.State.Health.Log[lastEvent].Start),
			"end_date":   common.Time(container.State.Health.Log[lastEvent].End),
			"exit_code":  container.State.Health.Log[lastEvent].ExitCode,
			"output":     container.State.Health.Log[lastEvent].Output,
		},
	}
}

// hasHealthCheck detects if healthcheck is available for container
func hasHealthCheck(status string) bool {
	return strings.Contains(status, "(") && strings.Contains(status, ")")
}
