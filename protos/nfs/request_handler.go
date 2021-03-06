package nfs

// This file contains methods process RPC calls

import (
	"expvar"
	"fmt"
	"time"

	"packetbeat/libbeat/common"
	"packetbeat/protos/tcp"
)

const nfsProgramNumber = 100003

var acceptStatus = [...]string{
	"success",
	"prog_unavail",
	"prog_mismatch",
	"proc_unavail",
	"garbage_args",
	"system_err",
}

var (
	unmatchedRequests = expvar.NewInt("nfs.unmatched_requests")
)

// called by Cache, when re reply seen within expected time window
func (r *rpc) handleExpiredPacket(nfs *nfs) {
	nfs.event["status"] = "NO_REPLY"
	r.results.PublishTransaction(nfs.event)
	unmatchedRequests.Add(1)
}

// called when we process a RPC call
func (r *rpc) handleCall(xid string, xdr *xdr, ts time.Time, tcptuple *common.TCPTuple, dir uint8) {

	// eat rpc version number
	xdr.getUInt()
	rpcProg := xdr.getUInt()
	if rpcProg != nfsProgramNumber {
		// not a NFS request
		return
	}

	src := common.Endpoint{
		IP:   tcptuple.SrcIP.String(),
		Port: tcptuple.SrcPort,
	}
	dst := common.Endpoint{
		IP:   tcptuple.DstIP.String(),
		Port: tcptuple.DstPort,
	}

	// The direction of the stream is based in the direction of first packet seen.
	// if we have stored stream in reverse order, swap src and dst
	if dir == tcp.TCPDirectionReverse {
		src, dst = dst, src
	}

	event := common.MapStr{}
	event["@timestamp"] = common.Time(ts)
	event["status"] = common.OK_STATUS // all packages are OK for now
	event["src"] = &src
	event["dst"] = &dst

	nfsVers := xdr.getUInt()
	nfsProc := xdr.getUInt()

	// build event only if it's a nfs packet
	rpcInfo := common.MapStr{}
	rpcInfo["xid"] = xid
	rpcInfo["call_size"] = xdr.size()

	authFlavor := xdr.getUInt()
	authOpaque := xdr.getDynamicOpaque()
	switch authFlavor {
	case 0:
		rpcInfo["auth_flavor"] = "none"
	case 1:
		rpcInfo["auth_flavor"] = "unix"
		cred := common.MapStr{}
		credXdr := makeXDR(authOpaque)
		cred["stamp"] = credXdr.getUInt()
		machine := credXdr.getString()
		if machine == "" {
			machine = src.IP
		}
		cred["machinename"] = machine
		cred["uid"] = credXdr.getUInt()
		cred["gid"] = credXdr.getUInt()
		cred["gids"] = credXdr.getUIntVector()
		rpcInfo["cred"] = cred
	case 6:
		rpcInfo["auth_flavor"] = "rpcsec_gss"
	default:
		rpcInfo["auth_flavor"] = fmt.Sprintf("unknown (%d)", authFlavor)
	}

	// eat auth verifier
	xdr.getUInt()
	xdr.getDynamicOpaque()

	event["type"] = "nfs"
	event["rpc"] = rpcInfo
	nfs := nfs{vers: nfsVers, proc: nfsProc, event: event}
	event["nfs"] = nfs.getRequestInfo(xdr)

	// use xid+src ip to uniquely identify request
	reqID := xid + tcptuple.SrcIP.String()

	// populate cache to trace request reply
	r.callsSeen.Put(reqID, &nfs)
}

// called when we process a RPC reply
func (r *rpc) handleReply(xid string, xdr *xdr, ts time.Time, tcptuple *common.TCPTuple, dir uint8) {
	replyStatus := xdr.getUInt()
	// we are interested only in accepted rpc reply
	if replyStatus != 0 {
		return
	}

	// eat auth verifier
	xdr.getUInt()
	xdr.getDynamicOpaque()

	// xid+src ip is used to uniquely identify request.
	var reqID string
	if dir == tcp.TCPDirectionReverse {
		// stream in correct order: Src points to a client
		reqID = xid + tcptuple.SrcIP.String()
	} else {
		// stream in reverse order: Dst points to a client
		reqID = xid + tcptuple.DstIP.String()
	}

	// get cached request
	v := r.callsSeen.Delete(reqID)
	if v != nil {
		nfs := v.(*nfs)
		event := nfs.event
		rpcInfo := event["rpc"].(common.MapStr)
		rpcInfo["reply_size"] = xdr.size()
		rpcTime := ts.Sub(time.Time(event["@timestamp"].(common.Time)))
		rpcInfo["time"] = rpcTime
		// the same in human readable form
		rpcInfo["time_str"] = fmt.Sprintf("%v", rpcTime)
		status := int(xdr.getUInt())
		rpcInfo["status"] = acceptStatus[status]

		// populate nfs info for successfully executed requests
		if status == 0 {
			nfsInfo := event["nfs"].(common.MapStr)
			nfsInfo["status"] = nfs.getNFSReplyStatus(xdr)
		}
		r.results.PublishTransaction(event)
	}
}
