package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/rtp"
	"github.com/signalwire/signalwire-golang/signalwire"
	"github.com/zaf/g711"
	"gopkg.in/hraban/opus.v2"
)

// App environment settings
var (
	// required
	ProjectID = os.Getenv("ProjectID")
	TokenID   = os.Getenv("TokenID")
	// context required only for Inbound calls
	DefaultContext = os.Getenv("DefaultContext")
	// required
	ListenAddress = os.Getenv("ListenAddress")
	ListenPort    = os.Getenv("ListenPort")
	FromNumber    = os.Getenv("FromNumber")
	ToNumber      = os.Getenv("ToNumber")
	// SDK will use default if not set
	Host = os.Getenv("Host")
)

const (
	OPUS           = "OPUS"
	OGGOPUS        = "OGGOPUS"
	PCMA           = "PCMA"
	PCMU           = "PCMU"
	OGGMaxPageSize = 65307
	OGGMinPageSize = 2400 // this much data buffered before trying to open the incoming stream
	OGGDebug       = true
)

// Contexts not needed for only outbound calls
var Contexts = []string{DefaultContext}

// PProjectID passed from command-line
var PProjectID string

// PTokenID passed from command-line
var PTokenID string

// CallThisNumber get the callee phone number from command line
var CallThisNumber string

var defaultListenAddress = "127.0.0.1" // can be replaced through ENV var.

var defaultListenPort = 1234 // can be replaced through ENV var.

var address string

var port string

var ws = false

var secure = false

var justlisten = false

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
	addrStr := fmt.Sprintf("%s:%s", address, port)

	conn, err := net.ListenPacket("udp", addrStr)
	if err != nil {
		signalwire.Log.Fatal("%v", err)
	}

	defer conn.Close()

	var g711dec *g711.Decoder

	var opusdec *opus.Decoder

	b := new(bytes.Buffer)

	switch strings.ToUpper(codec) {
	case PCMU:
		g711dec, _ = g711.NewUlawDecoder(b)
	case PCMA:
		g711dec, _ = g711.NewAlawDecoder(b)
	case OPUS:
		if rate == 0 {
			rate = 48000
		}

		opusdec, _ = opus.NewDecoder(int(rate), 1)

	default:
		signalwire.Log.Warn("Unknown codec")
		return
	}

	/* this file can be imported in Audacity, 8000 hz, 1 channel, Little-Endian, Signed 16 bit PCM */
	out, err := os.Create("tapaudio_rtp_endpoint.raw")
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

func wsListen(codec string) {
	var addr = address + ":" + port

	var upgrader = websocket.Upgrader{
		ReadBufferSize: 1024,
	}

	var ptime = 20

	var rate uint

	var isstream bool

	//	var pageReader io.Reader

	var oggbuf *bytes.Buffer

	/* this file can be imported in Audacity, 8000 hz, 1 channel, Little-Endian, Signed 16 bit PCM */
	out, err := os.Create("tapaudio_ws_endpoint.raw")
	if err != nil {
		signalwire.Log.Fatal("%v", err)
	}

	defer out.Close()

	b := new(bytes.Buffer)

	var g711dec *g711.Decoder

	g711dec, _ = g711.NewUlawDecoder(b)

	var opusdec *opus.Decoder

	var opusstream *opus.Stream

	var p []byte

	switch strings.ToUpper(codec) {
	case PCMU:
		g711dec, _ = g711.NewUlawDecoder(b)
	case PCMA:
		g711dec, _ = g711.NewAlawDecoder(b)
	case OPUS:
		if rate == 0 {
			rate = 48000
		}

		opusdec, _ = opus.NewDecoder(int(rate), 1)
	case OGGOPUS:
		if rate == 0 {
			rate = 48000
		}

		upgrader = websocket.Upgrader{
			ReadBufferSize: OGGMaxPageSize,
		}

		oggbuf = new(bytes.Buffer)

		p = make([]byte, OGGMaxPageSize)

		isstream = true
	default:
		signalwire.Log.Warn("Unknown codec")
		return
	}

	var pcm []int16

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			signalwire.Log.Error("%v", err)
			return
		}

		if opusdec != nil {
			if ptime == 0 {
				ptime = 20
			}

			frameSize := uint(ptime) * rate / 1000

			pcm = make([]int16, frameSize)
		}

		rPipe, wPipe := io.Pipe()

		for {
			if isstream {
				//				done := make(chan bool)
				msgType, r, err := conn.NextReader()
				if err != nil {
					signalwire.Log.Error("%v", err)
					return
				}
				if err != nil {
					conn.Close()
					signalwire.Log.Fatal("NextReader(): %v\n", err)
				}
				if msgType != websocket.BinaryMessage {
					p := make([]byte, 1024)
					n, _ := r.Read(p)
					signalwire.Log.Info("received: %v\n", hex.Dump(p[:n]))
					w := []byte("listening")
					if errw := conn.WriteMessage(websocket.TextMessage, w); errw != nil {
						signalwire.Log.Fatal("cannot write answer to websocket: %v\n", errw)
					}
					continue
				}

				nr, _ := oggbuf.ReadFrom(r)
				signalwire.Log.Info("wrote to oggbuf [%d]\n", nr)

				if opusstream == nil && oggbuf.Len() < OGGMinPageSize {
					continue
				}

				n, _ := oggbuf.Read(p)

				go func() {
					pageReader := bytes.NewReader(p)
					io.Copy(wPipe, pageReader)
				}()

				signalwire.Log.Info("received OGG Opus page size [%d]\n", n)
				if OGGDebug {
					signalwire.Log.Info("rcvd [%d] bytes: %s\n", n, hex.Dump(p[:n]))
				}

				if opusstream == nil {
					opusstream, err = opus.NewStream(rPipe)
				}

				if opusstream == nil {
					signalwire.Log.Error("stream error [%v]\n", err)
					continue
				}

				pcmbuf := make([]int16, 8192)

				go func() {
					for {
						signalwire.Log.Info("read")
						n, err := opusstream.Read(pcmbuf)
						switch err {
						case io.EOF:
							fallthrough
						case nil:
							break
						default:
							signalwire.Log.Error("Error while decoding opus stream: %v", err)
						}
						if n == 0 {
							signalwire.Log.Info("No data, buffering.")
							break
						}
						signalwire.Log.Info("decoded OGG Opus [%v] bytes\n", n)
						err = binary.Write(b, binary.LittleEndian, pcmbuf[:n])
						if err != nil {
							signalwire.Log.Fatal("binary.Write failed:", err)
						}
						_, _ = io.Copy(out, b)
					}
				}()
			} else {
				msgType, msg, err := conn.ReadMessage()
				if err != nil {
					signalwire.Log.Error("%v", err)
					return
				}
				if msgType != websocket.BinaryMessage || len(msg) == 0 {
					continue
				}

				if g711dec != nil {
					b.Write(msg)

					_, err = io.Copy(out, g711dec)
					if err != nil {
						signalwire.Log.Fatal("Decoding failed: %v\n", err)
					}
				} else {
					signalwire.Log.Info("received Opus packet: %s\n", hex.Dump(msg))
					n, err := opusdec.Decode(msg, pcm)
					if err != nil {
						signalwire.Log.Fatal("Decoding failed: %v\n", err)
					}
					err = binary.Write(b, binary.LittleEndian, pcm[:n])
					if err != nil {
						signalwire.Log.Fatal("binary.Write failed:", err)
					}
					_, _ = io.Copy(out, b)
					signalwire.Log.Info("%s rcvd: %v bytes\n", conn.RemoteAddr(), len(msg))
				}
				signalwire.Log.Info("%s rcvd: %v bytes\n", conn.RemoteAddr(), len(msg))
			}
		}
	})

	if !secure {
		signalwire.Log.Fatal("%v", http.ListenAndServe(addr, nil))
	} else {
		// openssl genrsa -out tapserver.key 2048
		// openssl ecparam -genkey -name secp384r1 -out tapserver.key
		// don't forget to set FQDN (IP) !
		// openssl req -new -x509 -sha256 -key tapserver.key -out tapserver.crt -days 3650
		_, err := os.Stat("tapserver.key")
		if os.IsNotExist(err) {
			signalwire.Log.Fatal("%v", err)
		}
		_, err = os.Stat("tapserver.crt")
		if os.IsNotExist(err) {
			signalwire.Log.Fatal("%v", err)
		}

		signalwire.Log.Fatal("%v", http.ListenAndServeTLS(addr, "tapserver.crt", "tapserver.key", nil))
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

	if len(ListenAddress) > 0 {
		address = ListenAddress // env
	} else {
		address = defaultListenAddress
	}

	if len(ListenPort) > 0 {
		port = ListenPort // env
	} else {
		port = fmt.Sprintf("%d", defaultListenPort)
	}

	if len(CallThisNumber) > 0 {
		FromNumber = CallThisNumber
	}

	resultDial := consumer.Client.Calling.DialPhone(FromNumber, ToNumber)
	if !resultDial.Successful {
		if err := consumer.Stop(); err != nil {
			signalwire.Log.Error("Error occurred while trying to stop Consumer\n")
		}

		return
	}

	_, err := resultDial.Call.PlayAudioAsync("https://cdn.signalwire.com/default-music/welcome.mp3")
	if err != nil {
		signalwire.Log.Error("Error occurred while trying to play audio\n")
	}

	signalwire.Log.Info("Tap audio...\n")

	go spinner(100 * time.Millisecond)

	var tapdevice signalwire.TapDevice
	if !ws {
		tapdevice.Type = signalwire.TapRTP.String()
		tapdevice.Params.Addr = address

		listenPort, errx := strconv.Atoi(port)
		if errx != nil {
			signalwire.Log.Fatal("invalid port.")
		}

		tapdevice.Params.Port = uint16(listenPort)
	} else {
		tapdevice.Type = signalwire.TapWS.String()
		if !secure {
			tapdevice.Params.URI = "ws://"
		} else {
			tapdevice.Params.URI = "wss://"
		}
		tapdevice.Params.URI = tapdevice.Params.URI + address + ":" + port
	}

	tapdevice.Params.Codec = OPUS

	if !ws {
		go rtpListen(tapdevice.Params.Codec, tapdevice.Params.Ptime, tapdevice.Params.Rate)
	} else {
		go wsListen(tapdevice.Params.Codec)
	}

	/* direction can be TapDirectionListen, TapDirectionSpeak or TapDirectionBoth */
	tapAction, err := resultDial.Call.TapAudioAsync(signalwire.TapDirectionListen, &tapdevice)

	if err != nil {
		signalwire.Log.Fatal("Error occurred while trying to tap audio: %v\n", err)
	}

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
	flag.BoolVar(&ws, "w", false, " Enable websocket tap ")                     // mutually exclusive with the RTP tap
	flag.BoolVar(&secure, "s", false, " Enable secure websocket tap (wss URI)") // has meaning only if "-w" is set
	flag.BoolVar(&justlisten, "j", false, " Just listen on port, don't make a call")
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

	if !justlisten {
		consumer := new(signalwire.Consumer)

		signalwire.GlobalOverwriteHost = Host
		// setup the Client
		consumer.Setup(PProjectID, PTokenID, Contexts)
		// register callback
		consumer.Ready = MyReady
		// start
		if err := consumer.Run(); err != nil {
			signalwire.Log.Error("Error occurred while starting Signalwire Consumer\n")
		}
	} else {
		port = fmt.Sprintf("%d", defaultListenPort)
		wsListen("OGGOPUS")
	}
}
