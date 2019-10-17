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

// CallObj is the external Call object
type CallObj struct {
	call    *CallSession
	I       ICallObj
	Calling *Calling
	Context string

	OnStateChange       func()
	OnRinging           func()
	OnAnswered          func()
	OnEnding            func()
	OnEnded             func()
	OnPlayFinished      func(*PlayAction)
	OnPlayPaused        func(*PlayAction)
	OnPlayError         func(*PlayAction)
	OnPlayPlaying       func(*PlayAction)
	OnPlayStateChange   func(*PlayAction)
	OnRecordStateChange func(*RecordAction)
	OnRecordRecording   func(*RecordAction)
	OnRecordPaused      func(*RecordAction)
	OnRecordFinished    func(*RecordAction)
	OnRecordNoInput     func(*RecordAction)
	OnDetectUpdate      func(interface{})
	OnDetectError       func(interface{})
	OnDetectFinished    func(interface{})
	OnFaxFinished       func(*FaxAction)
	OnFaxPage           func(*FaxAction)
	OnFaxError          func(*FaxAction)
	Timeout             uint32 // ring timeout
	Active              bool
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
	if err := calling.Relay.RelayPhoneDial(calling.Ctx, newcall, fromNumber, toNumber, DefaultRingTimeout); err != nil {
		return *res
	}

	var I ICallObj = CallObjNew()

	c := &CallObj{I: I}
	c.call = newcall
	c.Calling = calling

	if ret := newcall.WaitCallStateInternal(calling.Ctx, Answered); !ret {
		Log.Debug("did not get Answered state\n")

		c.call = nil
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

	if err := calling.Relay.RelayPhoneDial(calling.Ctx, c.call, c.call.From, c.call.To, DefaultRingTimeout); err != nil {
		Log.Error("fields From or To not set for call\n")

		return *res
	}

	if ret := c.call.WaitCallStateInternal(calling.Ctx, Answered); !ret {
		Log.Debug("did not get Answered state\n")

		c.call = nil
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
