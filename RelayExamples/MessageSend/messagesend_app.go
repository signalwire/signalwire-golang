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
	FromNumber     = os.Getenv("FromNumber")
	ToNumber       = os.Getenv("ToNumber")
	// SDK will use default if not set
	Host = os.Getenv("Host")
)

// Contexts a list with Signalwire Contexts
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
	signalwire.GlobalOverwriteHost = Host
	// setup the Client
	consumer.Setup(PProjectID, PTokenID, Contexts)
	// register callback
	consumer.Ready = func(consumer *signalwire.Consumer) {
		go spinner(100 * time.Millisecond)

		/*prepare the msg first, then send*/
		text := "Hello from Signalwire !"

		if len(DefaultContext) == 0 {
			DefaultContext = "test"
		}

		context := DefaultContext

		if len(FromNumber) == 0 {
			FromNumber = "+132XXXXXXXX" // edit to set FromNumber if not set through env
		}

		if len(ToNumber) == 0 {
			ToNumber = "+166XXXXXXXX" // edit to set ToNumber if not set through env
		}

		message := consumer.Client.Messaging.NewMessage(context, FromNumber, ToNumber, text)
		message.OnMessageQueued = func(_ *signalwire.SendResult) {
			signalwire.Log.Info("Message Queued.\n")
		}

		message.OnMessageDelivered = func(_ *signalwire.SendResult) {
			signalwire.Log.Info("Message Delivered.\n")
		}

		signalwire.Log.Info("Sending first message using SendMsg()...\n")

		resultSend1 := consumer.Client.Messaging.SendMsg(message)

		if !resultSend1.GetSuccessful() {
			signalwire.Log.Error("Could not send message: %v\n", resultSend1.GetReason())
		} else {
			signalwire.Log.Info("Delivered. MsgID: %v\n", resultSend1.GetMsgID())
		}

		/* now just send a message using Send() with params */

		signalwire.Log.Info("Sending second message using Send()...\n")

		// may be marked as Spam
		resultSend2 := consumer.Client.Messaging.Send(FromNumber, ToNumber, context, "Hello again from Signalwire !")

		if !resultSend2.GetSuccessful() {
			signalwire.Log.Error("Could not send message: %v\n", resultSend2.GetReason())
		} else {
			signalwire.Log.Info("Delivered. MsgID: %v\n", resultSend2.GetMsgID())
		}

		// third Message: MMS
		text = "Nice pics"
		messageMms := consumer.Client.Messaging.NewMessage(context, FromNumber, ToNumber, text)
		list := []string{"https://spdf.gsfc.nasa.gov/pub/misc/photo_gallery/image/astro/hst_abell2218.jpg",
			"https://spdf.gsfc.nasa.gov/pub/misc/photo_gallery/image/astro/hst_antennae_9734a.jpg"}
		messageMms.SetMedia(list)
		messageMms.OnMessageQueued = func(_ *signalwire.SendResult) {
			signalwire.Log.Info("Message Queued.\n")
		}

		messageMms.OnMessageDelivered = func(_ *signalwire.SendResult) {
			signalwire.Log.Info("Message Delivered.\n")
		}

		signalwire.Log.Info("Sending third message using SendMsg()...\n")

		resultSend3 := consumer.Client.Messaging.SendMsg(messageMms)

		if !resultSend3.GetSuccessful() {
			signalwire.Log.Error("Could not send message: %v\n", resultSend3.GetReason())
		} else {
			signalwire.Log.Info("Delivered. MsgID: %v\n", resultSend3.GetMsgID())
		}

		if err := consumer.Stop(); err != nil {
			signalwire.Log.Error("Error occurred while stopping Signalwire Consumer: %v\n", err)
		}
	}
	// start
	if err := consumer.Run(); err != nil {
		signalwire.Log.Error("Error occurred while starting Signalwire Consumer: %v\n", err)
	}
}
