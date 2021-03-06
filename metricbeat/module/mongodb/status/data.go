package status

import (
	s "packetbeat/metricbeat/schema"
	c "packetbeat/metricbeat/schema/mapstriface"
)

var schema = s.Schema{
	"version": c.Str("version"),
	"uptime": s.Object{
		"ms": c.Int("uptimeMillis"),
	},
	"local_time":         c.Time("localTime"),
	"write_backs_queued": c.Bool("writeBacksQueued", s.Optional),
	"asserts": c.Dict("asserts", s.Schema{
		"regular":   c.Int("regular"),
		"warning":   c.Int("warning"),
		"msg":       c.Int("msg"),
		"user":      c.Int("user"),
		"rollovers": c.Int("rollovers"),
	}),
	"background_flushing": c.Dict("backgroundFlushing", s.Schema{
		"flushes": c.Int("flushes"),
		"total": s.Object{
			"ms": c.Int("total_ms"),
		},
		"average": s.Object{
			"ms": c.Int("average_ms"),
		},
		"last": s.Object{
			"ms": c.Int("last_ms"),
		},
		"last_finished": c.Time("last_finished"),
	}, c.DictOptional),
	"connections": c.Dict("connections", s.Schema{
		"current":       c.Int("current"),
		"available":     c.Int("available"),
		"total_created": c.Int("totalCreated"),
	}),
	"journaling": c.Dict("dur", s.Schema{
		"commits": c.Int("commits"),
		"journaled": s.Object{
			"mb": c.Int("journaledMB"),
		},
		"write_to_data_files": s.Object{
			"mb": c.Int("writeToDataFilesMB"),
		},
		"compression":           c.Int("compression"),
		"commits_in_write_lock": c.Int("commitsInWriteLock"),
		"early_commits":         c.Int("earlyCommits"),
		"times": c.Dict("timeMs", s.Schema{
			"dt":                    s.Object{"ms": c.Int("dt")},
			"prep_log_buffer":       s.Object{"ms": c.Int("prepLogBuffer")},
			"write_to_journal":      s.Object{"ms": c.Int("writeToJournal")},
			"write_to_data_files":   s.Object{"ms": c.Int("writeToDataFiles")},
			"remap_private_view":    s.Object{"ms": c.Int("remapPrivateView")},
			"commits":               s.Object{"ms": c.Int("commits")},
			"commits_in_write_lock": s.Object{"ms": c.Int("commitsInWriteLock")},
		}),
	}, c.DictOptional),
	"extra_info": c.Dict("extra_info", s.Schema{
		"heap_usage":  s.Object{"bytes": c.Int("heap_usage_bytes", s.Optional)},
		"page_faults": c.Int("page_faults"),
	}),
	"network": c.Dict("network", s.Schema{
		"in":       s.Object{"bytes": c.Int("bytesIn")},
		"out":      s.Object{"bytes": c.Int("bytesOut")},
		"requests": c.Int("numRequests"),
	}),
	"memory": c.Dict("mem", s.Schema{
		"bits":                c.Int("bits"),
		"resident":            s.Object{"mb": c.Int("resident")},
		"virtual":             s.Object{"mb": c.Int("virtual")},
		"mapped":              s.Object{"mb": c.Int("mapped")},
		"mapped_with_journal": s.Object{"mb": c.Int("mappedWithJournal")},
	}),
	"opcounters": c.Dict("opcounters", s.Schema{
		"insert":  c.Int("insert"),
		"query":   c.Int("query"),
		"update":  c.Int("update"),
		"delete":  c.Int("delete"),
		"getmore": c.Int("getmore"),
		"command": c.Int("command"),
	}),
	"opcounters_replicated": c.Dict("opcountersRepl", s.Schema{
		"insert":  c.Int("insert"),
		"query":   c.Int("query"),
		"update":  c.Int("update"),
		"delete":  c.Int("delete"),
		"getmore": c.Int("getmore"),
		"command": c.Int("command"),
	}),
	"storage_engine": c.Dict("storageEngine", s.Schema{
		"name": c.Str("name"),
	}),
	"wired_tiger": c.Dict("wiredTiger", wiredTigerSchema, c.DictOptional),
}

var wiredTigerSchema = s.Schema{
	"concurrent_transactions": c.Dict("concurrentTransactions", s.Schema{
		"write": c.Dict("write", s.Schema{
			"out":           c.Int("out"),
			"available":     c.Int("available"),
			"total_tickets": c.Int("totalTickets"),
		}),
		"read": c.Dict("write", s.Schema{
			"out":           c.Int("out"),
			"available":     c.Int("available"),
			"total_tickets": c.Int("totalTickets"),
		}),
	}),
	"cache": c.Dict("cache", s.Schema{
		"maximum": s.Object{"bytes": c.Int("maximum bytes configured")},
		"used":    s.Object{"bytes": c.Int("bytes currently in the cache")},
		"dirty":   s.Object{"bytes": c.Int("tracked dirty bytes in the cache")},
		"pages": s.Object{
			"read":    c.Int("pages read into cache"),
			"write":   c.Int("pages written from cache"),
			"evicted": c.Int("unmodified pages evicted"),
		},
	}),
	"log": c.Dict("log", s.Schema{
		"size":          s.Object{"bytes": c.Int("total log buffer size")},
		"write":         s.Object{"bytes": c.Int("log bytes written")},
		"max_file_size": s.Object{"bytes": c.Int("maximum log file size")},
		"flushes":       c.Int("log flush operations"),
		"writes":        c.Int("log write operations"),
		"scans":         c.Int("log scan operations"),
		"syncs":         c.Int("log sync operations"),
	}),
}

var eventMapping = schema.Apply
