package signalwire

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
)

// CallConnectState TODO DESCRIPTION
type CallConnectState int

// call.connect states
const (
	CallConnectFailed CallConnectState = iota
	CallConnectConnecting
	CallConnectConnected
	CallConnectDisconnected
)

func (s CallConnectState) String() string {
	return [...]string{"Failed", "Connecting", "Connected", "Disconnected"}[s]
}

// ConnectResult TODO DESCRIPTION
type ConnectResult struct {
	Successful bool
	Event      json.RawMessage
	CallObj    *CallObj
}

// ConnectAction TODO DESCRIPTION
type ConnectAction struct {
	CallObj   *CallObj
	ControlID string
	Completed bool
	Result    ConnectResult
	State     CallConnectState
	Payload   *json.RawMessage
	err       error
	sync.RWMutex
}

// Connect TODO DESCRIPTION
func (callobj *CallObj) Connect(ringback *[]RingbackStruct, devices *[][]DeviceStruct) (*ConnectResult, error) {
	a := new(ConnectAction)
	res := &a.Result

	a.Result.CallObj = callobj

	if callobj.Calling == nil {
		return res, errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return res, errors.New("nil Relay object")
	}

	if err := callobj.Calling.Relay.RelayConnect(callobj.Calling.Ctx, callobj.call, ringback, devices, nil); err != nil {
		return res, err
	}

	callobj.callbacksRunConnect(callobj.Calling.Ctx, a, true)

	return res, nil
}

// ConnectAsync TODO DESCRIPTION
func (callobj *CallObj) ConnectAsync(fromNumber, toNumber string) (*ConnectAction, error) {
	res := new(ConnectAction)

	if callobj.Calling == nil {
		return res, errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return res, errors.New("nil Relay object")
	}

	res.CallObj = callobj

	res.Result.CallObj = callobj

	done := make(chan struct{}, 1)

	go func() {
		go func() {
			callobj.callbacksRunConnect(callobj.Calling.Ctx, res, false)
		}()

		err := callobj.Calling.Relay.RelayPhoneConnect(callobj.Calling.Ctx, callobj.call, fromNumber, toNumber, &res.Payload)

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

// callbacksRunConnect TODO DESCRIPTION
func (callobj *CallObj) callbacksRunConnect(ctx context.Context, res *ConnectAction, norunCB bool) {
	var out bool

	for {
		select {
		case connectstate := <-callobj.call.CallConnectStateChan:
			res.RLock()

			prevstate := res.State

			res.RUnlock()

			switch connectstate {
			case CallConnectDisconnected:
				res.Lock()

				res.State = connectstate
				res.Result.Successful = true
				res.Completed = true

				res.Unlock()

				out = true

				if callobj.OnConnectDisconnected != nil && !norunCB {
					callobj.OnConnectDisconnected(res)
				}

			case CallConnectConnecting:
				res.Lock()

				res.State = connectstate

				res.Unlock()

				if callobj.OnConnectConnecting != nil && !norunCB {
					callobj.OnConnectConnecting(res)
				}
			case CallConnectFailed:
				res.Lock()

				res.Completed = true
				res.State = connectstate

				res.Unlock()

				out = true

				if callobj.OnConnectFailed != nil && !norunCB {
					callobj.OnConnectFailed(res)
				}
			case CallConnectConnected:
				res.Lock()

				res.State = connectstate

				res.Unlock()

				if callobj.OnConnectConnected != nil && !norunCB {
					callobj.OnConnectConnected(res)
				}
			default:
				out = true
			}

			if prevstate != connectstate && callobj.OnConnectStateChange != nil && !norunCB {
				callobj.OnConnectStateChange(res)
			}
		case rawEvent := <-callobj.call.CallConnectRawEventChan:
			res.Lock()
			res.Result.Event = *rawEvent
			res.Unlock()
		case <-callobj.call.Hangup:
			out = true
		case <-ctx.Done():
			out = true
		}

		if out {
			break
		}
	}
}

// GetEvent TODO DESCRIPTION
func GetEvent(action *ConnectAction) *json.RawMessage {
	action.RLock()

	ret := &action.Result.Event

	action.RUnlock()

	return ret
}

// GetPayload TODO DESCRIPTION
func (action *ConnectAction) GetPayload() *json.RawMessage {
	action.RLock()

	ret := action.Payload

	action.RUnlock()

	return ret
}

// GetCall TODO DESCRIPTION
func (action *ConnectAction) GetCall() *CallObj {
	action.RLock()

	ret := action.CallObj

	action.RUnlock()

	return ret
}
