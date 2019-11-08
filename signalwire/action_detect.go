package signalwire

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
)

// DetectMachineEvent keeps the event of a detect action
type DetectMachineEvent int

// Machine detector event constants
const (
	DetectMachineUnknown DetectMachineEvent = iota
	DetectMachineMachine
	DetectMachineHuman
	DetectMachineReady
	DetectMachineNotReady
	DetectMachineFinished
)

func (s DetectMachineEvent) String() string {
	return [...]string{"Unknown", "Machine", "Human", "Ready", "Not_Ready", "Finished"}[s]
}

// DetectResultType TODO DESCRIPTION
type DetectResultType int

// Type of detector (used only in the Result)
const (
	DetectorMachine DetectResultType = iota
	DetectorHuman
	DetectorFax
	DetectorDTMF
	DetectorUnknown
	DetectorError
	DetectorFinished
)

func (s DetectResultType) String() string {
	return [...]string{"Machine", "Human", "Fax", "DTMF", "Unknown", "Error", Finished}[s]
}

// DetectResult TODO DESCRIPTION
type DetectResult struct {
	Successful bool
	Type       DetectResultType
	Event      json.RawMessage
}

// DetectMachineAction TODO DESCRIPTION
type DetectMachineAction struct {
	CallObj   *CallObj
	ControlID string
	Completed bool
	Result    DetectResult
	Event     DetectMachineEvent
	Payload   *json.RawMessage
	err       error
	done      chan bool
	sync.RWMutex
}

// DetectDigitEvent TODO DESCRIPTION
type DetectDigitEvent int

// Digit Detector events
const (
	DetectDigitZero DetectDigitEvent = iota
	DetectDigitOne
	DetectDigitTwo
	DetectDigitThree
	DetectDigitFour
	DetectDigitFive
	DetectDigitSix
	DetectDigitSeven
	DetectDigitEight
	DetectDigitNine
	DetectDigitPound
	DetectDigitStar
	DetectDigitFinished
)

func (s DetectDigitEvent) String() string {
	return [...]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "#", "*", "Finished"}[s]
}

// DetectFaxEvent TODO DESCRIPTION
type DetectFaxEvent int

// Call event constants
const (
	DetectFaxNone DetectFaxEvent = iota
	DetectFaxCED                 // Called Station Fax Tone
	DetectFaxCNG                 // Calling Station Fax Tone
	DetectFaxFinished
)

func (s DetectFaxEvent) String() string {
	return [...]string{"", "CED", "CNG", "Finished"}[s]
}

// DetectDigitAction TODO DESCRIPTION
type DetectDigitAction struct {
	CallObj   *CallObj
	ControlID string
	Completed bool
	Result    DetectResult
	Event     DetectDigitEvent
	Payload   *json.RawMessage
	err       error
	done      chan bool
	sync.RWMutex
}

// DetectFaxAction TODO DESCRIPTION
type DetectFaxAction struct {
	CallObj   *CallObj
	ControlID string
	Completed bool
	Result    DetectResult
	Event     DetectFaxEvent
	Payload   *json.RawMessage
	err       error
	done      chan bool
	sync.RWMutex
}

// IDetectMachineAction TODO DESCRIPTION
type IDetectMachineAction interface {
	detectAsyncStop() error
	Stop()
}

// IDetectDigitAction TODO DESCRIPTION
type IDetectDigitAction interface {
	detectAsyncStop() error
	Stop()
}

// IDetectFaxAction TODO DESCRIPTION
type IDetectFaxAction interface {
	detectAsyncStop() error
	Stop()
}

// DetectMachine TODO DESCRIPTION
func (callobj *CallObj) DetectMachine(det *DetectMachineParams) (*DetectResult, error) {
	a := new(DetectMachineAction)

	if callobj.Calling == nil {
		return &a.Result, errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return &a.Result, errors.New("nil Relay object")
	}

	ctrlID, _ := GenUUIDv4()

	err := callobj.Calling.Relay.RelayDetectMachine(callobj.Calling.Ctx, callobj.call, ctrlID, det, nil)

	if err != nil {
		return &a.Result, err
	}

	callobj.callbacksRunDetectMachine(callobj.Calling.Ctx, ctrlID, a)

	return &a.Result, nil
}

// DetectFax TODO DESCRIPTION
func (callobj *CallObj) DetectFax(det *DetectFaxParams) (*DetectResult, error) {
	a := new(DetectFaxAction)

	if callobj.Calling == nil {
		return &a.Result, errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return &a.Result, errors.New("nil Relay object")
	}

	ctrlID, _ := GenUUIDv4()

	err := callobj.Calling.Relay.RelayDetectFax(callobj.Calling.Ctx, callobj.call, ctrlID, det.Tone, nil)

	if err != nil {
		return &a.Result, err
	}

	callobj.callbacksRunDetectFax(callobj.Calling.Ctx, ctrlID, a)

	return &a.Result, nil
}

// DetectDigit TODO DESCRIPTION
func (callobj *CallObj) DetectDigit(det *DetectDigitParams) (*DetectResult, error) {
	a := new(DetectDigitAction)

	if callobj.Calling == nil {
		return &a.Result, errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return &a.Result, errors.New("nil Relay object")
	}

	ctrlID, _ := GenUUIDv4()

	err := callobj.Calling.Relay.RelayDetectDigit(callobj.Calling.Ctx, callobj.call, ctrlID, det.Digits, nil)

	if err != nil {
		return &a.Result, err
	}

	callobj.callbacksRunDetectDigit(callobj.Calling.Ctx, ctrlID, a)

	return &a.Result, nil
}

// DetectStop TODO DESCRIPTION
func (callobj *CallObj) DetectStop(ctrlID *string) error {
	if callobj.Calling == nil {
		return errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return errors.New("nil Relay object")
	}

	return callobj.Calling.Relay.RelayDetectStop(callobj.Calling.Ctx, callobj.call, ctrlID, nil)
}

// callbacksRunDetectMachine TODO DESCRIPTION
func (callobj *CallObj) callbacksRunDetectMachine(ctx context.Context, ctrlID string, res *DetectMachineAction) {
	for {
		var out bool

		select {
		// get detect events
		case detectevent := <-callobj.call.CallDetectMachineChans[ctrlID]:
			if detectevent == DetectMachineFinished {
				out = true
			}

			res.RLock()

			prevevent := res.Event

			res.RUnlock()

			Log.Debug("Got detectevent %s. ctrlID: %s\n", detectevent.String(), ctrlID)

			switch detectevent {
			case DetectMachineFinished:
				res.Lock()

				res.Result.Successful = true
				res.Completed = true

				res.Unlock()

				Log.Debug("Detect finished. ctrlID: %s\n", ctrlID)

				if callobj.OnDetectFinished != nil {
					callobj.OnDetectFinished(res)
				}
			case DetectMachineMachine:
				res.Result.Type = DetectorMachine
				fallthrough
			case DetectMachineHuman:
				res.Result.Type = DetectorHuman
				fallthrough
			case DetectMachineUnknown:
				fallthrough
			case DetectMachineReady:
				fallthrough
			case DetectMachineNotReady:
				res.Lock()

				res.Event = detectevent

				res.Unlock()
			}

			if prevevent != detectevent && callobj.OnDetectUpdate != nil {
				callobj.OnDetectUpdate(res)
			}
		case rawEvent := <-callobj.call.CallDetectRawEventChans[ctrlID]:
			res.Lock()
			res.Result.Event = *rawEvent
			res.Unlock()
		case <-callobj.call.Hangup:
			out = true
		case <-ctx.Done():
			out = true
		}

		if out {
			Log.Debug("blocked")
			res.done <- res.Result.Successful
			Log.Debug("not blocked")
			break
		}
	}
}

// callbacksRunDetectFax TODO DESCRIPTION
func (callobj *CallObj) callbacksRunDetectFax(ctx context.Context, ctrlID string, res *DetectFaxAction) {
	for {
		var out bool

		select {
		// get detect events
		case detectevent := <-callobj.call.CallDetectFaxChans[ctrlID]:
			if detectevent == DetectFaxFinished {
				out = true
			}

			res.RLock()

			prevevent := res.Event

			res.RUnlock()

			Log.Debug("Got detectevent %s. ctrlID: %s\n", detectevent.String(), ctrlID)

			switch detectevent {
			case DetectFaxFinished:
				res.Lock()

				res.Result.Successful = true
				res.Completed = true

				res.Unlock()

				Log.Debug("Detect finished. ctrlID: %s\n", ctrlID)

				if callobj.OnDetectFinished != nil {
					callobj.OnDetectFinished(res)
				}
			case DetectFaxCED:
				fallthrough
			case DetectFaxCNG:
				res.Lock()

				res.Event = detectevent
				res.Result.Type = DetectorFax

				res.Unlock()
			}

			if prevevent != detectevent && callobj.OnDetectUpdate != nil {
				callobj.OnDetectUpdate(res)
			}
		case rawEvent := <-callobj.call.CallDetectRawEventChans[ctrlID]:
			res.Lock()
			res.Result.Event = *rawEvent
			res.Unlock()
		case <-callobj.call.Hangup:
			out = true
		case <-ctx.Done():
			out = true
		}

		if out {
			res.done <- res.Result.Successful
			break
		}
	}
}

// callbacksRunDetectDigit TODO DESCRIPTION
func (callobj *CallObj) callbacksRunDetectDigit(ctx context.Context, ctrlID string, res *DetectDigitAction) {
	for {
		var out bool

		select {
		// get detect events
		case detectevent := <-callobj.call.CallDetectDigitChans[ctrlID]:
			if detectevent == DetectDigitFinished {
				out = true
			}

			res.RLock()

			prevevent := res.Event

			res.RUnlock()

			Log.Debug("Got detectevent %s. ctrlID: %s\n", detectevent.String(), ctrlID)

			switch detectevent {
			case DetectDigitFinished:
				res.Lock()

				res.Result.Successful = true
				res.Completed = true

				res.Unlock()

				Log.Debug("Detect finished. ctrlID: %s\n", ctrlID)

				if callobj.OnDetectFinished != nil {
					callobj.OnDetectFinished(res)
				}
			case DetectDigitZero:
				fallthrough
			case DetectDigitOne:
				fallthrough
			case DetectDigitTwo:
				fallthrough
			case DetectDigitThree:
				fallthrough
			case DetectDigitFour:
				fallthrough
			case DetectDigitFive:
				fallthrough
			case DetectDigitSix:
				fallthrough
			case DetectDigitSeven:
				fallthrough
			case DetectDigitEight:
				fallthrough
			case DetectDigitNine:
				fallthrough
			case DetectDigitPound:
				fallthrough
			case DetectDigitStar:
				res.Lock()

				res.Event = detectevent

				res.Unlock()
			}

			if prevevent != detectevent && callobj.OnDetectUpdate != nil {
				callobj.OnDetectUpdate(res)
			}
		case rawEvent := <-callobj.call.CallDetectRawEventChans[ctrlID]:
			res.Lock()
			res.Result.Event = *rawEvent
			res.Unlock()
		case <-callobj.call.Hangup:
			out = true
		case <-ctx.Done():
			out = true
		}

		if out {
			res.done <- res.Result.Successful
			break
		}
	}
}

// DetectMachineAsync TODO DESCRIPTION
func (callobj *CallObj) DetectMachineAsync(det *DetectMachineParams) (*DetectMachineAction, error) {
	res := new(DetectMachineAction)

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
			ctrlID := <-callobj.call.CallDetectMachineControlID

			callobj.callbacksRunDetectMachine(callobj.Calling.Ctx, ctrlID, res)
		}()

		newCtrlID, _ := GenUUIDv4()
		res.Lock()
		res.ControlID = newCtrlID
		res.Unlock()

		err := callobj.Calling.Relay.RelayDetectMachine(callobj.Calling.Ctx, callobj.call, newCtrlID, det, &res.Payload)
		if err != nil {
			res.Lock()
			res.err = err
			res.Completed = true
			res.Unlock()
		}
		done <- struct{}{}
	}()

	<-done

	return res, nil
}

// DetectDigitAsync TODO DESCRIPTION
func (callobj *CallObj) DetectDigitAsync(det *DetectDigitParams) (*DetectDigitAction, error) {
	res := new(DetectDigitAction)
	res.Result.Type = DetectorDTMF

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
			ctrlID := <-callobj.call.CallDetectDigitControlID

			callobj.callbacksRunDetectDigit(callobj.Calling.Ctx, ctrlID, res)
		}()

		newCtrlID, _ := GenUUIDv4()
		res.Lock()
		res.ControlID = newCtrlID
		res.Unlock()

		err := callobj.Calling.Relay.RelayDetectDigit(callobj.Calling.Ctx, callobj.call, newCtrlID, det.Digits, &res.Payload)
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

// DetectFaxAsync TODO DESCRIPTION
func (callobj *CallObj) DetectFaxAsync(det *DetectFaxParams) (*DetectFaxAction, error) {
	res := new(DetectFaxAction)
	res.Result.Type = DetectorFax

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
			ctrlID := <-callobj.call.CallDetectFaxControlID

			callobj.callbacksRunDetectFax(callobj.Calling.Ctx, ctrlID, res)
		}()

		newCtrlID, _ := GenUUIDv4()
		res.Lock()
		res.ControlID = newCtrlID
		res.Unlock()

		err := callobj.Calling.Relay.RelayDetectFax(callobj.Calling.Ctx, callobj.call, newCtrlID, det.Tone, &res.Payload)
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

// DetectAction TODO DESCRIPTION
type DetectAction interface{}

func detectInternalStop(v interface{}) error {
	m, ok := v.(*DetectMachineAction)
	d, ok1 := v.(*DetectDigitAction)
	f, ok2 := v.(*DetectFaxAction)

	var call *CallSession

	var ctrlID string

	if ok {
		if m.CallObj.Calling == nil {
			return errors.New("nil Calling object")
		}

		if m.CallObj.Calling.Relay == nil {
			return errors.New("nil Relay object")
		}

		m.RLock()

		if len(m.ControlID) == 0 {
			m.RUnlock()
			Log.Error("no controlID\n")

			return errors.New("no controlID")
		}

		ctrlID = m.ControlID

		m.RUnlock()

		call = m.CallObj.call

		return m.CallObj.Calling.Relay.RelayDetectStop(m.CallObj.Calling.Ctx, call, &ctrlID, &m.Payload)
	}

	if ok1 {
		if d.CallObj.Calling == nil {
			return errors.New("nil Calling object")
		}

		if d.CallObj.Calling.Relay == nil {
			return errors.New("nil Relay object")
		}

		d.RLock()

		if len(d.ControlID) == 0 {
			d.RUnlock()

			Log.Error("no controlID\n")

			return errors.New("no controlID")
		}

		ctrlID = d.ControlID

		d.RUnlock()

		call = d.CallObj.call

		return d.CallObj.Calling.Relay.RelayDetectStop(d.CallObj.Calling.Ctx, call, &ctrlID, &d.Payload)
	}

	if ok2 {
		if f.CallObj.Calling == nil {
			return errors.New("nil Calling object")
		}

		if f.CallObj.Calling.Relay == nil {
			return errors.New("nil Relay object")
		}

		f.RLock()

		if len(f.ControlID) == 0 {
			f.RUnlock()

			Log.Error("no controlID\n")

			return errors.New("no controlID")
		}

		ctrlID = f.ControlID

		f.RUnlock()

		call = f.CallObj.call

		return f.CallObj.Calling.Relay.RelayDetectStop(f.CallObj.Calling.Ctx, call, &ctrlID, &f.Payload)
	}

	return errors.New("type assertion failed")
}

// detectAsyncStop TODO DESCRIPTION
func (detectaction *DetectMachineAction) detectAsyncStop() error {
	return detectInternalStop(detectaction)
}

// Stop TODO DESCRIPTION
func (detectaction *DetectMachineAction) Stop() StopResult {
	res := new(StopResult)

	Log.Debug("*DetectMachineAction) Stop():::")
	detectaction.err = detectaction.detectAsyncStop()

	if detectaction.err == nil {
		res.Successful = <-detectaction.done
	}

	Log.Debug("*DetectMachineAction) Stop()")
	return *res
}

// detectAsyncStop TODO DESCRIPTION
func (detectaction *DetectDigitAction) detectAsyncStop() error {
	return detectInternalStop(detectaction)
}

// Stop TODO DESCRIPTION
func (detectaction *DetectDigitAction) Stop() StopResult {
	res := new(StopResult)

	detectaction.err = detectaction.detectAsyncStop()

	if detectaction.err == nil {
		res.Successful = <-detectaction.done
	}

	Log.Debug("*DetectDigitAction) Stop()")
	return *res
}

// detectAsyncStop TODO DESCRIPTION
func (detectaction *DetectFaxAction) detectAsyncStop() error {
	return detectInternalStop(detectaction)
}

// Stop TODO DESCRIPTION
func (detectaction *DetectFaxAction) Stop() StopResult {
	res := new(StopResult)
	detectaction.err = detectaction.detectAsyncStop()

	if detectaction.err == nil {
		res.Successful = <-detectaction.done
	}

	Log.Debug("*DetectFaxAction) Stop()")
	return *res
}

// GetCompleted TODO DESCRIPTION
func (detectaction *DetectMachineAction) GetCompleted() bool {
	detectaction.RLock()

	ret := detectaction.Completed

	detectaction.RUnlock()

	return ret
}

// GetResult TODO DESCRIPTION
func (detectaction *DetectMachineAction) GetResult() DetectResult {
	detectaction.RLock()

	ret := detectaction.Result

	detectaction.RUnlock()

	return ret
}

// GetDetectorEvent TODO DESCRIPTION
func (detectaction *DetectMachineAction) GetDetectorEvent() DetectMachineEvent {
	detectaction.RLock()

	ret := detectaction.Event

	detectaction.RUnlock()

	return ret
}

// GetCompleted TODO DESCRIPTION
func (detectaction *DetectDigitAction) GetCompleted() bool {
	detectaction.RLock()

	ret := detectaction.Completed

	detectaction.RUnlock()

	return ret
}

// GetResult TODO DESCRIPTION
func (detectaction *DetectDigitAction) GetResult() DetectResult {
	detectaction.RLock()

	ret := detectaction.Result

	detectaction.RUnlock()

	return ret
}

// GetDetectorEvent TODO DESCRIPTION
func (detectaction *DetectDigitAction) GetDetectorEvent() DetectDigitEvent {
	detectaction.RLock()

	ret := detectaction.Event

	detectaction.RUnlock()

	return ret
}

// GetCompleted TODO DESCRIPTION
func (detectaction *DetectFaxAction) GetCompleted() bool {
	detectaction.RLock()

	ret := detectaction.Completed

	detectaction.RUnlock()

	return ret
}

// GetResult TODO DESCRIPTION
func (detectaction *DetectFaxAction) GetResult() DetectResult {
	detectaction.RLock()

	ret := detectaction.Result

	detectaction.RUnlock()

	return ret
}

// GetDetectorEvent TODO DESCRIPTION
func (detectaction *DetectFaxAction) GetDetectorEvent() DetectFaxEvent {
	detectaction.RLock()

	ret := detectaction.Event

	detectaction.RUnlock()

	return ret
}

// GetEvent TODO DESCRIPTION
func (detectaction *DetectMachineAction) GetEvent() *json.RawMessage {
	detectaction.RLock()

	ret := &detectaction.Result.Event

	detectaction.RUnlock()

	return ret
}

// GetEvent TODO DESCRIPTION
func (detectaction *DetectFaxAction) GetEvent() *json.RawMessage {
	detectaction.RLock()

	ret := &detectaction.Result.Event

	detectaction.RUnlock()

	return ret
}

// GetEvent TODO DESCRIPTION
func (detectaction *DetectDigitAction) GetEvent() *json.RawMessage {
	detectaction.RLock()

	ret := &detectaction.Result.Event

	detectaction.RUnlock()

	return ret
}

// GetPayload TODO DESCRIPTION
func (detectaction *DetectMachineAction) GetPayload() *json.RawMessage {
	detectaction.RLock()

	ret := detectaction.Payload

	detectaction.RUnlock()

	return ret
}

// GetPayload TODO DESCRIPTION
func (detectaction *DetectFaxAction) GetPayload() *json.RawMessage {
	detectaction.RLock()

	ret := detectaction.Payload

	detectaction.RUnlock()

	return ret
}

// GetPayload TODO DESCRIPTION
func (detectaction *DetectDigitAction) GetPayload() *json.RawMessage {
	detectaction.RLock()

	ret := detectaction.Payload

	detectaction.RUnlock()

	return ret
}
