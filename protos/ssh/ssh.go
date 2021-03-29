package ssh

import (
	"time"

	"packetbeat/libbeat/common"
	"packetbeat/libbeat/logp"
	"packetbeat/protos/tcp"

	"packetbeat/protos"
	"packetbeat/publish"
)

// sshPlugin application level protocol analyzer plugin
type sshPlugin struct {
	ports        protos.PortsConfig
	parserConfig parserConfig
	transConfig  transactionConfig
	pub          transPub
}

// Application Layer tcp stream data to be stored on tcp connection context.
type connection struct {
	streams [2]*stream
	trans   transactions
}

// Uni-directioal tcp stream state for parsing messages.
type stream struct {
	parser parser
}

var (
	debugf = logp.MakeDebug("ssh")

	// use isDebug/isDetailed to guard debugf/detailedf to minimize allocations
	// (garbage collection) when debug log is disabled.
	isDebug = false
)

func init() {
	protos.Register("ssh", New)
}

// New create and initializes a new ssh protocol analyzer instance.
func New(
	testMode bool,
	results publish.Transactions,
	cfg *common.Config,
) (protos.Plugin, error) {
	p := &sshPlugin{}
	config := defaultConfig
	if !testMode {
		if err := cfg.Unpack(&config); err != nil {
			return nil, err
		}
	}

	if err := p.init(results, &config); err != nil {
		return nil, err
	}
	return p, nil
}

func (ssh *sshPlugin) init(results publish.Transactions, config *sshConfig) error {
	if err := ssh.setFromConfig(config); err != nil {
		return err
	}
	ssh.pub.results = results

	isDebug = logp.IsDebug("ssh")
	return nil
}

func (ssh *sshPlugin) setFromConfig(config *sshConfig) error {

	// set module configuration
	if err := ssh.ports.Set(config.Ports); err != nil {
		return err
	}

	// set parser configuration

	parser := &ssh.parserConfig
	parser.maxBytes = tcp.TCPMaxDataInStream

	// set transaction correlator configuration
	trans := &ssh.transConfig
	trans.transactionTimeout = config.TransactionTimeout

	// set transaction publisher configuration
	pub := &ssh.pub
	pub.sendRequest = config.SendRequest
	pub.sendResponse = config.SendResponse

	return nil
}

// ConnectionTimeout returns the per stream connection timeout.
// Return <=0 to set default tcp module transaction timeout.
func (ssh *sshPlugin) ConnectionTimeout() time.Duration {
	return ssh.transConfig.transactionTimeout
}

// GetPorts returns the ports numbers packets shall be processed for.
func (ssh *sshPlugin) GetPorts() []int {
	return ssh.ports.Ports
}

// Parse processes a TCP packet. Return nil if connection
// state shall be dropped (e.g. parser not in sync with tcp stream)
func (ssh *sshPlugin) Parse(
	pkt *protos.Packet,
	tcptuple *common.TCPTuple, dir uint8,
	private protos.ProtocolData,
) protos.ProtocolData {
	defer logp.Recover("Parse sshPlugin exception")

	conn := ssh.ensureConnection(private)
	st := conn.streams[dir]

	if st == nil {
		st = &stream{}
		st.parser.init(&ssh.parserConfig, func(msg *message) error {
			return conn.trans.onMessage(tcptuple.IPPort(), dir, msg)
		})
		conn.streams[dir] = st
	}

	if err := st.parser.feed(pkt.Ts, pkt.Payload); err != nil {
		debugf("%v, dropping TCP stream for error in direction %v.", err, dir)
		ssh.onDropConnection(conn)
		return nil
	}

	return conn
}

// ReceivedFin handles TCP-FIN packet.
func (ssh *sshPlugin) ReceivedFin(
	tcptuple *common.TCPTuple, dir uint8,
	private protos.ProtocolData,
) protos.ProtocolData {
	return private
}

// GapInStream handles lost packets in tcp-stream.
func (ssh *sshPlugin) GapInStream(tcptuple *common.TCPTuple, dir uint8,
	nbytes int,
	private protos.ProtocolData,
) (protos.ProtocolData, bool) {
	conn := getConnection(private)
	if conn != nil {
		ssh.onDropConnection(conn)
	}

	return nil, true
}

// onDropConnection processes and optionally sends incomplete
// transaction in case of connection being dropped due to error
func (ssh *sshPlugin) onDropConnection(conn *connection) {
}

func (ssh *sshPlugin) ensureConnection(private protos.ProtocolData) *connection {
	conn := getConnection(private)
	if conn == nil {
		conn = &connection{}
		conn.trans.init(&ssh.transConfig, ssh.pub.onTransaction)
	}
	return conn
}

func (conn *connection) dropStreams() {
	conn.streams[0] = nil
	conn.streams[1] = nil
}

func getConnection(private protos.ProtocolData) *connection {
	if private == nil {
		return nil
	}

	priv, ok := private.(*connection)

	if !ok {
		logp.Warn("ssh connection type error")
		return nil
	}
	if priv == nil {
		logp.Warn("Unexpected: ssh connection data not set")
		return nil
	}
	return priv
}
