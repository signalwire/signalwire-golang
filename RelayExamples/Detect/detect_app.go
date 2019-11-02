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
	ProjectID      = os.Getenv("ProjectID")
	TokenID        = os.Getenv("TokenID")
	DefaultContext = os.Getenv("DefaultContext")
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

// MyOnDetectFinished ran when Detect Action finishes
func MyOnDetectFinished(_ interface{}) {
	signalwire.Log.Info("Detect finished.\n")
}

// MyOnDetectUpdate ran on Detector update
func MyOnDetectUpdate(det interface{}) {
	signalwire.Log.Info("Detect update.\n")

	detectAction, ok := det.(*signalwire.DetectMachineAction)
	if ok {
		signalwire.Log.Info("Machine Detect Action.\n")
		// stop the Machine detector if READY
		if detectAction.GetDetectorEvent() == signalwire.DetectMachineReady {
			signalwire.Log.Info("Machine READY.\n")
			detectAction.Stop()
		}
	}

	_, ok2 := det.(*signalwire.DetectFaxAction)
	if ok2 {
		signalwire.Log.Info("Fax Detect Action.\n")
	}

	_, ok3 := det.(*signalwire.DetectDigitAction)
	if ok3 {
		signalwire.Log.Info("Digits Detect Action.\n")
	}
}

// MyOnDetectError ran on Detector error
func MyOnDetectError(_ interface{}) {
	signalwire.Log.Error("Detect error.\n")
}

// MyReady - gets executed when Blade is successfully setup (after signalwire.receive)
func MyReady(consumer *signalwire.Consumer) {
	signalwire.Log.Info("calling out...\n")

	fromNumber := "+132XXXXXXXX"

	var toNumber = "+132XXXXXXXX"

	if len(CallThisNumber) > 0 {
		toNumber = CallThisNumber
	}

	resultDial := consumer.Client.Calling.DialPhone(fromNumber, toNumber)
	if !resultDial.Successful {
		if err := consumer.Stop(); err != nil {
			signalwire.Log.Error("Error occurred while trying to stop Consumer")
		}

		return
	}

	resultDial.Call.OnDetectError = MyOnDetectError
	resultDial.Call.OnDetectFinished = MyOnDetectFinished
	resultDial.Call.OnDetectUpdate = MyOnDetectUpdate

	var det signalwire.DetectMachineParams

	detectMachineAction, err := resultDial.Call.DetectMachineAsync(&det)
	if err != nil {
		signalwire.Log.Error("Error occurred while trying to start answering machine detector")
	}

	var det2 signalwire.DetectDigitParams
	detectDigitAction, err := resultDial.Call.DetectDigitAsync(&det2)

	if err != nil {
		signalwire.Log.Error("Error occurred while trying to start digit detector")
	}

	var det3 signalwire.DetectFaxParams
	det3.Tone = "CED"
	detectFaxAction, err := resultDial.Call.DetectFaxAsync(&det3)

	if err != nil {
		signalwire.Log.Error("Error occurred while trying to start fax detector")
	}

	signalwire.Log.Info("Detecting...")

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
			signalwire.Log.Info("Machine Detection Successful(%v) State %v\n", detectMachineAction.Result.Successful, detectMachineAction.Event.String())
			break
		}

		signalwire.Log.Info("Last Machine event: %s", detectMachineAction.GetDetectorEvent().String())
	}

	for {
		time.Sleep(1 * time.Second)

		if detectDigitAction.GetCompleted() {
			signalwire.Log.Info("Digit Detection Successful(%v) State %v\n", detectDigitAction.Result.Successful, detectDigitAction.Event.String())
			break
		}

		signalwire.Log.Info("Last Digit event: %s", detectDigitAction.GetDetectorEvent().String())
	}

	for {
		time.Sleep(1 * time.Second)

		if detectFaxAction.GetCompleted() {
			signalwire.Log.Info("Fax Detection Successful(%v) State %v\n", detectFaxAction.Result.Successful, detectFaxAction.Event.String())
			break
		}

		signalwire.Log.Info("Last Fax event: %s", detectFaxAction.GetDetectorEvent().String())
	}

	if resultDial.Call.GetCallState() != signalwire.Ending && resultDial.Call.GetCallState() != signalwire.Ended {
		if _, err := resultDial.Call.Hangup(); err != nil {
			signalwire.Log.Error("Error occurred while trying to hangup call. Err: %v\n", err)
		}
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

	consumer := signalwire.NewConsumer()
	// setup the Client
	consumer.Setup(PProjectID, PTokenID, Contexts)
	// register callback
	consumer.Ready = MyReady
	// start
	if err := consumer.Run(); err != nil {
		consumer.Log.Error("Error occurred while starting Signalwire Consumer")
	}
}
