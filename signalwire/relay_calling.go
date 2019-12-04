package signalwire

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// RelayPhoneDial make outbound phone call
func (relay *RelaySession) RelayPhoneDial(ctx context.Context, call *CallSession, fromNumber string, toNumber string, timeout uint, payload **json.RawMessage) error {
	var err error

	if relay == nil {
		return errors.New("empty relay object")
	}

	if relay.Blade == nil {
		return errors.New("blade server object not defined")
	}

	if call == nil {
		return errors.New("empty call object")
	}

	call.TagID, err = GenUUIDv4()
	if err != nil {
		return err
	}

	if timeout == 0 {
		call.SetTimeout(DefaultRingTimeout)
	} else {
		call.SetTimeout(timeout)
	}

	call.CallInit(ctx)
	call.SetType(CallTypePhone)

	v := ParamsBladeExecuteStruct{
		Protocol: relay.Blade.Protocol,
		Method:   "calling.begin",
		Params: ParamsCallingBeginStruct{
			Device: DeviceStruct{
				Type: "phone",
				Params: DevicePhoneParams{
					ToNumber:   toNumber,
					FromNumber: fromNumber,
					Timeout:    call.Timeout,
				},
			},
			Tag: call.TagID,
		},
	}

	savePayload(payload, v)

	var ReplyBladeExecuteDecode ReplyBladeExecute

	reply, err := relay.Blade.I.BladeExecute(ctx, &v, &ReplyBladeExecuteDecode)
	if err != nil {
		return err
	}

	r, ok := reply.(*ReplyBladeExecute)
	if !ok {
		return errors.New("type assertion failed")
	}

	/* use tag as Call-ID*/
	Log.Debug("call [%p] tag_id [%s]\n", call, call.TagID)
	Log.Debug("reply ReplyBladeExecuteDecode: %v\n", r)

	if r.Result.Code != okCode {
		return errors.New(r.Result.Message)
	}

	return relay.Blade.EventCalling.Cache.SetCallCache(call.TagID, call)
}

// RelayPhoneConnect TODO DESCRIPTION
func (relay *RelaySession) RelayPhoneConnect(ctx context.Context, call *CallSession, fromNumber string, toNumber string, payload **json.RawMessage) error {
	if relay == nil {
		return errors.New("empty relay object")
	}

	if relay.Blade == nil {
		return errors.New("blade server object not defined")
	}

	if call == nil {
		return errors.New("empty call object")
	}

	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	deviceParams := DevicePhoneParams{
		ToNumber:   toNumber,
		FromNumber: fromNumber,
		Timeout:    call.Timeout,
	}

	devices := [][]DeviceStruct{{
		DeviceStruct{
			Type:   "phone",
			Params: deviceParams,
		},
	}}

	v := ParamsBladeExecuteStruct{
		Protocol: relay.Blade.Protocol,
		Method:   "calling.connect",
		Params: ParamsCallConnectStruct{
			Devices: &devices,
			NodeID:  call.NodeID,
			CallID:  call.CallID,
		},
	}

	savePayload(payload, v)

	var ReplyBladeExecuteDecode ReplyBladeExecute

	reply, err := relay.Blade.I.BladeExecute(ctx, &v, &ReplyBladeExecuteDecode)
	if err != nil {
		return err
	}

	r, ok := reply.(*ReplyBladeExecute)
	if !ok {
		return errors.New("type assertion failed")
	}

	Log.Debug("reply ReplyBladeExecuteDecode: %v\n", r)

	if r.Result.Code != okCode {
		return errors.New(r.Result.Message)
	}

	return nil
}

// RelayConnect TODO DESCRIPTION
func (relay *RelaySession) RelayConnect(ctx context.Context, call *CallSession, ringback *[]RingbackStruct, devices *[][]DeviceStruct, payload **json.RawMessage) error {
	if relay == nil {
		return errors.New("empty relay object")
	}

	if relay.Blade == nil {
		return errors.New("blade server object not defined")
	}

	if call == nil {
		return errors.New("empty call object")
	}

	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	v := ParamsBladeExecuteStruct{
		Protocol: relay.Blade.Protocol,
		Method:   "calling.connect",
		Params: ParamsCallConnectStruct{
			Ringback: ringback,
			Devices:  devices,
			NodeID:   call.NodeID,
			CallID:   call.CallID,
		},
	}

	savePayload(payload, v)

	var ReplyBladeExecuteDecode ReplyBladeExecute

	reply, err := relay.Blade.I.BladeExecute(ctx, &v, &ReplyBladeExecuteDecode)
	if err != nil {
		return err
	}

	r, ok := reply.(*ReplyBladeExecute)
	if !ok {
		return errors.New("type assertion failed")
	}

	Log.Debug("reply ReplyBladeExecuteDecode: %v\n", r)

	if r.Result.Code != okCode {
		return errors.New(r.Result.Message)
	}

	return nil
}

// RelayCallAnswer TODO DESCRIPTION
func (relay *RelaySession) RelayCallAnswer(ctx context.Context, call *CallSession, payload **json.RawMessage) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	v := ParamsBladeExecuteStruct{
		Protocol: relay.Blade.Protocol,
		Method:   "calling.answer",
		Params: ParamsCallAnswer{
			NodeID: call.NodeID,
			CallID: call.CallID,
		},
	}

	savePayload(payload, v)

	var ReplyBladeExecuteDecode ReplyBladeExecute

	reply, err := relay.Blade.BladeExecute(ctx, &v, &ReplyBladeExecuteDecode)
	if err != nil {
		return err
	}

	r, ok := reply.(*ReplyBladeExecute)
	if !ok {
		return errors.New("type assertion failed")
	}

	Log.Debug("reply ReplyBladeExecuteDecode: %v\n", r)

	if r.Result.Code != okCode {
		return errors.New(r.Result.Message)
	}

	return nil
}

// RelayCallEnd TODO DESCRIPTION
func (relay *RelaySession) RelayCallEnd(ctx context.Context, call *CallSession, payload **json.RawMessage) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	v := ParamsBladeExecuteStruct{
		Protocol: relay.Blade.Protocol,
		Method:   "calling.end",
		Params: ParamsCallEndStruct{
			Reason: "hangup",
			NodeID: call.NodeID,
			CallID: call.CallID,
		},
	}

	savePayload(payload, v)

	var ReplyBladeExecuteDecode ReplyBladeExecute

	reply, err := relay.Blade.BladeExecute(ctx, &v, &ReplyBladeExecuteDecode)
	if err != nil {
		return err
	}

	r, ok := reply.(*ReplyBladeExecute)
	if !ok {
		return errors.New("type assertion failed")
	}

	Log.Debug("reply ReplyBladeExecuteDecode: %v\n", r)

	if r.Result.Code != okCode {
		return errors.New(r.Result.Message)
	}

	return nil
}

// RelayStop TODO DESCRIPTION
func (relay *RelaySession) RelayStop(ctx context.Context) error {
	// TODO: hangup all calls, cleanup
	return relay.Blade.BladeDisconnect(ctx)
}

// RelayOnInboundAnswer TODO DESCRIPTION
func (relay *RelaySession) RelayOnInboundAnswer(ctx context.Context) (*CallSession, error) {
	relay.Blade.BladeSetupInbound(ctx)

	call, err := relay.Blade.BladeWaitInboundCall(ctx)
	if err != nil {
		return nil, err
	}

	if call != nil {
		err := relay.RelayCallAnswer(ctx, call, nil)
		if err != nil {
			return nil, err
		}
	}

	return call, nil
}

// RelayPlayAudio TODO DESCRIPTION
func (relay *RelaySession) RelayPlayAudio(ctx context.Context, call *CallSession, ctrlID string, url string, payload **json.RawMessage) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	playAudioParams := PlayAudioParams{
		URL: url,
	}

	play := []PlayStruct{{
		Type:   "audio",
		Params: playAudioParams,
	}}

	return relay.RelayPlay(ctx, call, ctrlID, play, payload)
}

// RelayPlayTTS TODO DESCRIPTION
func (relay *RelaySession) RelayPlayTTS(ctx context.Context, call *CallSession, ctrlID string, tts *TTSParamsInternal, payload **json.RawMessage) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	playTTSParams := PlayTTSParams{
		Text:     tts.text,
		Language: tts.language,
		Gender:   tts.gender,
	}

	play := []PlayStruct{{
		Type:   "tts",
		Params: playTTSParams,
	}}

	return relay.RelayPlay(ctx, call, ctrlID, play, payload)
}

// RelayPlayRingtone TODO DESCRIPTION
func (relay *RelaySession) RelayPlayRingtone(ctx context.Context, call *CallSession, ctrlID string, name string, duration float64, payload **json.RawMessage) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	playRingtoneParams := PlayRingtoneParams{
		Name:     name,
		Duration: duration,
	}

	play := []PlayStruct{{
		Type:   "ringtone",
		Params: playRingtoneParams,
	}}

	return relay.RelayPlay(ctx, call, ctrlID, play, payload)
}

// RelayPlaySilence TODO DESCRIPTION
func (relay *RelaySession) RelayPlaySilence(ctx context.Context, call *CallSession, ctrlID string, duration float64, payload **json.RawMessage) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	playSilenceParams := PlaySilenceParams{
		Duration: duration,
	}

	play := []PlayStruct{{
		Type:   "silence",
		Params: playSilenceParams,
	}}

	return relay.RelayPlay(ctx, call, ctrlID, play, payload)
}

// RelayPlay TODO DESCRIPTION
func (relay *RelaySession) RelayPlay(ctx context.Context, call *CallSession, controlID string, play []PlayStruct, payload **json.RawMessage) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	v := ParamsBladeExecuteStruct{
		Protocol: relay.Blade.Protocol,
		Method:   "calling.play",
		Params: ParamsCallPlay{
			NodeID:    call.NodeID,
			CallID:    call.CallID,
			ControlID: controlID,
			Play:      play,
		},
	}

	/* prepare payload per Action in case user want to inspect it. We'll not send this, jsonrpc2 lib will do it's own marshaling on v */
	savePayload(payload, v)

	call.Lock()

	call.CallPlayChans[controlID] = make(chan PlayState, EventQueue)
	call.CallPlayEventChans[controlID] = make(chan ParamsEventCallingCallPlay, EventQueue)
	call.CallPlayReadyChans[controlID] = make(chan struct{})
	call.CallPlayRawEventChans[controlID] = make(chan *json.RawMessage, EventQueue)

	call.Unlock()

	select {
	case call.CallPlayControlIDs <- controlID:
		// send the ctrlID to go routine that fires Consumer callbacks
		Log.Debug("sent controlID to go routine\n")
	default:
		Log.Debug("controlID was not sent to go routine\n")
	}

	var ReplyBladeExecuteDecode ReplyBladeExecute

	reply, err := relay.Blade.BladeExecute(ctx, &v, &ReplyBladeExecuteDecode /*, payload*/)
	if err != nil {
		return err
	}

	r, ok := reply.(*ReplyBladeExecute)
	if !ok {
		return errors.New("type assertion failed")
	}

	Log.Debug("reply ReplyBladeExecuteDecode: %v\n", r)

	if r.Result.Code != okCode {
		return errors.New(r.Result.Message)
	}

	return nil
}

// RelayPlayVolume TODO DESCRIPTION
func (relay *RelaySession) RelayPlayVolume(ctx context.Context, call *CallSession, ctrlID *string, vol float64, payload **json.RawMessage) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	v := ParamsBladeExecuteStruct{
		Protocol: relay.Blade.Protocol,
		Method:   "calling.play.volume",
		Params: ParamsCallPlayVolume{
			NodeID:    call.NodeID,
			CallID:    call.CallID,
			ControlID: *ctrlID,
			Volume:    vol,
		},
	}

	savePayload(payload, v)

	var ReplyBladeExecuteDecode ReplyBladeExecute

	reply, err := relay.Blade.BladeExecute(ctx, &v, &ReplyBladeExecuteDecode)
	if err != nil {
		return err
	}

	r, ok := reply.(*ReplyBladeExecute)
	if !ok {
		return errors.New("type assertion failed")
	}

	Log.Debug("reply ReplyBladeExecuteDecode: %v\n", r)

	if r.Result.Code != okCode {
		return errors.New(r.Result.Message)
	}

	return nil
}

// RelayPlayResume TODO DESCRIPTION
func (relay *RelaySession) RelayPlayResume(ctx context.Context, call *CallSession, ctrlID *string, payload **json.RawMessage) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	v := ParamsBladeExecuteStruct{
		Protocol: relay.Blade.Protocol,
		Method:   "calling.play.resume",
		Params: ParamsCallPlayResume{
			NodeID:    call.NodeID,
			CallID:    call.CallID,
			ControlID: *ctrlID,
		},
	}

	savePayload(payload, v)

	var ReplyBladeExecuteDecode ReplyBladeExecute

	reply, err := relay.Blade.BladeExecute(ctx, &v, &ReplyBladeExecuteDecode)
	if err != nil {
		return err
	}

	r, ok := reply.(*ReplyBladeExecute)
	if !ok {
		return errors.New("type assertion failed")
	}

	Log.Debug("reply ReplyBladeExecuteDecode: %v\n", r)

	if r.Result.Code != okCode {
		return errors.New(r.Result.Message)
	}

	return nil
}

// RelayPlayPause TODO DESCRIPTION
func (relay *RelaySession) RelayPlayPause(ctx context.Context, call *CallSession, ctrlID *string, payload **json.RawMessage) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	v := ParamsBladeExecuteStruct{
		Protocol: relay.Blade.Protocol,
		Method:   "calling.play.pause",
		Params: ParamsCallPlayPause{
			NodeID:    call.NodeID,
			CallID:    call.CallID,
			ControlID: *ctrlID,
		},
	}

	savePayload(payload, v)

	var ReplyBladeExecuteDecode ReplyBladeExecute

	reply, err := relay.Blade.BladeExecute(ctx, &v, &ReplyBladeExecuteDecode)
	if err != nil {
		return err
	}

	r, ok := reply.(*ReplyBladeExecute)
	if !ok {
		return errors.New("type assertion failed")
	}

	Log.Debug("reply ReplyBladeExecuteDecode: %v\n", r)

	if r.Result.Code != okCode {
		return errors.New(r.Result.Message)
	}

	return nil
}

// RelayPlayStop TODO DESCRIPTION
func (relay *RelaySession) RelayPlayStop(ctx context.Context, call *CallSession, ctrlID *string, payload **json.RawMessage) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	v := ParamsBladeExecuteStruct{
		Protocol: relay.Blade.Protocol,
		Method:   "calling.play.stop",
		Params: ParamsCallPlayStop{
			NodeID:    call.NodeID,
			CallID:    call.CallID,
			ControlID: *ctrlID,
		},
	}

	savePayload(payload, v)

	var ReplyBladeExecuteDecode ReplyBladeExecute

	reply, err := relay.Blade.BladeExecute(ctx, &v, &ReplyBladeExecuteDecode)
	if err != nil {
		return err
	}

	r, ok := reply.(*ReplyBladeExecute)
	if !ok {
		return errors.New("type assertion failed")
	}

	Log.Debug("reply ReplyBladeExecuteDecode: %v\n", r)

	if r.Result.Code != okCode {
		return errors.New(r.Result.Message)
	}

	return nil
}

// RelayRecordAudio TODO DESCRIPTION
func (relay *RelaySession) RelayRecordAudio(ctx context.Context, call *CallSession, controlID string, rec *RecordParams, payload **json.RawMessage) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	recordAudioParams := RecordParams{
		Beep:              rec.Beep,
		Format:            rec.Format,
		Direction:         rec.Direction,
		Stereo:            rec.Stereo,
		InitialTimeout:    rec.InitialTimeout,
		EndSilenceTimeout: rec.EndSilenceTimeout,
		Terminators:       rec.Terminators,
	}

	record := RecordStruct{
		Audio: recordAudioParams,
	}

	v := ParamsBladeExecuteStruct{
		Protocol: relay.Blade.Protocol,
		Method:   "calling.record",
		Params: ParamsCallRecord{
			NodeID:    call.NodeID,
			CallID:    call.CallID,
			ControlID: controlID,
			Record:    record,
		},
	}

	savePayload(payload, v)

	call.Lock()

	call.CallRecordChans[controlID] = make(chan RecordState, EventQueue)
	call.CallRecordEventChans[controlID] = make(chan ParamsEventCallingCallRecord, EventQueue)
	call.CallRecordReadyChans[controlID] = make(chan struct{})
	call.CallRecordRawEventChans[controlID] = make(chan *json.RawMessage, EventQueue)

	call.Unlock()

	select {
	case call.CallRecordControlIDs <- controlID:
		// send the ctrlID to go routine that fires Consumer callbacks
		Log.Debug("sent controlID to go routine\n")
	default:
		Log.Debug("controlID was not sent to go routine\n")
	}

	var ReplyBladeExecuteDecode ReplyBladeExecute

	reply, err := relay.Blade.BladeExecute(ctx, &v, &ReplyBladeExecuteDecode)
	if err != nil {
		return err
	}

	r, ok := reply.(*ReplyBladeExecute)
	if !ok {
		return errors.New("type assertion failed")
	}

	Log.Debug("reply ReplyBladeExecuteDecode: %v\n", r)

	if r.Result.Code != okCode {
		return errors.New(r.Result.Message)
	}

	return nil
}

// RelayRecordAudioStop TODO DESCRIPTION
func (relay *RelaySession) RelayRecordAudioStop(ctx context.Context, call *CallSession, ctrlID *string, payload **json.RawMessage) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	v := ParamsBladeExecuteStruct{
		Protocol: relay.Blade.Protocol,
		Method:   "calling.record.stop",
		Params: ParamsCallRecordStop{
			NodeID:    call.NodeID,
			CallID:    call.CallID,
			ControlID: *ctrlID,
		},
	}

	savePayload(payload, v)

	var ReplyBladeExecuteDecode ReplyBladeExecute

	reply, err := relay.Blade.BladeExecute(ctx, &v, &ReplyBladeExecuteDecode)
	if err != nil {
		return err
	}

	r, ok := reply.(*ReplyBladeExecute)
	if !ok {
		return errors.New("type assertion failed")
	}

	Log.Debug("reply ReplyBladeExecuteDecode: %v\n", r)

	if r.Result.Code != okCode {
		return errors.New(r.Result.Message)
	}

	return nil
}

// RelayDetectDigit TODO DESCRIPTION
func (relay *RelaySession) RelayDetectDigit(ctx context.Context, call *CallSession, controlID string, digits string, timeout float64, payload **json.RawMessage) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	detectDigitParams := DetectDigitParamsInternal{
		Digits: digits,
	}
	detect := DetectStruct{
		Type:   "digit",
		Params: detectDigitParams,
	}

	call.Lock()

	call.CallDetectDigitChans[controlID] = make(chan DetectDigitEvent, EventQueue)

	call.Unlock()

	select {
	case call.CallDetectDigitControlID <- controlID:
		// send the ctrlID to go routine that fires Consumer callbacks
		Log.Debug("sent controlID to go routine\n")
	default:
		Log.Debug("controlID was not sent to go routine\n")
	}

	return relay.RelayDetect(ctx, call, controlID, detect, timeout, payload)
}

// RelayDetectFax TODO DESCRIPTION
func (relay *RelaySession) RelayDetectFax(ctx context.Context, call *CallSession, controlID string, faxtone string, timeout float64, payload **json.RawMessage) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	detectFaxParams := DetectFaxParamsInternal{
		Tone: faxtone,
	}
	detect := DetectStruct{
		Type:   "fax",
		Params: detectFaxParams,
	}

	call.Lock()

	call.CallDetectFaxChans[controlID] = make(chan DetectFaxEvent, EventQueue)

	call.Unlock()

	select {
	case call.CallDetectFaxControlID <- controlID:
		// send the ctrlID to go routine that fires Consumer callbacks
		Log.Debug("sent controlID to go routine\n")
	default:
		Log.Debug("controlID was not sent to go routine\n")
	}

	return relay.RelayDetect(ctx, call, controlID, detect, timeout, payload)
}

// RelayDetectMachine TODO DESCRIPTION
func (relay *RelaySession) RelayDetectMachine(ctx context.Context, call *CallSession, controlID string, det *DetectMachineParamsInternal, timeout float64, payload **json.RawMessage) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	detectMachineParams := DetectMachineParamsInternal{
		InitialTimeout:        det.InitialTimeout,
		EndSilenceTimeout:     det.EndSilenceTimeout,
		MachineVoiceThreshold: det.MachineVoiceThreshold,
		MachineWordsThreshold: det.MachineWordsThreshold,
	}
	detect := DetectStruct{
		Type:   "machine",
		Params: detectMachineParams,
	}

	call.Lock()

	call.CallDetectMachineChans[controlID] = make(chan DetectMachineEvent, EventQueue)

	call.Unlock()

	select {
	case call.CallDetectMachineControlID <- controlID:
		// send the ctrlID to go routine that fires Consumer callbacks
		Log.Debug("sent controlID to go routine\n")
	default:
		Log.Debug("controlID was not sent to go routine\n")
	}

	return relay.RelayDetect(ctx, call, controlID, detect, timeout, payload)
}

// RelayDetect TODO DESCRIPTION
func (relay *RelaySession) RelayDetect(ctx context.Context, call *CallSession, controlID string, detect DetectStruct, timeout float64, payload **json.RawMessage) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	call.Lock()

	call.CallDetectReadyChans[controlID] = make(chan struct{})
	call.CallDetectRawEventChans[controlID] = make(chan *json.RawMessage, EventQueue)

	call.Unlock()

	var detTimeout float64

	if timeout != 0 {
		detTimeout = timeout
	} else {
		detTimeout = DefaultActionTimeout
	}

	v := ParamsBladeExecuteStruct{
		Protocol: relay.Blade.Protocol,
		Method:   "calling.detect",
		Params: ParamsCallDetect{
			NodeID:    call.NodeID,
			CallID:    call.CallID,
			ControlID: controlID,
			Detect:    detect,
			Timeout:   detTimeout,
		},
	}

	savePayload(payload, v)

	var ReplyBladeExecuteDecode ReplyBladeExecute

	reply, err := relay.Blade.BladeExecute(ctx, &v, &ReplyBladeExecuteDecode)
	if err != nil {
		return err
	}

	r, ok := reply.(*ReplyBladeExecute)
	if !ok {
		return errors.New("type assertion failed")
	}

	Log.Debug("reply ReplyBladeExecuteDecode: %v\n", r)

	if r.Result.Code != okCode {
		return errors.New(r.Result.Message)
	}

	return nil
}

// RelayDetectStop TODO DESCRIPTION
func (relay *RelaySession) RelayDetectStop(ctx context.Context, call *CallSession, ctrlID *string, payload **json.RawMessage) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	v := ParamsBladeExecuteStruct{
		Protocol: relay.Blade.Protocol,
		Method:   "calling.detect.stop",
		Params: ParamsCallDetectStop{
			NodeID:    call.NodeID,
			CallID:    call.CallID,
			ControlID: *ctrlID,
		},
	}

	savePayload(payload, v)

	var ReplyBladeExecuteDecode ReplyBladeExecute

	reply, err := relay.Blade.BladeExecute(ctx, &v, &ReplyBladeExecuteDecode)
	if err != nil {
		return err
	}

	r, ok := reply.(*ReplyBladeExecute)
	if !ok {
		return errors.New("type assertion failed")
	}

	Log.Debug("reply ReplyBladeExecuteDecode: %v\n", r)

	if r.Result.Code != okCode {
		return errors.New(r.Result.Message)
	}

	return nil
}

// RelaySendFax TODO DESCRIPTION
func (relay *RelaySession) RelaySendFax(ctx context.Context, call *CallSession, ctrlID *string, fax *FaxParamsInternal, payload **json.RawMessage) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	v := ParamsBladeExecuteStruct{
		Protocol: relay.Blade.Protocol,
		Method:   "calling.send_fax",
		Params: ParamsSendFax{
			NodeID:     call.NodeID,
			CallID:     call.CallID,
			ControlID:  *ctrlID,
			Document:   fax.doc,
			Identity:   fax.id,
			HeaderInfo: fax.headerInfo,
		},
	}

	savePayload(payload, v)

	select {
	case call.CallFaxControlID <- *ctrlID:
		// send the ctrlID to go routine that fires Consumer callbacks
		Log.Debug("sent controlID to go routine\n")
	default:
		Log.Debug("controlID was not sent to go routine\n")
	}

	var ReplyBladeExecuteDecode ReplyBladeExecute

	reply, err := relay.Blade.BladeExecute(ctx, &v, &ReplyBladeExecuteDecode)
	if err != nil {
		return err
	}

	r, ok := reply.(*ReplyBladeExecute)
	if !ok {
		return errors.New("type assertion failed")
	}

	Log.Debug("reply ReplyBladeExecuteDecode: %v\n", r)

	if r.Result.Code != okCode {
		return errors.New(r.Result.Message)
	}

	return nil
}

// RelayReceiveFax TODO DESCRIPTION
func (relay *RelaySession) RelayReceiveFax(ctx context.Context, call *CallSession, ctrlID *string, payload **json.RawMessage) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	v := ParamsBladeExecuteStruct{
		Protocol: relay.Blade.Protocol,
		Method:   "calling.receive_fax",
		Params: ParamsSendFax{
			NodeID:    call.NodeID,
			CallID:    call.CallID,
			ControlID: *ctrlID,
		},
	}

	savePayload(payload, v)

	select {
	case call.CallFaxControlID <- *ctrlID:
		// send the ctrlID to go routine that fires Consumer callbacks
		Log.Debug("sent controlID to go routine\n")
	default:
		Log.Debug("controlID was not sent to go routine\n")
	}

	var ReplyBladeExecuteDecode ReplyBladeExecute

	reply, err := relay.Blade.BladeExecute(ctx, &v, &ReplyBladeExecuteDecode)
	if err != nil {
		return err
	}

	r, ok := reply.(*ReplyBladeExecute)
	if !ok {
		return errors.New("type assertion failed")
	}

	Log.Debug("reply ReplyBladeExecuteDecode: %v\n", r)

	if r.Result.Code != okCode {
		return errors.New(r.Result.Message)
	}

	return nil
}

// RelaySendFaxStop TODO DESCRIPTION
func (relay *RelaySession) RelaySendFaxStop(ctx context.Context, call *CallSession, ctrlID *string, payload **json.RawMessage) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	v := ParamsBladeExecuteStruct{
		Protocol: relay.Blade.Protocol,
		Method:   "calling.send_fax.stop",
		Params: ParamsFaxStop{
			NodeID:    call.NodeID,
			CallID:    call.CallID,
			ControlID: *ctrlID,
		},
	}

	savePayload(payload, v)

	var ReplyBladeExecuteDecode ReplyBladeExecute

	reply, err := relay.Blade.BladeExecute(ctx, &v, &ReplyBladeExecuteDecode)
	if err != nil {
		return err
	}

	r, ok := reply.(*ReplyBladeExecute)
	if !ok {
		return errors.New("type assertion failed")
	}

	Log.Debug("reply ReplyBladeExecuteDecode: %v\n", r)

	if r.Result.Code != okCode {
		return errors.New(r.Result.Message)
	}

	return nil
}

// RelayReceiveFaxStop TODO DESCRIPTION
func (relay *RelaySession) RelayReceiveFaxStop(ctx context.Context, call *CallSession, ctrlID *string, payload **json.RawMessage) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	v := ParamsBladeExecuteStruct{
		Protocol: relay.Blade.Protocol,
		Method:   "calling.receive_fax.stop",
		Params: ParamsFaxStop{
			NodeID:    call.NodeID,
			CallID:    call.CallID,
			ControlID: *ctrlID,
		},
	}

	savePayload(payload, v)

	var ReplyBladeExecuteDecode ReplyBladeExecute

	reply, err := relay.Blade.BladeExecute(ctx, &v, &ReplyBladeExecuteDecode)
	if err != nil {
		return err
	}

	r, ok := reply.(*ReplyBladeExecute)
	if !ok {
		return errors.New("type assertion failed")
	}

	Log.Debug("reply ReplyBladeExecuteDecode: %v\n", r)

	if r.Result.Code != okCode {
		return errors.New(r.Result.Message)
	}

	return nil
}

// RelayTapAudio TODO DESCRIPTION
func (relay *RelaySession) RelayTapAudio(ctx context.Context, call *CallSession, ctrlID, direction string, device *TapDevice, payload **json.RawMessage) (TapDevice, error) {
	tapAudioParams := TapAudioParams{
		Direction: direction,
	}

	tap := TapStruct{
		Type:   "audio",
		Params: tapAudioParams,
	}

	// return source Tap device
	return relay.RelayTap(ctx, call, ctrlID, tap, device, payload)
}

// RelayTap TODO DESCRIPTION
func (relay *RelaySession) RelayTap(ctx context.Context, call *CallSession, controlID string, tap TapStruct, device *TapDevice, payload **json.RawMessage) (TapDevice, error) {
	var srcDevice TapDevice

	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return srcDevice, fmt.Errorf("no CallID for call [%p]", call)
	}

	v := ParamsBladeExecuteStruct{
		Protocol: relay.Blade.Protocol,
		Method:   "calling.tap",
		Params: ParamsCallTap{
			NodeID:    call.NodeID,
			CallID:    call.CallID,
			ControlID: controlID,
			Tap:       tap,
			Device:    *device,
		},
	}

	savePayload(payload, v)

	call.Lock()

	call.CallTapChans[controlID] = make(chan TapState, EventQueue)
	call.CallTapEventChans[controlID] = make(chan ParamsEventCallingCallTap, EventQueue)
	call.CallTapReadyChans[controlID] = make(chan struct{})
	call.CallTapRawEventChans[controlID] = make(chan *json.RawMessage, EventQueue)

	call.Unlock()

	select {
	case call.CallTapControlIDs <- controlID:
		// send the ctrlID to go routine that fires Consumer callbacks
		Log.Debug("sent controlID to go routine\n")
	default:
		Log.Debug("controlID was not sent to go routine\n")
	}

	var ReplyBladeExecuteDecode ReplyBladeExecuteTap

	reply, err := relay.Blade.BladeExecute(ctx, &v, &ReplyBladeExecuteDecode)
	if err != nil {
		return srcDevice, err
	}

	r, ok := reply.(*ReplyBladeExecuteTap)
	if !ok {
		return srcDevice, errors.New("type assertion failed")
	}

	Log.Debug("reply ReplyBladeExecuteDecode: %v\n", r)

	if r.Result.Code != okCode {
		return srcDevice, errors.New(r.Result.Message)
	}

	return r.Result.SourceDevice, nil
}

// RelayTapStop TODO DESCRIPTION
func (relay *RelaySession) RelayTapStop(ctx context.Context, call *CallSession, ctrlID *string, payload **json.RawMessage) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	v := ParamsBladeExecuteStruct{
		Protocol: relay.Blade.Protocol,
		Method:   "calling.tap.stop",
		Params: ParamsCallPlayStop{
			NodeID:    call.NodeID,
			CallID:    call.CallID,
			ControlID: *ctrlID,
		},
	}

	savePayload(payload, v)

	var ReplyBladeExecuteDecode ReplyBladeExecute

	reply, err := relay.Blade.BladeExecute(ctx, &v, &ReplyBladeExecuteDecode)
	if err != nil {
		return err
	}

	r, ok := reply.(*ReplyBladeExecute)
	if !ok {
		return errors.New("type assertion failed")
	}

	Log.Debug("reply ReplyBladeExecuteDecode: %v\n", r)

	if r.Result.Code != okCode {
		return errors.New(r.Result.Message)
	}

	return nil
}

// RelaySendDigits TODO DESCRIPTION
func (relay *RelaySession) RelaySendDigits(ctx context.Context, call *CallSession, controlID, digits string, payload **json.RawMessage) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	v := ParamsBladeExecuteStruct{
		Protocol: relay.Blade.Protocol,
		Method:   "calling.send_digits",
		Params: ParamsCallSendDigits{
			NodeID:    call.NodeID,
			CallID:    call.CallID,
			ControlID: controlID,
			Digits:    digits,
		},
	}

	savePayload(payload, v)

	call.Lock()

	call.CallSendDigitsChans[controlID] = make(chan SendDigitsState, EventQueue)
	call.CallSendDigitsReadyChans[controlID] = make(chan struct{})
	call.CallSendDigitsRawEventChans[controlID] = make(chan *json.RawMessage, EventQueue)

	call.Unlock()

	select {
	case call.CallSendDigitsControlIDs <- controlID:
		// send the ctrlID to go routine that fires Consumer callbacks
		Log.Debug("sent controlID to go routine\n")
	default:
		Log.Debug("controlID was not sent to go routine\n")
	}

	var ReplyBladeExecuteDecode ReplyBladeExecute

	reply, err := relay.Blade.BladeExecute(ctx, &v, &ReplyBladeExecuteDecode)
	if err != nil {
		return err
	}

	r, ok := reply.(*ReplyBladeExecute)
	if !ok {
		return errors.New("type assertion failed")
	}

	Log.Debug("reply ReplyBladeExecuteDecode: %v\n", r)

	if r.Result.Code != okCode {
		return errors.New(r.Result.Message)
	}

	return nil
}

// RelayPlayAndCollect TODO DESCRIPTION
func (relay *RelaySession) RelayPlayAndCollect(ctx context.Context, call *CallSession, controlID string, playlist *[]PlayStruct, collect *CollectStruct, payload **json.RawMessage) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	v := ParamsBladeExecuteStruct{
		Protocol: relay.Blade.Protocol,
		Method:   "calling.play_and_collect",
		Params: ParamsCallPlayAndCollect{
			NodeID:    call.NodeID,
			CallID:    call.CallID,
			ControlID: controlID,
			Play:      *playlist,
			Collect:   *collect,
		},
	}

	savePayload(payload, v)

	call.Lock()

	call.CallPlayAndCollectChans[controlID] = make(chan CollectResultType, EventQueue)
	call.CallPlayAndCollectEventChans[controlID] = make(chan ParamsEventCallingCallPlayAndCollect, EventQueue)
	call.CallPlayAndCollectReadyChans[controlID] = make(chan struct{})
	call.CallPlayAndCollectRawEventChans[controlID] = make(chan *json.RawMessage, EventQueue)

	call.CallPlayChans[controlID] = make(chan PlayState, EventQueue)
	call.CallPlayEventChans[controlID] = make(chan ParamsEventCallingCallPlay, EventQueue)
	call.CallPlayReadyChans[controlID] = make(chan struct{})
	call.CallPlayRawEventChans[controlID] = make(chan *json.RawMessage, EventQueue)

	call.Unlock()

	select {
	case call.CallPlayAndCollectControlID <- controlID:
		// send the ctrlID to go routine that fires Consumer callbacks
		Log.Debug("sent controlID to go routine\n")
	default:
		Log.Debug("controlID was not sent to go routine\n")
	}

	var ReplyBladeExecuteDecode ReplyBladeExecute

	reply, err := relay.Blade.BladeExecute(ctx, &v, &ReplyBladeExecuteDecode)
	if err != nil {
		return err
	}

	r, ok := reply.(*ReplyBladeExecute)
	if !ok {
		return errors.New("type assertion failed")
	}

	Log.Debug("reply ReplyBladeExecuteDecode: %v\n", r)

	if r.Result.Code != okCode {
		return errors.New(r.Result.Message)
	}

	return nil
}

// RelayPlayAndCollectVolume TODO DESCRIPTION
func (relay *RelaySession) RelayPlayAndCollectVolume(ctx context.Context, call *CallSession, ctrlID *string, vol float64, payload **json.RawMessage) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	v := ParamsBladeExecuteStruct{
		Protocol: relay.Blade.Protocol,
		Method:   "calling.play_and_collect.volume",
		Params: ParamsCallPlayVolume{
			NodeID:    call.NodeID,
			CallID:    call.CallID,
			ControlID: *ctrlID,
			Volume:    vol,
		},
	}

	savePayload(payload, v)

	var ReplyBladeExecuteDecode ReplyBladeExecute

	reply, err := relay.Blade.BladeExecute(ctx, &v, &ReplyBladeExecuteDecode)
	if err != nil {
		return err
	}

	r, ok := reply.(*ReplyBladeExecute)
	if !ok {
		return errors.New("type assertion failed")
	}

	Log.Debug("reply ReplyBladeExecuteDecode: %v\n", r)

	if r.Result.Code != "200" {
		return errors.New(r.Result.Message)
	}

	return nil
}

// RelayPlayAndCollectStop TODO DESCRIPTION
func (relay *RelaySession) RelayPlayAndCollectStop(ctx context.Context, call *CallSession, ctrlID *string, payload **json.RawMessage) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	v := ParamsBladeExecuteStruct{
		Protocol: relay.Blade.Protocol,
		Method:   "calling.play_and_collect.stop",
		Params: ParamsCallPlayStop{
			NodeID:    call.NodeID,
			CallID:    call.CallID,
			ControlID: *ctrlID,
		},
	}

	savePayload(payload, v)

	var ReplyBladeExecuteDecode ReplyBladeExecute

	reply, err := relay.Blade.BladeExecute(ctx, &v, &ReplyBladeExecuteDecode)
	if err != nil {
		return err
	}

	r, ok := reply.(*ReplyBladeExecute)
	if !ok {
		return errors.New("type assertion failed")
	}

	Log.Debug("reply ReplyBladeExecuteDecode: %v\n", r)

	if r.Result.Code != "200" {
		return errors.New(r.Result.Message)
	}

	return nil
}

type placeHolder struct {
	Params json.RawMessage
}

func savePayload(payload **json.RawMessage, v interface{}) {
	if payload != nil {
		placeholder := new(placeHolder)

		b, err := json.Marshal(v)
		if err != nil {
			Log.Error("payload: cannot marshal")
		}

		err = json.Unmarshal(b, placeholder)
		if err != nil {
			Log.Error("payload: cannot unmarshal to RawMessage")
		}

		*payload = &placeholder.Params
	}
}
