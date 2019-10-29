package signalwire

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// EventCalling  TODO DESCRIPTION
type EventCalling struct {
	blade *BladeSession
	Cache BCache
	I     IEventCalling
}

// EventCallingNew TODO DESCRIPTION
func EventCallingNew() *EventCalling {
	return &EventCalling{}
}

// IEventCalling  TODO DESCRIPTION
type IEventCalling interface {
	callConnectStateFromStr(s string) (CallConnectState, error)
	callStateFromStr(s string) (CallState, error)
	callPlayStateFromStr(s string) (PlayState, error)
	callRecordStateFromStr(s string) (RecordState, error)
	callDetectEventFromStr(event, detType string) (interface{}, error)
	callFaxEventFromStr(t string) (FaxEventType, error)
	callDisconnectReasonFromStr(r string) (CallDisconnectReason, error)
	callTapStateFromStr(s string) (TapState, error)
	callSendDigitsStateFromStr(s string) (SendDigitsState, error)
	callDirectionFromStr(s string) (CallDirection, error)
	callPlayAndCollectStateFromStr(s string) (CollectResultType, error)
	dispatchStateNotif(ctx context.Context, callParams CallParams) error
	dispatchConnectStateNotif(ctx context.Context, callParams CallParams, peer PeerDeviceStruct, ccstate CallConnectState) error
	dispatchPlayState(ctx context.Context, callID, ctrlID string, playState PlayState) error
	dispatchRecordState(ctx context.Context, callID, ctrlID string, recordState RecordState) error
	dispatchRecordEventParams(ctx context.Context, callID, ctrlID string, params ParamsEventCallingCallRecord) error
	dispatchDetect(ctx context.Context, callID, ctrlID string, v interface{}) error
	dispatchFax(ctx context.Context, callID, ctrlID string, faxType FaxEventType) error
	dispatchFaxEventParams(ctx context.Context, callID, ctrlID string, params ParamsEventCallingFax) error
	dispatchTapEventParams(ctx context.Context, callID, ctrlID string, params ParamsEventCallingCallTap) error
	dispatchTapState(ctx context.Context, callID, ctrlID string, tapState TapState) error
	dispatchSendDigitsState(ctx context.Context, callID, ctrlID string, sendDigitsState SendDigitsState) error
	dispatchPlayAndCollectEventParams(ctx context.Context, callID, ctrlID string, params ParamsEventCallingCallPlayAndCollect) error
	dispatchPlayAndCollectResType(ctx context.Context, callID, ctrlID string, resType CollectResultType) error
	getCall(ctx context.Context, tag, callID string) (*CallSession, error)
	getBroadcastParams(ctx context.Context, in, out interface{}) error
	onCallingEventConnect(ctx context.Context, broadcast NotifParamsBladeBroadcast) error
	onCallingEventReceive(ctx context.Context, broadcast NotifParamsBladeBroadcast) error
	onCallingEventState(ctx context.Context, broadcast NotifParamsBladeBroadcast) error
	onCallingEventPlay(ctx context.Context, broadcast NotifParamsBladeBroadcast) error
	onCallingEventCollect(ctx context.Context, broadcast NotifParamsBladeBroadcast) error
	onCallingEventRecord(ctx context.Context, broadcast NotifParamsBladeBroadcast) error
	onCallingEventTap(ctx context.Context, broadcast NotifParamsBladeBroadcast) error
	onCallingEventDetect(ctx context.Context, broadcast NotifParamsBladeBroadcast) error
	onCallingEventFax(ctx context.Context, broadcast NotifParamsBladeBroadcast) error
	onCallingEventSendDigits(ctx context.Context, broadcast NotifParamsBladeBroadcast) error
}

// EventMessaging  TODO DESCRIPTION
type EventMessaging struct {
	blade *BladeSession
	Cache BCache
	I     IEventMessaging
}

// EventMessagingNew TODO DESCRIPTION
func EventMessagingNew() *EventMessaging {
	return &EventMessaging{}
}

// IEventMessaging  TODO DESCRIPTION
type IEventMessaging interface {
	onMessagingEventState(ctx context.Context, broadcast NotifParamsBladeBroadcast) error
	dispatchMsgStateNotif(ctx context.Context, msgParams MsgParams) error
	msgStateFromStr(s string) (MsgState, error)
	msgDirectionFromStr(s string) (MsgDirection, error)
	getMsg(ctx context.Context, msgID string) (*MsgSession, error)
}

// the 'finished' keyword
const (
	Finished   = "finished"
	InboundStr = "inbound"
)

// callConnectStateFromStr TODO DESCRIPTION
func (*EventCalling) callConnectStateFromStr(s string) (CallConnectState, error) {
	var state CallConnectState

	switch strings.ToLower(s) {
	case "connecting":
		state = CallConnectConnecting
	case "connected":
		state = CallConnectConnected
	case "disconnected":
		state = CallConnectDisconnected
	case "failed":
		state = CallConnectFailed
	default:
		return state, errors.New("invalid CallConnectState")
	}

	Log.Debug("state [%s] [%s]\n", s, state.String())

	return state, nil
}

func (*EventCalling) callDirectionFromStr(s string) (CallDirection, error) {
	var dir CallDirection

	switch strings.ToLower(s) {
	case InboundStr:
		dir = CallInbound
	case "outbound":
		dir = CallOutbound
	default:
		return dir, errors.New("invalid Call Direction")
	}

	return dir, nil
}

func (*EventCalling) callStateFromStr(s string) (CallState, error) {
	var state CallState

	switch strings.ToLower(s) {
	case "created":
		state = Created
	case "ringing":
		state = Ringing
	case "answered":
		state = Answered
	case "ending":
		state = Ending
	case "ended":
		state = Ended
	default:
		return state, errors.New("invalid CallState")
	}

	Log.Debug("callstate [%s] [%s]\n", s, state.String())

	return state, nil
}

// callConnectStateFromStr TODO DESCRIPTION
func (*EventCalling) callPlayStateFromStr(s string) (PlayState, error) {
	var state PlayState

	switch strings.ToLower(s) {
	case "playing":
		state = PlayPlaying
	case "paused":
		state = PlayPaused
	case StrError:
		state = PlayError
	case Finished:
		state = PlayFinished
	default:
		return state, errors.New("invalid PlayState")
	}

	Log.Debug("state [%s] [%s]\n", s, state.String())

	return state, nil
}

// callRecordStateFromStr TODO DESCRIPTION
func (*EventCalling) callRecordStateFromStr(s string) (RecordState, error) {
	var state RecordState

	switch strings.ToLower(s) {
	case "recording":
		state = RecordRecording
	case Finished:
		state = RecordFinished
	case "no_input":
		state = RecordNoInput
	default:
		return state, errors.New("invalid RecordState")
	}

	Log.Debug("state [%s] [%s]\n", s, state.String())

	return state, nil
}

// callDetectEventFromStr TODO DESCRIPTION
func (*EventCalling) callDetectEventFromStr(event, detType string) (interface{}, error) {
	if len(detType) == 0 {
		return nil, errors.New("invalid Detector type")
	}

	switch detType {
	case "machine":
		var outevent DetectMachineEvent

		switch strings.ToLower(event) {
		case "unknown":
			outevent = DetectMachineUnknown
		case Finished:
			outevent = DetectMachineFinished
		case "machine":
			outevent = DetectMachineMachine
		case "human":
			outevent = DetectMachineHuman
		case "ready":
			outevent = DetectMachineReady
		case "not_ready":
			outevent = DetectMachineNotReady
		default:
			return nil, errors.New("invalid Detector Event")
		}

		return outevent, nil
	case "digit":
		var outevent DetectDigitEvent

		switch event {
		case "0":
			outevent = DetectDigitZero
		case "1":
			outevent = DetectDigitOne
		case "2":
			outevent = DetectDigitTwo
		case "3":
			outevent = DetectDigitThree
		case "4":
			outevent = DetectDigitFour
		case "5":
			outevent = DetectDigitFive
		case "6":
			outevent = DetectDigitSix
		case "7":
			outevent = DetectDigitSeven
		case "8":
			outevent = DetectDigitEight
		case "9":
			outevent = DetectDigitNine
		case "#":
			outevent = DetectDigitPound
		case "*":
			outevent = DetectDigitStar
		case Finished:
			outevent = DetectDigitFinished
		default:
			return nil, errors.New("invalid Detector Event")
		}

		return outevent, nil
	case "fax":
		var outevent DetectFaxEvent

		switch strings.ToLower(event) {
		case Finished:
			outevent = DetectFaxFinished
		case "ced":
			outevent = DetectFaxCED
		case "cng":
			outevent = DetectFaxCNG
		default:
			return nil, errors.New("invalid Detector Event")
		}

		return outevent, nil
	default:
		return nil, errors.New("invalid Detector Type")
	}
}

func (*EventCalling) callFaxEventFromStr(t string) (FaxEventType, error) {
	var faxtype FaxEventType

	switch strings.ToLower(t) {
	case "page":
		faxtype = FaxPage
	case Finished:
		faxtype = FaxFinished
	case "error":
		faxtype = FaxError
	default:
		return faxtype, errors.New("invalid Fax Detector Event")
	}

	return faxtype, nil
}

func (*EventCalling) callDisconnectReasonFromStr(r string) (CallDisconnectReason, error) {
	var reason CallDisconnectReason

	switch strings.ToLower(r) {
	case "hangup":
		reason = CallHangup
	case "cancel":
		reason = CallCancel
	case "busy":
		reason = CallBusy
	case "noanswer":
		reason = CallNoAnswer
	case "decline":
		reason = CallDecline
	case "error":
		reason = CallGenericError
	default:
		return reason, errors.New("invalid Disconnect Reason")
	}

	return reason, nil
}

// callTapStateFromStr TODO DESCRIPTION
func (*EventCalling) callTapStateFromStr(s string) (TapState, error) {
	var state TapState

	switch strings.ToLower(s) {
	case "tapping":
		state = TapTapping
	case Finished:
		state = TapFinished
	default:
		return state, errors.New("invalid Tap State")
	}

	Log.Debug("state [%s] [%s]\n", s, state.String())

	return state, nil
}

// callSendDigitsStateFromStr TODO DESCRIPTION
func (*EventCalling) callSendDigitsStateFromStr(s string) (SendDigitsState, error) {
	var state SendDigitsState

	switch strings.ToLower(s) {
	case Finished:
		state = SendDigitsFinished
	default:
		return state, errors.New("invalid Send Digits State")
	}

	Log.Debug("state [%s] [%s]\n", s, state.String())

	return state, nil
}

// callPlayAndCollectStateFromStr TODO DESCRIPTION
func (*EventCalling) callPlayAndCollectStateFromStr(s string) (CollectResultType, error) {
	var resType CollectResultType

	switch strings.ToLower(s) {
	case StrError:
		resType = CollectResultError
	case "no_input":
		resType = CollectResultNoInput
	case "no_match":
		resType = CollectResultNoMatch
	case "digit":
		resType = CollectResultDigit
	case "speech":
		resType = CollectResultSpeech
	case "start_of_speech":
		resType = CollectResultStartOfSpeech
	default:
		return resType, errors.New("invalid PlayAndCollect result type")
	}

	Log.Debug("resType [%s] [%s]\n", s, resType.String())

	return resType, nil
}

func (*EventMessaging) msgDirectionFromStr(s string) (MsgDirection, error) {
	var dir MsgDirection

	switch strings.ToLower(s) {
	case InboundStr:
		dir = MsgInbound
	case "outbound":
		dir = MsgOutbound
	default:
		return dir, errors.New("invalid Msg Direction")
	}

	return dir, nil
}

func (*EventMessaging) msgStateFromStr(s string) (MsgState, error) {
	var state MsgState

	switch strings.ToLower(s) {
	case "queued":
		state = MsgQueued
	case "initiated":
		state = MsgInitiated
	case "sent":
		state = MsgSent
	case "delivered":
		state = MsgDelivered
	case "undelivered":
		state = MsgUndelivered
	case "failed":
		state = MsgFailed
	case "received":
		state = MsgReceived
	default:
		return state, errors.New("invalid MsgState")
	}

	Log.Debug("msgstate [%s] [%s]\n", s, state.String())

	return state, nil
}

func getBroadcastGeneric(_ context.Context, in, out interface{}) error {
	var (
		jsonData []byte
		err      error
	)

	jsonData, err = json.Marshal(in)
	if err != nil {
		Log.Error("error marshaling Params\n")

		return err
	}

	if err = json.Unmarshal(jsonData, out); err != nil {
		Log.Error("error unmarshaling\n")

		return err
	}

	return nil
}

func (*EventCalling) getBroadcastParams(ctx context.Context, in, out interface{}) error {
	return getBroadcastGeneric(ctx, in, out)
}

func (*EventMessaging) getBroadcastParams(ctx context.Context, in, out interface{}) error {
	return getBroadcastGeneric(ctx, in, out)
}

func (calling *EventCalling) onCallingEventConnect(ctx context.Context, broadcast NotifParamsBladeBroadcast) error {
	var params ParamsEventCallingCallConnect

	if err := calling.getBroadcastParams(ctx, broadcast.Params.Params, &params); err != nil {
		return err
	}

	Log.Debug("broadcast.Params.Params.CallID: %v\n", params.CallID)
	Log.Debug("broadcast.Params.Params.NodeID: %v\n", params.NodeID)
	Log.Debug("broadcast.Params.Params.TagID: %v\n", params.TagID)
	Log.Debug("params.CallState: %v\n", params.ConnectState)

	state, err := calling.I.callConnectStateFromStr(params.ConnectState)
	if err != nil {
		return err
	}

	var callParams CallParams

	callParams.TagID = params.TagID
	callParams.CallID = params.CallID
	callParams.NodeID = params.NodeID

	return calling.I.dispatchConnectStateNotif(
		ctx,
		callParams,
		params.Peer,
		state,
	)
}

func (calling *EventCalling) onCallingEventState(ctx context.Context, broadcast NotifParamsBladeBroadcast) error {
	var params ParamsEventCallingCallState

	if err := calling.getBroadcastParams(ctx, broadcast.Params.Params, &params); err != nil {
		return err
	}

	state, err := calling.I.callStateFromStr(params.CallState)
	if err != nil {
		return err
	}

	var callParams CallParams /*keep params until we identify the call*/

	callParams.TagID = params.TagID
	callParams.CallID = params.CallID
	callParams.NodeID = params.NodeID
	callParams.Direction = params.Direction
	callParams.ToNumber = params.Device.Params.ToNumber
	callParams.FromNumber = params.Device.Params.FromNumber
	callParams.CallState = state
	callParams.EndReason = params.EndReason

	return calling.I.dispatchStateNotif(ctx, callParams)
}

// onCallingEvent_Receive  this is almost the same as onCallingEvent_State
func (calling *EventCalling) onCallingEventReceive(ctx context.Context, broadcast NotifParamsBladeBroadcast) error {
	var params ParamsEventCallingCallReceive

	if err := calling.getBroadcastParams(ctx, broadcast.Params.Params, &params); err != nil {
		return err
	}

	state, err := calling.I.callStateFromStr(params.CallState)
	if err != nil {
		return err
	}

	var callParams CallParams /*keep params until we identify the call*/

	callParams.TagID = params.TagID
	callParams.CallID = params.CallID
	callParams.NodeID = params.NodeID
	callParams.Direction = params.Direction
	callParams.ToNumber = params.Device.Params.ToNumber
	callParams.FromNumber = params.Device.Params.FromNumber
	callParams.CallState = state
	callParams.Context = params.Context

	// only state Created
	return calling.I.dispatchStateNotif(ctx, callParams)
}

func (calling *EventCalling) onCallingEventPlay(ctx context.Context, broadcast NotifParamsBladeBroadcast) error {
	Log.Debug("ctx: %p calling %p %v\n", ctx, calling, broadcast)

	var params ParamsEventCallingCallPlay

	if err := calling.getBroadcastParams(ctx, broadcast.Params.Params, &params); err != nil {
		return err
	}

	state, err := calling.I.callPlayStateFromStr(params.PlayState)
	if err != nil {
		return err
	}

	return calling.I.dispatchPlayState(
		ctx,
		params.CallID,
		params.ControlID,
		state,
	)
}

func (calling *EventCalling) onCallingEventCollect(ctx context.Context, broadcast NotifParamsBladeBroadcast) error {
	Log.Debug("ctx: %p calling %p %v\n", ctx, calling, broadcast)

	var params ParamsEventCallingCallPlayAndCollect

	if err := calling.getBroadcastParams(ctx, broadcast.Params.Params, &params); err != nil {
		return err
	}

	state, err := calling.I.callPlayAndCollectStateFromStr(params.Result.Type)
	if err != nil {
		return err
	}

	if err := calling.I.dispatchPlayAndCollectEventParams(ctx, params.CallID, params.ControlID, params); err != nil {
		return err
	}

	return calling.I.dispatchPlayAndCollectResType(
		ctx,
		params.CallID,
		params.ControlID,
		state,
	)
}

func (calling *EventCalling) onCallingEventRecord(ctx context.Context, broadcast NotifParamsBladeBroadcast) error {
	Log.Debug("ctx: %p calling %p %v\n", ctx, calling, broadcast)

	var params ParamsEventCallingCallRecord

	if err := calling.getBroadcastParams(ctx, broadcast.Params.Params, &params); err != nil {
		return err
	}

	state, err := calling.I.callRecordStateFromStr(params.RecordState)
	if err != nil {
		return err
	}

	if err := calling.I.dispatchRecordEventParams(ctx, params.CallID, params.ControlID, params); err != nil {
		return err
	}

	return calling.I.dispatchRecordState(
		ctx,
		params.CallID,
		params.ControlID,
		state,
	)
}

func (calling *EventCalling) onCallingEventTap(ctx context.Context, broadcast NotifParamsBladeBroadcast) error {
	Log.Debug("ctx: %p calling %p %v\n", ctx, calling, broadcast)

	var params ParamsEventCallingCallTap

	if err := calling.getBroadcastParams(ctx, broadcast.Params.Params, &params); err != nil {
		return err
	}

	state, err := calling.I.callTapStateFromStr(params.TapState)
	if err != nil {
		return err
	}

	if err := calling.I.dispatchTapEventParams(ctx, params.CallID, params.ControlID, params); err != nil {
		return err
	}

	return calling.I.dispatchTapState(
		ctx,
		params.CallID,
		params.ControlID,
		state,
	)
}

func (calling *EventCalling) onCallingEventDetect(ctx context.Context, broadcast NotifParamsBladeBroadcast) error {
	Log.Debug("ctx: %p calling %p %v\n", ctx, calling, broadcast)

	var params ParamsEventCallingCallDetect

	if err := calling.getBroadcastParams(ctx, broadcast.Params.Params, &params); err != nil {
		return err
	}

	event, err := calling.I.callDetectEventFromStr(params.Detect.Params.Event, params.Detect.Type)
	if err != nil {
		return err
	}

	return calling.I.dispatchDetect(
		ctx,
		params.CallID,
		params.ControlID,
		event,
	)
}

func (calling *EventCalling) onCallingEventFax(ctx context.Context, broadcast NotifParamsBladeBroadcast) error {
	Log.Debug("ctx: %p calling %p %v\n", ctx, calling, broadcast)

	var params ParamsEventCallingFax

	if err := calling.getBroadcastParams(ctx, broadcast.Params.Params, &params); err != nil {
		return err
	}

	if err := calling.I.dispatchFaxEventParams(ctx, params.CallID, params.ControlID, params); err != nil {
		return err
	}

	event, err := calling.I.callFaxEventFromStr(params.Fax.EventType)
	if err != nil {
		return err
	}

	return calling.I.dispatchFax(
		ctx,
		params.CallID,
		params.ControlID,
		event,
	)
}

func (calling *EventCalling) onCallingEventSendDigits(ctx context.Context, broadcast NotifParamsBladeBroadcast) error {
	Log.Debug("ctx: %p calling %p %v\n", ctx, calling, broadcast)

	var params ParamsEventCallingCallSendDigits

	if err := calling.getBroadcastParams(ctx, broadcast.Params.Params, &params); err != nil {
		return err
	}

	state, err := calling.I.callSendDigitsStateFromStr(params.SendDigitsState)
	if err != nil {
		return err
	}

	return calling.I.dispatchSendDigitsState(
		ctx,
		params.CallID,
		params.ControlID,
		state,
	)
}

func (messaging *EventMessaging) onMessagingEventState(ctx context.Context, broadcast NotifParamsBladeBroadcast) error {
	var params ParamsEventMessagingState

	if err := messaging.getBroadcastParams(ctx, broadcast.Params.Params, &params); err != nil {
		return err
	}

	state, err := messaging.I.msgStateFromStr(params.MessageState)
	if err != nil {
		return err
	}

	dir, err := messaging.I.msgDirectionFromStr(params.Direction)
	if err != nil {
		return err
	}

	var msgParams MsgParams

	msgParams.MsgID = params.MsgID
	msgParams.Context = params.Context
	msgParams.Direction = dir
	msgParams.To = params.ToNumber
	msgParams.From = params.FromNumber
	msgParams.MsgState = state
	msgParams.Segments = params.Segments
	msgParams.Body = params.Body
	msgParams.Tags = params.Tags
	msgParams.Media = params.Media

	return messaging.I.dispatchMsgStateNotif(ctx, msgParams)
}

func (calling *EventCalling) getCall(ctx context.Context, tag, callID string) (*CallSession, error) {
	var (
		call *CallSession
		err  error
	)

	Log.Debug("tag [%s] callid [%s] [%p]\n", tag, callID, calling)

	/* some events don't have the tag */
	if len(callID) > 0 && len(tag) > 0 {
		call, err = calling.Cache.GetCallCache(tag)
		if err != nil {
			Log.Debug("GetCallCache failed: %v", err)
		}

		if call != nil {
			// remove call object from the mapping and read with the call_id as key
			if err = calling.Cache.DeleteCallCache(tag); err != nil {
				Log.Debug("DeleteCallCache failed: %v", err)
			}

			if err = calling.Cache.SetCallCache(callID, call); err != nil {
				Log.Debug("SetCallCache failed: %v", err)
			}
		}
	}

	call, err = calling.Cache.GetCallCache(callID)
	if call == nil {
		// new inbound call
		call = new(CallSession)
		if err = calling.Cache.SetCallCache(callID, call); err != nil {
			Log.Debug("SetCallCache failed: %v\n", err)
		}

		call.CallInit(ctx)

		Log.Debug("new inbound call: [%p]\n", call)
	}

	return call, err
}

func (messaging *EventMessaging) getMsg(ctx context.Context, msgID string) (*MsgSession, error) {
	var (
		msg *MsgSession
		err error
	)

	Log.Debug("msgid [%s] [%p]\n", msgID, messaging)

	msg, err = messaging.Cache.GetMsgCache(msgID)
	if msg == nil {
		// new inbound msg
		msg = new(MsgSession)
		if err = messaging.Cache.SetMsgCache(msgID, msg); err != nil {
			Log.Debug("SetMsgCache failed: %v\n", err)
		}

		msg.MsgInit(ctx)

		Log.Debug("new inbound msg: [%p]\n", msg)
	}

	return msg, err
}

func (calling *EventCalling) dispatchStateNotif(ctx context.Context, callParams CallParams) error {
	Log.Debug("tag [%s] callstate [%s] blade [%p] direction: %s\n", callParams.TagID, callParams.CallState.String(), calling.blade, callParams.Direction)
	Log.Debug("direction : %v\n", callParams.Direction)

	call, _ := calling.I.getCall(ctx, callParams.TagID, callParams.CallID)
	if call == nil {
		return fmt.Errorf("error, nil CallSession")
	}

	Log.Debug("call [%p]\n", call)

	direction, err := calling.I.callDirectionFromStr(callParams.Direction)
	if err != nil {
		return err
	}

	call.SetParams(callParams.CallID, callParams.NodeID, callParams.ToNumber, callParams.FromNumber, callParams.Context, direction)

	call.UpdateCallState(callParams.CallState)

	call.Blade = calling.blade

	if (callParams.CallState == Created) && (callParams.Direction == InboundStr) {
		calling.blade.I.handleInboundCall(ctx, callParams.CallID)
	}

	if callParams.CallState == Ended {
		call.SetActive(false)

		disconnectReason, err := calling.I.callDisconnectReasonFromStr(callParams.EndReason)
		if err != nil {
			return err
		}

		call.SetDisconnectReason(disconnectReason)
	}
	select {
	case call.CallStateChan <- callParams.CallState:
		Log.Debug("sent callstate\n")
	default:
		Log.Debug("no callstate sent\n")
	}

	return nil
}

func (calling *EventCalling) dispatchConnectStateNotif(ctx context.Context, callParams CallParams, peer PeerDeviceStruct, ccstate CallConnectState) error {
	Log.Debug("tag [%s] [%s] [%p]\n", callParams.TagID, ccstate.String(), calling.blade)

	call, _ := calling.I.getCall(ctx, callParams.TagID, callParams.CallID)
	if call == nil {
		return fmt.Errorf("error, nil CallSession")
	}

	Log.Debug("call [%p]\n", call)

	call.UpdateCallConnectState(ccstate)

	if ccstate == CallConnectConnected {
		call.UpdateConnectPeer(peer)
	}

	select {
	case call.CallConnectStateChan <- ccstate:
		Log.Debug("sent connstate\n")
	default:
		Log.Debug("no connstate sent\n")
	}

	return nil
}

func (calling *EventCalling) dispatchPlayState(ctx context.Context, callID, ctrlID string, playState PlayState) error {
	Log.Debug("callid [%s] playstate [%s] blade [%p] ctrlID: %s\n", callID, playState, calling.blade, ctrlID)

	call, _ := calling.I.getCall(ctx, "", callID)
	if call == nil {
		return fmt.Errorf("error, nil CallSession")
	}

	Log.Debug("call [%p]\n", call)

	select {
	case call.CallPlayChans[ctrlID] <- playState:
		Log.Debug("sent playstate\n")
	default:
		Log.Debug("no playstate sent\n")
	}

	return nil
}

func (calling *EventCalling) dispatchRecordState(ctx context.Context, callID, ctrlID string, recordState RecordState) error {
	Log.Debug("callid [%s] recordstate [%s] blade [%p] ctrlID: %s\n", callID, recordState.String(), calling.blade, ctrlID)

	call, _ := calling.I.getCall(ctx, "", callID)
	if call == nil {
		return fmt.Errorf("error, nil CallSession")
	}

	Log.Debug("call [%p]\n", call)

	/* // possibly redundant, we have a map of Recording channels with key as ctrlID
	if call.GetActionState(ctrlID) == "" {
		return fmt.Errorf("error, unknown control ID: %s", ctrlID)
	}*/
	<-call.CallRecordReadyChans[ctrlID]
	select {
	case call.CallRecordChans[ctrlID] <- recordState:
		Log.Debug("sent recordstate\n")
	default:
		Log.Debug("no recordstate sent\n")
	}

	return nil
}

func (calling *EventCalling) dispatchRecordEventParams(ctx context.Context, callID, ctrlID string, params ParamsEventCallingCallRecord) error {
	Log.Debug("callid [%s] blade [%p] ctrlID: %s\n", callID, calling.blade, ctrlID)

	call, _ := calling.I.getCall(ctx, "", callID)
	if call == nil {
		return fmt.Errorf("error, nil CallSession")
	}

	Log.Debug("call [%p]\n", call)

	select {
	case call.CallRecordEventChans[ctrlID] <- params:
		Log.Debug("sent params (event)\n")
	default:
		Log.Debug("no params (event) sent\n")
	}

	return nil
}

func (calling *EventCalling) dispatchDetect(ctx context.Context, callID, ctrlID string, v interface{}) error {
	Log.Debug("callid [%s] detectevent [%v] blade [%p] ctrlID: %s\n", callID, v, calling.blade, ctrlID)

	call, _ := calling.I.getCall(ctx, "", callID)
	if call == nil {
		return fmt.Errorf("error, nil CallSession")
	}

	Log.Debug("call [%p]\n", call)

	detectEventMachine, ok1 := v.(DetectMachineEvent)
	if ok1 {
		select {
		case call.CallDetectMachineChans[ctrlID] <- detectEventMachine:
			Log.Debug("sent detectevent Machine\n")
		default:
			Log.Debug("no detectevent sent - Machine\n")
		}

		return nil
	}

	detectEventDigit, ok2 := v.(DetectDigitEvent)
	if ok2 {
		select {
		case call.CallDetectDigitChans[ctrlID] <- detectEventDigit:
			Log.Debug("sent detectevent Digit\n")
		default:
			Log.Debug("no detectevent sent - Digit\n")
		}

		return nil
	}

	detectEventFax, ok3 := v.(DetectFaxEvent)
	if ok3 {
		select {
		case call.CallDetectFaxChans[ctrlID] <- detectEventFax:
			Log.Debug("sent detectevent Fax\n")
		default:
			Log.Debug("no detectevent sent - Fax\n")
		}

		return nil
	}

	Log.Error("type assertion failed (detector event)\n")

	return nil
}

// todo : Event in the action
func (calling *EventCalling) dispatchDetectEventParams(ctx context.Context, callID, ctrlID string, params ParamsEventCallingCallDetect) error {
	Log.Debug("callid [%s] blade [%p] ctrlID: %s\n", callID, calling.blade, ctrlID)

	call, _ := calling.I.getCall(ctx, "", callID)
	if call == nil {
		return fmt.Errorf("error, nil CallSession")
	}

	Log.Debug("call [%p]\n", call)

	select {
	case call.CallDetectEventChans[ctrlID] <- params:
		Log.Debug("sent params (event)\n")
	default:
		Log.Debug("no params (event) sent\n")
	}

	return nil
}

func (calling *EventCalling) dispatchFax(ctx context.Context, callID, ctrlID string, faxType FaxEventType) error {
	Log.Debug("callid [%s] faxtype [%s] blade [%p] ctrlID: %s\n", callID, faxType.String(), calling.blade, ctrlID)

	call, _ := calling.I.getCall(ctx, "", callID)
	if call == nil {
		return fmt.Errorf("error, nil CallSession")
	}

	Log.Debug("call [%p]\n", call)

	<-call.CallFaxReadyChan
	select {
	case call.CallFaxChan <- faxType:
		Log.Debug("sent faxType\n")
	default:
		Log.Debug("no faxType sent\n")
	}

	return nil
}

func (calling *EventCalling) dispatchFaxEventParams(ctx context.Context, callID, ctrlID string, params ParamsEventCallingFax) error {
	Log.Debug("callid [%s] blade [%p] ctrlID: %s\n", callID, calling.blade, ctrlID)

	call, _ := calling.I.getCall(ctx, "", callID)
	if call == nil {
		return fmt.Errorf("error, nil CallSession")
	}

	Log.Debug("call [%p]\n", call)

	select {
	case call.CallFaxEventChan <- params.Fax:
		Log.Debug("sent params (event)\n")
	default:
		Log.Debug("no params (event) sent\n")
	}

	return nil
}

func (calling *EventCalling) dispatchTapEventParams(ctx context.Context, callID, ctrlID string, params ParamsEventCallingCallTap) error {
	Log.Debug("callid [%s] blade [%p] ctrlID: %s\n", callID, calling.blade, ctrlID)

	call, _ := calling.I.getCall(ctx, "", callID)
	if call == nil {
		return fmt.Errorf("error, nil CallSession")
	}

	Log.Debug("call [%p]\n", call)

	select {
	case call.CallTapEventChans[ctrlID] <- params:
		Log.Debug("sent params (event)\n")
	default:
		Log.Debug("no params (event) sent\n")
	}

	return nil
}

func (calling *EventCalling) dispatchTapState(ctx context.Context, callID, ctrlID string, tapState TapState) error {
	Log.Debug("callid [%s] tapstate [%s] blade [%p] ctrlID: %s\n", callID, tapState.String(), calling.blade, ctrlID)

	call, _ := calling.I.getCall(ctx, "", callID)
	if call == nil {
		return fmt.Errorf("error, nil CallSession")
	}

	Log.Debug("call [%p]\n", call)

	<-call.CallTapReadyChans[ctrlID]
	select {
	case call.CallTapChans[ctrlID] <- tapState:
		Log.Debug("sent recordstate\n")
	default:
		Log.Debug("no recordstate sent\n")
	}

	return nil
}

func (calling *EventCalling) dispatchSendDigitsState(ctx context.Context, callID, ctrlID string, sendDigitsState SendDigitsState) error {
	Log.Debug("callid [%s] sendDigitsState [%s] blade [%p] ctrlID: %s\n", callID, sendDigitsState.String(), calling.blade, ctrlID)

	call, _ := calling.I.getCall(ctx, "", callID)
	if call == nil {
		return fmt.Errorf("error, nil CallSession")
	}

	Log.Debug("call [%p]\n", call)

	select {
	case call.CallSendDigitsChans[ctrlID] <- sendDigitsState:
		Log.Debug("sent recordstate\n")
	default:
		Log.Debug("no recordstate sent\n")
	}

	return nil
}

func (calling *EventCalling) dispatchPlayAndCollectResType(ctx context.Context, callID, ctrlID string, resType CollectResultType) error {
	Log.Debug("callid [%s] resType [%s] blade [%p] ctrlID: %s\n", callID, resType.String(), calling.blade, ctrlID)

	call, _ := calling.I.getCall(ctx, "", callID)
	if call == nil {
		return fmt.Errorf("error, nil CallSession")
	}

	Log.Debug("call [%p]\n", call)

	<-call.CallPlayAndCollectReadyChans[ctrlID]
	select {
	case call.CallPlayAndCollectChans[ctrlID] <- resType:
		Log.Debug("sent collect resType\n")
	default:
		Log.Debug("no collect resType sent\n")
	}

	return nil
}

func (calling *EventCalling) dispatchPlayAndCollectEventParams(ctx context.Context, callID, ctrlID string, params ParamsEventCallingCallPlayAndCollect) error {
	Log.Debug("callid [%s] blade [%p] ctrlID: %s\n", callID, calling.blade, ctrlID)

	call, _ := calling.I.getCall(ctx, "", callID)
	if call == nil {
		return fmt.Errorf("error, nil CallSession")
	}

	Log.Debug("call [%p]\n", call)

	select {
	case call.CallPlayAndCollectEventChans[ctrlID] <- params:
		Log.Debug("sent params (event)\n")
	default:
		Log.Debug("no params (event) sent\n")
	}

	return nil
}

func (messaging *EventMessaging) dispatchMsgStateNotif(ctx context.Context, msgParams MsgParams) error {
	Log.Debug("msgID [%s] msgstate [%s] blade [%p] direction: %s\n", msgParams.MsgID, msgParams.MsgState.String(), messaging.blade, msgParams.Direction)
	Log.Debug("direction : %v\n", msgParams.Direction)

	msg, _ := messaging.I.getMsg(ctx, msgParams.MsgID)
	if msg == nil {
		return fmt.Errorf("error, nil MsgSession")
	}

	msg.SetParams(msgParams.MsgID, msgParams.To, msgParams.From, msgParams.Context, msgParams.Direction)

	msg.UpdateMsgState(msgParams.MsgState)

	msg.Blade = messaging.blade

	if (msgParams.MsgState == MsgReceived) && (msgParams.Direction == MsgInbound) {
		messaging.blade.I.handleInboundMessage(ctx, msgParams.MsgID)
	}

	select {
	case msg.MsgStateChan <- msgParams.MsgState:
		Log.Debug("sent msgstate\n")
	default:
		Log.Debug("no msgstate sent\n")
	}

	return nil
}
