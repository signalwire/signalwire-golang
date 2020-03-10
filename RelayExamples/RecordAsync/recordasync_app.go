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

// App environment settings
var (
	// required
	ProjectID = os.Getenv("ProjectID")
	TokenID   = os.Getenv("TokenID")
	// context required only for Inbound calls.
	DefaultContext = os.Getenv("DefaultContext")
	FromNumber     = os.Getenv("FromNumber")
	ToNumber       = os.Getenv("ToNumber")
	// SDK will use default if not set
	Host = os.Getenv("Host")
)

// Contexts not needed for only outbound calls
var Contexts = []string{DefaultContext}

// PProjectID passed from command-line
var PProjectID string

// PTokenID passed from command-line
var PTokenID string

// CallThisNumber get the callee phone number from command line
var CallThisNumber string

/*gopl.io spinner*/
func spinner(delay time.Duration) {
	for {
		for _, r := range `-\|/` {
			fmt.Printf("\r%c", r)
			time.Sleep(delay)
		}
	}
}

// MyOnRecordFinished ran when Record Action finishes
func MyOnRecordFinished(recordAction *signalwire.RecordAction) {
	if recordAction.State == signalwire.RecordFinished {
		signalwire.Log.Info("Recording audio stopped.\n")
	}
}

// MyOnRecordRecording ran when Recording starts on the call
func MyOnRecordRecording(recordAction *signalwire.RecordAction) {
	if recordAction.State == signalwire.RecordRecording {
		signalwire.Log.Info("Recording audio\n")
	}
}

// MyOnRecordStateChange ran when Record State changes, eg: Recording->Finished
func MyOnRecordStateChange(recordAction *signalwire.RecordAction) {
	signalwire.Log.Info("Recording State changed.\n")

	switch recordAction.State {
	case signalwire.RecordRecording:
	case signalwire.RecordFinished:
	case signalwire.RecordNoInput:
	}
}

// MyReady - gets executed when Blade is successfully setup (after signalwire.receive)
func MyReady(consumer *signalwire.Consumer) {
	signalwire.Log.Info("calling out...\n")

	if len(FromNumber) == 0 {
		FromNumber = "+132XXXXXXXX" // edit to set FromNumber if not set through env
	}

	if len(ToNumber) == 0 {
		ToNumber = "+166XXXXXXXX" // edit to set ToNumber if not set through env
	}

	if len(CallThisNumber) > 0 {
		ToNumber = CallThisNumber
	}

	resultDial := consumer.Client.Calling.DialPhone(FromNumber, ToNumber)
	if !resultDial.Successful {
		if err := consumer.Stop(); err != nil {
			signalwire.Log.Error("Error occurred while trying to stop Consumer\n")
		}

		return
	}

	resultDial.Call.OnRecordRecording = MyOnRecordRecording
	resultDial.Call.OnRecordFinished = MyOnRecordFinished
	resultDial.Call.OnRecordStateChange = MyOnRecordStateChange

	var rec signalwire.RecordParams

	rec.Beep = true
	rec.Format = "wav"
	rec.Stereo = false
	rec.Direction = signalwire.RecordDirectionBoth.String()
	rec.InitialTimeout = 10
	rec.EndSilenceTimeout = 3
	rec.Terminators = "#*"

	recordAction, err := resultDial.Call.RecordAudioAsync(&rec)
	if err != nil {
		signalwire.Log.Error("Error occurred while trying to record audio\n")
	}

	if _, err := resultDial.Call.PlayTTSAsync("Welcome to Signalwire!", "en-US", "female"); err != nil {
		signalwire.Log.Error("Error occurred while trying to play TTS. Err: %v\n", err)
	}

	go spinner(100 * time.Millisecond)
	time.Sleep(3 * time.Second)
	/*
		for !recordAction.GetCompleted() {
			// if the callee does not play/say anything the recording would stop on timeout, after EndSilenceTimeout
			time.Sleep(1 * time.Second)
		}
	*/

	signalwire.Log.Info("Stopping recording.\n")
	recordAction.Stop()

	if _, err := resultDial.Call.PlaySilence(3); err != nil {
		signalwire.Log.Error("Error occurred while trying to play silence. Err: %v\n", err)
	}

	if _, err := resultDial.Call.Hangup(); err != nil {
		signalwire.Log.Error("Error occurred while trying to hangup call. Err: %v\n", err)
	}

	signalwire.Log.Info("Recording is at: %s\n", recordAction.URL)

	if err := consumer.Stop(); err != nil {
		signalwire.Log.Error("Error occurred while trying to stop Consumer. Err: %v\n", err)
	}
}

func main() {
	var printVersion bool

	var verbose bool

	flag.BoolVar(&printVersion, "v", false, " Show version ")
	flag.StringVar(&CallThisNumber, "n", "", " Number to call ")
	flag.StringVar(&PProjectID, "p", ProjectID, " ProjectID ")
	flag.StringVar(&PTokenID, "t", TokenID, " TokenID ")
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
				signalwire.Log.Info("Exit")
				os.Exit(0)
			}
		}
	}()

	consumer := new(signalwire.Consumer)
	signalwire.GlobalOverwriteHost = Host
	// setup the Client
	consumer.Setup(PProjectID, PTokenID, Contexts)
	// register callback
	consumer.Ready = MyReady
	// start
	if err := consumer.Run(); err != nil {
		signalwire.Log.Error("Error occurred while starting Signalwire Consumer: %v\n", err)
	}
}
