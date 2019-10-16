package signalwire

import (
	"context"
	"errors"
	"sync"
)

/*
type SwireAPI func()
type CallingAPI func()
type MessagingAPI func()
type TaskingAPI func()
*/

// ClientSession TODO DESCRIPTION
type ClientSession struct {
	Project     string
	Token       string
	Host        string
	Agent       string
	Relay       RelaySession
	Calling     Calling
	Ctx         context.Context
	Cancel      context.CancelFunc
	Operational chan struct{}
	I           IClientSession
}

// IClientSession TODO DESCRIPTION
type IClientSession interface {
	SetAuth(project, token string)
	SetClient(host string, contexts []string)
	Connect(ctx context.Context, cancel context.CancelFunc, runWG *sync.WaitGroup) error
	Disconnect() error
	SetupInbound()
	WaitInbound() (*CallSession, error)
	//	OnReady()
	//	OnDisconnected()
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
	blade := client.Relay.Blade
	client.Calling.Relay = new(RelaySession)
	client.Calling.Relay.Blade = blade

	if err := blade.BladeInit(ctx, client.Host); err != nil {
		log.Debugf("cannot init Blade: %v\n", err)
		return err
	}

	if err := blade.BladeConnect(ctx, &blade.bladeAuth); err != nil {
		log.Debugf("cannot connect to Blade Network: %v\n", err)
		return err
	}

	if blade.SessionState != BladeConnected {
		log.Debugf("not in connected state\n")
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

	log.Debugf("execute Setup\n")

	if err := blade.BladeSetup(ctx); err != nil {
		log.Debugf("cannot setup protocol on Blade Network: %v\n", err)
		return err
	}

	log.Debugf("waiting for Netcast (protocol.add)...")

	wg.Wait()

	if sProtocol != blade.Protocol {
		log.Debugf("cannot setup protocol on Blade Network / different protocol received [%s:%s]\n", sProtocol, blade.Protocol)
		return errors.New("different protocol received (netcast)")
	}

	blade.SignalwireChannels = []string{"notifications"}
	if err := blade.BladeAddSubscription(ctx, blade.SignalwireChannels); err != nil {
		log.Debugf("cannot subscribe to notifications on Blade Network: %v\n", err)
		return err
	}

	if err := blade.BladeSignalwireReceive(ctx, blade.SignalwireContexts); err != nil {
		log.Debugf("cannot subscribe to inbound context on Blade Network: %v\n", err)
		return err
	}

	client.Operational <- struct{}{}

	blade.BladeWaitDisconnect(ctx)
	log.Debugf("got Disconnect")
	cancel()
	runWG.Done()

	return nil
}

// Disconnect TODO DESCRIPTION
func (client *ClientSession) Disconnect() error {
	blade := client.Relay.Blade

	if err := blade.BladeDisconnect(client.Ctx); err != nil {
		log.Debugf("Blade: error disconnecting\n")
		return err
	}

	select {
	case blade.InboundDone <- struct{}{}:
		Logger.Debugf("sent InboundDone to go routine\n")
	default:
		Logger.Debugf("cannot send InboundDone to go routine\n")
	}

	select {
	case blade.DisconnectChan <- struct{}{}:
		Logger.Debugf("sent DisconnectChan to go routine\n")
	default:
		Logger.Debugf("cannot send DisconnectChan to go routine\n")
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
func (client *ClientSession) WaitInbound() (*CallSession, error) {
	blade := client.Relay.Blade
	call, err := blade.BladeWaitInboundCall(client.Ctx)

	return call, err
}

// ClientNew TODO DESCRIPTION
func ClientNew() *ClientSession {
	return &ClientSession{}
}
