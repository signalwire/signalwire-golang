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
var Contexts = []string{"test-golang"}

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

// MyOnDetectFinished ran when Detect Action finishes
func MyOnDetectFinished(_ interface{}) {
	log.Printf("Detect finished.\n")
}

// MyOnDetectUpdate ran on Detector update
func MyOnDetectUpdate(_ interface{}) {
	log.Printf("Detect update.\n")
}

// MyOnDetectError ran on Detector error
func MyOnDetectError(_ interface{}) {
	log.Errorf("Detect error.\n")
}

// MyReady - gets executed when Blade is successfully setup (after signalwire.receive)
func MyReady(consumer *signalwire.Consumer) {
	log.Printf("calling out...\n")

	fromNumber := "+132XXXXXXXX"

	var toNumber = "+132XXXXXXXX"

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

	resultDial.Call.OnDetectError = MyOnDetectError
	resultDial.Call.OnDetectFinished = MyOnDetectFinished
	resultDial.Call.OnDetectUpdate = MyOnDetectUpdate

	var det signalwire.DetectMachineParams

	detectMachineAction, err := resultDial.Call.DetectMachineAsync(&det)
	if err != nil {
		log.Errorf("Error occurred while trying to start answering machine detector")
	}

	var det2 signalwire.DetectDigitParams
	detectDigitAction, err := resultDial.Call.DetectDigitAsync(&det2)

	if err != nil {
		log.Errorf("Error occurred while trying to start digit detector")
	}

	var det3 signalwire.DetectFaxParams
	det3.Tone = "CED"
	detectFaxAction, err := resultDial.Call.DetectFaxAsync(&det3)

	if err != nil {
		log.Errorf("Error occurred while trying to start fax detector")
	}

	log.Printf("Detecting...")

	go spinner(100 * time.Millisecond)

	time.Sleep(15 * time.Second)

	if !detectFaxAction.GetCompleted() {
		detectFaxAction.Stop()
	}

	if !detectMachineAction.GetCompleted() {
		detectMachineAction.Stop()
	}

	if !detectDigitAction.GetCompleted() {
		detectDigitAction.Stop()
	}

	for {
		time.Sleep(1 * time.Second)

		if detectMachineAction.GetCompleted() {
			log.Printf("Machine Detection Successful(%v) State %v\n", detectMachineAction.Result.Successful, detectMachineAction.Event.String())
			break
		}

		log.Printf("Last Machine event: %s", detectMachineAction.GetEvent().String())
	}

	for {
		time.Sleep(1 * time.Second)

		if detectDigitAction.GetCompleted() {
			log.Printf("Digit Detection Successful(%v) State %v\n", detectDigitAction.Result.Successful, detectDigitAction.Event.String())
			break
		}

		log.Printf("Last Digit event: %s", detectDigitAction.GetEvent().String())
	}

	for {
		time.Sleep(1 * time.Second)

		if detectFaxAction.GetCompleted() {
			log.Printf("Fax Detection Successful(%v) State %v\n", detectFaxAction.Result.Successful, detectFaxAction.Event.String())
			break
		}

		log.Printf("Last Fax event: %s", detectFaxAction.GetEvent().String())
	}

	if resultDial.Call.GetCallState() != signalwire.Ending && resultDial.Call.GetCallState() != signalwire.Ended {
		if err := resultDial.Call.Hangup(); err != nil {
			log.Errorf("Error occurred while trying to hangup call. Err: %v\n", err)
		}
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
