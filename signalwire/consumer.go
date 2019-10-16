package signalwire

import (
	"context"
	"errors"
	"fmt"
	"path"
	"runtime"
	"sync"

	"github.com/sirupsen/logrus"
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
}

// IConsumer TODO DESCRIPTION
type IConsumer interface {
	Setup(projectID, token string)
	Teardown()
	Stop()
	Run()
}

var log *logrus.Logger

func loginit() {
	log = logrus.New()

	log.SetFormatter(
		&logrus.TextFormatter{
			DisableColors: false,
			FullTimestamp: true,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				filename := path.Base(f.File)
				return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filename, f.Line)
			},
		},
	)
	log.SetReportCaller(true)
	//	log.SetLevel(logrus.DebugLevel)

	Logger = log
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

	loginit()

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
		log.Debugf("cannot setup Blade: %v", err)

		return errors.New("cannot setup Blade")
	}

	log.Debugf("Blade Ready...\n")

	consumer.Client.SetupInbound()

	if consumer.Ready != nil {
		consumer.Ready(consumer)
	}

	if consumer.OnIncomingCall != nil {
		go func() {
			for {
				call, ierr := consumer.Client.I.WaitInbound()
				if ierr != nil {
					log.Errorf("Error processing incoming call: %v\n", ierr)
				} else {
					var I ICallObj = CallObjNew()

					c := &CallObj{I: I}
					c.call = call
					c.Calling = &consumer.Client.Calling
					consumer.OnIncomingCall(consumer, c)
				}
			}
		}()
		log.Debugf("OnIncomingCall CB enabled\n")
	}

	wg.Wait()

	log.Debug("consumer()/Run() stopped.\n")

	return err
}

// Stop TODO DESCRIPTION
func (consumer *Consumer) Stop() error {
	return consumer.Client.I.Disconnect()
}
