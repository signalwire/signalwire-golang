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
	ProjectID = "replaceme"
	TokenID   = "replaceme" // nolint: gosec
)

// Contexts not needed for only outbound calls
var Contexts = []string{"replaceme"}

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

// MyOnPlayFinished ran when Play Action finishes
func MyOnPlayFinished(playAction *signalwire.PlayAction) {
	if playAction.State == signalwire.PlayFinished {
		signalwire.Log.Info("Playing audio stopped.\n")
	}
}

// MyOnPlayPlaying ran when Playing starts on the call
func MyOnPlayPlaying(playAction *signalwire.PlayAction) {
	if playAction.State == signalwire.PlayPlaying {
		signalwire.Log.Info("Playing audio\n")
	}
}

// MyOnPlayStateChange ran when Play State changes, eg: Playing->Finished
func MyOnPlayStateChange(playAction *signalwire.PlayAction) {
	signalwire.Log.Info("Playing State changed.\n")

	switch playAction.State {
	case signalwire.PlayPlaying:
	case signalwire.PlayFinished:
	case signalwire.PlayError:
	}
}

// MyOnRecordFinished ran when Record Action finishes
func MyOnRecordFinished(recordAction *signalwire.RecordAction) {
	if recordAction.State == signalwire.RecordFinished {
		signalwire.Log.Info("Recording audio stopped.\n")
	}

	signalwire.Log.Info("Recording is at: %s\n", recordAction.Result.URL)
	signalwire.Log.Info("Recording Duration: %d\n", recordAction.Result.Duration)
	signalwire.Log.Info("Recording File Size: %d\n", recordAction.Result.Size)
}

// MyReady - gets executed when Blade is successfully setup (after signalwire.receive)
func MyReady(consumer *signalwire.Consumer) {
	signalwire.Log.Info("calling out...\n")

	fromNumber := "+132XXXXXXXX"

	var toNumber = "+166XXXXXXXX"

	if len(CallThisNumber) > 0 {
		toNumber = CallThisNumber
	}

	resultDial := consumer.Client.Calling.DialPhone(fromNumber, toNumber)
	if !resultDial.Successful {
		if err := consumer.Stop(); err != nil {
			signalwire.Log.Error("Error occurred while trying to stop Consumer\n")
		}

		return
	}

	resultDial.Call.OnPlayPlaying = MyOnPlayPlaying
	resultDial.Call.OnPlayFinished = MyOnPlayFinished
	resultDial.Call.OnPlayStateChange = MyOnPlayStateChange

	resultDial.Call.OnRecordFinished = MyOnRecordFinished

	var rec signalwire.RecordParams

	rec.Beep = true
	rec.Format = "wav"
	rec.Stereo = false
	rec.Direction = "both"
	rec.InitialTimeout = 10
	rec.EndSilenceTimeout = 3
	rec.Terminators = "#*"

	_, err := resultDial.Call.RecordAudioAsync(&rec)
	if err != nil {
		signalwire.Log.Error("Error occurred while trying to record audio\n")
	}

	PlayAction, err := resultDial.Call.PlayAudioAsync("https://www.phatdrumloops.com/audio/wav/space_funk1.wav")
	if err != nil {
		signalwire.Log.Error("Error occurred while trying to Play audio\n")
	}

	PlayAction2, err2 := resultDial.Call.PlayAudioAsync("https://www.phatdrumloops.com/audio/wav/smokinatt2.wav")
	if err2 != nil {
		signalwire.Log.Error("Error occurred while trying to Play audio\n")
	}

	_, err3 := resultDial.Call.PlayTTSAsync("Welcome to Signalwire !", "en-US", "female")
	if err3 != nil {
		signalwire.Log.Error("Error occurred while trying to Play audio\n")
	}

	go spinner(100 * time.Millisecond)
	time.Sleep(2 * time.Second)

	signalwire.Log.Info("Stopping first Playing...\n")

	if !PlayAction.Completed {
		PlayAction.Stop()
	}

	signalwire.Log.Debug("App - Play finished. ctrlID: %s res [%p] Completed [%v] Successful [%v]\n", PlayAction2.ControlID, PlayAction, PlayAction.Completed, PlayAction.Result.Successful)

	for ok := true; ok; ok = !(PlayAction.State == signalwire.PlayFinished) {
		signalwire.Log.Info("Completed 1: %v\n", PlayAction.GetCompleted())
		time.Sleep(1 * time.Second)
	}

	signalwire.Log.Info("...Done.\n")

	time.Sleep(2 * time.Second)

	signalwire.Log.Info("Stopping second Playing...\n")

	if !PlayAction2.Completed {
		PlayAction2.Stop()
	}

	signalwire.Log.Debug("App2 - Play finished. ctrlID: %s res [%p] Completed [%v] Successful [%v]\n", PlayAction2.ControlID, PlayAction2, PlayAction2.Completed, PlayAction2.Result.Successful)

	for ok := true; ok; ok = !(PlayAction2.State == signalwire.PlayFinished) {
		signalwire.Log.Info("Completed 2: %v\n", PlayAction2.GetCompleted())
		time.Sleep(1 * time.Second)
	}

	signalwire.Log.Info("...Done.\n")

	if _, err := resultDial.Call.Hangup(); err != nil {
		signalwire.Log.Error("Error occurred while trying to hangup call. Err: %v\n", err)
	}

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
	// setup the Client
	consumer.Setup(PProjectID, PTokenID, Contexts)
	// register callback
	consumer.Ready = MyReady
	// start
	if err := consumer.Run(); err != nil {
		signalwire.Log.Error("Error occurred while starting Signalwire Consumer\n")
	}
}
