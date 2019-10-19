package signalwire

import (
	"context"
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
}

// ConnectAction TODO DESCRIPTION
type ConnectAction struct {
	CallObj   *CallObj
	ControlID string
	Completed bool
	Result    ConnectResult
	State     CallConnectState
	err       error
	sync.RWMutex
}

// Connect TODO DESCRIPTION
func (callobj *CallObj) Connect(ringback *[]RingbackStruct, devices *[][]DeviceStruct) (*ConnectResult, error) {
	a := new(ConnectAction)
	res := &a.Result

	if callobj.Calling == nil {
		return res, errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return res, errors.New("nil Relay object")
	}

	if err := callobj.Calling.Relay.RelayConnect(callobj.Calling.Ctx, callobj.call, ringback, devices); err != nil {
		return res, err
	}

	callobj.callbacksRunConnect(callobj.Calling.Ctx, a)
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

	go func() {
		go func() {
			callobj.callbacksRunConnect(callobj.Calling.Ctx, res)
		}()

		err := callobj.Calling.Relay.RelayPhoneConnect(callobj.Calling.Ctx, callobj.call, fromNumber, toNumber)

		if err != nil {
			res.Lock()

			res.err = err

			res.Completed = true

			res.Unlock()
		}
	}()

	return res, nil
}

// callbacksRunConnect TODO DESCRIPTION
func (callobj *CallObj) callbacksRunConnect(_ context.Context, res *ConnectAction) {
	for {
		var out bool

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

				if callobj.OnConnectDisconnected != nil {
					callobj.OnConnectDisconnected(res)
				}

			case CallConnectConnecting:
				res.Lock()

				res.State = connectstate

				res.Unlock()

				if callobj.OnConnectConnecting != nil {
					callobj.OnConnectConnecting(res)
				}
			case CallConnectFailed:

				res.Lock()

				res.Completed = true
				res.State = connectstate

				res.Unlock()

				out = true

				if callobj.OnConnectFailed != nil {
					callobj.OnConnectFailed(res)
				}
			case CallConnectConnected:
				res.Lock()

				res.State = connectstate

				res.Unlock()

				if callobj.OnConnectConnected != nil {
					callobj.OnConnectConnected(res)
				}
			default:
				out = true
			}

			if prevstate != connectstate && callobj.OnConnectStateChange != nil {
				callobj.OnConnectStateChange(res)
			}
		case <-callobj.call.Hangup:
			out = true
		}

		if out {
			break
		}
	}
}
