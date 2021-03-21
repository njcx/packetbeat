package lb

import (
	"testing"

	"packetbeat/libbeat/common"
	"packetbeat/libbeat/logp"
	"packetbeat/libbeat/outputs"
)

var (
	testNoOpts     = outputs.Options{}
	testGuaranteed = outputs.Options{Guaranteed: true}

	testEvent = common.MapStr{
		"msg": "hello world",
	}
)

func enableLogging(selectors []string) {
	if testing.Verbose() {
		logp.LogInit(logp.LOG_DEBUG, "", false, true, selectors)
	}
}
