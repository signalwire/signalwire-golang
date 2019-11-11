package signalwire

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

// PlayState keeps the state of a play action
type PlayState int

// Call state constants
const (
	PlayPlaying PlayState = iota
	PlayPaused
	PlayError
	PlayFinished
)

func (s PlayState) String() string {
	return [...]string{"Playing", "Paused", "Error", "Finished"}[s]
}

// PlayResult TODO DESCRIPTION
type PlayResult struct {
	Successful bool
	Event      json.RawMessage
}

// PlayVolumeResult TODO DESCRIPTION
type PlayVolumeResult struct {
	Successful bool
	Event      json.RawMessage
}

// PlayPauseResult TODO DESCRIPTION
type PlayPauseResult struct {
	Successful bool
	Event      json.RawMessage
}

// PlayResumeResult TODO DESCRIPTION
type PlayResumeResult struct {
	Successful bool
	Event      json.RawMessage
}

// PlayAction TODO DESCRIPTION
type PlayAction struct {
	CallObj   *CallObj
	ControlID string
	Completed bool
	Result    PlayResult
	State     PlayState
	Payload   *json.RawMessage
	err       error
	done      chan bool
	sync.RWMutex
}

// IPlayAction TODO DESCRIPTION
type IPlayAction interface {
	playAsyncStop() error
	Stop()
	GetCompleted() bool
	GetResult() PlayResult
	Volume() error
	Pause() error
	Resume() error
}

// PlayAudio TODO DESCRIPTION
type PlayAudio struct {
	URL string
}

// PlayTTS TODO DESCRIPTION
type PlayTTS struct {
	Text     string
	Language string
	Gender   string
}

// PlaySilence TODO DESCRIPTION
type PlaySilence struct {
	Duration float64
}

// PlayGenericParams TODO DESCRIPTION
type PlayGenericParams struct {
	SpecificParams interface{}
}

// TTSParamsInternal  TODO DESCRIPTION
type TTSParamsInternal struct {
	text     string
	language string
	gender   string
}

// PlayAudio TODO DESCRIPTION
func (callobj *CallObj) PlayAudio(s string) (*PlayResult, error) {
	a := new(PlayAction)

	if callobj.Calling == nil {
		return &a.Result, errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return &a.Result, errors.New("nil Relay object")
	}

	ctrlID, _ := GenUUIDv4()

	err := callobj.Calling.Relay.RelayPlayAudio(callobj.Calling.Ctx, callobj.call, ctrlID, s, nil)

	if err != nil {
		return &a.Result, err
	}

	callobj.callbacksRunPlay(callobj.Calling.Ctx, ctrlID, a, true)

	return &a.Result, nil
}

// PlayTTS TODO DESCRIPTION
func (callobj *CallObj) PlayTTS(text, language, gender string) (*PlayResult, error) {
	a := new(PlayAction)

	if callobj.Calling == nil {
		return &a.Result, errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return &a.Result, errors.New("nil Relay object")
	}

	ctrlID, _ := GenUUIDv4()

	var tts TTSParamsInternal

	tts.text = text
	tts.language = language
	tts.gender = gender

	err := callobj.Calling.Relay.RelayPlayTTS(callobj.Calling.Ctx, callobj.call, ctrlID, &tts, nil)

	if err != nil {
		return &a.Result, err
	}

	callobj.callbacksRunPlay(callobj.Calling.Ctx, ctrlID, a, true)

	return &a.Result, nil
}

// PlaySilence TODO DESCRIPTION
func (callobj *CallObj) PlaySilence(duration float64) (*PlayResult, error) {
	a := new(PlayAction)

	if callobj.Calling == nil {
		return &a.Result, errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return &a.Result, errors.New("nil Relay object")
	}

	ctrlID, _ := GenUUIDv4()
	err := callobj.Calling.Relay.RelayPlaySilence(callobj.Calling.Ctx, callobj.call, ctrlID, duration, nil)

	if err != nil {
		return &a.Result, err
	}

	callobj.callbacksRunPlay(callobj.Calling.Ctx, ctrlID, a, true)

	return &a.Result, nil
}

// PlayRingtone TODO DESCRIPTION
func (callobj *CallObj) PlayRingtone(name string, duration float64) (*PlayResult, error) {
	a := new(PlayAction)

	if callobj.Calling == nil {
		return &a.Result, errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return &a.Result, errors.New("nil Relay object")
	}

	ctrlID, _ := GenUUIDv4()
	err := callobj.Calling.Relay.RelayPlayRingtone(callobj.Calling.Ctx, callobj.call, ctrlID, name, duration, nil)

	if err != nil {
		return &a.Result, err
	}

	callobj.callbacksRunPlay(callobj.Calling.Ctx, ctrlID, a, true)

	return &a.Result, nil
}

// PlayStop TODO DESCRIPTION
func (callobj *CallObj) PlayStop(ctrlID *string) error {
	if callobj.Calling == nil {
		return errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return errors.New("nil Relay object")
	}

	return callobj.Calling.Relay.RelayPlayStop(callobj.Calling.Ctx, callobj.call, ctrlID, nil)
}

// callbacksRunPlay TODO DESCRIPTION
func (callobj *CallObj) callbacksRunPlay(ctx context.Context, ctrlID string, res *PlayAction, norunCB bool) {
	var out bool

	timer := time.NewTimer(BroadcastEventTimeout * time.Second)

	for {
		select {
		case <-timer.C:
			out = true
		// get play states
		case playstate := <-callobj.call.CallPlayChans[ctrlID]:
			res.RLock()

			prevstate := res.State

			res.RUnlock()

			switch playstate {
			case PlayFinished:
				res.Lock()

				res.State = playstate
				res.Result.Successful = true
				res.Completed = true

				res.Unlock()

				Log.Debug("Play finished. ctrlID: %s res [%p] Completed [%v] Successful [%v]\n", ctrlID, res, res.Completed, res.Result.Successful)

				out = true

				if callobj.OnPlayFinished != nil && !norunCB {
					callobj.OnPlayFinished(res)
				}

			case PlayPlaying:
				timer.Reset(MaxCallDuration * time.Second)
				res.Lock()

				res.State = playstate

				res.Unlock()

				Log.Debug("Playing. ctrlID: %s\n", ctrlID)

				if callobj.OnPlayPlaying != nil && !norunCB {
					callobj.OnPlayPlaying(res)
				}
			case PlayError:
				Log.Debug("Play error. ctrlID: %s\n", ctrlID)

				res.Lock()

				res.Completed = true
				res.State = playstate

				res.Unlock()

				out = true

				if callobj.OnPlayError != nil && !norunCB {
					callobj.OnPlayError(res)
				}
			case PlayPaused:
				timer.Reset(MaxCallDuration * time.Second)
				res.Lock()

				res.State = playstate

				res.Unlock()

				Log.Debug("Play paused. ctrlID: %s\n", ctrlID)

				if callobj.OnPlayPaused != nil && !norunCB {
					callobj.OnPlayPaused(res)
				}
			default:
				Log.Debug("Unknown state. ctrlID: %s\n", ctrlID)
			}

			if prevstate != playstate && callobj.OnPlayStateChange != nil && !norunCB {
				callobj.OnPlayStateChange(res)
			}
		case rawEvent := <-callobj.call.CallPlayRawEventChans[ctrlID]:
			res.Lock()
			res.Result.Event = *rawEvent
			res.Unlock()

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

// PlaySilenceAsync TODO DESCRIPTION
func (callobj *CallObj) PlaySilenceAsync(duration float64) (*PlayAction, error) {
	res := new(PlayAction)

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
			ctrlID := <-callobj.call.CallPlayControlIDs
			callobj.callbacksRunPlay(callobj.Calling.Ctx, ctrlID, res, false)
		}()

		newCtrlID, _ := GenUUIDv4()

		res.Lock()

		res.ControlID = newCtrlID

		res.Unlock()

		err := callobj.Calling.Relay.RelayPlaySilence(callobj.Calling.Ctx, callobj.call, newCtrlID, duration, &res.Payload)

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

// PlayRingtoneAsync TODO DESCRIPTION
func (callobj *CallObj) PlayRingtoneAsync(name string, duration float64) (*PlayAction, error) {
	res := new(PlayAction)

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
			ctrlID := <-callobj.call.CallPlayControlIDs

			callobj.callbacksRunPlay(callobj.Calling.Ctx, ctrlID, res, false)
		}()

		newCtrlID, _ := GenUUIDv4()

		res.Lock()

		res.ControlID = newCtrlID

		res.Unlock()

		err := callobj.Calling.Relay.RelayPlayRingtone(callobj.Calling.Ctx, callobj.call, newCtrlID, name, duration, &res.Payload)

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

// PlayTTSAsync TODO DESCRIPTION
func (callobj *CallObj) PlayTTSAsync(text, language, gender string) (*PlayAction, error) {
	res := new(PlayAction)

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
			ctrlID := <-callobj.call.CallPlayControlIDs

			callobj.callbacksRunPlay(callobj.Calling.Ctx, ctrlID, res, false)
		}()

		newCtrlID, _ := GenUUIDv4()

		res.Lock()

		res.ControlID = newCtrlID

		res.Unlock()

		var tts TTSParamsInternal

		tts.text = text
		tts.language = language
		tts.gender = gender

		err := callobj.Calling.Relay.RelayPlayTTS(callobj.Calling.Ctx, callobj.call, newCtrlID, &tts, &res.Payload)

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

// PlayAudioAsync TODO DESCRIPTION
func (callobj *CallObj) PlayAudioAsync(url string) (*PlayAction, error) {
	res := new(PlayAction)

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
			ctrlID := <-callobj.call.CallPlayControlIDs

			callobj.callbacksRunPlay(callobj.Calling.Ctx, ctrlID, res, false)
		}()

		newCtrlID, _ := GenUUIDv4()

		res.Lock()

		res.ControlID = newCtrlID

		res.Unlock()

		err := callobj.Calling.Relay.RelayPlayAudio(callobj.Calling.Ctx, callobj.call, newCtrlID, url, &res.Payload)

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

// ctrlIDCopy TODO DESCRIPTION
func (playaction *PlayAction) ctrlIDCopy() (string, error) {
	playaction.RLock()

	if len(playaction.ControlID) == 0 {
		playaction.RUnlock()
		return "", errors.New("no controlID")
	}

	c := playaction.ControlID

	playaction.RUnlock()

	return c, nil
}

// PlayAudioAsyncStop TODO DESCRIPTION
func (playaction *PlayAction) playAsyncStop() error {
	if playaction.CallObj.Calling == nil {
		return errors.New("nil Calling object")
	}

	if playaction.CallObj.Calling.Relay == nil {
		return errors.New("nil Relay object")
	}

	c, err := playaction.ctrlIDCopy()
	if err != nil {
		return err
	}

	call := playaction.CallObj.call

	return playaction.CallObj.Calling.Relay.RelayPlayStop(playaction.CallObj.Calling.Ctx, call, &c, &playaction.Payload)
}

// Stop TODO DESCRIPTION
func (playaction *PlayAction) Stop() StopResult {
	res := new(StopResult)
	playaction.err = playaction.playAsyncStop()

	if playaction.err == nil {
		waitStop(res, playaction.done)
	}

	return *res
}

// GetCompleted TODO DESCRIPTION
func (playaction *PlayAction) GetCompleted() bool {
	playaction.RLock()

	ret := playaction.Completed

	playaction.RUnlock()

	return ret
}

// GetPayload TODO DESCRIPTION
func (playaction *PlayAction) GetPayload() *json.RawMessage {
	playaction.RLock()

	ret := playaction.Payload

	playaction.RUnlock()

	return ret
}

// GetResult TODO DESCRIPTION
func (playaction *PlayAction) GetResult() PlayResult {
	playaction.RLock()

	ret := playaction.Result

	playaction.RUnlock()

	return ret
}

// GetError TODO DESCRIPTION
func (playaction *PlayAction) GetError() error {
	playaction.RLock()

	ret := playaction.err

	playaction.RUnlock()

	return ret
}

// GetState TODO DESCRIPTION
func (playaction *PlayAction) GetState() PlayState {
	playaction.RLock()

	ret := playaction.State

	playaction.RUnlock()

	return ret
}

// GetSuccessful TODO DESCRIPTION
func (playaction *PlayAction) GetSuccessful() bool {
	playaction.RLock()

	ret := playaction.Result.Successful

	playaction.RUnlock()

	return ret
}

// GetEvent TODO DESCRIPTION
func (playaction *PlayAction) GetEvent() *json.RawMessage {
	playaction.RLock()

	ret := &playaction.Result.Event

	playaction.RUnlock()

	return ret
}

// PlayVolume TODO DESCRIPTION
func (playaction *PlayAction) PlayVolume(vol float64) (*PlayVolumeResult, error) {
	res := new(PlayVolumeResult)

	if playaction.CallObj.Calling == nil {
		return res, errors.New("nil Calling object")
	}

	if playaction.CallObj.Calling.Relay == nil {
		return res, errors.New("nil Relay object")
	}

	c, err := playaction.ctrlIDCopy()
	if err != nil {
		return res, err
	}

	call := playaction.CallObj.call

	err = playaction.CallObj.Calling.Relay.RelayPlayVolume(playaction.CallObj.Calling.Ctx, call, &c, vol, nil)

	if err != nil {
		return res, err
	}

	res.Successful = true

	return res, nil
}

// PlayPause TODO DESCRIPTION
func (playaction *PlayAction) PlayPause() (*PlayVolumeResult, error) {
	res := new(PlayVolumeResult)

	if playaction.CallObj.Calling == nil {
		return res, errors.New("nil Calling object")
	}

	if playaction.CallObj.Calling.Relay == nil {
		return res, errors.New("nil Relay object")
	}

	c, err := playaction.ctrlIDCopy()
	if err != nil {
		return res, err
	}

	call := playaction.CallObj.call

	err = playaction.CallObj.Calling.Relay.RelayPlayPause(playaction.CallObj.Calling.Ctx, call, &c, &playaction.Payload)

	if err != nil {
		return res, err
	}

	res.Successful = true

	return res, nil
}

// PlayResume TODO DESCRIPTION
func (playaction *PlayAction) PlayResume() (*PlayVolumeResult, error) {
	res := new(PlayVolumeResult)

	if playaction.CallObj.Calling == nil {
		return res, errors.New("nil Calling object")
	}

	if playaction.CallObj.Calling.Relay == nil {
		return res, errors.New("nil Relay object")
	}

	c, err := playaction.ctrlIDCopy()
	if err != nil {
		return res, err
	}

	call := playaction.CallObj.call

	err = playaction.CallObj.Calling.Relay.RelayPlayResume(playaction.CallObj.Calling.Ctx, call, &c, &playaction.Payload)

	if err != nil {
		return res, err
	}

	res.Successful = true

	return res, nil
}

// Play TODO DESCRIPTION
func (callobj *CallObj) Play(g [MaxPlay]PlayGenericParams) ([]PlayResult, error) {
	var result []PlayResult

	for _, playParams := range g {
		var ok bool

		_, ok = playParams.SpecificParams.(*PlayAudio)
		if ok {
			params, _ := playParams.SpecificParams.(*PlayAudio)

			res, err := callobj.PlayAudio(params.URL)
			if err != nil {
				return result, err
			}

			result = append(result, *res)
		}

		_, ok = playParams.SpecificParams.(*PlayTTS)
		if ok {
			params, _ := playParams.SpecificParams.(*PlayTTS)

			res, err := callobj.PlayTTS(params.Text, params.Language, params.Gender)
			if err != nil {
				return result, err
			}

			result = append(result, *res)
		}

		_, ok = playParams.SpecificParams.(*PlaySilence)
		if ok {
			params, _ := playParams.SpecificParams.(*PlaySilence)

			res, err := callobj.PlaySilence(params.Duration)
			if err != nil {
				return result, err
			}

			result = append(result, *res)
		}
	}

	return result, nil
}

// PlayAsync TODO DESCRIPTION
func (callobj *CallObj) PlayAsync(g [MaxPlay]PlayGenericParams) ([]*PlayAction, error) {
	var result []*PlayAction

	for _, playParams := range g {
		var ok bool
		_, ok = playParams.SpecificParams.(*PlayAudio)

		if ok {
			params, _ := playParams.SpecificParams.(*PlayAudio)

			res, err := callobj.PlayAudioAsync(params.URL)
			if err != nil {
				return result, err
			}

			result = append(result, res)
		}

		_, ok = playParams.SpecificParams.(*PlayTTS)

		if ok {
			params, _ := playParams.SpecificParams.(*PlayTTS)

			res, err := callobj.PlayTTSAsync(params.Text, params.Language, params.Gender)
			if err != nil {
				return result, err
			}

			result = append(result, res)
		}

		_, ok = playParams.SpecificParams.(*PlaySilence)

		if ok {
			params, _ := playParams.SpecificParams.(*PlaySilence)

			res, err := callobj.PlaySilenceAsync(params.Duration)
			if err != nil {
				return result, err
			}

			result = append(result, res)
		}
	}

	return result, nil
}
