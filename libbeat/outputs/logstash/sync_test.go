// +build !integration

package logstash

import (
	"sync"
	"testing"
	"time"

	"packetbeat/libbeat/outputs"
	"packetbeat/libbeat/outputs/mode"
	"packetbeat/libbeat/outputs/transport"
	"packetbeat/libbeat/outputs/transport/transptest"
)

type testSyncDriver struct {
	client  mode.ProtocolClient
	ch      chan testDriverCommand
	returns []testClientReturn
	wg      sync.WaitGroup
}

type clientServer struct {
	*transptest.MockServer
}

func TestClientSendZero(t *testing.T) {
	testSendZero(t, makeTestClient(nil))
}

func TestClientSimpleEvent(t *testing.T) {
	testSimpleEvent(t, makeTestClient(nil))
}

func TestClientStructuredEvent(t *testing.T) {
	testStructuredEvent(t, makeTestClient(nil))
}

func TestClientMultiFailMaxTimeouts(t *testing.T) {
	testMultiFailMaxTimeouts(t, makeTestClient(nil))
}

func newClientServerTCP(t *testing.T, to time.Duration) *clientServer {
	return &clientServer{transptest.NewMockServerTCP(t, to, "", nil)}
}

func makeTestClient(settings map[string]interface{}) func(*transport.Client, string) testClientDriver {
	return func(conn *transport.Client, host string) testClientDriver {
		return newClientTestDriver(newLumberjackTestClient(conn, host, settings))
	}
}

func newClientTestDriver(client mode.ProtocolClient) *testSyncDriver {
	driver := &testSyncDriver{
		client:  client,
		ch:      make(chan testDriverCommand),
		returns: nil,
	}

	driver.wg.Add(1)
	go func() {
		defer driver.wg.Done()

		for {
			cmd, ok := <-driver.ch
			if !ok {
				return
			}

			switch cmd.code {
			case driverCmdQuit:
				return
			case driverCmdConnect:
				driver.client.Connect(1 * time.Second)
			case driverCmdClose:
				driver.client.Close()
			case driverCmdPublish:
				events, err := driver.client.PublishEvents(cmd.data)
				n := len(cmd.data) - len(events)
				driver.returns = append(driver.returns, testClientReturn{n, err})
			}
		}
	}()

	return driver
}

func (t *testSyncDriver) Stop() {
	if t.ch != nil {
		t.ch <- testDriverCommand{code: driverCmdQuit}
		t.wg.Wait()
		close(t.ch)
		t.client.Close()
		t.ch = nil
	}
}

func (t *testSyncDriver) Connect() {
	t.ch <- testDriverCommand{code: driverCmdConnect}
}

func (t *testSyncDriver) Close() {
	t.ch <- testDriverCommand{code: driverCmdClose}
}

func (t *testSyncDriver) Publish(data []outputs.Data) {
	t.ch <- testDriverCommand{code: driverCmdPublish, data: data}
}

func (t *testSyncDriver) Returns() []testClientReturn {
	return t.returns
}
