package signalwire

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// TapState keeps the state of a tap action
type TapState int

// Tap state constants
const (
	TapTapping TapState = iota
	TapFinished
)

func (s TapState) String() string {
	return [...]string{"Tappping", "Finished"}[s]
}

// TapDeviceType  TODO DESCRIPTION
type TapDeviceType int

// Tap state constants
const (
	TapRTP TapDeviceType = iota
	TapWS
)

func (s TapDeviceType) String() string {
	return [...]string{"rtp", "ws"}[s]
}

// TapDirection TODO DESCRIPTION
type TapDirection int

// Tap state constants
const (
	TapDirectionListen TapDirection = iota
	TapDirectionSpeak
)

func (s TapDirection) String() string {
	return [...]string{"listen", "speak"}[s]
}

// TapType TODO DESCRIPTION
type TapType int

// Tap state constants
const (
	TapAudio TapType = iota
)

func (s TapType) String() string {
	return [...]string{"audio"}[s]
}

// Tap TODO DESCRIPTION
type Tap struct {
	TapType       TapType
	TapDirection  TapDirection
	TapDeviceType TapDeviceType
}

// TapResult TODO DESCRIPTION
type TapResult struct {
	Successful        bool
	SourceDevice      TapDevice
	DestinationDevice TapDevice
	Tap               Tap
	Event             json.RawMessage
}

// TapAction TODO DESCRIPTION
type TapAction struct {
	CallObj   *CallObj
	ControlID string
	Completed bool
	Result    TapResult
	State     TapState
	Payload   *json.RawMessage
	err       error
	done      chan bool
	sync.RWMutex
}

// ITapAction TODO DESCRIPTION
type ITapAction interface {
	tapAsyncStop() error
	Stop()
	GetCompleted() bool
	GetResult() TapResult
	GetTap() Tap
	GetSourceDevice() TapDevice
	GetDestinationDevice() TapDevice
}

// TapAudio TODO DESCRIPTION
func (callobj *CallObj) TapAudio(direction fmt.Stringer, tapdev *TapDevice) (*TapResult, error) {
	a := new(TapAction)

	if callobj.Calling == nil {
		return &a.Result, errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return &a.Result, errors.New("nil Relay object")
	}

	ctrlID, _ := GenUUIDv4()

	var err error

	a.Result.SourceDevice, err = callobj.Calling.Relay.RelayTapAudio(callobj.Calling.Ctx, callobj.call, ctrlID, direction.String(), tapdev, nil)

	if err != nil {
		return &a.Result, err
	}

	callobj.callbacksRunTap(callobj.Calling.Ctx, ctrlID, a, true)

	return &a.Result, nil
}

// TapStop TODO DESCRIPTION
func (callobj *CallObj) TapStop(ctrlID *string) error {
	if callobj.Calling == nil {
		return errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return errors.New("nil Relay object")
	}

	return callobj.Calling.Relay.RelayTapStop(callobj.Calling.Ctx, callobj.call, ctrlID, nil)
}

// callbacksRunTap TODO DESCRIPTION
func (callobj *CallObj) callbacksRunTap(ctx context.Context, ctrlID string, res *TapAction, norunCB bool) {
	var out bool

	timer := time.NewTimer(BroadcastEventTimeout * time.Second)

	for {
		select {
		case <-timer.C:
			out = true
		// get tap states
		case tapstate := <-callobj.call.CallTapChans[ctrlID]:
			res.RLock()

			prevstate := res.State

			res.RUnlock()

			switch tapstate {
			case TapFinished:
				res.Lock()

				res.State = tapstate
				res.Result.Successful = true
				res.Completed = true

				res.Unlock()

				Log.Debug("Tap finished. ctrlID: %s res [%p] Completed [%v] Successful [%v]\n", ctrlID, res, res.Completed, res.Result.Successful)

				out = true

				if callobj.OnTapFinished != nil && !norunCB {
					callobj.OnTapFinished(res)
				}

			case TapTapping:
				timer.Reset(MaxCallDuration * time.Second)
				res.Lock()

				res.State = tapstate

				res.Unlock()

				Log.Debug("Tapping. ctrlID: %s\n", ctrlID)

				if callobj.OnTapTapping != nil && !norunCB {
					callobj.OnTapTapping(res)
				}
			default:
				Log.Debug("Unknown state. ctrlID: %s\n", ctrlID)
			}

			if prevstate != tapstate && callobj.OnTapStateChange != nil && !norunCB {
				callobj.OnTapStateChange(res)
			}

		case params := <-callobj.call.CallTapEventChans[ctrlID]:
			Log.Debug("got params for ctrlID : %s %v\n", ctrlID, params)

			res.Lock()

			switch params.Tap.Type {
			case "audio":
				res.Result.Tap.TapType = TapAudio

				switch params.Tap.Params.Direction {
				case "listen":
					res.Result.Tap.TapDirection = TapDirectionListen
				case "speak":
					res.Result.Tap.TapDirection = TapDirectionSpeak
				default:
					res.err = errors.New("invalid tap direction")
					res.Completed = true
					out = true

					goto ready
				}

			default:
				res.err = errors.New("invalid tap type")
				res.Completed = true
				out = true

				goto ready
			}

			switch params.Device.Type {
			case "rtp":
				res.Result.Tap.TapDeviceType = TapRTP
			default:
				res.err = errors.New("invalid tap device type")
				res.Completed = true
				out = true

				goto ready
			}

			res.Result.DestinationDevice.Params.Addr = params.Device.Params.Addr
			res.Result.DestinationDevice.Params.Port = params.Device.Params.Port
			res.Result.DestinationDevice.Params.Codec = params.Device.Params.Codec
			res.Result.DestinationDevice.Params.Ptime = params.Device.Params.Ptime
			res.Result.DestinationDevice.Params.Rate = params.Device.Params.Rate

		ready:
			res.Unlock()

			callobj.call.CallTapReadyChans[ctrlID] <- struct{}{}

		case rawEvent := <-callobj.call.CallTapRawEventChans[ctrlID]:
			res.Lock()
			res.Result.Event = *rawEvent
			res.Unlock()

			callobj.call.CallTapReadyChans[ctrlID] <- struct{}{}
		case <-callobj.call.Hangup:
			out = true
		case <-ctx.Done():
			out = true
		}

		if out {
			if !norunCB {
				res.done <- res.Result.Successful
			}

			break
		}
	}
}

// TapAudioAsync TODO DESCRIPTION
func (callobj *CallObj) TapAudioAsync(direction fmt.Stringer, tapdev *TapDevice) (*TapAction, error) {
	res := new(TapAction)

	if callobj.Calling == nil {
		return res, errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return res, errors.New("nil Relay object")
	}

	res.CallObj = callobj
	done := make(chan struct{}, 1)

	go func() {
		go func() {
			res.done = make(chan bool, 2)
			// wait to get control ID (buffered channel)
			ctrlID := <-callobj.call.CallTapControlIDs

			callobj.callbacksRunTap(callobj.Calling.Ctx, ctrlID, res, false)
		}()

		newCtrlID, _ := GenUUIDv4()
		res.Lock()
		res.ControlID = newCtrlID
		res.Unlock()

		srcDevice, err := callobj.Calling.Relay.RelayTapAudio(callobj.Calling.Ctx, callobj.call, newCtrlID, direction.String(), tapdev, &res.Payload)

		if err != nil {
			res.Lock()

			res.err = err

			res.Completed = true

			res.Unlock()
		} else {
			res.Lock()

			res.Result.SourceDevice = srcDevice

			res.Unlock()
		}
		done <- struct{}{}
	}()

	<-done

	return res, res.err
}

// ctrlIDCopy TODO DESCRIPTION
func (tapaction *TapAction) ctrlIDCopy() (string, error) {
	tapaction.RLock()

	if len(tapaction.ControlID) == 0 {
		tapaction.RUnlock()
		return "", errors.New("no controlID")
	}

	c := tapaction.ControlID

	tapaction.RUnlock()

	return c, nil
}

// tapAsyncStop TODO DESCRIPTION
func (tapaction *TapAction) tapAsyncStop() error {
	if tapaction.CallObj.Calling == nil {
		return errors.New("nil Calling object")
	}

	if tapaction.CallObj.Calling.Relay == nil {
		return errors.New("nil Relay object")
	}

	c, err := tapaction.ctrlIDCopy()
	if err != nil {
		return err
	}

	call := tapaction.CallObj.call

	return tapaction.CallObj.Calling.Relay.RelayTapStop(tapaction.CallObj.Calling.Ctx, call, &c, &tapaction.Payload)
}

// Stop TODO DESCRIPTION
func (tapaction *TapAction) Stop() StopResult {
	res := new(StopResult)
	tapaction.err = tapaction.tapAsyncStop()

	if tapaction.err == nil {
		res.Successful = <-tapaction.done
	}

	return *res
}

// GetCompleted TODO DESCRIPTION
func (tapaction *TapAction) GetCompleted() bool {
	tapaction.RLock()

	ret := tapaction.Completed

	tapaction.RUnlock()

	return ret
}

// GetResult TODO DESCRIPTION
func (tapaction *TapAction) GetResult() TapResult {
	tapaction.RLock()

	ret := tapaction.Result

	tapaction.RUnlock()

	return ret
}

// GetState TODO DESCRIPTION
func (tapaction *TapAction) GetState() TapState {
	tapaction.RLock()

	ret := tapaction.State

	tapaction.RUnlock()

	return ret
}

// GetTap TODO DESCRIPTION
func (tapaction *TapAction) GetTap() *Tap {
	return &tapaction.Result.Tap
}

// GetSourceDevice TODO DESCRIPTION
func (tapaction *TapAction) GetSourceDevice() *TapDevice {
	return &tapaction.Result.SourceDevice
}

// GetDestinationDevice TODO DESCRIPTION
func (tapaction *TapAction) GetDestinationDevice() *TapDevice {
	return &tapaction.Result.DestinationDevice
}

// GetEvent TODO DESCRIPTION
func (tapaction *TapAction) GetEvent() *json.RawMessage {
	tapaction.RLock()

	ret := &tapaction.Result.Event

	tapaction.RUnlock()

	return ret
}

// GetPayload TODO DESCRIPTION
func (tapaction *TapAction) GetPayload() *json.RawMessage {
	tapaction.RLock()

	ret := tapaction.Payload

	tapaction.RUnlock()

	return ret
}
