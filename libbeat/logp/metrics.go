package logp

import (
	"bytes"
	"fmt"
	"sort"
	"time"

	"packetbeat/libbeat/monitoring"
)

// logMetrics logs at Info level the integer expvars that have changed in the
// last interval. For each expvar, the delta from the beginning of the interval
// is logged.
func logMetrics(metricsCfg *LoggingMetricsConfig) {
	if metricsCfg.Enabled != nil && *metricsCfg.Enabled == false {
		Info("Metrics logging disabled")
		return
	}
	if metricsCfg.Period == nil {
		metricsCfg.Period = &defaultMetricsPeriod
	}
	Info("Metrics logging every %s", metricsCfg.Period)

	ticker := time.NewTicker(*metricsCfg.Period)

	prevVals := monitoring.MakeFlatSnapshot()
	for range ticker.C {
		snapshot := snapshotMetrics()
		delta := snapshotDelta(prevVals, snapshot)
		prevVals = snapshot

		if len(delta) == 0 {
			Info("No non-zero metrics in the last %s", metricsCfg.Period)
			continue
		}

		metrics := formatMetrics(delta)
		Info("Non-zero metrics in the last %s:%s", metricsCfg.Period, metrics)
	}
}

// LogTotalExpvars logs all registered expvar metrics.
func LogTotalExpvars(cfg *Logging) {
	if cfg.Metrics.Enabled != nil && *cfg.Metrics.Enabled == false {
		return
	}

	zero := monitoring.MakeFlatSnapshot()
	metrics := formatMetrics(snapshotDelta(zero, snapshotMetrics()))
	Info("Total non-zero values: %s", metrics)
	Info("Uptime: %s", time.Now().Sub(startTime))
}

func snapshotMetrics() monitoring.FlatSnapshot {
	return monitoring.CollectFlatSnapshot(monitoring.Default, monitoring.Full, true)
}

func snapshotDelta(prev, cur monitoring.FlatSnapshot) map[string]interface{} {
	out := map[string]interface{}{}

	for k, b := range cur.Bools {
		if p, ok := prev.Bools[k]; !ok || p != b {
			out[k] = b
		}
	}

	for k, s := range cur.Strings {
		if p, ok := prev.Strings[k]; !ok || p != s {
			out[k] = s
		}
	}

	for k, i := range cur.Ints {
		if p := prev.Ints[k]; p != i {
			out[k] = i - p
		}
	}

	for k, f := range cur.Floats {
		if p := prev.Floats[k]; p != f {
			out[k] = f - p
		}
	}

	return out
}

func formatMetrics(ms map[string]interface{}) string {
	keys := make([]string, 0, len(ms))
	for key := range ms {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	var buf bytes.Buffer
	for _, key := range keys {
		buf.WriteByte(' ')
		buf.WriteString(key)
		buf.WriteString("=")
		buf.WriteString(fmt.Sprintf("%v", ms[key]))
	}
	return buf.String()
}
