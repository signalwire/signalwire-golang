package signalwire

import (
	"context"
	"errors"
	"sync"
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
	Ctx         context.Context
	Cancel      context.CancelFunc
	Operational chan struct{}
	I           IClientSession

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
	SetAuth(project, token string)
	SetClient(host string, contexts []string)
	Connect(ctx context.Context, cancel context.CancelFunc, runWG *sync.WaitGroup) error
	Disconnect() error
	SetupInbound()
	WaitInbound(ctx context.Context) (*CallSession, error)
	WaitInboundMsg(ctx context.Context) (*MsgSession, error)
}

// SetClient TODO DESCRIPTION
func (client *ClientSession) SetClient(host string, contexts []string) {
	client.Host = host

	var I IBlade = BladeNew()

	blade := &BladeSession{I: I}
	blade.I = blade
	client.Relay.Blade = blade
	client.Relay.Blade.SignalwireContexts = contexts
	client.Operational = make(chan struct{})
}

// SetAuth TODO DESCRIPTION
func (client *ClientSession) SetAuth(project, token string) {
	bladeAuth := new(BladeAuth)
	bladeAuth.ProjectID = project
	bladeAuth.TokenID = token

	client.Relay.Blade.bladeAuth = *bladeAuth
}

// Connect TODO DESCRIPTION
func (client *ClientSession) Connect(ctx context.Context, cancel context.CancelFunc, runWG *sync.WaitGroup) error {
	client.Ctx = ctx
	client.Cancel = cancel
	client.Calling.Ctx = client.Ctx
	client.Calling.Cancel = client.Cancel

	client.Messaging.Ctx = client.Ctx
	client.Messaging.Cancel = client.Cancel

	blade := client.Relay.Blade
	client.Calling.Relay = new(RelaySession)
	client.Calling.Relay.Blade = blade

	client.Messaging.Relay = client.Calling.Relay
	client.Messaging.Relay.Blade = client.Calling.Relay.Blade

	if err := blade.BladeInit(ctx, client.Host); err != nil {
		Log.Debug("cannot init Blade: %v\n", err)

		return err
	}

	if err := blade.BladeConnect(ctx, &blade.bladeAuth); err != nil {
		Log.Debug("cannot connect to Blade Network: %v\n", err)

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

	if err := blade.BladeSignalwireReceive(ctx, blade.SignalwireContexts); err != nil {
		Log.Debug("cannot subscribe to inbound context on Blade Network: %v\n", err)

		return err
	}

	client.Operational <- struct{}{}

	blade.BladeWaitDisconnect(ctx)
	Log.Debug("got Disconnect\n")

	cancel()
	runWG.Done()

	return nil
}

// Disconnect TODO DESCRIPTION
func (client *ClientSession) Disconnect() error {
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

// SetupInbound TODO DESCRIPTION
func (client *ClientSession) SetupInbound() {
	blade := client.Relay.Blade
	blade.BladeSetupInbound(client.Ctx)
}

// WaitInbound TODO DESCRIPTION
func (client *ClientSession) WaitInbound(_ context.Context) (*CallSession, error) {
	blade := client.Relay.Blade
	call, err := blade.BladeWaitInboundCall(client.Ctx)

	return call, err
}

// WaitInboundMsg TODO DESCRIPTION
func (client *ClientSession) WaitInboundMsg(_ context.Context) (*MsgSession, error) {
	blade := client.Relay.Blade
	call, err := blade.BladeWaitInboundMsg(client.Ctx)

	return call, err
}

// ClientNew TODO DESCRIPTION
func ClientNew() *ClientSession {
	return &ClientSession{}
}
