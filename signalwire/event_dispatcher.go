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
	callingNotif(ctx context.Context, b NotifParamsBladeBroadcast) error
	callConnectStateFromStr(s string) (CallConnectState, error)
	callStateFromStr(s string) (CallState, error)
	callPlayStateFromStr(s string) (PlayState, error)
	callRecordStateFromStr(s string) (RecordState, error)
	callDetectEventFromStr(event, detType string) (interface{}, error)
	callFaxEventFromStr(t string) (FaxEventType, error)
	dispatchStateNotif(ctx context.Context, callParams CallParams) error
	dispatchConnectStateNotif(ctx context.Context, callParams CallParams, peer PeerDeviceStruct, ccstate CallConnectState) error
	dispatchPlayState(ctx context.Context, callID, ctrlID string, playState PlayState) error
	dispatchRecordState(ctx context.Context, callID, ctrlID string, recordState RecordState) error
	dispatchRecordEventParams(ctx context.Context, callID, ctrlID string, params ParamsEventCallingCallRecord) error
	dispatchDetect(ctx context.Context, callID, ctrlID string, v interface{}) error
	dispatchFax(ctx context.Context, callID, ctrlID string, faxType FaxEventType) error
	dispatchFaxEventParams(ctx context.Context, callID, ctrlID string, params ParamsEventCallingFax) error
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

const (
	Finished = "finished"
)

// callConnectStateFromStr TODO DESCRIPTION
func (*EventCalling) callConnectStateFromStr(s string) (CallConnectState, error) {
	var state CallConnectState

	switch strings.ToLower(s) {
	case "connecting":
		state = Connecting
	case "connected":
		state = Connected
	case "disconnected":
		state = Disconnected
	case "failed":
		state = Failed
	default:
		return state, errors.New("invalid CallConnectState")
	}

	Logger.Debugf("state [%s] [%s]\n", s, state.String())

	return state, nil
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

	Logger.Debugf("callstate [%s] [%s]\n", s, state.String())

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
	case "error":
		state = PlayError
	case Finished:
		state = PlayFinished
	default:
		return state, errors.New("invalid PlayState")
	}

	Logger.Debugf("state [%s] [%s]\n", s, state.String())

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

	Logger.Debugf("state [%s] [%s]\n", s, state.String())

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
		return faxtype, errors.New("invalid FaxEventType")
	}

	return faxtype, nil
}

func (*EventCalling) getBroadcastParams(_ context.Context, in, out interface{}) error {
	var (
		jsonData []byte
		err      error
	)

	jsonData, err = json.Marshal(in)
	if err != nil {
		Logger.Errorf("error marshaling Params\n")

		return err
	}

	if err = json.Unmarshal(jsonData, out); err != nil {
		Logger.Errorf("error unmarshaling\n")

		return err
	}

	return nil
}

func (calling *EventCalling) onCallingEventConnect(ctx context.Context, broadcast NotifParamsBladeBroadcast) error {
	var params ParamsEventCallingCallConnect

	if err := calling.getBroadcastParams(ctx, broadcast.Params.Params, &params); err != nil {
		return err
	}

	Logger.Debugf("broadcast.Params.Params.CallID: %v\n", params.CallID)
	Logger.Debugf("broadcast.Params.Params.NodeID: %v\n", params.NodeID)
	Logger.Debugf("broadcast.Params.Params.TagID: %v\n", params.TagID)
	Logger.Debugf("params.ConnectState: %v\n", params.ConnectState)

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

	// only state Created
	return calling.I.dispatchStateNotif(ctx, callParams)
}

func (calling *EventCalling) onCallingEventPlay(ctx context.Context, broadcast NotifParamsBladeBroadcast) error {
	Logger.Debugf("ctx: %p calling %p %v\n", ctx, calling, broadcast)

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
	Logger.Debugf("ctx: %p calling %p %v\n", ctx, calling, broadcast)
	return nil
}

func (calling *EventCalling) onCallingEventRecord(ctx context.Context, broadcast NotifParamsBladeBroadcast) error {
	Logger.Debugf("ctx: %p calling %p %v\n", ctx, calling, broadcast)

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
	Logger.Debugf("ctx: %p calling %p %v\n", ctx, calling, broadcast)
	return nil
}

func (calling *EventCalling) onCallingEventDetect(ctx context.Context, broadcast NotifParamsBladeBroadcast) error {
	Logger.Debugf("ctx: %p calling %p %v\n", ctx, calling, broadcast)

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
	Logger.Debugf("ctx: %p calling %p %v\n", ctx, calling, broadcast)

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
	Logger.Debugf("ctx: %p calling %p %v\n", ctx, calling, broadcast)
	return nil
}

// HandleBladeBroadcast TODO DESCRIPTION
func (calling *EventCalling) callingNotif(ctx context.Context, broadcast NotifParamsBladeBroadcast) error {
	switch broadcast.Event {
	case "queuing.relay.events":
		switch broadcast.Params.EventType {
		case "calling.call.connect":
			if err := calling.onCallingEventConnect(ctx, broadcast); err != nil {
				return err
			}
		case "calling.call.state":
			if err := calling.onCallingEventState(ctx, broadcast); err != nil {
				return err
			}
		case "calling.call.receive":
			if err := calling.onCallingEventReceive(ctx, broadcast); err != nil {
				return err
			}
		case "calling.call.play":
			if err := calling.onCallingEventPlay(ctx, broadcast); err != nil {
				return err
			}
		case "calling.call.collect":
			if err := calling.onCallingEventCollect(ctx, broadcast); err != nil {
				return err
			}
		case "calling.call.record":
			if err := calling.onCallingEventRecord(ctx, broadcast); err != nil {
				return err
			}
		case "calling.call.tap":
			if err := calling.onCallingEventTap(ctx, broadcast); err != nil {
				return err
			}
		case "calling.call.detect":
			if err := calling.onCallingEventDetect(ctx, broadcast); err != nil {
				return err
			}
		case "calling.call.fax":
			if err := calling.onCallingEventFax(ctx, broadcast); err != nil {
				return err
			}
		case "calling.call.send_digits":
			if err := calling.onCallingEventSendDigits(ctx, broadcast); err != nil {
				return err
			}
		default:
			Logger.Debugf("got event_type %s\n", broadcast.Params.EventType)
		}
	case "relay":
		Logger.Debugf("got RELAY event\n")
	default:
		Logger.Debugf("got event %s . unsupported\n", broadcast.Event)
		return fmt.Errorf("unsupported event")
	}

	Logger.Debugf("broadcast: %v\n", broadcast)

	return nil
}

func (calling *EventCalling) getCall(ctx context.Context, tag, callID string) (*CallSession, error) {
	var (
		call *CallSession
		err  error
	)

	Logger.Debugf("tag [%s] callid [%s] [%p]\n", tag, callID, calling)

	/* some events don't have the tag */
	if len(callID) > 0 && len(tag) > 0 {
		call, err = calling.Cache.GetCallCache(tag)
		if err != nil {
			Logger.Debugf("GetCallCache failed: %v", err)
		}

		if call != nil {
			// remove call object from the mapping and read with the call_id as key
			if err = calling.Cache.DeleteCallCache(tag); err != nil {
				Logger.Debugf("DeleteCallCache failed: %v", err)
			}

			if err = calling.Cache.SetCallCache(callID, call); err != nil {
				Logger.Debugf("SetCallCache failed: %v", err)
			}
		}
	}

	call, err = calling.Cache.GetCallCache(callID)
	if call == nil {
		// new inbound call
		call = new(CallSession)
		if err = calling.Cache.SetCallCache(callID, call); err != nil {
			Logger.Debugf("SetCallCache failed: %v", err)
		}

		call.CallInit(ctx)

		Logger.Debugf("new inbound call: [%p]\n", call)
	}

	return call, err
}

func (calling *EventCalling) dispatchStateNotif(ctx context.Context, callParams CallParams) error {
	Logger.Debugf("tag [%s] callstate [%s] blade [%p] direction: %s\n", callParams.TagID, callParams.CallState.String(), calling.blade, callParams.Direction)
	Logger.Debugf("direction : %v\n", callParams.Direction)

	call, _ := calling.I.getCall(ctx, callParams.TagID, callParams.CallID)
	if call == nil {
		return fmt.Errorf("error, nil CallSession")
	}

	Logger.Debugf("call [%p]\n", call)

	call.SetParams(callParams.CallID, callParams.NodeID, callParams.Direction, callParams.ToNumber, callParams.FromNumber)

	call.UpdateCallState(callParams.CallState)

	call.Blade = calling.blade

	if (callParams.CallState == Created) && (callParams.Direction == "inbound") {
		calling.blade.I.handleInboundCall(ctx, callParams.CallID)
	}

	select {
	case call.CallStateChan <- callParams.CallState:
		Logger.Debugf("sent callstate\n")
	default:
		Logger.Debugf("no callstate sent\n")
	}

	return nil
}

func (calling *EventCalling) dispatchConnectStateNotif(ctx context.Context, callParams CallParams, peer PeerDeviceStruct, ccstate CallConnectState) error {
	Logger.Debugf("tag [%s] [%s] [%p]\n", callParams.TagID, ccstate.String(), calling.blade)

	call, _ := calling.I.getCall(ctx, callParams.TagID, callParams.CallID)
	if call == nil {
		return fmt.Errorf("error, nil CallSession")
	}

	Logger.Debugf("call [%p]\n", call)

	call.UpdateCallConnectState(ccstate)

	if ccstate == Connected {
		call.UpdateConnectPeer(peer)
	}

	select {
	case call.CallConnectStateChan <- ccstate:
		Logger.Debugf("sent connstate\n")
	default:
		Logger.Debugf("no connstate sent\n")
	}

	return nil
}

func (calling *EventCalling) dispatchPlayState(ctx context.Context, callID, ctrlID string, playState PlayState) error {
	Logger.Debugf("callid [%s] playstate [%s] blade [%p] ctrlID: %s\n", callID, playState, calling.blade, ctrlID)

	call, _ := calling.I.getCall(ctx, "", callID)
	if call == nil {
		return fmt.Errorf("error, nil CallSession")
	}

	Logger.Debugf("call [%p]\n", call)

	select {
	case call.CallPlayChans[ctrlID] <- playState:
		Logger.Debugf("sent playstate\n")
	default:
		Logger.Debugf("no playstate sent\n")
	}

	return nil
}

func (calling *EventCalling) dispatchRecordState(ctx context.Context, callID, ctrlID string, recordState RecordState) error {
	Logger.Debugf("callid [%s] recordstate [%s] blade [%p] ctrlID: %s\n", callID, recordState, calling.blade, ctrlID)

	call, _ := calling.I.getCall(ctx, "", callID)
	if call == nil {
		return fmt.Errorf("error, nil CallSession")
	}

	Logger.Debugf("call [%p]\n", call)

	/* // possibly redundant, we have a map of Recording channels with key as ctrlID
	if call.GetActionState(ctrlID) == "" {
		return fmt.Errorf("error, unknown control ID: %s", ctrlID)
	}*/
	<-call.CallRecordReadyChans[ctrlID]
	select {
	case call.CallRecordChans[ctrlID] <- recordState:
		Logger.Debugf("sent recordstate\n")
	default:
		Logger.Debugf("no recordstate sent\n")
	}

	return nil
}

func (calling *EventCalling) dispatchRecordEventParams(ctx context.Context, callID, ctrlID string, params ParamsEventCallingCallRecord) error {
	Logger.Debugf("callid [%s] blade [%p] ctrlID: %s\n", callID, calling.blade, ctrlID)

	call, _ := calling.I.getCall(ctx, "", callID)
	if call == nil {
		return fmt.Errorf("error, nil CallSession")
	}

	Logger.Debugf("call [%p]\n", call)

	select {
	case call.CallRecordEventChans[ctrlID] <- params:
		Logger.Debugf("sent params (event)\n")
	default:
		Logger.Debugf("no params (event) sent\n")
	}

	return nil
}

func (calling *EventCalling) dispatchDetect(ctx context.Context, callID, ctrlID string, v interface{}) error {
	Logger.Debugf("callid [%s] detectevent [%v] blade [%p] ctrlID: %s\n", callID, v, calling.blade, ctrlID)

	call, _ := calling.I.getCall(ctx, "", callID)
	if call == nil {
		return fmt.Errorf("error, nil CallSession")
	}

	Logger.Debugf("call [%p]\n", call)

	detectEventMachine, ok1 := v.(DetectMachineEvent)
	if ok1 {
		select {
		case call.CallDetectMachineChans[ctrlID] <- detectEventMachine:
			Logger.Debugf("sent detectevent Machine\n")
		default:
			Logger.Debugf("no detectevent sent - Machine\n")
		}

		return nil
	}

	detectEventDigit, ok2 := v.(DetectDigitEvent)
	if ok2 {
		select {
		case call.CallDetectDigitChans[ctrlID] <- detectEventDigit:
			Logger.Debugf("sent detectevent Digit\n")
		default:
			Logger.Debugf("no detectevent sent - Digit\n")
		}

		return nil
	}

	detectEventFax, ok3 := v.(DetectFaxEvent)
	if ok3 {
		select {
		case call.CallDetectFaxChans[ctrlID] <- detectEventFax:
			Logger.Debugf("sent detectevent Fax\n")
		default:
			Logger.Debugf("no detectevent sent - Fax\n")
		}

		return nil
	}

	Logger.Errorf("type assertion failed (detector event)\n")

	return nil
}

// todo : Event in the action
func (calling *EventCalling) dispatchDetectEventParams(ctx context.Context, callID, ctrlID string, params ParamsEventCallingCallDetect) error {
	Logger.Debugf("callid [%s] blade [%p] ctrlID: %s\n", callID, calling.blade, ctrlID)

	call, _ := calling.I.getCall(ctx, "", callID)
	if call == nil {
		return fmt.Errorf("error, nil CallSession")
	}

	Logger.Debugf("call [%p]\n", call)

	select {
	case call.CallDetectEventChans[ctrlID] <- params:
		Logger.Debugf("sent params (event)\n")
	default:
		Logger.Debugf("no params (event) sent\n")
	}

	return nil
}

func (calling *EventCalling) dispatchFax(ctx context.Context, callID, ctrlID string, faxType FaxEventType) error {
	Logger.Debugf("callid [%s] faxtype [%s] blade [%p] ctrlID: %s\n", callID, faxType.String(), calling.blade, ctrlID)

	call, _ := calling.I.getCall(ctx, "", callID)
	if call == nil {
		return fmt.Errorf("error, nil CallSession")
	}

	Logger.Debugf("call [%p]\n", call)

	<-call.CallFaxReadyChan
	select {
	case call.CallFaxChan <- faxType:
		Logger.Debugf("sent faxType\n")
	default:
		Logger.Debugf("no faxType sent\n")
	}

	return nil
}

func (calling *EventCalling) dispatchFaxEventParams(ctx context.Context, callID, ctrlID string, params ParamsEventCallingFax) error {
	Logger.Debugf("callid [%s] blade [%p] ctrlID: %s\n", callID, calling.blade, ctrlID)

	call, _ := calling.I.getCall(ctx, "", callID)
	if call == nil {
		return fmt.Errorf("error, nil CallSession")
	}

	Logger.Debugf("call [%p]\n", call)

	select {
	case call.CallFaxEventChan <- params.Fax:
		Logger.Debugf("sent params (event)\n")
	default:
		Logger.Debugf("no params (event) sent\n")
	}

	return nil
}
