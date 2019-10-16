package signalwire

import (
	"context"
	"errors"
	"sync"
)

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
	OnIncomingMessage    func(*Consumer)
	OnMessageStateChange func(*Consumer)
	OnTask               func(*Consumer)

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
	Teardown()
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

	c := &ClientSession{I: I}
	c.I = c
	consumer.Client = c
}

// Run TODO DESCRIPTION
func (consumer *Consumer) Run() error {
	consumer.Client.SetClient(consumer.Host, consumer.Contexts)
	consumer.Client.SetAuth(consumer.Project, consumer.Token)

	ctx, cancel := context.WithCancel(context.Background())

	var (
		err error
		wg  sync.WaitGroup
	)

	if consumer.OnIncomingCall != nil {
		wg.Add(2)
	} else {
		wg.Add(1)
	}

	go func() {
		err = consumer.Client.I.Connect(ctx, cancel, &wg)
	}()

	<-consumer.Client.Operational

	if err != nil {
		Log.Debug("cannot setup Blade: %v\n", err)

		return errors.New("cannot setup Blade")
	}

	Log.Debug("Blade Ready...\n")

	consumer.Client.SetupInbound()

	if consumer.Ready != nil {
		consumer.Ready(consumer)
	}

	if consumer.OnIncomingCall != nil {
		go func() {
			for {
				call, ierr := consumer.Client.I.WaitInbound()
				if ierr != nil {
					Log.Error("Error processing incoming call: %v\n", ierr)
				} else {
					var I ICallObj = CallObjNew()

					c := &CallObj{I: I}
					c.call = call
					c.Calling = &consumer.Client.Calling
					consumer.OnIncomingCall(consumer, c)
				}
			}
		}()

		Log.Debug("OnIncomingCall CB enabled\n")
	}

	wg.Wait()

	Log.Debug("consumer()/Run() stopped.\n")

	return err
}

// Stop TODO DESCRIPTION
func (consumer *Consumer) Stop() error {
	return consumer.Client.I.Disconnect()
}
