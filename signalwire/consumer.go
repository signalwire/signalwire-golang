package signalwire

import (
	"context"
	"errors"
	"sync"
	"time"
)

// GlobalOverwriteHost TODO DESCRIPTION
var GlobalOverwriteHost string

// Consumer TODO DESCRIPTION
type Consumer struct {
	Project              string
	Token                string
	Contexts             []string
	Host                 string
	Client               *ClientSession
	Ready                func(*Consumer)
	OnIncomingCall       func(*Consumer, *CallObj)
	OnIncomingMessage    func(*Consumer, *MsgObj)
	OnMessageStateChange func(*Consumer, *MsgObj)
	OnTask               func(*Consumer, ParamsEventTaskingTask)
	Teardown             func(*Consumer)

	Log LoggerWrapper
}

// NewConsumer TODO DESCRIPTION
func NewConsumer() *Consumer {
	return &Consumer{
		Log: Log,
	}
}

// IConsumer TODO DESCRIPTION
type IConsumer interface {
	Setup(projectID, token string)
	Stop()
	Run()
}

// Setup TODO DESCRIPTION
func (consumer *Consumer) Setup(project, token string, contexts []string) {
	consumer.Project = project
	consumer.Token = token

	if len(GlobalOverwriteHost) == 0 {
		consumer.Host = WssHost
	} else {
		consumer.Host = GlobalOverwriteHost
	}

	consumer.Contexts = contexts

	var I IClientSession = ClientNew()

	c := &ClientSession{I: I, Consumer: consumer}
	c.I = c
	consumer.Client = c
}

func (consumer *Consumer) runOnIncomingCall(_ context.Context, call *CallSession) {
	var I ICallObj = CallObjNew()

	c := &CallObj{I: I}
	c.call = call
	c.Calling = &consumer.Client.Calling
	consumer.OnIncomingCall(consumer, c)
}

func (consumer *Consumer) runOnIncomingMessage(_ context.Context, msg *MsgSession) {
	var I IMsgObj = MsgObjNew()

	m := &MsgObj{I: I}
	m.msg = msg
	m.Messaging = &consumer.Client.Messaging
	consumer.OnIncomingMessage(consumer, m)
}

func (consumer *Consumer) incomingCall(ctx context.Context, wg *sync.WaitGroup) {
	for {
		call, ierr := consumer.Client.I.waitInbound(ctx)
		if ierr != nil {
			Log.Error("Error processing incoming call: %v\n", ierr)
		} else if call == nil && ierr == nil {
			wg.Done()
			return
		}

		if call != nil {
			go consumer.runOnIncomingCall(ctx, call)
		}
	}
}

func (consumer *Consumer) incomingMessage(ctx context.Context, wg *sync.WaitGroup) {
	for {
		msg, ierr := consumer.Client.I.waitInboundMsg(ctx)
		if ierr != nil {
			Log.Error("Error processing incoming msg: %v\n", ierr)
		} else if msg == nil && ierr == nil {
			wg.Done()
			return
		}

		if msg != nil {
			go consumer.runOnIncomingMessage(ctx, msg)
		}
	}
}

// Run TODO DESCRIPTION
func (consumer *Consumer) Run() error {
	consumer.Client.setClient(consumer.Host, consumer.Contexts)
	consumer.Client.setAuth(consumer.Project, consumer.Token)

	ctx, cancel := context.WithCancel(context.Background())

	var (
		err error
		wg  sync.WaitGroup
	)

	var haveIncomingMsg int

	var haveIncomingCalls int

	// will start go routines for incoming calls and incoming messages.
	// will prepare waitGroup
	if consumer.OnIncomingMessage != nil {
		haveIncomingMsg = 1
	}

	if consumer.OnIncomingCall != nil {
		haveIncomingCalls = 1
	}

	wg.Add(haveIncomingMsg + haveIncomingCalls + 1)

	go func() {
		err = consumer.Client.I.connectInternal(ctx, cancel, &wg)
		if err != nil {
			Log.Error("Cannot setup Blade: %v\n", err)
		}
	}()

	timer := time.NewTimer(WSTimeOut * time.Second)
	select {
	case <-consumer.Client.Operational:
	case <-timer.C:
		return errors.New("cannot setup Blade (timeout)")
	}

	Log.Debug("Blade Ready...\n")

	consumer.Client.setupInbound()
	consumer.Client.setupInboundMsg()

	if consumer.Ready != nil {
		consumer.Ready(consumer)
	}

	if consumer.OnIncomingCall != nil {
		go consumer.incomingCall(ctx, &wg)
		Log.Debug("OnIncomingCall CB enabled\n")
	}

	if consumer.OnIncomingMessage != nil {
		go consumer.incomingMessage(ctx, &wg)
		Log.Debug("OnIncomingMessage CB enabled\n")
	}

	wg.Wait()

	Log.Debug("consumer()/Run() stopped.\n")

	return nil
}

// Stop TODO DESCRIPTION
func (consumer *Consumer) Stop() error {
	if consumer.Teardown != nil {
		consumer.Teardown(consumer)
	}

	return consumer.Client.I.disconnectInternal()
}
