package signalwire

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

// RecordState keeps the state of a play action
type RecordState int

// Recording state constants
const (
	RecordRecording RecordState = iota
	RecordFinished
	RecordNoInput
	RecordPaused
)

func (s RecordState) String() string {
	return [...]string{"Recording", "Finished", "No_input"}[s]
}

// RecordDirection keeps the direction of a recording
type RecordDirection int

// Recording state constants
const (
	RecordDirectionListen RecordDirection = iota
	RecordDirectionSpeak
	RecordDirectionBoth
)

func (s RecordDirection) String() string {
	return [...]string{"Listen", "Speak", "Both"}[s]
}

// RecordResult TODO DESCRIPTION
type RecordResult struct {
	Successful bool
	URL        string
	Duration   uint
	Size       uint
	Event      json.RawMessage
}

// RecordAction TODO DESCRIPTION
type RecordAction struct {
	CallObj   *CallObj
	ControlID string
	Completed bool
	Result    RecordResult
	State     RecordState
	URL       string
	Payload   *json.RawMessage
	err       error
	done      chan bool
	sync.RWMutex
}

// IRecordAction TODO DESCRIPTION
type IRecordAction interface {
	recordAudioAsyncStop() error
	Stop()
}

// RecordAudio TODO DESCRIPTION
func (callobj *CallObj) RecordAudio(rec *RecordParams) (*RecordResult, error) {
	a := new(RecordAction)

	if callobj.Calling == nil {
		return &a.Result, errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return &a.Result, errors.New("nil Relay object")
	}

	ctrlID, _ := GenUUIDv4()
	err := callobj.Calling.Relay.RelayRecordAudio(callobj.Calling.Ctx, callobj.call, ctrlID, rec, nil)

	if err != nil {
		return &a.Result, err
	}

	callobj.callbacksRunRecord(callobj.Calling.Ctx, ctrlID, a, true)

	return &a.Result, nil
}

// RecordAudioStop TODO DESCRIPTION
func (callobj *CallObj) RecordAudioStop(ctrlID *string) error {
	if callobj.Calling == nil {
		return errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return errors.New("nil Relay object")
	}

	return callobj.Calling.Relay.RelayRecordAudioStop(callobj.Calling.Ctx, callobj.call, ctrlID, nil)
}

func (callobj *CallObj) callbacksRunRecord(ctx context.Context, ctrlID string, res *RecordAction, norunCB bool) {
	var out bool

	timer := time.NewTimer(BroadcastEventTimeout * time.Second)

	for {
		select {
		case <-timer.C:
			out = true
		case state := <-callobj.call.CallRecordChans[ctrlID]:
			res.RLock()

			prevstate := res.State

			res.RUnlock()

			switch state {
			case RecordFinished:
				res.Lock()

				res.Result.Successful = true
				res.Completed = true
				res.State = state

				res.Unlock()

				Log.Debug("Record finished. ctrlID: %s\n", ctrlID)

				callobj.call.RemoveAction(ctrlID)

				out = true

				if callobj.OnRecordFinished != nil && !norunCB {
					callobj.OnRecordFinished(res)
				}
			case RecordRecording:
				timer.Reset(MaxCallDuration * time.Second)
				res.Lock()

				res.State = state

				res.Unlock()

				Log.Debug("Recording. ctrlID: %s\n", ctrlID)

				if callobj.OnRecordRecording != nil && !norunCB {
					callobj.OnRecordRecording(res)
				}
			case RecordNoInput:
				Log.Debug("No input for recording. ctrlID: %s\n", ctrlID)
				res.Lock()

				res.Completed = true
				res.State = state

				res.Unlock()

				out = true

				if callobj.OnRecordNoInput != nil && !norunCB {
					callobj.OnRecordNoInput(res)
				}
			case RecordPaused:
				timer.Reset(MaxCallDuration * time.Second)
				res.Lock()

				res.State = state

				res.Unlock()

				Log.Debug("Recording paused. ctrlID: %s\n", ctrlID)

				out = true

				if callobj.OnRecordPaused != nil {
					callobj.OnRecordPaused(res)
				}
			}

			if prevstate != state && callobj.OnRecordStateChange != nil && !norunCB {
				callobj.OnRecordStateChange(res)
			}
		case params := <-callobj.call.CallRecordEventChans[ctrlID]:
			Log.Debug("got params for ctrlID : %s\n", ctrlID)

			res.Lock()

			if len(params.URL) > 0 {
				res.URL = params.URL
				res.Result.URL = params.URL
			}

			if params.Duration > 0 {
				res.Result.Duration = params.Duration
			}

			if params.Size > 0 {
				res.Result.Size = params.Size
			}

			res.Unlock()

			callobj.call.CallRecordReadyChans[ctrlID] <- struct{}{}
		case rawEvent := <-callobj.call.CallRecordRawEventChans[ctrlID]:
			res.Lock()
			res.Result.Event = *rawEvent
			res.Unlock()

			callobj.call.CallRecordReadyChans[ctrlID] <- struct{}{}
		case <-callobj.call.Hangup:
			out = true
		case <-ctx.Done():
			out = true
		}

		if out {
			if !norunCB {
				res.done <- res.Result.Successful
			}

			break
		}
	}
}

// RecordAudioAsync TODO DESCRIPTION
func (callobj *CallObj) RecordAudioAsync(rec *RecordParams) (*RecordAction, error) {
	res := new(RecordAction)

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
			ctrlID := <-callobj.call.CallRecordControlIDs

			// states && callbacks
			callobj.callbacksRunRecord(callobj.Calling.Ctx, ctrlID, res, false)
		}()

		newCtrlID, _ := GenUUIDv4()

		res.Lock()

		res.ControlID = newCtrlID

		res.Unlock()

		err := callobj.Calling.Relay.I.RelayRecordAudio(callobj.Calling.Ctx, callobj.call, newCtrlID, rec, &res.Payload)
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

// RecordAudioAsyncStop TODO DESCRIPTION
func (recordaction *RecordAction) recordAudioAsyncStop() error {
	if recordaction.CallObj.Calling == nil {
		return errors.New("nil Calling object")
	}

	if recordaction.CallObj.Calling.Relay == nil {
		return errors.New("nil Relay object")
	}

	recordaction.RLock()

	if len(recordaction.ControlID) == 0 {
		recordaction.RUnlock()
		Log.Error("no controlID\n")

		return errors.New("no controlID")
	}

	c := recordaction.ControlID

	recordaction.RUnlock()

	call := recordaction.CallObj.call

	return recordaction.CallObj.Calling.Relay.RelayRecordAudioStop(recordaction.CallObj.Calling.Ctx, call, &c, &recordaction.Payload)
}

// Stop TODO DESCRIPTION
func (recordaction *RecordAction) Stop() StopResult {
	res := new(StopResult)
	recordaction.err = recordaction.recordAudioAsyncStop()

	if recordaction.err == nil {
		waitStop(res, recordaction.done)
	}

	return *res
}

// GetCompleted TODO DESCRIPTION
func (recordaction *RecordAction) GetCompleted() bool {
	recordaction.RLock()

	ret := recordaction.Completed

	recordaction.RUnlock()

	return ret
}

// GetResult TODO DESCRIPTION
func (recordaction *RecordAction) GetResult() RecordResult {
	recordaction.RLock()

	ret := recordaction.Result

	recordaction.RUnlock()

	return ret
}

// GetSuccessful TODO DESCRIPTION
func (recordaction *RecordAction) GetSuccessful() bool {
	recordaction.RLock()

	ret := recordaction.Result.Successful

	recordaction.RUnlock()

	return ret
}

// GetURL TODO DESCRIPTION
func (recordaction *RecordAction) GetURL() string {
	recordaction.RLock()

	ret := recordaction.Result.URL

	recordaction.RUnlock()

	return ret
}

// GetDuration TODO DESCRIPTION
func (recordaction *RecordAction) GetDuration() uint {
	recordaction.RLock()

	ret := recordaction.Result.Duration

	recordaction.RUnlock()

	return ret
}

// GetSize TODO DESCRIPTION
func (recordaction *RecordAction) GetSize() uint {
	recordaction.RLock()

	ret := recordaction.Result.Size

	recordaction.RUnlock()

	return ret
}

// GetState TODO DESCRIPTION
func (recordaction *RecordAction) GetState() RecordState {
	recordaction.RLock()

	ret := recordaction.State

	recordaction.RUnlock()

	return ret
}

// GetEvent TODO DESCRIPTION
func (recordaction *RecordAction) GetEvent() *json.RawMessage {
	recordaction.RLock()

	ret := &recordaction.Result.Event

	recordaction.RUnlock()

	return ret
}

// GetPayload TODO DESCRIPTION
func (recordaction *RecordAction) GetPayload() *json.RawMessage {
	recordaction.RLock()

	ret := recordaction.Payload

	recordaction.RUnlock()

	return ret
}

// GetControlID TODO DESCRIPTION
func (recordaction *RecordAction) GetControlID() string {
	recordaction.RLock()

	ret := recordaction.ControlID

	recordaction.RUnlock()

	return ret
}
