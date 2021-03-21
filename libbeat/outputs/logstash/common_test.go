package logstash

import (
	"sync"
	"testing"

	"packetbeat/libbeat/logp"
)

var enableLoggingOnce sync.Once

func enableLogging(selectors []string) {
	if testing.Verbose() {
		enableLoggingOnce.Do(func() {
			logp.LogInit(logp.LOG_DEBUG, "", false, true, selectors)
		})
	}
}
