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

// MyOnReady - gets executed when Blade (Client) is successfully setup (after signalwire.receive)
func MyOnReady(client *signalwire.ClientSession) {
	if err := client.Disconnect(); err != nil {
		signalwire.Log.Error("Error occurred while trying to stop Client\n")
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
				signalwire.Log.Info("Exit\n")
				os.Exit(0)
			}
		}
	}()

	go spinner(100 * time.Millisecond)

	var counter int

	var max int = 100

	var maxroutines int = 10

	var i int

	for i = 0; i < maxroutines; i++ {
		go func() {
			for {
				// connect/disconnect stress
				counter++

				signalwireContexts := []string{"replaceme"}

				client := signalwire.Client(PProjectID, PTokenID, "" /*host, empty for default*/, signalwireContexts)
				// register callback
				client.OnReady = MyOnReady
				// start
				err := client.Connect()

				if err != nil {
					signalwire.Log.Error("Error occurred while starting Signalwire Client\n")
				}

				if counter == max {
					break
				}
			}
		}()
	}

	/*keep the program running*/
	<-make(chan struct{})
}
