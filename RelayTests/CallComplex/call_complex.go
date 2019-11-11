package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"runtime/trace"
	"syscall"

	"github.com/signalwire/signalwire-golang/signalwire"
)

// App consts
const (
	// testConn  = 1
	testCalls = 1
)

// App environment settings
var (
	ProjectID = os.Getenv("ProjectID")
	TokenID   = os.Getenv("TokenID")
)

// ProfilesStruct TODO DESCRIPTION
type ProfilesStruct struct {
	CallingComplex CallingComplexStruct `json:"Calling_Complex"`
}

// CallingComplexStruct TODO DESCRIPTION
type CallingComplexStruct struct {
	CommandName          string
	EnvironmentVariables EnvironmentVariablesStruct
}

// EnvironmentVariablesStruct TODO DESCRIPTION
type EnvironmentVariablesStruct struct {
	TestHost       string `json:"TEST_HOST"`
	TestProject    string `json:"TEST_PROJECT"`
	TestToken      string `json:"TEST_TOKEN"`
	TestContext    string `json:"TEST_CONTEXT"`
	TestFromNumber string `json:"TEST_FROM_NUMBER"`
	TestToNumber   string `json:"TEST_TO_NUMBER"`
}

// LocalSettings TODO DESCRIPTION
type LocalSettings struct {
	Profiles ProfilesStruct
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
		signalwire.Log.SetLevel(signalwire.DebugLevelLog)
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

				signalwire.Log.Info("Exit\n")

				os.Exit(0)
			}
		}
	}()

	jLaunch, err := os.Open("launchSettings_complex.json")
	if err != nil {
		signalwire.Log.Fatal("%v\n", err)
	}

	b, err := ioutil.ReadAll(jLaunch)
	if err != nil {
		signalwire.Log.Fatal("%v\n", err)
	}

	var settings LocalSettings
	if err = json.Unmarshal(b, &settings); err != nil {
		signalwire.Log.Fatal("%v\n", err)
	}

	signalwire.Log.Info("Launch settings: %v\n", settings)

	toNumber := settings.Profiles.CallingComplex.EnvironmentVariables.TestToNumber
	fromNumber := settings.Profiles.CallingComplex.EnvironmentVariables.TestFromNumber
	testContext := settings.Profiles.CallingComplex.EnvironmentVariables.TestContext
	testHost := settings.Profiles.CallingComplex.EnvironmentVariables.TestHost

	if len(testHost) == 0 {
		testHost = signalwire.WssHost
	}

	signalwire.Log.Info("ToNumber: %s\n", toNumber)
	signalwire.Log.Info("FromNumber: %s\n", fromNumber)

	defer jLaunch.Close()

	var DemoSession signalwire.BladeSession

	var I signalwire.IBlade = signalwire.BladeNew()

	DemoSession = signalwire.BladeSession{I: I}
	blade := &DemoSession
	blade.I = blade

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := blade.BladeInit(ctx, testHost); err != nil {
		signalwire.Log.Fatal("cannot init Blade: %v\n", err)
	}

	var bladeAuth signalwire.BladeAuth

	bladeAuth.ProjectID = ProjectID
	bladeAuth.TokenID = TokenID

	if err := blade.BladeConnect(ctx, &bladeAuth); err != nil {
		signalwire.Log.Fatal("cannot connect to Blade Network: %v\n", err)
	}

	if blade.SessionState != signalwire.BladeConnected {
		signalwire.Log.Fatal("not in connected state\n")
	}

	if err := blade.BladeSetup(ctx); err != nil {
		signalwire.Log.Fatal("cannot setup protocol on Blade Network: %v\n", err)
	}

	DemoSession.SignalwireChannels = []string{"notifications"}
	if err := blade.BladeAddSubscription(ctx, DemoSession.SignalwireChannels); err != nil {
		signalwire.Log.Fatal("cannot subscribe to notifications on Blade Network: %v\n", err)
	}

	/* can be skipped for outbound call */
	DemoSession.SignalwireContexts = []string{testContext}
	if err := blade.BladeSignalwireReceive(ctx, DemoSession.SignalwireContexts); err != nil {
		signalwire.Log.Fatal("cannot subscribe to inbound context on Blade Network: %v\n", err)
	}

	signalwire.Log.Info("dialing...\n")

	for j := 1; j <= testCalls; j++ {
		call := &DemoSession.Calls[j]

		var Relay signalwire.RelaySession

		Relay.Blade = blade

		var (
			err1B  error
			call1B *signalwire.CallSession /*call 1, B leg */
		)

		go func() {
			call1B, err1B = Relay.RelayOnInboundAnswer(ctx)
		}()

		if err := Relay.RelayPhoneDial(ctx, call, fromNumber, toNumber, 10, nil); err != nil {
			signalwire.Log.Fatal("cannot dial phone number: %v\n", err)
		}

		// wait for "Answered"
		signalwire.Log.Info("wait for 'Answered' on originated call tag [%s]...\n", call.TagID)

		if ret := call.WaitCallStateInternal(ctx, signalwire.Answered, 3); !ret {
			signalwire.Log.Fatal("did not get Answered state\n")
		}

		if err := Relay.RelayPhoneConnect(ctx, call, fromNumber, toNumber, nil); err != nil {
			signalwire.Log.Fatal("call.connect error: %v\n", err)
		}

		var (
			err2B  error
			call2B *signalwire.CallSession /*call 2, B leg */
		)

		go func() {
			call2B, err2B = Relay.RelayOnInboundAnswer(ctx)
		}()

		// wait for "Connected"
		signalwire.Log.Info("wait for 'Connected'...\n")

		if ret := call.WaitCallConnectState(ctx, signalwire.CallConnectConnected); !ret {
			signalwire.Log.Fatal("did not get CallConnected state\n")
		}

		if err := Relay.RelayCallEnd(ctx, call, nil); err != nil {
			signalwire.Log.Fatal("call.end error: %v\n", err)
		}

		if ret := call.WaitCallStateInternal(ctx, signalwire.Ended, 3); !ret {
			signalwire.Log.Fatal("did not get Ended state\n")
		}

		signalwire.Log.Info("show CallSession for 1st call [%p] [%v]\n", &call, call)

		peercall, _ := call.GetPeer(ctx)
		if peercall == nil {
			signalwire.Log.Fatal("cannot get peer call\n")
		}

		signalwire.Log.Info("peercall CallID: [%s]\n", peercall.CallID)

		if peercall.CallState != signalwire.Ended && peercall.CallState != signalwire.Ending {
			if err := Relay.RelayCallEnd(ctx, peercall, nil); err != nil {
				signalwire.Log.Fatal("call.end error: %v\n", err)
			}

			if ret := peercall.WaitCallStateInternal(ctx, signalwire.Ended, 3); !ret {
				signalwire.Log.Fatal("did not get Ended state\n")
			}
		}

		signalwire.Log.Info("show CallSession for peer call [%p] [%v]\n", peercall, peercall)

		signalwire.Log.Info("show CallSession for call 1 B leg  [%p] [%v]\n", call1B, call1B)

		signalwire.Log.Info("show CallSession for call 2 B leg  [%p] [%v]\n", call2B, call2B)

		if err := Relay.RelayStop(ctx); err != nil {
			signalwire.Log.Fatal("RelayStop error: %v\n", err)
		}

		if err1B != nil || err2B != nil {
			signalwire.Log.Fatal("err1B: [%v] err2B: [%v]\n", err1B, err2B)
		}

		signalwire.Log.Info("Test Passed\n")

		call.CallCleanup(ctx)
		peercall.CallCleanup(ctx)
	}

	/* pass context to go routine */
	go blade.BladeWaitDisconnect(DemoSession.Ctx)
}
