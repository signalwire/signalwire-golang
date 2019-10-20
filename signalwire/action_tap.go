package signalwire

import (
	"context"
	"errors"
	"sync"
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

// TapResult TODO DESCRIPTION
type TapResult struct {
	Successful bool
}

// TapAction TODO DESCRIPTION
type TapAction struct {
	CallObj   *CallObj
	ControlID string
	Completed bool
	Result    TapResult
	State     TapState
	err       error
	sync.RWMutex
}

// ITapAction TODO DESCRIPTION
type ITapAction interface {
	tapAsyncStop() error
	Stop()
	GetCompleted() bool
	GetResult() TapResult
}

func (callobj *CallObj) checkTapFinished(_ context.Context, ctrlID string, res *TapResult) (*TapResult, error) {
	var out bool

	for {
		select {
		case tapstate := <-callobj.call.CallTapChans[ctrlID]:
			if tapstate == TapFinished {
				out = true
				res.Successful = true
			}
		case <-callobj.call.Hangup:
			out = true
		}

		if out {
			break
		}
	}

	return res, nil
}

// TapAudio TODO DESCRIPTION
func (callobj *CallObj) TapAudio(direction string, tapdev *TapDevice) (*TapResult, error) {
	res := new(TapResult)

	if callobj.Calling == nil {
		return res, errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return res, errors.New("nil Relay object")
	}

	ctrlID, _ := GenUUIDv4()
	err := callobj.Calling.Relay.RelayTapAudio(callobj.Calling.Ctx, callobj.call, ctrlID, direction, tapdev)

	if err != nil {
		return res, err
	}

	return callobj.checkTapFinished(callobj.Calling.Ctx, ctrlID, res)
}

// TapStop TODO DESCRIPTION
func (callobj *CallObj) TapStop(ctrlID *string) error {
	if callobj.Calling == nil {
		return errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return errors.New("nil Relay object")
	}

	return callobj.Calling.Relay.RelayTapStop(callobj.Calling.Ctx, callobj.call, ctrlID)
}

// callbacksRunTap TODO DESCRIPTION
func (callobj *CallObj) callbacksRunTap(_ context.Context, ctrlID string, res *TapAction) {
	for {
		var out bool

		select {
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

				if callobj.OnTapFinished != nil {
					callobj.OnTapFinished(res)
				}

			case TapTapping:
				res.Lock()

				res.State = tapstate

				res.Unlock()

				Log.Debug("Tapping. ctrlID: %s\n", ctrlID)

				if callobj.OnTapTapping != nil {
					callobj.OnTapTapping(res)
				}
			default:
				Log.Debug("Unknown state. ctrlID: %s\n", ctrlID)
			}

			if prevstate != tapstate && callobj.OnTapStateChange != nil {
				callobj.OnTapStateChange(res)
			}
		case <-callobj.call.Hangup:
			out = true
		}

		if out {
			break
		}
	}
}

// TapAudioAsync TODO DESCRIPTION
func (callobj *CallObj) TapAudioAsync(direction string, tapdev *TapDevice) (*TapAction, error) {
	res := new(TapAction)

	if callobj.Calling == nil {
		return res, errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return res, errors.New("nil Relay object")
	}

	res.CallObj = callobj

	go func() {
		go func() {
			// wait to get control ID (buffered channel)
			ctrlID := <-callobj.call.CallTapControlIDs

			callobj.callbacksRunTap(callobj.Calling.Ctx, ctrlID, res)
		}()

		newCtrlID, _ := GenUUIDv4()
		res.Lock()
		res.ControlID = newCtrlID
		res.Unlock()

		err := callobj.Calling.Relay.RelayTapAudio(callobj.Calling.Ctx, callobj.call, newCtrlID, direction, tapdev)

		if err != nil {
			res.Lock()

			res.err = err

			res.Completed = true

			res.Unlock()
		}
	}()

	return res, nil
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

	return tapaction.CallObj.Calling.Relay.RelayTapStop(tapaction.CallObj.Calling.Ctx, call, &c)
}

// Stop TODO DESCRIPTION
func (tapaction *TapAction) Stop() {
	tapaction.err = tapaction.tapAsyncStop()
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
