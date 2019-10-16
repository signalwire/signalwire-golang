package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"runtime"
	"runtime/trace"

	//	"sync"
	"syscall"

	"github.com/signalwire/signalwire-golang/signalwire"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// App consts
const (
	ProjectID = "replaceme"
	TokenID   = "replaceme" // nolint: gosec
)

// ProfilesStruct TODO DESCRIPTION
type ProfilesStruct struct {
	CallingInbound CallingInboundStruct `json:"Calling_Inbound"`
}

// CallingInboundStruct TODO DESCRIPTION
type CallingInboundStruct struct {
	CommandName          string
	EnvironmentVariables EnvironmentVariablesStruct
}

// EnvironmentVariablesStruct TODO DESCRIPTION
type EnvironmentVariablesStruct struct {
	TestHost    string `json:"TEST_HOST"`
	TestProject string `json:"TEST_PROJECT"`
	TestToken   string `json:"TEST_TOKEN"`
	TestContext string `json:"TEST_CONTEXT"`
}

// LocalSettings TODO DESCRIPTION
type LocalSettings struct {
	Profiles ProfilesStruct
}

func init() {
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

	signalwire.Logger = log
}

func main() {
	var (
		printVersion bool
		verbose      bool
		doTrace      bool
	)

	flag.BoolVar(&verbose, "d", false, " Enable debug mode ")
	flag.BoolVar(&printVersion, "v", false, " Show version ")
	flag.BoolVar(&doTrace, "t", false, " Generate trace (trace.out) ")
	flag.Parse()

	if printVersion {
		fmt.Printf("%s\n", signalwire.SDKVersion)
		fmt.Printf("App built with GO Lang version: " + fmt.Sprintf("%s\n", runtime.Version()))

		os.Exit(0)
	}

	if verbose {
		log.SetLevel(logrus.DebugLevel)
	}

	var traceFh *os.File

	if doTrace {
		var err error

		traceFh, err = os.Create("trace.out")
		if err != nil {
			fmt.Printf("Cound not create the trace file\n")

			os.Exit(0)
		}

		defer traceFh.Close()

		if err = trace.Start(traceFh); err != nil {
			fmt.Printf("Trace error: %v", err)
		}
	}

	go func() {
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)
		for {
			s := <-interrupt

			switch s {
			case syscall.SIGHUP:
				fallthrough
			case syscall.SIGTERM:
				fallthrough
			case syscall.SIGINT:
				if doTrace {
					trace.Stop()
					traceFh.Close()
				}
				log.Printf("Exit")

				os.Exit(0)
			}
		}
	}()

	jLaunch, openErr := os.Open("launchSettings_inbound.json")
	if openErr != nil {
		log.Fatal(openErr)
	}

	b, ierr := ioutil.ReadAll(jLaunch)
	if ierr != nil {
		log.Fatal(ierr)
	}

	var settings LocalSettings
	if err := json.Unmarshal(b, &settings); err != nil {
		log.Fatal(err)
	}

	log.Printf("Launch settings: %v\n", settings)

	testContext := settings.Profiles.CallingInbound.EnvironmentVariables.TestContext
	testHost := settings.Profiles.CallingInbound.EnvironmentVariables.TestHost

	if len(testHost) == 0 {
		testHost = signalwire.WssHost
	}

	defer jLaunch.Close()

	var DemoSession signalwire.BladeSession

	var I signalwire.IBlade = signalwire.BladeNew()

	DemoSession = signalwire.BladeSession{I: I}
	blade := &DemoSession
	blade.I = blade

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := blade.BladeInit(ctx, testHost); err != nil {
		log.Fatalf("cannot init Blade: %v\n", err)
	}

	var bladeAuth signalwire.BladeAuth

	bladeAuth.ProjectID = ProjectID
	bladeAuth.TokenID = TokenID

	if err := blade.BladeConnect(ctx, &bladeAuth); err != nil {
		log.Fatalf("cannot connect to Blade Network: %v\n", err)
	}

	if blade.SessionState != signalwire.BladeConnected {
		log.Fatalf("not in connected state\n")
	}

	if err := blade.BladeSetup(ctx); err != nil {
		log.Fatalf("cannot setup protocol on Blade Network: %v\n", err)
	}

	DemoSession.SignalwireChannels = []string{"notifications"}
	if err := blade.BladeAddSubscription(ctx, DemoSession.SignalwireChannels); err != nil {
		log.Fatalf("cannot subscribe to notifications on Blade Network: %v\n", err)
	}

	DemoSession.SignalwireContexts = []string{testContext}
	if err := blade.BladeSignalwireReceive(ctx, DemoSession.SignalwireContexts); err != nil {
		log.Fatalf("cannot subscribe to inbound context on Blade Network: %v\n", err)
	}

	var Relay signalwire.RelaySession

	Relay.Blade = blade

	var (
		call *signalwire.CallSession
		err  error
	)

	log.Printf("Waiting for incoming call...\n")

	// blocking
	call, err = Relay.RelayOnInboundAnswer(ctx)
	if err != nil {
		log.Fatal("cannot answer call\n")
	}

	log.Printf("Answered call. [%v]\n", call)

	// 'answer' state event may have already come before we get the 200 for call.answer
	if call.CallState != signalwire.Answered {
		if ret := call.WaitCallStateInternal(ctx, signalwire.Answered); !ret {
			log.Warn("did not get Answered state")
		}
	}

	if err := Relay.RelayPlayAudio(ctx, call, "1234abcdef", "https://cdn.signalwire.com/default-music/welcome.mp3"); err != nil {
		log.Fatalf("cannot play audio on call: %v\n", err)
	}

	// wait for the other side to hangup, otherwise we hangup
	log.Infof("wait for Ending...")

	if ret := call.WaitCallStateInternal(ctx, signalwire.Ending); !ret {
		log.Warn("did not get Ending state")
	}

	if ret := call.WaitCallStateInternal(ctx, signalwire.Ended); !ret {
		log.Warn("did not get Ended state")
	}

	if call.CallState != signalwire.Ending && call.CallState != signalwire.Ended {
		if err := Relay.RelayCallEnd(ctx, call); err != nil {
			log.Fatalf("call.end error: %v\n", err)
		}
	}

	log.Printf("show CallSession for call [%p] [%v]\n", &call, call)

	if err := Relay.RelayStop(ctx); err != nil {
		log.Fatalf("RelayStop error: %v\n", err)
	}

	log.Printf("Test Passed\n")

	call.CallCleanup(ctx)

	/*pass context to go routine*/
	go blade.BladeWaitDisconnect(DemoSession.Ctx)
}
