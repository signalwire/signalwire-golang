package signalwire

import (
	"context"
	"errors"
	"strings"
	"sync"
)

// SendDigitsState keeps the state of a SendDigits action
type SendDigitsState int

// TODO DESCRIPTION
const (
	SendDigitsFinished SendDigitsState = iota
)

func (s SendDigitsState) String() string {
	return [...]string{"Finished"}[s]
}

// SendDigitsResult TODO DESCRIPTION
type SendDigitsResult struct {
	Successful bool
}

// SendDigitsAction TODO DESCRIPTION
type SendDigitsAction struct {
	CallObj   *CallObj
	ControlID string
	Completed bool
	Result    SendDigitsResult
	State     SendDigitsState
	err       error
	sync.RWMutex
}

// ISendDigits TODO DESCRIPTION
type ISendDigits interface {
	GetCompleted() bool
	GetResult() SendDigitsResult
}

func (callobj *CallObj) checkSendDigitsFinished(_ context.Context, ctrlID string, res *SendDigitsResult) (*SendDigitsResult, error) {
	var out bool

	for {
		select {
		case state := <-callobj.call.CallSendDigitsChans[ctrlID]:
			if state == SendDigitsFinished {
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

func checkDtmf(s string) bool {
	allowed := "wW1234567890*#ABCD"

	for _, c := range s {
		if !strings.Contains(allowed, string(c)) {
			return false
		}
	}

	return true
}

// SendDigits TODO DESCRIPTION
func (callobj *CallObj) SendDigits(digits string) (*SendDigitsResult, error) {
	if !checkDtmf(digits) {
		return nil, errors.New("invalid DTMF")
	}

	res := new(SendDigitsResult)

	if callobj.Calling == nil {
		return res, errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return res, errors.New("nil Relay object")
	}

	ctrlID, _ := GenUUIDv4()
	err := callobj.Calling.Relay.RelaySendDigits(callobj.Calling.Ctx, callobj.call, ctrlID, digits)

	if err != nil {
		return res, err
	}

	return callobj.checkSendDigitsFinished(callobj.Calling.Ctx, ctrlID, res)
}

// callbacksRunSendDigits TODO DESCRIPTION
func (callobj *CallObj) callbacksRunSendDigits(_ context.Context, ctrlID string, res *SendDigitsAction) {
	var out bool

	for {
		select {
		case state := <-callobj.call.CallSendDigitsChans[ctrlID]:
			res.RLock()

			prevstate := res.State

			res.RUnlock()

			switch state {
			case SendDigitsFinished:
				res.Lock()

				res.State = state
				res.Result.Successful = true
				res.Completed = true

				res.Unlock()

				Log.Debug("SendDigits finished. ctrlID: %s res [%p] Completed [%v] Successful [%v]\n", ctrlID, res, res.Completed, res.Result.Successful)

				out = true

				if callobj.OnSendDigitsFinished != nil {
					callobj.OnSendDigitsFinished(res)
				}

			default:
				Log.Debug("Unknown state. ctrlID: %s\n", ctrlID)
			}

			if prevstate != state && callobj.OnSendDigitsStateChange != nil {
				callobj.OnSendDigitsStateChange(res)
			}

		case <-callobj.call.Hangup:
			out = true
		}

		if out {
			break
		}
	}
}

// SendDigitsAsync TODO DESCRIPTION
func (callobj *CallObj) SendDigitsAsync(digits string) (*SendDigitsAction, error) {
	if !checkDtmf(digits) {
		return nil, errors.New("invalid DTMF")
	}

	res := new(SendDigitsAction)

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
			// wait to get control ID (buffered channel)
			ctrlID := <-callobj.call.CallSendDigitsControlIDs

			callobj.callbacksRunSendDigits(callobj.Calling.Ctx, ctrlID, res)
		}()

		newCtrlID, _ := GenUUIDv4()
		res.Lock()
		res.ControlID = newCtrlID
		res.Unlock()

		err := callobj.Calling.Relay.RelaySendDigits(callobj.Calling.Ctx, callobj.call, newCtrlID, digits)

		if err != nil {
			res.Lock()

			res.err = err

			res.Completed = true

			res.Unlock()
		}
		done <- struct{}{}
	}()

	<-done

	return res, res.err
}

// GetCompleted TODO DESCRIPTION
func (action *SendDigitsAction) GetCompleted() bool {
	action.RLock()

	ret := action.Completed

	action.RUnlock()

	return ret
}

// GetResult TODO DESCRIPTION
func (action *SendDigitsAction) GetResult() SendDigitsResult {
	action.RLock()

	ret := action.Result

	action.RUnlock()

	return ret
}
