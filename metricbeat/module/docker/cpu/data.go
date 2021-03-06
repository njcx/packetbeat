package cpu

import (
	"packetbeat/libbeat/common"
	"packetbeat/metricbeat/mb"
)

func eventsMapping(cpuStatsList []CPUStats) []common.MapStr {
	events := []common.MapStr{}
	for _, cpuStats := range cpuStatsList {
		events = append(events, eventMapping(&cpuStats))
	}
	return events
}

func eventMapping(stats *CPUStats) common.MapStr {
	event := common.MapStr{
		mb.ModuleData: common.MapStr{
			"container": stats.Container.ToMapStr(),
		},
		"core": stats.PerCpuUsage,
		"total": common.MapStr{
			"pct": stats.TotalUsage,
		},
		"kernel": common.MapStr{
			"ticks": stats.UsageInKernelmode,
			"pct":   stats.UsageInKernelmodePercentage,
		},
		"user": common.MapStr{
			"ticks": stats.UsageInUsermode,
			"pct":   stats.UsageInUsermodePercentage,
		},
		"system": common.MapStr{
			"ticks": stats.SystemUsage,
			"pct":   stats.SystemUsagePercentage,
		},
	}

	return event
}
