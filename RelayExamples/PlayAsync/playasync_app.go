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

// MyOnPlayFinished ran when Play Action finishes
func MyOnPlayFinished(playAction *signalwire.PlayAction) {
	if playAction.State == signalwire.PlayFinished {
		log.Printf("Playing audio stopped.\n")
	}
}

// MyOnPlayPlaying ran when Playing starts on the call
func MyOnPlayPlaying(playAction *signalwire.PlayAction) {
	if playAction.State == signalwire.PlayPlaying {
		log.Printf("Playing audio\n")
	}
}

// MyOnPlayStateChange ran when Play State changes, eg: Playing->Finished
func MyOnPlayStateChange(playAction *signalwire.PlayAction) {
	log.Printf("Play State changed.\n")

	switch playAction.State {
	case signalwire.PlayPlaying:
	case signalwire.PlayFinished:
	case signalwire.PlayError:
	case signalwire.PlayPaused:
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

	resultDial.Call.OnPlayPlaying = MyOnPlayPlaying
	resultDial.Call.OnPlayFinished = MyOnPlayFinished
	resultDial.Call.OnPlayStateChange = MyOnPlayStateChange

	playAction, err := resultDial.Call.PlayAudioAsync("https://cdn.signalwire.com/default-music/welcome.mp3")
	if err != nil {
		log.Errorf("Error occurred while trying to play audio")
	}

	// do something here
	go spinner(100 * time.Millisecond)
	time.Sleep(3 * time.Second)

	playAction2, err := resultDial.Call.PlayTTSAsync("Hello from Signalwire!", "en-US", "female")
	if err != nil {
		log.Errorf("Error occurred while trying to play tts")
	}

	playAction.Stop()

	// do something more here
	time.Sleep(3 * time.Second)

	playAction2.Stop()

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
