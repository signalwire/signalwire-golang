package signalwire

import (
	"context"
	"errors"
	"fmt"
)

// RelaySession TODO DESCRIPTION
type RelaySession struct {
	Blade *BladeSession
}

// IRelay TODO DESCRIPTION
type IRelay interface {
	RelayPhoneDial(ctx context.Context, call *CallSession, fromNumber string, toNumber string, timeout uint) error
	RelayPhoneConnect(ctx context.Context, call *CallSession, fromNumber string, toNumber string) error
	RelayCallEnd(ctx context.Context, call *CallSession) error
	RelayStop(ctx context.Context) error
	RelayOnInboundAnswer(ctx context.Context)
	RelayPlayAudio(ctx context.Context, call *CallSession, ctrlID string, url string) error
	RelayPlayAudioStop(ctx context.Context, call *CallSession, ctrlID string) error
	RelayRecordAudio(ctx context.Context, call *CallSession, ctrlID string, rec RecordParams) error
	RelayRecordAudioStop(ctx context.Context, call *CallSession, ctrlID string) error
}

// RelayPhoneDial make outbound phone call
func (relay *RelaySession) RelayPhoneDial(ctx context.Context, call *CallSession, fromNumber string, toNumber string, timeout uint) error {
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

	return relay.Blade.EventCalling.Cache.SetCallCache(call.TagID, call)
}

// RelayPhoneConnect TODO DESCRIPTION
func (relay *RelaySession) RelayPhoneConnect(ctx context.Context, call *CallSession, fromNumber string, toNumber string) error {
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
			Devices: devices,
			NodeID:  call.NodeID,
			CallID:  call.CallID,
		},
	}

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

	return nil
}

// RelayConnect TODO DESCRIPTION
func (relay *RelaySession) RelayConnect(ctx context.Context, call *CallSession, ringback *[]RingbackStruct, devices *[][]DeviceStruct) error {
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

	var v interface{}
	if ringback == nil {
		v = ParamsBladeExecuteStruct{
			Protocol: relay.Blade.Protocol,
			Method:   "calling.connect",
			Params: ParamsCallConnectStruct{
				Devices: *devices,
				NodeID:  call.NodeID,
				CallID:  call.CallID,
			},
		}
	} else {
		v = ParamsBladeExecuteStruct{
			Protocol: relay.Blade.Protocol,
			Method:   "calling.connect",
			Params: ParamsCallConnectStruct{
				Ringback: *ringback,
				Devices:  *devices,
				NodeID:   call.NodeID,
				CallID:   call.CallID,
			},
		}
	}

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

	return nil
}

// RelayCallAnswer TODO DESCRIPTION
func (relay *RelaySession) RelayCallAnswer(ctx context.Context, call *CallSession) error {
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

	return nil
}

// RelayCallEnd TODO DESCRIPTION
func (relay *RelaySession) RelayCallEnd(ctx context.Context, call *CallSession) error {
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
		err := relay.RelayCallAnswer(ctx, call)
		if err != nil {
			return nil, err
		}
	}

	return call, nil
}

// RelayPlayAudio TODO DESCRIPTION
func (relay *RelaySession) RelayPlayAudio(ctx context.Context, call *CallSession, ctrlID string, url string) error {
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

	return relay.RelayPlay(ctx, call, ctrlID, play)
}

// RelayPlayTTS TODO DESCRIPTION
func (relay *RelaySession) RelayPlayTTS(ctx context.Context, call *CallSession, ctrlID string, text, language, gender string) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	playTTSParams := PlayTTSParams{
		Text:     text,
		Language: language,
		Gender:   gender,
	}

	play := []PlayStruct{{
		Type:   "tts",
		Params: playTTSParams,
	}}

	return relay.RelayPlay(ctx, call, ctrlID, play)
}

// RelayPlayRingtone TODO DESCRIPTION
func (relay *RelaySession) RelayPlayRingtone(ctx context.Context, call *CallSession, ctrlID string, name string, duration float64) error {
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

	return relay.RelayPlay(ctx, call, ctrlID, play)
}

// RelayPlaySilence TODO DESCRIPTION
func (relay *RelaySession) RelayPlaySilence(ctx context.Context, call *CallSession, ctrlID string, duration float64) error {
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

	return relay.RelayPlay(ctx, call, ctrlID, play)
}

// RelayPlay TODO DESCRIPTION
func (relay *RelaySession) RelayPlay(ctx context.Context, call *CallSession, controlID string, play []PlayStruct) error {
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

	call.Lock()

	call.CallPlayChans[controlID] = make(chan PlayState, EventQueue)
	call.CallPlayEventChans[controlID] = make(chan ParamsEventCallingCallPlay, EventQueue)
	call.CallPlayReadyChans[controlID] = make(chan struct{})

	call.Unlock()

	select {
	case call.CallPlayControlIDs <- controlID:
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

	return nil
}

// RelayPlayVolume TODO DESCRIPTION
func (relay *RelaySession) RelayPlayVolume(ctx context.Context, call *CallSession, ctrlID *string, vol float64) error {
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

	return nil
}

// RelayPlayResume TODO DESCRIPTION
func (relay *RelaySession) RelayPlayResume(ctx context.Context, call *CallSession, ctrlID *string) error {
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

	return nil
}

// RelayPlayPause TODO DESCRIPTION
func (relay *RelaySession) RelayPlayPause(ctx context.Context, call *CallSession, ctrlID *string) error {
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

	return nil
}

// RelayPlayStop TODO DESCRIPTION
func (relay *RelaySession) RelayPlayStop(ctx context.Context, call *CallSession, ctrlID *string) error {
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

	return nil
}

// RelayRecordAudio TODO DESCRIPTION
func (relay *RelaySession) RelayRecordAudio(ctx context.Context, call *CallSession, controlID string, rec *RecordParams) error {
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

	call.Lock()

	call.CallRecordChans[controlID] = make(chan RecordState, EventQueue)
	call.CallRecordEventChans[controlID] = make(chan ParamsEventCallingCallRecord, EventQueue)
	call.CallRecordReadyChans[controlID] = make(chan struct{})

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

	call.AddAction(controlID, "init")

	return nil
}

// RelayRecordAudioStop TODO DESCRIPTION
func (relay *RelaySession) RelayRecordAudioStop(ctx context.Context, call *CallSession, ctrlID *string) error {
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

	return nil
}

// RelayDetectDigit TODO DESCRIPTION
func (relay *RelaySession) RelayDetectDigit(ctx context.Context, call *CallSession, controlID string, digits string) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	detectDigitParams := DetectDigitParams{
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

	return relay.RelayDetect(ctx, call, controlID, detect)
}

// RelayDetectFax TODO DESCRIPTION
func (relay *RelaySession) RelayDetectFax(ctx context.Context, call *CallSession, controlID string, faxtone string) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	detectFaxParams := DetectFaxParams{
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

	return relay.RelayDetect(ctx, call, controlID, detect)
}

// RelayDetectMachine TODO DESCRIPTION
func (relay *RelaySession) RelayDetectMachine(ctx context.Context, call *CallSession, controlID string, det *DetectMachineParams) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	detectMachineParams := DetectMachineParams{
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

	return relay.RelayDetect(ctx, call, controlID, detect)
}

// RelayDetect TODO DESCRIPTION
func (relay *RelaySession) RelayDetect(ctx context.Context, call *CallSession, controlID string, detect DetectStruct) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	v := ParamsBladeExecuteStruct{
		Protocol: relay.Blade.Protocol,
		Method:   "calling.detect",
		Params: ParamsCallDetect{
			NodeID:    call.NodeID,
			CallID:    call.CallID,
			ControlID: controlID,
			Detect:    detect,
			Timeout:   DefaultActionTimeout,
		},
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

	return nil
}

// RelayDetectStop TODO DESCRIPTION
func (relay *RelaySession) RelayDetectStop(ctx context.Context, call *CallSession, ctrlID *string) error {
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

	return nil
}

// RelaySendFax TODO DESCRIPTION
func (relay *RelaySession) RelaySendFax(ctx context.Context, call *CallSession, ctrlID *string, doc, id, headerInfo string) error {
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
			Document:   doc,
			Identity:   id,
			HeaderInfo: headerInfo,
		},
	}

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

	return nil
}

// RelayReceiveFax TODO DESCRIPTION
func (relay *RelaySession) RelayReceiveFax(ctx context.Context, call *CallSession, ctrlID *string) error {
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

	return nil
}

// RelaySendFaxStop TODO DESCRIPTION
func (relay *RelaySession) RelaySendFaxStop(ctx context.Context, call *CallSession, ctrlID *string) error {
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

	return nil
}

// RelayReceiveFaxStop TODO DESCRIPTION
func (relay *RelaySession) RelayReceiveFaxStop(ctx context.Context, call *CallSession, ctrlID *string) error {
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

	return nil
}

// RelayTapAudio TODO DESCRIPTION
func (relay *RelaySession) RelayTapAudio(ctx context.Context, call *CallSession, ctrlID, direction string, device *TapDevice) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
	}

	tapAudioParams := TapAudioParams{
		Direction: direction,
	}

	tap := TapStruct{
		Type:   "audio",
		Params: tapAudioParams,
	}

	return relay.RelayTap(ctx, call, ctrlID, tap, device)
}

// RelayTap TODO DESCRIPTION
func (relay *RelaySession) RelayTap(ctx context.Context, call *CallSession, controlID string, tap TapStruct, device *TapDevice) error {
	if len(call.CallID) == 0 {
		Log.Error("no CallID\n")

		return fmt.Errorf("no CallID for call [%p]", call)
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

	call.Lock()

	call.CallTapChans[controlID] = make(chan TapState, EventQueue)
	call.CallTapEventChans[controlID] = make(chan ParamsEventCallingCallTap, EventQueue)
	call.CallTapReadyChans[controlID] = make(chan struct{})

	call.Unlock()

	select {
	case call.CallTapControlIDs <- controlID:
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

	return nil
}

// RelayTapStop TODO DESCRIPTION
func (relay *RelaySession) RelayTapStop(ctx context.Context, call *CallSession, ctrlID *string) error {
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

	return nil
}

// RelaySendDigits TODO DESCRIPTION
func (relay *RelaySession) RelaySendDigits(ctx context.Context, call *CallSession, controlID, digits string) error {
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

	call.Lock()

	call.CallSendDigitsChans[controlID] = make(chan SendDigitsState, EventQueue)

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

	return nil
}

// RelayPlayAndCollect TODO DESCRIPTION
func (relay *RelaySession) RelayPlayAndCollect(ctx context.Context, call *CallSession, controlID string, playlist *[]PlayStruct, collect *CollectStruct) error {
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

	call.Lock()

	call.CallPlayAndCollectChans[controlID] = make(chan CollectResultType, EventQueue)
	call.CallPlayAndCollectEventChans[controlID] = make(chan ParamsEventCallingCallPlayAndCollect, EventQueue)
	call.CallPlayAndCollectReadyChans[controlID] = make(chan struct{})

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
func (relay *RelaySession) RelayPlayAndCollectVolume(ctx context.Context, call *CallSession, ctrlID *string, vol float64) error {
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
func (relay *RelaySession) RelayPlayAndCollectStop(ctx context.Context, call *CallSession, ctrlID *string) error {
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
