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
	CallingOutbound CallingOutboundStruct `json:"Calling_Outbound"`
}

// CallingOutboundStruct TODO DESCRIPTION
type CallingOutboundStruct struct {
	CommandName          string
	EnvironmentVariables EnvironmentVariablesStruct
}

// EnvironmentVariablesStruct TODO DESCRIPTION
type EnvironmentVariablesStruct struct {
	TestHost       string `json:"TEST_HOST"`
	TestProject    string `json:"TEST_PROJECT"`
	TestToken      string `json:"TEST_TOKEN"`
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
			fmt.Printf("Trace error: %v\n", err)
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

	jLaunch, err := os.Open("launchSettings_outbound.json")
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

	toNumber := settings.Profiles.CallingOutbound.EnvironmentVariables.TestToNumber
	fromNumber := settings.Profiles.CallingOutbound.EnvironmentVariables.TestFromNumber
	testHost := settings.Profiles.CallingOutbound.EnvironmentVariables.TestHost

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

	signalwire.Log.Info("dialing...\n")

	for j := 1; j <= testCalls; j++ {
		call := &DemoSession.Calls[j]

		var Relay signalwire.RelaySession

		Relay.Blade = blade

		if err := Relay.RelayPhoneDial(ctx, call, fromNumber, toNumber, 10); err != nil {
			signalwire.Log.Fatal("cannot dial phone number: %v\n", err)
		}

		// wait for "Answered"
		signalwire.Log.Info("wait for 'Answered'...\n")

		if ret := call.WaitCallStateInternal(ctx, signalwire.Answered); !ret {
			signalwire.Log.Fatal("did not get Answered state\n")
		}

		if err := Relay.RelayPlayAudio(ctx, call, "1234abc", "https://cdn.signalwire.com/default-music/welcome.mp3"); err != nil {
			signalwire.Log.Fatal("cannot play audio on call: %v\n", err)
		}

		signalwire.Log.Info("wait for 'Ending'...\n")

		if ret := call.WaitCallStateInternal(ctx, signalwire.Ending); !ret {
			signalwire.Log.Warn("did not get Ending state\n")
		}

		if ret := call.WaitCallStateInternal(ctx, signalwire.Ended); !ret {
			signalwire.Log.Warn("did not get Ended state\n")
		}

		if call.CallState != signalwire.Ending && call.CallState != signalwire.Ended {
			if err := Relay.RelayCallEnd(ctx, call); err != nil {
				signalwire.Log.Fatal("call.end error: %v\n", err)
			}
		}

		signalwire.Log.Info("show CallSession for call [%v]\n", call)

		if err := Relay.RelayStop(ctx); err != nil {
			signalwire.Log.Fatal("RelayStop error: %v\n", err)
		}

		signalwire.Log.Info("Test Passed\n")
		call.CallCleanup(ctx)
	}

	/* pass context to go routine */
	go blade.BladeWaitDisconnect(DemoSession.Ctx)
}
