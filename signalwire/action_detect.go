package signalwire

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
)

// DetectorType type of running detector
type DetectorType int

// TODO DESCRIPTION
const (
	MachineDetector DetectorType = iota
	FaxDetector
	DigitDetector
)

func (s DetectorType) String() string {
	return [...]string{"Machine", "Fax", "Digit"}[s]
}

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
	Result     string
	Event      json.RawMessage
}

// DetectAction TODO DESCRIPTION
type DetectAction struct {
	CallObj      *CallObj
	ControlID    string
	Result       DetectResult
	detEvent     interface{}
	DetectorType DetectorType
	Payload      *json.RawMessage
	err          error
	done         chan bool
	sync.RWMutex
	waitForBeep bool
	Completed   bool
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

// IDetectAction TODO DESCRIPTION
type IDetectAction interface {
	detectAsyncStop() error
	Stop()
}

// DetectMachineParams TODO DESCRIPTION
type DetectMachineParams struct {
	InitialTimeout        float64
	EndSilenceTimeout     float64
	MachineVoiceThreshold float64
	MachineWordsThreshold float64
	WaitForBeep           bool // special param that does not get sent
}

// AMD TODO DESCRIPTION
func (callobj *CallObj) AMD(det *DetectMachineParams) (*DetectResult, error) {
	return callobj.DetectMachine(det)
}

// AMDAsync TODO DESCRIPTION
func (callobj *CallObj) AMDAsync(det *DetectMachineParams) (*DetectAction, error) {
	return callobj.DetectMachineAsync(det)
}

// DetectMachine TODO DESCRIPTION
func (callobj *CallObj) DetectMachine(det *DetectMachineParams) (*DetectResult, error) {
	a := new(DetectAction)

	if callobj.Calling == nil {
		return &a.Result, errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return &a.Result, errors.New("nil Relay object")
	}

	ctrlID, _ := GenUUIDv4()

	var detInternal DetectMachineParamsInternal
	detInternal.InitialTimeout = det.InitialTimeout
	detInternal.EndSilenceTimeout = det.EndSilenceTimeout
	detInternal.MachineVoiceThreshold = det.MachineVoiceThreshold
	detInternal.MachineWordsThreshold = det.MachineWordsThreshold

	err := callobj.Calling.Relay.RelayDetectMachine(callobj.Calling.Ctx, callobj.call, ctrlID, &detInternal, nil)

	if err != nil {
		return &a.Result, err
	}

	a.waitForBeep = det.WaitForBeep
	callobj.callbacksRunDetectMachine(callobj.Calling.Ctx, ctrlID, a)

	return &a.Result, nil
}

// DetectFax TODO DESCRIPTION
func (callobj *CallObj) DetectFax(det *DetectFaxParams) (*DetectResult, error) {
	a := new(DetectAction)

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
	a := new(DetectAction)

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
func (callobj *CallObj) callbacksRunDetectMachine(ctx context.Context, ctrlID string, res *DetectAction) {
	for {
		var out bool

		select {
		// get detect events
		case detectevent := <-callobj.call.CallDetectMachineChans[ctrlID]:
			if detectevent == DetectMachineFinished {
				out = true
			}

			res.RLock()

			prevevent := res.detEvent

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
				if !res.waitForBeep {
					res.Lock()

					res.detEvent = detectevent
					res.Result.Result = detectevent.String()
					res.Result.Type = DetectorMachine

					res.Result.Successful = true
					res.Completed = true

					res.Unlock()

					res.Stop()

					out = true

					if callobj.OnDetectFinished != nil {
						callobj.OnDetectFinished(res)
					}
				}

			case DetectMachineHuman:
				res.Lock()

				res.detEvent = detectevent
				res.Result.Result = detectevent.String()
				res.Result.Type = DetectorHuman

				res.Result.Successful = true
				res.Completed = true

				res.Unlock()

				res.Stop()

				out = true

				if callobj.OnDetectFinished != nil {
					callobj.OnDetectFinished(res)
				}

			case DetectMachineUnknown:
				res.Lock()

				res.detEvent = detectevent
				res.Result.Result = detectevent.String()
				res.Result.Type = DetectorUnknown

				res.Result.Successful = true
				res.Completed = true

				res.Unlock()

				res.Stop()

				out = true

				if callobj.OnDetectFinished != nil {
					callobj.OnDetectFinished(res)
				}

			case DetectMachineReady:
				if res.waitForBeep {
					res.Lock()

					res.detEvent = detectevent
					res.Result.Result = detectevent.String()
					res.Result.Type = DetectorUnknown

					res.Result.Successful = true
					res.Completed = true

					res.Unlock()

					res.Stop()

					out = true

					if callobj.OnDetectFinished != nil {
						callobj.OnDetectFinished(res)
					}
				}

			case DetectMachineNotReady:
				res.Lock()

				res.detEvent = detectevent
				res.Result.Result = detectevent.String()

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

// callbacksRunDetectFax TODO DESCRIPTION
func (callobj *CallObj) callbacksRunDetectFax(ctx context.Context, ctrlID string, res *DetectAction) {
	for {
		var out bool

		select {
		// get detect events
		case detectevent := <-callobj.call.CallDetectFaxChans[ctrlID]:
			if detectevent == DetectFaxFinished {
				out = true
			}

			res.RLock()

			prevevent := res.detEvent

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

				res.detEvent = detectevent
				res.Result.Type = DetectorFax
				res.Result.Result = detectevent.String()

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
func (callobj *CallObj) callbacksRunDetectDigit(ctx context.Context, ctrlID string, res *DetectAction) {
	for {
		var out bool

		select {
		// get detect events
		case detectevent := <-callobj.call.CallDetectDigitChans[ctrlID]:
			if detectevent == DetectDigitFinished {
				out = true
			}

			res.RLock()

			prevevent := res.detEvent

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

				res.detEvent = detectevent
				res.Result.Result = detectevent.String()
				res.Result.Type = DetectorDTMF

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
func (callobj *CallObj) DetectMachineAsync(det *DetectMachineParams) (*DetectAction, error) {
	res := new(DetectAction)

	if callobj.Calling == nil {
		return res, errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return res, errors.New("nil Relay object")
	}

	res.DetectorType = MachineDetector
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

		var detInternal DetectMachineParamsInternal
		detInternal.InitialTimeout = det.InitialTimeout
		detInternal.EndSilenceTimeout = det.EndSilenceTimeout
		detInternal.MachineVoiceThreshold = det.MachineVoiceThreshold
		detInternal.MachineWordsThreshold = det.MachineWordsThreshold

		err := callobj.Calling.Relay.RelayDetectMachine(callobj.Calling.Ctx, callobj.call, newCtrlID, &detInternal, &res.Payload)
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
func (callobj *CallObj) DetectDigitAsync(det *DetectDigitParams) (*DetectAction, error) {
	res := new(DetectAction)
	res.Result.Type = DetectorDTMF

	if callobj.Calling == nil {
		return res, errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return res, errors.New("nil Relay object")
	}

	res.CallObj = callobj
	res.DetectorType = DigitDetector
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
func (callobj *CallObj) DetectFaxAsync(det *DetectFaxParams) (*DetectAction, error) {
	res := new(DetectAction)
	res.Result.Type = DetectorFax

	if callobj.Calling == nil {
		return res, errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return res, errors.New("nil Relay object")
	}

	res.CallObj = callobj
	res.DetectorType = FaxDetector
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

func detectInternalStop(v interface{}) error {
	m, ok := v.(*DetectAction)

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

	return errors.New("type assertion failed")
}

// detectAsyncStop TODO DESCRIPTION
func (detectaction *DetectAction) detectAsyncStop() error {
	return detectInternalStop(detectaction)
}

// Stop TODO DESCRIPTION
func (detectaction *DetectAction) Stop() StopResult {
	res := new(StopResult)

	detectaction.err = detectaction.detectAsyncStop()

	if detectaction.err == nil {
		waitStop(res, detectaction.done)
	}

	return *res
}

// GetCompleted TODO DESCRIPTION
func (detectaction *DetectAction) GetCompleted() bool {
	detectaction.RLock()

	ret := detectaction.Completed

	detectaction.RUnlock()

	return ret
}

// GetResult TODO DESCRIPTION
func (detectaction *DetectAction) GetResult() DetectResult {
	detectaction.RLock()

	ret := detectaction.Result

	detectaction.RUnlock()

	return ret
}

// GetDetectorEvent TODO DESCRIPTION
func (detectaction *DetectAction) GetDetectorEvent() interface{} {
	var ret interface{}

	var ok bool

	detectaction.RLock()

	switch detectaction.DetectorType {
	case MachineDetector:
		ret, ok = detectaction.detEvent.(DetectMachineEvent)
		if !ok {
			Log.Error("type assertion failed")
		}
	case FaxDetector:
		ret, ok = detectaction.detEvent.(DetectFaxEvent)
		if !ok {
			Log.Error("type assertion failed")
		}
	case DigitDetector:
		ret, ok = detectaction.detEvent.(DetectDigitEvent)
		if !ok {
			Log.Error("type assertion failed")
		}
	}

	detectaction.RUnlock()

	return ret
}

// GetEvent TODO DESCRIPTION
func (detectaction *DetectAction) GetEvent() *json.RawMessage {
	detectaction.RLock()

	ret := &detectaction.Result.Event

	detectaction.RUnlock()

	return ret
}

// GetPayload TODO DESCRIPTION
func (detectaction *DetectAction) GetPayload() *json.RawMessage {
	detectaction.RLock()

	ret := detectaction.Payload

	detectaction.RUnlock()

	return ret
}

// GetControlID TODO DESCRIPTION
func (detectaction *DetectAction) GetControlID() string {
	detectaction.RLock()

	ret := detectaction.ControlID

	detectaction.RUnlock()

	return ret
}
