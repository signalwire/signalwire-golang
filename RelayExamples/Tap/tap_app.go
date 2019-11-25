package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/pion/rtp"
	"github.com/signalwire/signalwire-golang/signalwire"
	"github.com/zaf/g711"
	"gopkg.in/hraban/opus.v2"
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

var listenAddress = "127.0.0.1" // replace this

var listenPort = 1234

func depak(rawpacket []byte) (*[]byte, error) {
	packet := new(rtp.Packet)
	err := packet.Unmarshal(rawpacket)

	if err != nil {
		return nil, err
	}

	signalwire.Log.Info("%s", packet.String())

	return &packet.Payload, nil
}

func rtpListen(codec string, ptime uint8, rate uint) {
	addrStr := fmt.Sprintf("%s:%d", listenAddress, listenPort)

	conn, err := net.ListenPacket("udp", addrStr)
	if err != nil {
		signalwire.Log.Fatal("%v", err)
	}

	defer conn.Close()

	var g711dec *g711.Decoder

	var opusdec *opus.Decoder

	b := new(bytes.Buffer)

	switch strings.ToUpper(codec) {
	case "PCMU":
		g711dec, _ = g711.NewUlawDecoder(b)
	case "PCMA":
		g711dec, _ = g711.NewAlawDecoder(b)
	case "OPUS":
		if rate == 0 {
			rate = 48000
		}

		opusdec, _ = opus.NewDecoder(int(rate), 1)
	default:
		signalwire.Log.Warn("Unknown codec")
		return
	}

	/* this file can be imported in Audacity, 8000 hz, 1 channel, Little-Endian, Signed 16 bit PCM */
	out, err := os.Create("tapaudio.raw")
	if err != nil {
		signalwire.Log.Fatal("%v", err)
	}

	defer out.Close()

	var pcm []int16

	if opusdec != nil {
		if ptime == 0 {
			ptime = 20
		}

		frameSize := uint(ptime) * rate / 1000

		pcm = make([]int16, frameSize)
	}

	for {
		buf := make([]byte, 1024)

		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			signalwire.Log.Warn("%v", err)
			continue
		}

		payload, err := depak(buf[:n])
		if err != nil {
			signalwire.Log.Fatal("%v", err)
		}

		if g711dec != nil {
			b.Write(*payload)

			_, err = io.Copy(out, g711dec)
			if err != nil {
				signalwire.Log.Fatal("Decoding failed: %v\n", err)
			}
		} else if opusdec != nil {
			data := (*payload)[:n]

			n, err := opusdec.Decode(data, pcm)
			if err != nil {
				signalwire.Log.Fatal("Decoding failed: %v\n", err)
			}
			err = binary.Write(b, binary.LittleEndian, pcm[:n])
			if err != nil {
				signalwire.Log.Fatal("binary.Write failed:", err)
			}
			_, _ = io.Copy(out, b)
		}
	}
}

/*gopl.io spinner*/
func spinner(delay time.Duration) {
	for {
		for _, r := range `-\|/` {
			fmt.Printf("\r%c", r)
			time.Sleep(delay)
		}
	}
}

// MyReady - gets executed when Blade is successfully setup (after signalwire.receive)
func MyReady(consumer *signalwire.Consumer) {
	fmt.Printf("calling out...")

	fromNumber := "+13XXXXXXXXX"

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

	signalwire.Log.Info("Tap audio...\n")

	go spinner(100 * time.Millisecond)

	var tapdevice signalwire.TapDevice
	tapdevice.Type = signalwire.TapRTP.String()
	tapdevice.Params.Addr = listenAddress
	tapdevice.Params.Port = uint16(listenPort)
	tapdevice.Params.Codec = "PCMU"

	tapAction, err := resultDial.Call.TapAudioAsync(signalwire.TapDirectionListen, &tapdevice)

	if err != nil {
		signalwire.Log.Fatal("Error occurred while trying to tap audio: %v\n", err)
	}

	go rtpListen(tapdevice.Params.Codec, tapdevice.Params.Ptime, tapdevice.Params.Rate)

	time.Sleep(10 * time.Second)
	tapAction.Stop()

	signalwire.Log.Info("Tap: %v\n", tapAction.GetTap())
	signalwire.Log.Info("SourceDevice: %v\n", tapAction.GetSourceDevice())           // comes from the Signalwire platform
	signalwire.Log.Info("DestinationDevice: %v\n", tapAction.GetDestinationDevice()) // the device passed above

	if _, err := resultDial.Call.Hangup(); err != nil {
		signalwire.Log.Error("Error occurred while trying to hangup call\n")
	}

	if err := consumer.Stop(); err != nil {
		signalwire.Log.Error("Error occurred while trying to stop Consumer\n")
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
