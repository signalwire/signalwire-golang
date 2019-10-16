package signalwire

import (
	"errors"
)

// CallConnectState TODO DESCRIPTION
type CallConnectState int

// call.connect states
const (
	Failed CallConnectState = iota
	Connecting
	Connected
	Disconnected
)

func (s CallConnectState) String() string {
	return [...]string{"Failed", "Connecting", "Connected", "Disconnected"}[s]
}

// ConnectResult TODO DESCRIPTION
type ConnectResult struct {
	Successful bool
}

func (callobj *CallObj) Connect(fromNumber, toNumber string) (*ConnectResult, error) {
	res := new(ConnectResult)

	if callobj.Calling == nil {
		return res, errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return res, errors.New("nil Relay object")
	}

	if err := callobj.Calling.Relay.RelayPhoneConnect(callobj.Calling.Ctx, callobj.call, fromNumber, toNumber); err != nil {
		return res, err
	}

	res.Successful = true

	return res, nil
}
