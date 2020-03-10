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
	ProjectID      = os.Getenv("ProjectID")
	TokenID        = os.Getenv("TokenID")
	DefaultContext = os.Getenv("DefaultContext")
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

// MyOnFaxFinished ran when Faxing Action finishes
func MyOnFaxFinished(_ *signalwire.FaxAction) {
	signalwire.Log.Info("Faxing finished.\n")
}

// MyOnFaxPage ran when a document page is sent/received
func MyOnFaxPage(_ *signalwire.FaxAction) {
	signalwire.Log.Info("Fax page event\n")
}

// MyOnFaxError used for logging
func MyOnFaxError(_ *signalwire.FaxAction) {
	signalwire.Log.Info("Faxing error.\n")
}

// MyOnIncomingCall - gets executed when we receive an incoming call
func MyOnIncomingCall(consumer *signalwire.Consumer, call *signalwire.CallObj) {
	resultAnswer, _ := call.Answer()
	if !resultAnswer.Successful {
		if err := consumer.Stop(); err != nil {
			signalwire.Log.Error("Error occurred while trying to stop Consumer\n")
		}

		return
	}

	call.OnFaxError = MyOnFaxError
	call.OnFaxFinished = MyOnFaxFinished
	call.OnFaxPage = MyOnFaxPage

	go spinner(100 * time.Millisecond)

	faxResult, err := call.ReceiveFax()
	if err != nil {
		signalwire.Log.Error("Error occurred while trying to receive fax\n")
	}

	signalwire.Log.Info("Download Document from %s\n Pages #%d\n", faxResult.Document, faxResult.Pages)

	if _, err := call.Hangup(); err != nil {
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
	signalwire.GlobalOverwriteHost = Host
	// setup the Client
	consumer.Setup(PProjectID, PTokenID, Contexts)
	// register callback
	consumer.OnIncomingCall = MyOnIncomingCall

	signalwire.Log.Info("Wait incoming call..")
	// start
	if err := consumer.Run(); err != nil {
		signalwire.Log.Error("Error occurred while starting Signalwire Consumer: %v\n", err)
	}
}
