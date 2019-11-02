package main

import (
	"encoding/json"
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

type PassToTask struct {
	Foo uint   `json:"foo"`
	Bar string `json:"bar"`
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
	consumer.Ready = func(consumer *signalwire.Consumer) {
		go spinner(100 * time.Millisecond)

		done := make(chan struct{}, 1)

		consumer.OnTask = func(_ *signalwire.Consumer, ev signalwire.ParamsEventTaskingTask) {
			signalwire.Log.Info("Task Delivered. %v\n", ev)

			go func() {
				done <- struct{}{}
			}()
		}

		taskMsg := PassToTask{
			Foo: 123,
			Bar: "baz",
		}

		byteArray, err := json.MarshalIndent(taskMsg, "", "  ")
		if err != nil {
			signalwire.Log.Error("%v", err)
			return
		}

		signalwire.Log.Info(string(byteArray))

		if result := consumer.Client.Tasking.Deliver(DefaultContext, byteArray); !result {
			signalwire.Log.Error("Could not deliver task\n")

			go func() {
				done <- struct{}{}
			}()
		}

		// stop when task has been delivered or on error
		<-done

		if err := consumer.Stop(); err != nil {
			signalwire.Log.Error("Error occurred while stopping Signalwire Consumer\n")
		}
	}
	// start
	if err := consumer.Run(); err != nil {
		signalwire.Log.Error("Error occurred while starting Signalwire Consumer\n")
	}
}
