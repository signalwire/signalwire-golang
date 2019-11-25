package signalwire

import (
	"context"
	"encoding/json"
	"time"
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
	Payload *json.RawMessage // last command payload

	OnStateChange           func(*CallObj)
	OnRinging               func(*CallObj)
	OnAnswered              func(*CallObj)
	OnEnding                func(*CallObj)
	OnEnded                 func(*CallObj)
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
	OnDetectUpdate          func(*DetectAction)
	OnDetectError           func(*DetectAction)
	OnDetectFinished        func(*DetectAction)
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
	OnPrompt                func(*PromptAction)
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
	DetectMachineAsync(det *DetectMachineParams) (*DetectAction, error)
	DetectFax(det *DetectFaxParams) (*DetectResult, error)
	DetectFaxAsync(det *DetectFaxParams) (*DetectAction, error)
	DetectDigit(det *DetectDigitParams) (*DetectResult, error)
	DetectDigitAsync(det *DetectDigitParams) (*DetectAction, error)
	DetectStop(ctrlID *string) error
	ReceiveFax() (*FaxResult, error)
	SendFax(doc, id, headerInfo string) (*FaxResult, error)
	SendFaxStop(ctrlID *string) error
	ReceiveFaxAsync() (*FaxAction, error)
	SendFaxAsync(doc, id, headerInfo string) (*FaxAction, error)
	WaitFor(state CallState, timeout uint) bool
	WaitForRinging(timeout uint) bool
	WaitForAnswered(timeout uint) bool
	WaitForEnding(timeout uint) bool
	WaitForEnded(timeout uint) bool
	Active() bool
	GetState() CallState
	GetPrevState() CallState
	GetID() string
	GetTempID() string
	GetTo() string
	GetFrom() string
}

// ResultDial TODO DESCRIPTION
type ResultDial struct {
	Successful bool
	Call       *CallObj
	I          ICalling
	err        error
}

// ResultAnswer TODO DESCRIPTION
type ResultAnswer struct {
	Successful bool
}

// ResultHangup TODO DESCRIPTION
type ResultHangup struct {
	Successful bool
	Reason     CallDisconnectReason
	Event      *json.RawMessage
	err        error
}

// StopResult TODO DESCRIPTION
type StopResult struct {
	Successful bool
}

func waitStop(res *StopResult, done chan bool) {
	timer := time.NewTimer(BroadcastEventTimeout * time.Second)

	select {
	case <-timer.C:
	case res.Successful = <-done:
	default:
	}
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

	var I ICall = CallNew()

	newcall := &CallSession{I: I}
	newcall.I = newcall

	newcall.SetActive(true)

	var savePayload *json.RawMessage

	if err := calling.Relay.I.RelayPhoneDial(calling.Ctx, newcall, fromNumber, toNumber, DefaultRingTimeout, &savePayload); err != nil {
		newcall.SetActive(false)

		res.err = err

		return *res
	}

	var J ICallObj = CallObjNew()

	c := &CallObj{I: J}
	c.I = J
	c.call = newcall
	c.Payload = savePayload
	c.Calling = calling

	if ret := newcall.I.WaitCallStateInternal(calling.Ctx, Answered, DefaultRingTimeout); !ret {
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
		if err := callobj.Calling.Relay.RelayCallEnd(callobj.Calling.Ctx, call, &callobj.Payload); err != nil {
			res.err = err
			return res, err
		}
	}

	if ret := call.WaitCallStateInternal(callobj.Calling.Ctx, Ended, BroadcastEventTimeout); !ret {
		Log.Debug("did not get Ended state for call\n")
	}

	if call.CallState == Ended {
		res.Reason = call.CallDisconnectReason
		res.Event = call.Event
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

		if err := callobj.Calling.Relay.RelayCallAnswer(callobj.Calling.Ctx, call, &callobj.Payload); err != nil {
			Log.Debug("cannot answer call. err: %v\n", err)

			return res, err
		}
	}

	// 'Answered' state event may have already come before we get the 200 for calling.answer command.
	if call.CallState != Answered {
		if ret := call.WaitCallStateInternal(callobj.Calling.Ctx, Answered, BroadcastEventTimeout); !ret {
			Log.Debug("did not get Answered state for inbound call\n")

			return res, nil
		}
	}

	go func(ctx context.Context) {
		// states && callbacks
		callobj.callbacksRunCallState(ctx)
	}(callobj.Calling.Ctx)

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
func (callobj *CallObj) WaitFor(want CallState, timeout uint) bool {
	if ret := callobj.call.WaitCallStateInternal(callobj.Calling.Ctx, want, timeout); !ret {
		Log.Error("did not get %s state for call\n", want.String())
		return false
	}

	return true
}

// WaitForRinging TODO DESCRIPTION
func (callobj *CallObj) WaitForRinging(timeout uint) bool {
	if ret := callobj.call.WaitCallStateInternal(callobj.Calling.Ctx, Ringing, timeout); !ret {
		Log.Error("did not get Ringing state for call\n")
		return false
	}

	return true
}

// WaitForAnswered TODO DESCRIPTION
func (callobj *CallObj) WaitForAnswered(timeout uint) bool {
	if ret := callobj.call.WaitCallStateInternal(callobj.Calling.Ctx, Answered, timeout); !ret {
		Log.Error("did not get Answered state for call\n")
		return false
	}

	return true
}

// WaitForEnding TODO DESCRIPTION
func (callobj *CallObj) WaitForEnding(timeout uint) bool {
	if ret := callobj.call.WaitCallStateInternal(callobj.Calling.Ctx, Ending, timeout); !ret {
		Log.Error("did not get Ending state for call\n")
		return false
	}

	return true
}

// WaitForEnded TODO DESCRIPTION
func (callobj *CallObj) WaitForEnded(timeout uint) bool {
	if ret := callobj.call.WaitCallStateInternal(callobj.Calling.Ctx, Ended, timeout); !ret {
		Log.Error("did not get Ended state for call\n")
		return false
	}

	return true
}

// NewCall  TODO DESCRIPTION
func (calling *Calling) NewCall(from, to string) *CallObj {
	var I ICall = CallNew()

	newcall := &CallSession{I: I}
	newcall.I = newcall

	var J ICallObj = CallObjNew()

	c := &CallObj{I: J}
	c.call = newcall
	c.Calling = calling
	c.call.SetFrom(from)
	c.call.SetTo(to)

	return c
}

func (callobj *CallObj) callbacksRunCallState(ctx context.Context) {
	var out bool

	for {
		select {
		case rcvState := <-callobj.call.cbStateChan:
			if rcvState != callobj.call.GetPrevState() {
				if callobj.OnStateChange != nil {
					callobj.OnStateChange(callobj)
				}
			}

			switch rcvState {
			case Answered:
				if callobj.OnAnswered != nil {
					callobj.OnAnswered(callobj)
				}
			case Ringing:
				if callobj.OnRinging != nil {
					callobj.OnRinging(callobj)
				}
			case Ending:
				if callobj.OnEnding != nil {
					callobj.OnEnding(callobj)
				}
			case Ended:
				if callobj.OnEnded != nil {
					callobj.OnEnded(callobj)
				}

				out = true
			}
		case <-ctx.Done():
			out = true
		}

		if out {
			break
		}
	}
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

	if err := calling.Relay.I.RelayPhoneDial(calling.Ctx, c.call, c.call.From, c.call.To, c.call.Timeout, &c.Payload); err != nil {
		res.err = err

		c.call.SetActive(false)

		return *res
	}

	if ret := c.call.I.WaitCallStateInternal(calling.Ctx, Answered, c.call.GetTimeout()); !ret {
		Log.Debug("did not get Answered state\n")

		c.call.SetActive(false)
		res.Call = c

		return *res
	}

	go func(ctx context.Context) {
		// states && callbacks
		c.callbacksRunCallState(ctx)
	}(calling.Ctx)

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

// GetError TODO DESCRIPTION
func (resultHangup *ResultHangup) GetError() bool {
	return resultHangup.Successful
}

// GetEvent TODO DESCRIPTION
func (resultHangup *ResultHangup) GetEvent() *json.RawMessage {
	return resultHangup.Event
}

// GetSuccessful TODO DESCRIPTION
func (resultAnswer *ResultAnswer) GetSuccessful() bool {
	return resultAnswer.Successful
}

// GetSuccessful TODO DESCRIPTION
func (resultDial *ResultDial) GetSuccessful() bool {
	return resultDial.Successful
}

// GetError TODO DESCRIPTION
func (resultDial *ResultDial) GetError() error {
	return resultDial.err
}

// Active TODO DESCRIPTION
func (callobj *CallObj) Active() bool {
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

// GetID TODO DESCRIPTION
func (callobj *CallObj) GetID() string {
	return callobj.call.GetCallID()
}

// GetTempID TODO DESCRIPTION
func (callobj *CallObj) GetTempID() string {
	return callobj.call.GetTagID()
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

// Answered TODO DESCRIPTION
func (callobj *CallObj) Answered() bool {
	return callobj.call.GetState() == Answered
}

// Busy TODO DESCRIPTION
func (callobj *CallObj) Busy() bool {
	return callobj.call.CallDisconnectReason == CallBusy
}

// Failed TODO DESCRIPTION
func (callobj *CallObj) Failed() bool {
	return callobj.call.CallDisconnectReason == CallGenericError
}

// Ended TODO DESCRIPTION
func (callobj *CallObj) Ended() bool {
	return callobj.call.GetState() == Ended
}

// GetType TODO DESCRIPTION
func (callobj *CallObj) GetType() string {
	return callobj.call.GetType()
}

// GetEvent TODO DESCRIPTION
func (callobj *CallObj) GetEvent() *json.RawMessage {
	return callobj.call.GetEventPayload()
}

// GetEventName this is expensive
func (callobj *CallObj) GetEventName() string {
	return callobj.call.GetEventName()
}
