package signalwire

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"
)

// ClientSession TODO DESCRIPTION
type ClientSession struct {
	Project     string
	Token       string
	Host        string
	Agent       string
	Relay       RelaySession
	Calling     Calling
	Messaging   Messaging
	Tasking     Tasking
	Ctx         context.Context
	Cancel      context.CancelFunc
	Operational chan struct{}
	I           IClientSession
	Consumer    *Consumer
	OnReady     func(*ClientSession)

	Log LoggerWrapper
}

// NewClientSession TODO DESCRIPTION
func NewClientSession() *ClientSession {
	return &ClientSession{
		Log: Log,
	}
}

// IClientSession TODO DESCRIPTION
type IClientSession interface {
	setAuth(project, token string)
	setClient(host string, contexts []string)
	connectInternal(ctx context.Context, cancel context.CancelFunc, runWG *sync.WaitGroup, t *time.Timer) error
	disconnectInternal() error
	setupInbound()
	waitInbound(ctx context.Context) (*CallSession, error)
	waitInboundMsg(ctx context.Context) (*MsgSession, error)
}

// SetClient TODO DESCRIPTION
func (client *ClientSession) setClient(host string, contexts []string) {
	client.Host = host

	var I IBlade = BladeNew()

	blade := &BladeSession{I: I}
	blade.I = blade
	client.Relay.Blade = blade
	client.Relay.Blade.SignalwireContexts = contexts
	client.Operational = make(chan struct{})
}

// SetAuth TODO DESCRIPTION
func (client *ClientSession) setAuth(project, token string) {
	bladeAuth := new(BladeAuth)
	bladeAuth.ProjectID = project
	bladeAuth.TokenID = token

	client.Relay.Blade.bladeAuth = *bladeAuth
}

// connectInternal TODO DESCRIPTION
func (client *ClientSession) connectInternal(ctx context.Context, cancel context.CancelFunc, runWG *sync.WaitGroup, t *time.Timer) error {
	client.Ctx = ctx
	client.Cancel = cancel
	client.Calling.Ctx = client.Ctx
	client.Calling.Cancel = client.Cancel

	client.Messaging.Ctx = client.Ctx
	client.Messaging.Cancel = client.Cancel

	client.Tasking.Ctx = client.Ctx
	client.Tasking.Cancel = client.Cancel

	blade := client.Relay.Blade

	var I IRelay = RelayNew()

	relay := &RelaySession{I: I}
	relay.I = relay

	client.Calling.Relay = relay
	client.Calling.Relay.Blade = blade

	client.Messaging.Relay = client.Calling.Relay
	client.Messaging.Relay.Blade = client.Calling.Relay.Blade

	client.Tasking.Relay = client.Calling.Relay
	client.Tasking.Relay.Blade = client.Calling.Relay.Blade

	client.Tasking.Consumer = client.Consumer

	var count = BladeConnectionRetries

again:
	if err := blade.BladeInit(ctx, client.Host); err != nil {
		Log.Debug("cannot init Blade: %v\n", err)

		return err
	}

	if err := blade.BladeConnect(ctx, &blade.bladeAuth); err != nil {
		Log.Debug("cannot connect to Blade Network. Error Code: [%v] Message: [%v]\n", blade.LastJRPCError.Code, blade.LastJRPCError.Message)

		if blade.LastJRPCError.Code == -32000 && strings.Contains(blade.LastJRPCError.Message, "Timeout") {
			_ = client.disconnectInternal()

			if t != nil {
				t.Reset(GlobalConnectTimeout * time.Second)
			}

			if count != 0 {
				time.Sleep(5 * time.Second)

				if count != -1 {
					count--
				}

				goto again
			}
		}

		return err
	}

	if blade.SessionState != BladeConnected {
		Log.Debug("not in connected state\n")

		return errors.New("not in connected state")
	}

	var (
		wg        sync.WaitGroup
		sProtocol string
	)

	wg.Add(1)

	go func() {
		sProtocol = <-blade.Netcast

		wg.Done()
	}()

	Log.Debug("execute Setup\n")

	if err := blade.BladeSetup(ctx); err != nil {
		Log.Debug("cannot setup protocol on Blade Network: %v\n", err)

		return err
	}

	Log.Debug("waiting for Netcast (protocol.add)...\n")

	wg.Wait()

	if sProtocol != blade.Protocol {
		Log.Debug("cannot setup protocol on Blade Network / different protocol received [%s:%s]\n", sProtocol, blade.Protocol)

		return errors.New("different protocol received (netcast)")
	}

	blade.SignalwireChannels = []string{"notifications"}
	if err := blade.BladeAddSubscription(ctx, blade.SignalwireChannels); err != nil {
		Log.Debug("cannot subscribe to notifications on Blade Network: %v\n", err)

		return err
	}

	if len(blade.SignalwireContexts) > 0 {
		if err := blade.BladeSignalwireReceive(ctx, blade.SignalwireContexts); err != nil {
			Log.Debug("cannot subscribe to inbound context on Blade Network: %v\n", err)

			return err
		}
	}

	client.Tasking.TaskChan = make(chan ParamsEventTaskingTask, 1)
	if err := blade.EventTasking.Cache.SetTasking("tasking", &client.Tasking); err != nil {
		return err
	}
	client.Operational <- struct{}{}

	blade.BladeWaitDisconnect(ctx)
	Log.Debug("got Disconnect\n")

	close(client.Tasking.TaskChan)
	cancel()
	runWG.Done()

	return nil
}

// disconnect TODO DESCRIPTION
func (client *ClientSession) disconnectInternal() error {
	blade := client.Relay.Blade

	if err := blade.BladeDisconnect(client.Ctx); err != nil {
		Log.Debug("Blade: error disconnecting\n")

		return err
	}

	select {
	case blade.InboundDone <- struct{}{}:
		Log.Debug("sent InboundDone to go routine\n")
	default:
		Log.Debug("cannot send InboundDone to go routine\n")
	}

	select {
	case blade.DisconnectChan <- struct{}{}:
		Log.Debug("sent DisconnectChan to go routine\n")
	default:
		Log.Debug("cannot send DisconnectChan to go routine\n")
	}

	client.Cancel()
	close(client.Operational)

	return nil
}

// setupInbound TODO DESCRIPTION
func (client *ClientSession) setupInbound() {
	blade := client.Relay.Blade
	blade.BladeSetupInbound(client.Ctx)
}

// waitInbound TODO DESCRIPTION
func (client *ClientSession) waitInbound(_ context.Context) (*CallSession, error) {
	blade := client.Relay.Blade
	call, err := blade.BladeWaitInboundCall(client.Ctx)

	return call, err
}

// setupInboundMsg TODO DESCRIPTION
func (client *ClientSession) setupInboundMsg() {
	blade := client.Relay.Blade
	blade.BladeSetupInboundMsg(client.Ctx)
}

// waitInboundMsg TODO DESCRIPTION
func (client *ClientSession) waitInboundMsg(_ context.Context) (*MsgSession, error) {
	blade := client.Relay.Blade
	call, err := blade.BladeWaitInboundMsg(client.Ctx)

	return call, err
}

// ClientNew TODO DESCRIPTION
func ClientNew() *ClientSession {
	return &ClientSession{}
}

// Client TODO DESCRIPTION
func Client(project, token, host string, signalwireContexts []string) *ClientSession {
	if len(host) == 0 {
		host = WssHost
	}

	client := NewClientSession()
	client.setClient(host, signalwireContexts)
	client.setAuth(project, token)

	ctx, cancel := context.WithCancel(context.Background())

	client.Ctx = ctx
	client.Cancel = cancel
	client.Calling.Ctx = client.Ctx
	client.Calling.Cancel = client.Cancel

	client.Messaging.Ctx = client.Ctx
	client.Messaging.Cancel = client.Cancel

	client.Tasking.Ctx = client.Ctx
	client.Tasking.Cancel = client.Cancel

	blade := client.Relay.Blade

	var I IRelay = RelayNew()

	relay := &RelaySession{I: I}
	relay.I = relay

	client.Calling.Relay = relay
	client.Calling.Relay.Blade = blade

	client.Messaging.Relay = client.Calling.Relay
	client.Messaging.Relay.Blade = client.Calling.Relay.Blade

	client.Tasking.Relay = client.Calling.Relay
	client.Tasking.Relay.Blade = client.Calling.Relay.Blade

	//	client.Tasking.Consumer = client.Consumer

	return client
}

// Connect TODO DESCRIPTION
func (client *ClientSession) Connect() error {
	var (
		err error
		wg  sync.WaitGroup
	)

	wg.Add(1)

	go func() {
		err = client.connectInternal(client.Ctx, client.Cancel, &wg, nil)
	}()

	<-client.Operational

	if err != nil {
		Log.Debug("cannot setup Blade: %v\n", err)

		return errors.New("cannot setup Blade")
	}

	Log.Debug("Blade Client Ready...\n")

	if client.OnReady != nil {
		client.OnReady(client)
	}

	return err
}

// Disconnect TODO DESCRIPTION
func (client *ClientSession) Disconnect() error {
	return client.disconnectInternal()
}
