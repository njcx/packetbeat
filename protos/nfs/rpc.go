// Package rpc provides support for parsing RPC messages and reporting the
// results. This package supports the RPC v2 protocol as defined by RFC 5531
// (RFC 1831).

package nfs

import (
	"encoding/binary"
	"time"

	"packetbeat/libbeat/common"
	"packetbeat/libbeat/logp"

	"fmt"

	"packetbeat/protos"
	"packetbeat/protos/tcp"
	"packetbeat/publish"
)

var debugf = logp.MakeDebug("rpc")

const (
	rpcLastFrag = 0x80000000
	rpcSizeMask = 0x7fffffff
)

const (
	rpcCall  = 0
	rpcReply = 1
)

type rpcStream struct {
	tcpTuple *common.TCPTuple
	rawData  []byte
}

type rpcConnectionData struct {
	streams [2]*rpcStream
}

type rpc struct {
	// Configuration data.
	ports              []int
	callsSeen          *common.Cache
	transactionTimeout time.Duration

	results publish.Transactions // Channel where results are pushed.
}

func init() {
	protos.Register("nfs", New)
}

func New(
	testMode bool,
	results publish.Transactions,
	cfg *common.Config,
) (protos.Plugin, error) {
	p := &rpc{}
	config := defaultConfig
	if !testMode {
		if err := cfg.Unpack(&config); err != nil {
			logp.Warn("failed to read config")
			return nil, err
		}
	}

	if err := p.init(results, &config); err != nil {
		logp.Warn("failed to init")
		return nil, err
	}
	return p, nil
}

func (r *rpc) init(results publish.Transactions, config *rpcConfig) error {
	r.setFromConfig(config)
	r.results = results
	r.callsSeen = common.NewCacheWithRemovalListener(
		r.transactionTimeout,
		protos.DefaultTransactionHashSize,
		func(k common.Key, v common.Value) {
			nfs, ok := v.(*nfs)
			if !ok {
				logp.Err("Expired value is not a MapStr (%T).", v)
				return
			}
			r.handleExpiredPacket(nfs)
		})

	r.callsSeen.StartJanitor(r.transactionTimeout)
	return nil
}

func (r *rpc) setFromConfig(config *rpcConfig) error {
	r.ports = config.Ports
	r.transactionTimeout = config.TransactionTimeout
	return nil
}

func (r *rpc) GetPorts() []int {
	return r.ports
}

// Called when TCP payload data is available for parsing.
func (r *rpc) Parse(
	pkt *protos.Packet,
	tcptuple *common.TCPTuple,
	dir uint8,
	private protos.ProtocolData,
) protos.ProtocolData {

	defer logp.Recover("ParseRPC exception")

	conn := ensureRPCConnection(private)

	conn = r.handleRPCFragment(conn, pkt, tcptuple, dir)
	if conn == nil {
		return nil
	}
	return conn
}

// Called when the FIN flag is seen in the TCP stream.
func (r *rpc) ReceivedFin(tcptuple *common.TCPTuple, dir uint8,
	private protos.ProtocolData) protos.ProtocolData {

	defer logp.Recover("ReceivedFinRpc exception")

	// forced by TCP interface
	return private
}

// Called when a packets are missing from the tcp
// stream.
func (r *rpc) GapInStream(tcptuple *common.TCPTuple, dir uint8,
	nbytes int, private protos.ProtocolData) (priv protos.ProtocolData, drop bool) {

	defer logp.Recover("GapInRpcStream exception")

	// forced by TCP interface
	return private, false
}

// ConnectionTimeout returns the per stream connection timeout.
// Return <=0 to set default tcp module transaction timeout.
func (r *rpc) ConnectionTimeout() time.Duration {
	// forced by TCP interface
	return r.transactionTimeout
}

func ensureRPCConnection(private protos.ProtocolData) *rpcConnectionData {
	conn := getRPCConnection(private)
	if conn == nil {
		conn = &rpcConnectionData{}
	}
	return conn
}

func getRPCConnection(private protos.ProtocolData) *rpcConnectionData {
	if private == nil {
		return nil
	}

	priv, ok := private.(*rpcConnectionData)
	if !ok {
		logp.Warn("rpc connection data type error")
		return nil
	}
	if priv == nil {
		logp.Warn("Unexpected: rpc connection data not set")
		return nil
	}

	return priv
}

// Parse function is used to process TCP payloads.
func (r *rpc) handleRPCFragment(
	conn *rpcConnectionData,
	pkt *protos.Packet,
	tcptuple *common.TCPTuple,
	dir uint8,
) *rpcConnectionData {

	st := conn.streams[dir]
	if st == nil {
		st = newStream(pkt, tcptuple)
		conn.streams[dir] = st
	} else {
		// concatenate bytes
		st.rawData = append(st.rawData, pkt.Payload...)
		if len(st.rawData) > tcp.TCPMaxDataInStream {
			debugf("Stream data too large, dropping TCP stream")
			conn.streams[dir] = nil
			return conn
		}
	}

	for len(st.rawData) > 0 {

		if len(st.rawData) < 4 {
			debugf("Wainting for more data")
			break
		}

		marker := uint32(binary.BigEndian.Uint32(st.rawData[0:4]))
		size := int(marker & rpcSizeMask)
		islast := (marker & rpcLastFrag) != 0

		if len(st.rawData)-4 < size {
			debugf("Wainting for more data")
			break
		}

		if !islast {
			logp.Warn("multifragment rpc message")
			break
		}

		xdr := newXDR(st.rawData[4 : 4+size])
		// keep the rest of the next fragment
		st.rawData = st.rawData[4+size:]

		r.handleRPCPacket(xdr, pkt.Ts, tcptuple, dir)
	}

	return conn
}

func (r *rpc) handleRPCPacket(xdr *xdr, ts time.Time, tcptuple *common.TCPTuple, dir uint8) {

	xid := fmt.Sprintf("%.8x", xdr.getUInt())

	msgType := xdr.getUInt()

	switch msgType {
	case rpcCall:
		r.handleCall(xid, xdr, ts, tcptuple, dir)
	case rpcReply:
		r.handleReply(xid, xdr, ts, tcptuple, dir)
	default:
		logp.Warn("Bad RPC message")
	}
}

func newStream(pkt *protos.Packet, tcptuple *common.TCPTuple) *rpcStream {
	return &rpcStream{
		tcpTuple: tcptuple,
		rawData:  pkt.Payload,
	}
}
