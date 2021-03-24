package ssh

import (
	"expvar"
	"time"

	"packetbeat/libbeat/common"
	"packetbeat/libbeat/logp"

	"packetbeat/protos"
	"packetbeat/protos/tcp"
	"packetbeat/publish"
)

type sshPlugin struct {
	ports              []int
	sendRequest        bool
	sendResponse       bool
	transactionTimeout time.Duration
	results            publish.Transactions
}

type connection struct {
	streams   [2]*stream
	requests  messageList
	responses messageList
}

type stream struct {
	applayer.Stream
	parser   parser
	tcptuple *common.TCPTuple
}

var (
	unmatchedResponses = expvar.NewInt("ssh.unmatched_responses")
)

var (
	debugf  = logp.MakeDebug("ssh")
	isDebug = false
)

func init() {
	protos.Register("ssh", New)
}

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
	ssh.results = results

	isDebug = logp.IsDebug("ssh")
	return nil
}

func (ssh *sshPlugin) setFromConfig(config *sshConfig) error {

	// set module configuration
	ssh.ports = config.Ports
	ssh.sendRequest = config.SendRequest
	ssh.sendResponse = config.SendResponse
	ssh.transactionTimeout = config.TransactionTimeout

	return nil
}

func (ssh *sshPlugin) ConnectionTimeout() time.Duration {
	return ssh.transactionTimeout
}

func (ssh *sshPlugin) GetPorts() []int {
	return ssh.ports
}

func (ssh *sshPlugin) Parse(
	pkt *protos.Packet,
	tcptuple *common.TcpTuple, dir uint8,
	private protos.ProtocolData,
) protos.ProtocolData {
	defer logp.Recover("Parse sshPlugin exception")

	conn := ssh.ensureConnection(private)
	st := conn.streams[dir]
	if st == nil {
		st = &stream{}
		st.parser.init(&ssh.parserConfig, func(msg *message) error {
			return conn.trans.onMessage(tcptuple.IpPort(), dir, msg)
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

func (redis *redisPlugin) GapInStream(tcptuple *common.TCPTuple, dir uint8,
	nbytes int, private protos.ProtocolData) (priv protos.ProtocolData, drop bool) {

	// tsg: being packet loss tolerant is probably not very useful for Redis,
	// because most requests/response tend to fit in a single packet.

	return private, true
}

func (redis *redisPlugin) ReceivedFin(tcptuple *common.TCPTuple, dir uint8,
	private protos.ProtocolData) protos.ProtocolData {

	// TODO: check if we have pending data that we can send up the stack

	return private
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
		logp.Warn("{{ cookiecutter.module }} connection type error")
		return nil
	}
	if priv == nil {
		logp.Warn("Unexpected: {{ cookiecutter.module }} connection data not set")
		return nil
	}
	return priv
}

func (ml *messageList) empty() bool {
	return ml.head == nil
}

func (ml *messageList) pop() *message {
	if ml.head == nil {
		return nil
	}

	msg := ml.head
	ml.head = ml.head.next
	if ml.head == nil {
		ml.tail = nil
	}
	return msg
}

func (ml *messageList) first() *message {
	return ml.head
}

func (ml *messageList) last() *message {
	return ml.tail
}
