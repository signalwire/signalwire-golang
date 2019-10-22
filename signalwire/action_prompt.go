package signalwire

import (
	"context"
	"errors"
	"sync"
)

// CollectResultType keeps the result type of a Collect action
type CollectResultType int

// Collect Result Type constants
const (
	CollectResultError CollectResultType = iota
	CollectResultNoInput
	CollectResultNoMatch
	CollectResultDigit
	CollectResultSpeech
	CollectResultStartOfSpeech
)

func (s CollectResultType) String() string {
	return [...]string{"Error", "No_input", "No_match", "Digit", "Speech", "Start_of_speech"}[s]
}

// CollectResult TODO DESCRIPTION
type CollectResult struct {
	Successful bool
}

type PromptResult CollectResult

// CollectContinue  TODO DESCRIPTION
type CollectContinue int

// Collect Result Type constants
const (
	CollectPartial CollectContinue = iota
	CollectFinal
)

// PromptAction TODO DESCRIPTION
type PromptAction struct {
	CallObj   *CallObj
	ControlID string
	Completed bool
	Result    CollectResult
	State     CollectContinue
	err       error
	sync.RWMutex
}

// IPromptAction TODO DESCRIPTION
type IPromptAction interface {
	playAndCollectAsyncStop() error
	Stop()
	GetCompleted() bool
	GetResult() PlayResult
}

func (callobj *CallObj) checkPlayAndCollectFinished(_ context.Context, ctrlID string, res *CollectResult) (*CollectResult, error) {
	var out bool

	for {
		select {
		case state := <-callobj.call.CallPlayAndCollectChans[ctrlID]:
			if state == CollectFinal {
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

// Prompt TODO DESCRIPTION
func (callobj *CallObj) Prompt(playlist *[]PlayStruct, collect *CollectStruct) (*CollectResult, error) {
	res := new(CollectResult)

	if callobj.Calling == nil {
		return res, errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return res, errors.New("nil Relay object")
	}

	ctrlID, _ := GenUUIDv4()

	err := callobj.Calling.Relay.RelayPlayAndCollect(callobj.Calling.Ctx, callobj.call, ctrlID, playlist, collect)

	if err != nil {
		return res, err
	}

	return callobj.checkPlayAndCollectFinished(callobj.Calling.Ctx, ctrlID, res)
}

// PromptStop TODO DESCRIPTION
func (callobj *CallObj) PromptStop(ctrlID *string) error {
	if callobj.Calling == nil {
		return errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return errors.New("nil Relay object")
	}

	return callobj.Calling.Relay.RelayPlayAndCollectStop(callobj.Calling.Ctx, callobj.call, ctrlID)
}

// callbacksRunPlayAndCollect TODO DESCRIPTION
func (callobj *CallObj) callbacksRunPlayAndCollect(_ context.Context, ctrlID string, res *PromptAction) {
	for {
		var out bool
		select {
		case state := <-callobj.call.CallPlayAndCollectChans[ctrlID]:
			switch state {
			case CollectFinal:
				res.Lock()

				res.State = state
				res.Result.Successful = true
				res.Completed = true

				res.Unlock()

				Log.Debug("Prompt finished. ctrlID: %s res [%p] Completed [%v] Successful [%v]\n", ctrlID, res, res.Completed, res.Result.Successful)

				out = true

				if callobj.OnPrompt != nil {
					callobj.OnPrompt(res)
				}

			case CollectPartial:
			default:
				Log.Debug("Unknown state. ctrlID: %s\n", ctrlID)
			}

		case <-callobj.call.Hangup:
			out = true
		}

		if out {
			break
		}
	}
}

// PromptAsync TODO DESCRIPTION
func (callobj *CallObj) PromptAsync(playlist *[]PlayStruct, collect *CollectStruct) (*PromptAction, error) {
	res := new(PromptAction)

	if callobj.Calling == nil {
		return res, errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return res, errors.New("nil Relay object")
	}

	res.CallObj = callobj

	fired := make(chan struct{})

	go func() {
		go func() {
			// wait to get control ID (buffered channel)
			ctrlID := <-callobj.call.CallPlayAndCollectControlID

			callobj.callbacksRunPlayAndCollect(callobj.Calling.Ctx, ctrlID, res)
		}()

		newCtrlID, _ := GenUUIDv4()
		res.Lock()
		res.ControlID = newCtrlID
		res.Unlock()

		err := callobj.Calling.Relay.RelayPlayAndCollect(callobj.Calling.Ctx, callobj.call, newCtrlID, playlist, collect)

		if err != nil {
			res.Lock()

			res.err = err

			res.Completed = true

			res.Unlock()
		}
		fired <- struct{}{}
	}()

	<-fired

	return res, res.err
}

// ctrlIDCopy TODO DESCRIPTION
func (action *PromptAction) ctrlIDCopy() (string, error) {
	action.RLock()

	if len(action.ControlID) == 0 {
		action.RUnlock()
		return "", errors.New("no controlID")
	}

	c := action.ControlID

	action.RUnlock()

	return c, nil
}

// playAndCollectAsyncStop TODO DESCRIPTION
func (action *PromptAction) playAndCollectAsyncStop() error {
	if action.CallObj.Calling == nil {
		return errors.New("nil Calling object")
	}

	if action.CallObj.Calling.Relay == nil {
		return errors.New("nil Relay object")
	}

	c, err := action.ctrlIDCopy()
	if err != nil {
		return err
	}

	call := action.CallObj.call

	return action.CallObj.Calling.Relay.RelayPlayAndCollectStop(action.CallObj.Calling.Ctx, call, &c)
}

// Stop TODO DESCRIPTION
func (action *PromptAction) Stop() {
	action.err = action.playAndCollectAsyncStop()
}

// GetCompleted TODO DESCRIPTION
func (action *PromptAction) GetCompleted() bool {
	action.RLock()

	ret := action.Completed

	action.RUnlock()

	return ret
}

// GetResult TODO DESCRIPTION
func (action *PromptAction) GetResult() CollectResult {
	action.RLock()

	ret := action.Result

	action.RUnlock()

	return ret
}

// PromptVolume TODO DESCRIPTION
func (action *PromptAction) PromptVolume(vol float64) (*PlayVolumeResult, error) {
	res := new(PlayVolumeResult)

	if action.CallObj.Calling == nil {
		return res, errors.New("nil Calling object")
	}

	if action.CallObj.Calling.Relay == nil {
		return res, errors.New("nil Relay object")
	}

	c, err := action.ctrlIDCopy()
	if err != nil {
		return res, err
	}

	call := action.CallObj.call

	err = action.CallObj.Calling.Relay.RelayPlayAndCollectVolume(action.CallObj.Calling.Ctx, call, &c, vol)

	if err != nil {
		return res, err
	}

	res.Successful = true

	return res, nil
}
