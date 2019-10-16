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
	log "github.com/sirupsen/logrus"
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

// MyOnRecordFinished ran when Record Action finishes
func MyOnRecordFinished(recordAction *signalwire.RecordAction) {
	if recordAction.State == signalwire.RecordFinished {
		log.Printf("Recording audio stopped.\n")
	}

	log.Infof("Recording is at: %s\n", recordAction.Result.URL)
	log.Infof("Recording Duration: %d\n", recordAction.Result.Duration)
	log.Infof("Recording File Size: %d\n", recordAction.Result.Size)
}

// MyOnRecordRecording ran when Recording starts on the call
func MyOnRecordRecording(recordAction *signalwire.RecordAction) {
	if recordAction.State == signalwire.RecordRecording {
		log.Printf("Recording audio\n")
	}
}

// MyOnRecordStateChange ran when Record State changes, eg: Recording->Finished
func MyOnRecordStateChange(recordAction *signalwire.RecordAction) {
	log.Printf("Recording State changed.\n")

	switch recordAction.State {
	case signalwire.RecordRecording:
	case signalwire.RecordFinished:
	case signalwire.RecordNoInput:
	}
}

// MyReady - gets executed when Blade is successfully setup (after signalwire.receive)
func MyReady(consumer *signalwire.Consumer) {
	log.Printf("calling out...\n")

	fromNumber := "+132XXXXXXXX"

	var toNumber = "+166XXXXXXXX"

	if len(CallThisNumber) > 0 {
		toNumber = CallThisNumber
	}

	resultDial := consumer.Client.Calling.DialPhone(fromNumber, toNumber)
	if !resultDial.Successful {
		if err := consumer.Stop(); err != nil {
			log.Errorf("Error occurred while trying to stop Consumer")
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
	rec.Direction = "both"
	rec.InitialTimeout = 10
	rec.EndSilenceTimeout = 3
	rec.Terminators = "#*"

	/*example assumes audio will be played from the callee side. */
	recordAction, err := resultDial.Call.RecordAudioAsync(&rec)
	if err != nil {
		log.Errorf("Error occurred while trying to record audio")
	}

	var rec2 signalwire.RecordParams

	rec2.Beep = true
	rec2.Format = "wav"
	rec2.Stereo = true
	rec2.Direction = "both"
	rec2.InitialTimeout = 10
	rec2.EndSilenceTimeout = 3
	rec2.Terminators = "#*"

	recordAction2, err2 := resultDial.Call.RecordAudioAsync(&rec2)
	if err2 != nil {
		log.Errorf("Error occurred while trying to record audio")
	}

	go spinner(100 * time.Millisecond)
	time.Sleep(3 * time.Second)

	log.Infof("Stopping first recording...\n")
	recordAction.Stop()

	for {
		time.Sleep(1 * time.Second)

		if recordAction.Completed {
			break
		}
	}

	log.Infof("...Done.\n")

	time.Sleep(5 * time.Second)
	log.Infof("Stopping second recording...\n")
	recordAction2.Stop()

	for {
		time.Sleep(1 * time.Second)

		if recordAction2.Completed {
			break
		}
	}

	log.Infof("...Done.\n")

	if err := resultDial.Call.Hangup(); err != nil {
		log.Errorf("Error occurred while trying to hangup call. Err: %v\n", err)
	}

	if err := consumer.Stop(); err != nil {
		log.Errorf("Error occurred while trying to stop Consumer. Err: %v\n", err)
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
		log.SetLevel(log.DebugLevel)
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
				log.Printf("Exit")
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
		log.Errorf("Error occurred while starting Signalwire Consumer")
	}
}
