package ssh

import (
	"packetbeat/libbeat/common"

	"packetbeat/publish"
)

// Transaction Publisher.
type transPub struct {
	sendRequest  bool
	sendResponse bool

	results publish.Transactions
}

func (pub *transPub) onTransaction(requ, resp *message) error {
	if pub.results == nil {
		return nil
	}

	event := pub.createEvent(requ, resp)
	pub.results.PublishTransaction(event)
	return nil
}

func (pub *transPub) createEvent(requ, resp *message) common.MapStr {
	status := common.OK_STATUS

	// resp_time in milliseconds
	responseTime := int32(resp.Ts.Sub(requ.Ts).Nanoseconds() / 1e6)

	src := &common.Endpoint{
		IP:   requ.Tuple.SrcIP.String(),
		Port: requ.Tuple.SrcPort,
		Proc: string(requ.CmdlineTuple.Src),
	}
	dst := &common.Endpoint{
		IP:   requ.Tuple.DstIP.String(),
		Port: requ.Tuple.DstPort,
		Proc: string(requ.CmdlineTuple.Dst),
	}

	event := common.MapStr{
		"@timestamp":   common.Time(requ.Ts),
		"type":         "ssh",
		"status":       status,
		"responsetime": responseTime,
		"bytes_in":     requ.Size,
		"bytes_out":    resp.Size,
		"src":          src,
		"dst":          dst,
	}

	// add processing notes/errors to event
	if len(requ.Notes)+len(resp.Notes) > 0 {
		event["notes"] = append(requ.Notes, resp.Notes...)
	}

	if pub.sendRequest {
		// event["request"] =
	}
	if pub.sendResponse {
		// event["response"] =
	}

	return event
}
