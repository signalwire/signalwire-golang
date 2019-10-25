package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/signalwire/signalwire-golang/signalwire"
)

// App consts
const (
	ProjectID      = "replaceme"
	TokenID        = "replaceme" // nolint: gosec
	DefaultContext = "replaceme"
)

// Contexts needed for inbound calls
var Contexts = []string{}

// PProjectID passed from command-line
var PProjectID string

// PTokenID passed from command-line
var PTokenID string

// PContext passed from command line (just one being passed, although we support many)
var PContext string

/*gopl.io spinner*/
func spinner(delay time.Duration) {
	for {
		for _, r := range `-\|/` {
			fmt.Printf("\r%c", r)
			time.Sleep(delay)
		}
	}
}

func myOnConnectStateChange(_ *signalwire.ConnectAction) {
	signalwire.Log.Info("Connect State Change")
}

func myOnConnectFailed(_ *signalwire.ConnectAction) {
	signalwire.Log.Info("Connect State: Failed")
}

func myOnConnectConnecting(_ *signalwire.ConnectAction) {
	signalwire.Log.Info("Connect State: Connecting")
}

func myOnConnectConnected(_ *signalwire.ConnectAction) {
	signalwire.Log.Info("Connect State: Connected")
}

func myOnConnectDisconnected(_ *signalwire.ConnectAction) {
	signalwire.Log.Info("Connect State: Disconnected")
}

// CountIncomingCalls Conter for incoming calls.
var CountIncomingCalls int

// MyOnIncomingCall - gets executed when we receive an incoming call
func MyOnIncomingCall(consumer *signalwire.Consumer, call *signalwire.CallObj) {
	fmt.Printf("got incoming call.\n")
	CountIncomingCalls++
	/*if the callee number in Connect() is assigned
	to our context, the call originated by Connect()
	will end up here as well. Avoid a loop.*/
	if CountIncomingCalls > 2 {
		return
	}

	resultAnswer, _ := call.Answer()
	if !resultAnswer.Successful {
		if err := consumer.Stop(); err != nil {
			signalwire.Log.Error("Error occurred while trying to stop Consumer\n")
		}

		return
	}

	call.OnConnectStateChange = myOnConnectStateChange
	call.OnConnectFailed = myOnConnectFailed
	call.OnConnectConnecting = myOnConnectConnecting
	call.OnConnectConnected = myOnConnectConnected
	call.OnConnectDisconnected = myOnConnectDisconnected

	go spinner(100 * time.Millisecond)

	deviceParams := signalwire.DevicePhoneParams{
		FromNumber: "+13XXXXXXXXX",
		ToNumber:   "+16XXXXXXXXX",
		Timeout:    20, /*ring timeout 20 seconds*/
	}

	devices := [][]signalwire.DeviceStruct{{
		signalwire.DeviceStruct{
			Type:   "phone",
			Params: deviceParams,
		},
	}}

	ringback := []signalwire.RingbackStruct{
		{
			Type: "ringtone",
			Params: signalwire.RingbackRingtoneParams{
				Name:     "us",
				Duration: 5.0,
			},
		},
		{
			Type: "tts",
			Params: signalwire.RingbackTTSParams{
				Text: "Welcome to Signalwire!",
			},
		},
	}

	_, err := call.Connect(&ringback, &devices)
	if err != nil {
		signalwire.Log.Error("error running call Connect()\n")
	}

	time.Sleep(20 * time.Second)
	signalwire.Log.Info("Hangup call...\n")

	hangupResult, err := call.Hangup()
	if err != nil {
		// RELAY error
		signalwire.Log.Error("Error occurred while trying to hangup call\n")
	}

	if hangupResult.GetSuccessful() {
		signalwire.Log.Info("Call disconnect result: %s\n", hangupResult.GetReason().String())
	}
}

func main() {
	var printVersion bool

	var verbose bool

	flag.BoolVar(&printVersion, "v", false, " Show version ")
	flag.StringVar(&PProjectID, "p", ProjectID, " ProjectID ")
	flag.StringVar(&PTokenID, "t", TokenID, " TokenID ")
	flag.StringVar(&PContext, "c", DefaultContext, " Context ")
	flag.BoolVar(&verbose, "d", false, " Enable debug mode ")
	flag.Parse()

	if printVersion {
		fmt.Printf("%s\n", signalwire.SDKVersion)
		fmt.Printf("Blade version: %d.%d.%d\n", signalwire.BladeVersionMajor, signalwire.BladeVersionMinor, signalwire.BladeRevision)
		fmt.Printf("App built with GO Lang version: " + fmt.Sprintf("%s\n", runtime.Version()))

		os.Exit(0)
	}

	if verbose {
		signalwire.Log.SetLevel(signalwire.DebugLevelLog)
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
				signalwire.Log.Info("Exit\n")
				os.Exit(0)
			}
		}
	}()

	Contexts = append(Contexts, PContext)
	consumer := new(signalwire.Consumer)
	// setup the Client
	consumer.Setup(PProjectID, PTokenID, Contexts)
	// register callback
	consumer.OnIncomingCall = MyOnIncomingCall

	timer := time.NewTimer(25 * time.Second)

	signalwire.Log.Info("Wait incoming call..\n")

	// start
	if err := consumer.Run(); err != nil {
		signalwire.Log.Error("Error occurred while starting Signalwire Consumer\n")
	}

	<-timer.C

	if err := consumer.Stop(); err != nil {
		signalwire.Log.Error("Error occurred while trying to stop Consumer\n")
	}

	signalwire.Log.Info("End Program")
}
