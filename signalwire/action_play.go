package signalwire

import (
	"context"
	"errors"
	"sync"
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
}

// PlayVolumeResult TODO DESCRIPTION
type PlayVolumeResult struct {
	Successful bool
}

// PlayPauseResult TODO DESCRIPTION
type PlayPauseResult struct {
	Successful bool
}

// PlayResumeResult TODO DESCRIPTION
type PlayResumeResult struct {
	Successful bool
}

// PlayAction TODO DESCRIPTION
type PlayAction struct {
	CallObj   *CallObj
	ControlID string
	Completed bool
	Result    PlayResult
	State     PlayState
	err       error
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

func (callobj *CallObj) checkPlayFinished(ctx context.Context, ctrlID string, res *PlayResult) (*PlayResult, error) {
	if ret := callobj.call.WaitPlayState(ctx, ctrlID, PlayPlaying); !ret {
		Log.Debug("Playing did not start successfully. CtrlID: %s\n", ctrlID)

		return res, nil
	}

	var out bool

	for {
		select {
		case playstate := <-callobj.call.CallPlayChans[ctrlID]:
			if playstate == PlayFinished {
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

// PlayAudio TODO DESCRIPTION
func (callobj *CallObj) PlayAudio(s string) (*PlayResult, error) {
	res := new(PlayResult)

	if callobj.Calling == nil {
		return res, errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return res, errors.New("nil Relay object")
	}

	ctrlID, _ := GenUUIDv4()
	err := callobj.Calling.Relay.RelayPlayAudio(callobj.Calling.Ctx, callobj.call, ctrlID, s)

	if err != nil {
		return res, err
	}

	return callobj.checkPlayFinished(callobj.Calling.Ctx, ctrlID, res)
}

// PlayTTS TODO DESCRIPTION
func (callobj *CallObj) PlayTTS(text, language, gender string) (*PlayResult, error) {
	res := new(PlayResult)

	if callobj.Calling == nil {
		return res, errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return res, errors.New("nil Relay object")
	}

	ctrlID, _ := GenUUIDv4()
	err := callobj.Calling.Relay.RelayPlayTTS(callobj.Calling.Ctx, callobj.call, ctrlID, text, language, gender)

	if err != nil {
		return res, err
	}

	return callobj.checkPlayFinished(callobj.Calling.Ctx, ctrlID, res)
}

// PlaySilence TODO DESCRIPTION
func (callobj *CallObj) PlaySilence(duration float64) (*PlayResult, error) {
	res := new(PlayResult)

	if callobj.Calling == nil {
		return res, errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return res, errors.New("nil Relay object")
	}

	ctrlID, _ := GenUUIDv4()
	err := callobj.Calling.Relay.RelayPlaySilence(callobj.Calling.Ctx, callobj.call, ctrlID, duration)

	if err != nil {
		return res, err
	}

	return callobj.checkPlayFinished(callobj.Calling.Ctx, ctrlID, res)
}

// PlayRingtone TODO DESCRIPTION
func (callobj *CallObj) PlayRingtone(name string, duration float64) (*PlayResult, error) {
	res := new(PlayResult)

	if callobj.Calling == nil {
		return res, errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return res, errors.New("nil Relay object")
	}

	ctrlID, _ := GenUUIDv4()
	err := callobj.Calling.Relay.RelayPlayRingtone(callobj.Calling.Ctx, callobj.call, ctrlID, name, duration)

	if err != nil {
		return res, err
	}

	return callobj.checkPlayFinished(callobj.Calling.Ctx, ctrlID, res)
}

// PlayStop TODO DESCRIPTION
func (callobj *CallObj) PlayStop(ctrlID *string) error {
	if callobj.Calling == nil {
		return errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return errors.New("nil Relay object")
	}

	return callobj.Calling.Relay.RelayPlayStop(callobj.Calling.Ctx, callobj.call, ctrlID)
}

// callbacksRunPlay TODO DESCRIPTION
func (callobj *CallObj) callbacksRunPlay(_ context.Context, ctrlID string, res *PlayAction) {
	for {
		var out bool

		select {
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

				if callobj.OnPlayFinished != nil {
					callobj.OnPlayFinished(res)
				}

			case PlayPlaying:
				res.Lock()

				res.State = playstate

				res.Unlock()

				Log.Debug("Playing. ctrlID: %s\n", ctrlID)

				if callobj.OnPlayPlaying != nil {
					callobj.OnPlayPlaying(res)
				}
			case PlayError:
				Log.Debug("Play error. ctrlID: %s\n", ctrlID)

				res.Lock()

				res.Completed = true
				res.State = playstate

				res.Unlock()

				out = true

				if callobj.OnPlayError != nil {
					callobj.OnPlayError(res)
				}
			case PlayPaused:
				res.Lock()

				res.State = playstate

				res.Unlock()

				Log.Debug("Play paused. ctrlID: %s\n", ctrlID)

				if callobj.OnPlayPaused != nil {
					callobj.OnPlayPaused(res)
				}
			default:
				Log.Debug("Unknown state. ctrlID: %s\n", ctrlID)
			}

			if prevstate != playstate && callobj.OnPlayStateChange != nil {
				callobj.OnPlayStateChange(res)
			}
		case <-callobj.call.Hangup:
			out = true
		}

		if out {
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

	go func() {
		go func() {
			// wait to get control ID (buffered channel)
			ctrlID := <-callobj.call.CallPlayControlIDs
			callobj.callbacksRunPlay(callobj.Calling.Ctx, ctrlID, res)
		}()

		newCtrlID, _ := GenUUIDv4()
		res.Lock()
		res.ControlID = newCtrlID
		res.Unlock()

		err := callobj.Calling.Relay.RelayPlaySilence(callobj.Calling.Ctx, callobj.call, newCtrlID, duration)

		if err != nil {
			res.Lock()

			res.err = err

			res.Completed = true

			res.Unlock()
		}
	}()

	return res, nil
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

	go func() {
		go func() {
			// wait to get control ID (buffered channel)
			ctrlID := <-callobj.call.CallPlayControlIDs

			callobj.callbacksRunPlay(callobj.Calling.Ctx, ctrlID, res)
		}()

		newCtrlID, _ := GenUUIDv4()
		res.Lock()
		res.ControlID = newCtrlID
		res.Unlock()

		err := callobj.Calling.Relay.RelayPlayTTS(callobj.Calling.Ctx, callobj.call, newCtrlID, text, language, gender)

		if err != nil {
			res.Lock()

			res.err = err

			res.Completed = true

			res.Unlock()

			return
		}
	}()

	return res, nil
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

	go func() {
		go func() {
			// wait to get control ID (buffered channel)
			ctrlID := <-callobj.call.CallPlayControlIDs

			callobj.callbacksRunPlay(callobj.Calling.Ctx, ctrlID, res)
		}()

		newCtrlID, _ := GenUUIDv4()
		res.Lock()
		res.ControlID = newCtrlID
		res.Unlock()

		err := callobj.Calling.Relay.RelayPlayAudio(callobj.Calling.Ctx, callobj.call, newCtrlID, url)

		if err != nil {
			res.Lock()

			res.err = err

			res.Completed = true

			res.Unlock()
		}
	}()

	return res, nil
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

	return playaction.CallObj.Calling.Relay.RelayPlayStop(playaction.CallObj.Calling.Ctx, call, &c)
}

// Stop TODO DESCRIPTION
func (playaction *PlayAction) Stop() {
	playaction.err = playaction.playAsyncStop()
}

// GetCompleted TODO DESCRIPTION
func (playaction *PlayAction) GetCompleted() bool {
	playaction.RLock()

	ret := playaction.Completed

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

	err = playaction.CallObj.Calling.Relay.RelayPlayVolume(playaction.CallObj.Calling.Ctx, call, &c, vol)

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

	err = playaction.CallObj.Calling.Relay.RelayPlayPause(playaction.CallObj.Calling.Ctx, call, &c)

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

	err = playaction.CallObj.Calling.Relay.RelayPlayResume(playaction.CallObj.Calling.Ctx, call, &c)

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
