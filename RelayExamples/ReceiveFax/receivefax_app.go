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

// MyOnFaxFinished ran when Faxing Action finishes
func MyOnFaxFinished(faxAction *signalwire.FaxAction) {
	log.Printf("Faxing finished.\n")
}

// MyOnFaxPage ran when a document page is sent/received
func MyOnFaxPage(faxAction *signalwire.FaxAction) {
	log.Printf("Fax page event\n")
}

// MyOnFaxError
func MyOnFaxError(faxAction *signalwire.FaxAction) {
	log.Printf("Faxing error.\n")
}

// MyOnIncomingCall - gets executed when we receive an incoming call
func MyOnIncomingCall(consumer *signalwire.Consumer, call *signalwire.CallObj) {
	resultAnswer := call.Answer()
	if !resultAnswer.Successful {
		if err := consumer.Stop(); err != nil {
			log.Errorf("Error occurred while trying to stop Consumer")
		}

		return
	}

	call.OnFaxError = MyOnFaxError
	call.OnFaxFinished = MyOnFaxFinished
	call.OnFaxPage = MyOnFaxPage

	// do something here
	go spinner(100 * time.Millisecond)

	faxResult, err := call.ReceiveFax()
	if err != nil {
		log.Errorf("Error occurred while trying to receive fax")
	}

	log.Infof("Download Document from %s\n Pages #%d\n", faxResult.Document, faxResult.Pages)

	if err := call.Hangup(); err != nil {
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
	consumer.OnIncomingCall = MyOnIncomingCall

	log.Info("Wait incoming call..")
	// start
	if err := consumer.Run(); err != nil {
		log.Errorf("Error occurred while starting Signalwire Consumer")
	}
}
