package signalwire

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

// CallState keeps the state of a call
type CallState int

// Call state constants
const (
	Created CallState = iota
	Ringing
	Answered
	Ending
	Ended
)

func (s CallState) String() string {
	return [...]string{"Created", "Ringing", "Answered", "Ending", "Ended"}[s]
}

// CallDisconnectReason describes reason for call disconnection
type CallDisconnectReason int

// Call disconnection constants
const (
	CallHangup CallDisconnectReason = iota
	CallCancel
	CallBusy
	CallNoAnswer
	CallDecline
	CallGenericError
)

func (s CallDisconnectReason) String() string {
	return [...]string{"Hangup", "Cancel", "Busy", "NoAnswer", "Decline", "Error"}[s]
}

// CallDirection has the direction of a call
type CallDirection int

// Call state constants
const (
	CallInbound CallDirection = iota
	CallOutbound
)

func (s CallDirection) String() string {
	return [...]string{"Inbound", "Outbound"}[s]
}

// CallTypeState keeps the type of a call
type CallType int

// Call state constants
const (
	CallTypePhone CallType = iota
	CallTypeSIP
	CallWebrtc
	CallAgora
)

func (s CallType) String() string {
	return [...]string{"phone", "sip", "webrtc", "agora"}[s]
}

// CallSession internal representation of a call
type CallSession struct {
	Active                  bool
	To                      string
	From                    string
	TagID                   string
	Timeout                 uint // ring timeout
	CallID                  string
	NodeID                  string
	ProjectID               string
	SpaceID                 string
	Direction               CallDirection
	Context                 string
	CallState               CallState
	PrevCallState           CallState
	CallType                CallType
	CallDisconnectReason    CallDisconnectReason
	CallConnectState        CallConnectState
	CallStateChan           chan CallState
	cbStateChan             chan CallState
	CallConnectStateChan    chan CallConnectState
	CallConnectRawEventChan chan *json.RawMessage

	CallPlayChans         map[string](chan PlayState)
	CallPlayControlIDs    chan string
	CallPlayEventChans    map[string](chan ParamsEventCallingCallPlay)
	CallPlayReadyChans    map[string](chan struct{})
	CallPlayRawEventChans map[string](chan *json.RawMessage)

	CallRecordChans         map[string](chan RecordState)
	CallRecordControlIDs    chan string
	CallRecordEventChans    map[string](chan ParamsEventCallingCallRecord)
	CallRecordReadyChans    map[string](chan struct{})
	CallRecordRawEventChans map[string](chan *json.RawMessage)

	CallDetectMachineChans     map[string](chan DetectMachineEvent)
	CallDetectDigitChans       map[string](chan DetectDigitEvent)
	CallDetectFaxChans         map[string](chan DetectFaxEvent)
	CallDetectMachineControlID chan string
	CallDetectDigitControlID   chan string
	CallDetectFaxControlID     chan string
	CallDetectEventChans       map[string](chan ParamsEventCallingCallDetect)
	CallDetectReadyChans       map[string](chan struct{})
	CallDetectRawEventChans    map[string](chan *json.RawMessage)

	CallFaxChan         chan FaxEventType
	CallFaxControlID    chan string
	CallFaxEventChan    chan FaxEventStruct
	CallFaxReadyChan    chan struct{}
	CallFaxRawEventChan chan *json.RawMessage

	CallTapChans         map[string](chan TapState)
	CallTapControlIDs    chan string
	CallTapEventChans    map[string](chan ParamsEventCallingCallTap)
	CallTapReadyChans    map[string](chan struct{})
	CallTapRawEventChans map[string](chan *json.RawMessage)

	CallSendDigitsChans         map[string](chan SendDigitsState)
	CallSendDigitsControlIDs    chan string
	CallSenDigitsEventChans     map[string](chan ParamsEventCallingCallSendDigits)
	CallSendDigitsReadyChans    map[string](chan struct{})
	CallSendDigitsRawEventChans map[string](chan *json.RawMessage)

	CallPlayAndCollectChans         map[string](chan CollectResultType)
	CallPlayAndCollectControlID     chan string
	CallPlayAndCollectEventChans    map[string](chan ParamsEventCallingCallPlayAndCollect)
	CallPlayAndCollectReadyChans    map[string](chan struct{})
	CallPlayAndCollectRawEventChans map[string](chan *json.RawMessage)

	Hangup   chan struct{}
	CallPeer PeerDeviceStruct
	Actions  Actions
	Blade    *BladeSession
	I        ICall
	sync.RWMutex
}

// ICall Call Interface
type ICall interface {
	CallInit(ctx context.Context)
	CallCleanup(ctx context.Context)
	UpdateCallState(s CallState)
	UpdateCallConnectState(s CallConnectState)
	UpdateConnectPeer(p PeerDeviceStruct)
	WaitCallStateInternal(ctx context.Context, want CallState, timeoutSec uint) bool
	WaitCallConnectState(ctx context.Context, want CallConnectState) bool
	GetPeer(ctx context.Context) (*CallSession, error)
}

// CallTagToCallID map of tag to Call-ID
type CallTagToCallID struct {
	sync.RWMutex
	m map[string]string
}

// NewCallTagToCallID returns new map of tag to Call-ID
func NewCallTagToCallID() *CallTagToCallID {
	var obj CallTagToCallID

	obj.m = make(map[string]string)

	return &obj
}

// CallInit creates the channels communicating call states
func (c *CallSession) CallInit(_ context.Context) {
	c.CallStateChan = make(chan CallState, EventQueue)
	c.CallConnectStateChan = make(chan CallConnectState, EventQueue)
	c.CallConnectRawEventChan = make(chan *json.RawMessage)
	c.cbStateChan = make(chan CallState, EventQueue)

	c.CallPlayChans = make(map[string](chan PlayState))
	c.CallPlayControlIDs = make(chan string, SimActionsOfTheSameKind)
	c.CallPlayEventChans = make(map[string](chan ParamsEventCallingCallPlay))
	c.CallPlayReadyChans = make(map[string](chan struct{}))
	c.CallPlayRawEventChans = make(map[string](chan *json.RawMessage))

	c.CallRecordChans = make(map[string](chan RecordState))
	c.CallRecordControlIDs = make(chan string, SimActionsOfTheSameKind)
	c.CallRecordEventChans = make(map[string](chan ParamsEventCallingCallRecord))
	c.CallRecordReadyChans = make(map[string](chan struct{}))
	c.CallRecordRawEventChans = make(map[string](chan *json.RawMessage))

	c.CallDetectMachineControlID = make(chan string, 1)
	c.CallDetectDigitControlID = make(chan string, 1)
	c.CallDetectFaxControlID = make(chan string, 1)

	c.CallDetectMachineChans = make(map[string](chan DetectMachineEvent))
	c.CallDetectDigitChans = make(map[string](chan DetectDigitEvent))
	c.CallDetectFaxChans = make(map[string](chan DetectFaxEvent))
	c.CallDetectReadyChans = make(map[string](chan struct{}))
	c.CallDetectRawEventChans = make(map[string](chan *json.RawMessage))

	c.CallFaxChan = make(chan FaxEventType, EventQueue)
	c.CallFaxControlID = make(chan string, 1)
	c.CallFaxReadyChan = make(chan struct{})
	c.CallFaxEventChan = make(chan FaxEventStruct)
	c.CallFaxRawEventChan = make(chan *json.RawMessage)

	c.CallTapChans = make(map[string](chan TapState))
	c.CallTapControlIDs = make(chan string, 1)
	c.CallTapEventChans = make(map[string](chan ParamsEventCallingCallTap))
	c.CallTapReadyChans = make(map[string](chan struct{}))
	c.CallTapRawEventChans = make(map[string](chan *json.RawMessage))

	c.CallSendDigitsChans = make(map[string](chan SendDigitsState))
	c.CallSendDigitsControlIDs = make(chan string, 1)
	c.CallSendDigitsReadyChans = make(map[string](chan struct{}))
	c.CallSendDigitsRawEventChans = make(map[string](chan *json.RawMessage))

	c.CallPlayAndCollectChans = make(map[string](chan CollectResultType))
	c.CallPlayAndCollectControlID = make(chan string, 1)
	c.CallPlayAndCollectEventChans = make(map[string](chan ParamsEventCallingCallPlayAndCollect))
	c.CallPlayAndCollectReadyChans = make(map[string](chan struct{}))
	c.CallPlayAndCollectRawEventChans = make(map[string](chan *json.RawMessage))

	c.Hangup = make(chan struct{})
	c.Actions.m = make(map[string]string)
}

// CallCleanup close the channels, etc
// app writer should set CallObj and/or c to nil to clean it up
func (c *CallSession) CallCleanup(_ context.Context) {
	close(c.CallStateChan)
	close(c.CallConnectStateChan)
	close(c.Hangup)

	c.Actions.m = nil
}

// WaitCallStateInternal wait for a certain call state and return true when it arrives.
// return false if timeout.
func (c *CallSession) WaitCallStateInternal(ctx context.Context, want CallState, timeoutSec uint) bool {
	var ret bool

	for {
		var out bool

		select {
		case callstate := <-c.CallStateChan:
			if callstate == want {
				out = true
				ret = true
			} else if callstate == Ended {
				out = true
			}
		case <-time.After(time.Duration(timeoutSec) * time.Second):
			out = true
		case <-c.Hangup:
			out = true
		case <-ctx.Done():
			out = true
		}

		if out {
			break
		}
	}

	return ret
}

// WaitCallConnectState wait for a certain call connect state and return true when it arrives.
// return false if timeout.
func (c *CallSession) WaitCallConnectState(_ context.Context, want CallConnectState) bool {
	var ret bool

	for {
		var out bool

		select {
		case callconnectstate := <-c.CallConnectStateChan:
			if callconnectstate == want {
				out = true
				ret = true
			}
		case <-time.After(BroadcastEventTimeout * time.Second):
			out = true
		case <-c.Hangup:
			out = true
		}

		if out {
			break
		}
	}

	return ret
}

// GetPeer get the call peer if the peer is local
func (c *CallSession) GetPeer(_ context.Context) (*CallSession, error) {
	var (
		peercall *CallSession
		err      error
	)

	if len(c.CallPeer.CallID) > 0 && c.Blade != nil {
		peercall, err = c.Blade.EventCalling.Cache.GetCallCache(c.CallPeer.CallID)
	}

	return peercall, err
}

// UpdateCallState TODO DESCRIPTION
func (c *CallSession) UpdateCallState(s CallState) {
	c.Lock()
	c.PrevCallState = c.CallState
	c.CallState = s
	c.Unlock()
}

// SetParams setting Params that stay the same during the call*/
func (c *CallSession) SetParams(callID, nodeID, to, from, signalwireContext string, direction CallDirection) {
	c.Lock()
	c.CallID = callID
	c.NodeID = nodeID
	c.Direction = direction
	c.To = to
	c.From = from
	c.Context = signalwireContext
	c.Unlock()
}

// SetFrom TODO DESCRIPTION
func (c *CallSession) SetFrom(from string) {
	c.Lock()
	c.From = from
	c.Unlock()
}

// GetFrom TODO DESCRIPTION
func (c *CallSession) GetFrom() string {
	c.RLock()
	from := c.From
	c.RUnlock()

	return from
}

// SetTo TODO DESCRIPTION
func (c *CallSession) SetTo(to string) {
	c.Lock()
	c.To = to
	c.Unlock()
}

// GetTo TODO DESCRIPTION
func (c *CallSession) GetTo() string {
	c.RLock()
	to := c.To
	c.RUnlock()

	return to
}

// SetActive TODO DESCRIPTION
func (c *CallSession) SetActive(active bool) {
	c.Lock()
	c.Active = active
	c.Unlock()
}

// GetActive TODO DESCRIPTION
func (c *CallSession) GetActive() bool {
	c.RLock()
	a := c.Active
	c.RUnlock()

	return a
}

// SetTimeout TODO DESCRIPTION
func (c *CallSession) SetTimeout(t uint) {
	c.Lock()
	c.Timeout = t
	c.Unlock()
}

// GetTimeout TODO DESCRIPTION
func (c *CallSession) GetTimeout() uint {
	c.RLock()
	t := c.Timeout
	c.RUnlock()

	return t
}

// GetType TODO DESCRIPTION
func (c *CallSession) GetType() string {
	c.RLock()
	t := c.CallType.String()
	c.RUnlock()

	return t
}

// SetType TODO DESCRIPTION
func (c *CallSession) SetType(t CallType) {
	c.Lock()
	c.CallType = t
	c.Unlock()
}

// SetDisconnectReason TODO DESCRIPTION
func (c *CallSession) SetDisconnectReason(reason CallDisconnectReason) {
	c.Lock()
	c.CallDisconnectReason = reason
	c.Unlock()
}

// UpdateCallConnectState TODO DESCRIPTION
func (c *CallSession) UpdateCallConnectState(s CallConnectState) {
	Log.Debug("[%p] [%v]\n", c, s)

	c.CallConnectState = s
}

// UpdateConnectPeer TODO DESCRIPTION
func (c *CallSession) UpdateConnectPeer(p PeerDeviceStruct) {
	Log.Debug("[%p] [%v]\n", c, p)

	c.CallPeer.CallID = p.CallID
	c.CallPeer.NodeID = p.NodeID
	c.CallPeer.Device.Type = p.Device.Type
	c.CallPeer.Device.Params.ToNumber = p.Device.Params.ToNumber
	c.CallPeer.Device.Params.FromNumber = p.Device.Params.FromNumber
}

// GetState TODO DESCRIPTION
func (c *CallSession) GetState() CallState {
	c.RLock()
	s := c.CallState
	c.RUnlock()

	return s
}

// GetPrevState TODO DESCRIPTION
func (c *CallSession) GetPrevState() CallState {
	c.RLock()
	s := c.PrevCallState
	c.RUnlock()

	return s
}

// GetCallID TODO DESCRIPTION
func (c *CallSession) GetCallID() string {
	c.RLock()
	s := c.CallID
	c.RUnlock()

	return s
}

// GetTagID TODO DESCRIPTION
func (c *CallSession) GetTagID() string {
	c.RLock()
	s := c.TagID
	c.RUnlock()

	return s
}

// Actions TODO DESCRIPTION
type Actions struct {
	sync.RWMutex
	m map[string]string
}

// CallParams TODO DESCRIPTION (internal, to pass it around)
type CallParams struct {
	TagID      string
	CallID     string
	NodeID     string
	Direction  string
	ToNumber   string
	FromNumber string
	CallState  CallState
	EndReason  string
	Context    string // inbound
}

// ITagToCallID TODO DESCRIPTION
type ITagToCallID interface {
	addCallID(tag, callID string)
	getCallID(tag string) string
	removeCallID(tag string)
}

func (p *CallTagToCallID) addCallID(tag, callID string) {
	if p == nil {
		return
	}

	p.Lock()

	p.m[tag] = callID

	p.Unlock()
}

func (p *CallTagToCallID) getCallID(tag string) string {
	if p == nil {
		return ""
	}

	p.RLock()

	callID := p.m[tag]

	p.RUnlock()

	return callID
}

func (p *CallTagToCallID) removeCallID(tag string) {
	if p == nil {
		return
	}

	p.Lock()
	delete(p.m, tag)
	p.Unlock()
}

// AddAction TODO DESCRIPTION
func (c *CallSession) AddAction(ctrlID, state string) {
	if c == nil {
		return
	}

	c.Actions.Lock()

	c.Actions.m[ctrlID] = state

	c.Actions.Unlock()
}

// GetActionState TODO DESCRIPTION
func (c *CallSession) GetActionState(ctrlID string) string {
	if c == nil {
		return ""
	}

	c.Actions.RLock()

	state := c.Actions.m[ctrlID]

	c.Actions.RUnlock()

	return state
}

// RemoveAction TODO DESCRIPTION
func (c *CallSession) RemoveAction(ctrlID string) {
	if c == nil {
		return
	}

	c.Actions.Lock()
	delete(c.Actions.m, ctrlID)
	c.Actions.Unlock()
}

// UpdateAction TODO DESCRIPTION
func (c *CallSession) UpdateAction(ctrlID, state string) {
	if c == nil {
		return
	}

	c.Actions.Lock()

	c.Actions.m[ctrlID] = state

	c.Actions.Unlock()
}

// CallNew TODO DESCRIPTION
func CallNew() *CallSession {
	return &CallSession{}
}
