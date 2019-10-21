package signalwire

import (
	"context"
)

// Calling TODO DESCRIPTION
type Calling struct {
	Ctx    context.Context
	Cancel context.CancelFunc
	Relay  *RelaySession
}

// CallObj is the external Call object (as exposed to the user)
type CallObj struct {
	call    *CallSession
	I       ICallObj
	Calling *Calling

	OnStateChange           func()
	OnRinging               func()
	OnAnswered              func()
	OnEnding                func()
	OnEnded                 func()
	OnPlayFinished          func(*PlayAction)
	OnPlayPaused            func(*PlayAction)
	OnPlayError             func(*PlayAction)
	OnPlayPlaying           func(*PlayAction)
	OnPlayStateChange       func(*PlayAction)
	OnRecordStateChange     func(*RecordAction)
	OnRecordRecording       func(*RecordAction)
	OnRecordPaused          func(*RecordAction)
	OnRecordFinished        func(*RecordAction)
	OnRecordNoInput         func(*RecordAction)
	OnDetectUpdate          func(interface{})
	OnDetectError           func(interface{})
	OnDetectFinished        func(interface{})
	OnFaxFinished           func(*FaxAction)
	OnFaxPage               func(*FaxAction)
	OnFaxError              func(*FaxAction)
	OnConnectStateChange    func(*ConnectAction)
	OnConnectFailed         func(*ConnectAction)
	OnConnectConnecting     func(*ConnectAction)
	OnConnectConnected      func(*ConnectAction)
	OnConnectDisconnected   func(*ConnectAction)
	OnTapStateChange        func(*TapAction)
	OnTapFinished           func(*TapAction)
	OnTapTapping            func(*TapAction)
	OnSendDigitsFinished    func(*SendDigitsAction)
	OnSendDigitsStateChange func(*SendDigitsAction)
}

// ICallObj these are for unit-testing
type ICallObj interface {
	Hangup() (*ResultHangup, error)
	Answer() (*ResultAnswer, error)
	PlayAudio(url string) (*PlayResult, error)
	PlayStop(ctrlID *string) error
	PlayTTS(text, language, gender string) (*PlayResult, error)
	PlaySilence(duration float64) (*PlayResult, error)
	PlayRingtone(name string, duration float64) (*PlayResult, error)
	RecordAudio(r *RecordParams) (*RecordResult, error)
	RecordAudioAsync(r *RecordParams) (*RecordAction, error)
	RecordAudioStop(ctrlID *string) error
	DetectMachine(det *DetectMachineParams) (*DetectResult, error)
	DetectMachineAsync(det *DetectMachineParams) (*DetectMachineAction, error)
	DetectFax(det *DetectFaxParams) (*DetectResult, error)
	DetectFaxAsync(det *DetectFaxParams) (*DetectFaxAction, error)
	DetectDigit(det *DetectDigitParams) (*DetectResult, error)
	DetectDigitAsync(det *DetectDigitParams) (*DetectDigitAction, error)
	DetectStop(ctrlID *string) error
	ReceiveFax() (*FaxResult, error)
	SendFax(doc, id, headerInfo string) (*FaxResult, error)
	SendFaxStop(ctrlID *string) error
	ReceiveFaxAsync() (*FaxAction, error)
	SendFaxAsync(doc, id, headerInfo string) (*FaxAction, error)
	WaitFor(state CallState) bool
	WaitForRinging() bool
	WaitForAnswered() bool
	WaitForEnding() bool
	WaitForEnded() bool
	GetActive() bool
	GetState() CallState
	GetPrevState() CallState
	GetCallID() string
}

// ResultDial TODO DESCRIPTION
type ResultDial struct {
	Successful bool
	Call       *CallObj
	I          ICalling
}

// ResultAnswer TODO DESCRIPTION
type ResultAnswer struct {
	Successful bool
}

// ResultHangup TODO DESCRIPTION
type ResultHangup struct {
	Successful bool
	Reason     CallDisconnectReason
}

// ICalling object visible to the end user
type ICalling interface {
	DialPhone(fromNumber, toNumber string) ResultDial
	NewCall() *CallObj
	Dial(c *CallObj) ResultDial
}

// CallObjNew TODO DESCRIPTION
func CallObjNew() *CallObj {
	return &CallObj{}
}

// DialPhone  TODO DESCRIPTION
func (calling *Calling) DialPhone(fromNumber, toNumber string) ResultDial {
	res := new(ResultDial)

	if calling.Relay == nil {
		return *res
	}

	if calling.Ctx == nil {
		return *res
	}

	newcall := new(CallSession)
	newcall.SetActive(true)

	if err := calling.Relay.RelayPhoneDial(calling.Ctx, newcall, fromNumber, toNumber, DefaultRingTimeout); err != nil {
		newcall.SetActive(false)
		return *res
	}

	var I ICallObj = CallObjNew()

	c := &CallObj{I: I}
	c.call = newcall
	c.Calling = calling

	if ret := newcall.WaitCallStateInternal(calling.Ctx, Answered); !ret {
		Log.Debug("did not get Answered state\n")

		c.call.SetActive(false)
		res.Call = c

		return *res
	}

	res.Call = c
	res.Successful = true

	return *res
}

// Hangup TODO DESCRIPTION
func (callobj *CallObj) Hangup() (*ResultHangup, error) {
	res := new(ResultHangup)
	call := callobj.call

	if call.CallState != Ending && call.CallState != Ended {
		if err := callobj.Calling.Relay.RelayCallEnd(callobj.Calling.Ctx, call); err != nil {
			return res, err
		}
	}

	if ret := call.WaitCallStateInternal(callobj.Calling.Ctx, Ended); !ret {
		Log.Debug("did not get Ended state for call\n")
	}

	if call.CallState == Ended {
		// todo: handle race conds on hangup (don't write on closed channels)
		res.Reason = call.CallDisconnectReason
		call.CallCleanup(callobj.Calling.Ctx)
	}

	res.Successful = true

	return res, nil
}

// Answer TODO DESCRIPTION
func (callobj *CallObj) Answer() (*ResultAnswer, error) {
	call := callobj.call

	res := new(ResultAnswer)

	if call.CallState != Answered {
		Log.Info("Answering call [%p]\n", call)

		if err := callobj.Calling.Relay.RelayCallAnswer(callobj.Calling.Ctx, call); err != nil {
			Log.Debug("cannot answer call. err: %v\n", err)

			return res, err
		}
	}

	// 'Answered' state event may have already come before we get the 200 for calling.answer command.
	if call.CallState != Answered {
		if ret := call.WaitCallStateInternal(callobj.Calling.Ctx, Answered); !ret {
			Log.Debug("did not get Answered state for inbound call\n")

			return res, nil
		}
	}

	res.Successful = true

	return res, nil
}

// GetCallState TODO DESCRIPTION
func (callobj *CallObj) GetCallState() CallState {
	callobj.call.Lock()
	s := callobj.call.CallState
	callobj.call.Unlock()

	return s
}

// WaitFor TODO DESCRIPTION
func (callobj *CallObj) WaitFor(want CallState) bool {
	if ret := callobj.call.WaitCallStateInternal(callobj.Calling.Ctx, want); !ret {
		Log.Error("did not get %s state for call\n", want.String())
		return false
	}

	return true
}

// WaitForRinging TODO DESCRIPTION
func (callobj *CallObj) WaitForRinging() bool {
	if ret := callobj.call.WaitCallStateInternal(callobj.Calling.Ctx, Ringing); !ret {
		Log.Error("did not get Ringing state for call\n")
		return false
	}

	return true
}

// WaitForAnswered TODO DESCRIPTION
func (callobj *CallObj) WaitForAnswered() bool {
	if ret := callobj.call.WaitCallStateInternal(callobj.Calling.Ctx, Answered); !ret {
		Log.Error("did not get Answered state for call\n")
		return false
	}

	return true
}

// WaitForEnding TODO DESCRIPTION
func (callobj *CallObj) WaitForEnding() bool {
	if ret := callobj.call.WaitCallStateInternal(callobj.Calling.Ctx, Ending); !ret {
		Log.Error("did not get Ending state for call\n")
		return false
	}

	return true
}

// WaitForEnded TODO DESCRIPTION
func (callobj *CallObj) WaitForEnded() bool {
	if ret := callobj.call.WaitCallStateInternal(callobj.Calling.Ctx, Ended); !ret {
		Log.Error("did not get Ended state for call\n")
		return false
	}

	return true
}

// NewCall  TODO DESCRIPTION
func (calling *Calling) NewCall(from, to string) *CallObj {
	newcall := new(CallSession)

	var I ICallObj = CallObjNew()

	c := &CallObj{I: I}
	c.call = newcall
	c.Calling = calling
	c.call.SetFrom(from)
	c.call.SetTo(to)

	return c
}

// Dial TODO DESCRIPTION
func (calling *Calling) Dial(c *CallObj) ResultDial {
	res := new(ResultDial)

	if calling.Relay == nil {
		return *res
	}

	if calling.Ctx == nil {
		return *res
	}

	if len(c.call.From) == 0 || len(c.call.To) == 0 {
		return *res
	}

	c.call.SetActive(true)

	if err := calling.Relay.RelayPhoneDial(calling.Ctx, c.call, c.call.From, c.call.To, c.call.Timeout); err != nil {
		Log.Error("fields From or To not set for call\n")

		c.call.SetActive(false)

		return *res
	}

	if ret := c.call.WaitCallStateInternal(calling.Ctx, Answered); !ret {
		Log.Debug("did not get Answered state\n")

		c.call.SetActive(false)
		res.Call = c

		return *res
	}

	res.Call = c
	res.Successful = true

	return *res
}

// GetReason TODO DESCRIPTION
func (resultHangup *ResultHangup) GetReason() CallDisconnectReason {
	return resultHangup.Reason
}

// GetSuccessful TODO DESCRIPTION
func (resultHangup *ResultHangup) GetSuccessful() bool {
	return resultHangup.Successful
}

// GetSuccessful TODO DESCRIPTION
func (resultAnswer *ResultAnswer) GetSuccessful() bool {
	return resultAnswer.Successful
}

// GetSuccessful TODO DESCRIPTION
func (resultDial *ResultDial) GetSuccessful() bool {
	return resultDial.Successful
}

// GetActive TODO DESCRIPTION
func (callobj *CallObj) GetActive() bool {
	return callobj.call.GetActive()
}

// GetState TODO DESCRIPTION
func (callobj *CallObj) GetState() CallState {
	return callobj.call.GetState()
}

// GetPrevState TODO DESCRIPTION
func (callobj *CallObj) GetPrevState() CallState {
	return callobj.call.GetPrevState()
}

// GetCallID TODO DESCRIPTION
func (callobj *CallObj) GetCallID() string {
	return callobj.call.GetCallID()
}

// SetTimeout TODO DESCRIPTION
func (callobj *CallObj) SetTimeout(t uint) {
	callobj.call.SetTimeout(t)
}

// GetTimeout TODO DESCRIPTION
func (callobj *CallObj) GetTimeout() uint {
	return callobj.call.GetTimeout()
}

// GetFrom TODO DESCRIPTION
func (callobj *CallObj) GetFrom() string {
	return callobj.call.GetFrom()
}

// GetTo TODO DESCRIPTION
func (callobj *CallObj) GetTo() string {
	return callobj.call.GetTo()
}
