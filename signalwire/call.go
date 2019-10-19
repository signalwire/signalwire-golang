package signalwire

import (
	"context"
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

// CallSession TODO DESCRIPTION
type CallSession struct {
	Active               bool
	To                   string
	From                 string
	TagID                string
	Timeout              uint // ring timeout
	CallID               string
	NodeID               string
	ProjectID            string
	SpaceID              string
	Direction            string
	CallState            CallState
	PrevCallState        CallState
	CallDisconnectReason CallDisconnectReason
	CallConnectState     CallConnectState
	CallStateChan        chan CallState
	CallConnectStateChan chan CallConnectState

	CallPlayChans      map[string](chan PlayState)
	CallPlayControlIDs chan string
	CallPlayEventChans map[string](chan ParamsEventCallingCallPlay)
	CallPlayReadyChans map[string](chan struct{})

	CallRecordChans      map[string](chan RecordState)
	CallRecordControlIDs chan string
	CallRecordEventChans map[string](chan ParamsEventCallingCallRecord)
	CallRecordReadyChans map[string](chan struct{})

	CallDetectMachineChans     map[string](chan DetectMachineEvent)
	CallDetectDigitChans       map[string](chan DetectDigitEvent)
	CallDetectFaxChans         map[string](chan DetectFaxEvent)
	CallDetectMachineControlID chan string
	CallDetectDigitControlID   chan string
	CallDetectFaxControlID     chan string
	CallDetectEventChans       map[string](chan ParamsEventCallingCallDetect)
	CallDetectReadyChans       map[string](chan struct{})

	CallFaxChan      chan FaxEventType
	CallFaxControlID chan string
	CallFaxEventChan chan FaxEventStruct
	CallFaxReadyChan chan struct{}

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
	WaitCallState(ctx context.Context, want CallState) bool
	WaitCallConnectState(ctx context.Context, want CallConnectState) bool
	WaitPlayState(ctx context.Context, ctrlID string, want PlayState) bool
	WaitRecordState(ctx context.Context, ctrlID string, want RecordState) bool
	GetPeer(ctx context.Context)
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

	c.CallPlayChans = make(map[string](chan PlayState))
	c.CallPlayControlIDs = make(chan string, SimActionsOfTheSameKind)
	c.CallPlayEventChans = make(map[string](chan ParamsEventCallingCallPlay))
	c.CallPlayReadyChans = make(map[string](chan struct{}))

	c.CallRecordChans = make(map[string](chan RecordState))
	c.CallRecordControlIDs = make(chan string, SimActionsOfTheSameKind)
	c.CallRecordEventChans = make(map[string](chan ParamsEventCallingCallRecord))
	c.CallRecordReadyChans = make(map[string](chan struct{}))

	c.CallDetectMachineControlID = make(chan string, 1)
	c.CallDetectDigitControlID = make(chan string, 1)
	c.CallDetectFaxControlID = make(chan string, 1)

	c.CallDetectMachineChans = make(map[string](chan DetectMachineEvent))
	c.CallDetectDigitChans = make(map[string](chan DetectDigitEvent))
	c.CallDetectFaxChans = make(map[string](chan DetectFaxEvent))

	c.CallFaxChan = make(chan FaxEventType, EventQueue)
	c.CallFaxControlID = make(chan string, 1)
	c.CallFaxReadyChan = make(chan struct{})
	c.CallFaxEventChan = make(chan FaxEventStruct)

	c.Hangup = make(chan struct{})
	c.Actions.m = make(map[string]string)
}

// CallCleanup close the channels, etc
func (c *CallSession) CallCleanup(_ context.Context) {
	close(c.CallStateChan)
	close(c.CallConnectStateChan)
	close(c.Hangup)

	c.Actions.m = nil

	c = nil /*let the garbage collector clean it up */
}

// WaitCallStateInternal wait for a certain call state and return true when it arrives.
// return false if timeout.
func (c *CallSession) WaitCallStateInternal(_ context.Context, want CallState) bool {
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
				// signal all Action go routines to finish
				c.Hangup <- struct{}{}
			}
		case <-time.After(BroadcastEventTimeout * time.Second):
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

// WaitPlayState TODO DESCRIPTION
func (c *CallSession) WaitPlayState(_ context.Context, ctrlID string, want PlayState) bool {
	var ret bool

	for {
		var out bool

		select {
		case playstate := <-c.CallPlayChans[ctrlID]:
			if playstate == want {
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

// UpdatePlayState Never timeout while waiting for 'finished'
func (c *CallSession) UpdatePlayState(_ context.Context, ctrlID string, res *PlayAction) bool {
	var ret bool

	for {
		var out bool

		select {
		case playstate := <-c.CallPlayChans[ctrlID]:
			if playstate == PlayFinished {
				out = true
				ret = true
			}

			res.State = playstate
		case <-c.Hangup:
			out = true
		}

		if out {
			break
		}
	}

	return ret
}

// WaitRecordState TODO DESCRIPTION
func (c *CallSession) WaitRecordState(_ context.Context, ctrlID string, want RecordState) bool {
	var ret bool

	for {
		var out bool

		select {
		case recordstate := <-c.CallRecordChans[ctrlID]:
			if recordstate == want {
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

// UpdateRecordState Never timeout while waiting for 'finished'
func (c *CallSession) UpdateRecordState(_ context.Context, ctrlID string, res *RecordAction) bool {
	var ret bool

	for {
		var out bool

		select {
		case state := <-c.CallRecordChans[ctrlID]:
			if state == RecordFinished {
				out = true
				ret = true
			}

			res.State = state
		case <-c.Hangup:
			out = true
		}

		if out {
			break
		}
	}

	return ret
}

// WaitRecordStateFinished - wait for 'Finished', don't timeout.
// return on Hangup if 'Finished' did not arrive
func (c *CallSession) WaitRecordStateFinished(_ context.Context, ctrlID string) bool {
	var ret bool

	for {
		var out bool

		select {
		case recordstate := <-c.CallRecordChans[ctrlID]:
			if recordstate == RecordFinished {
				out = true
				ret = true
			}
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
	c.CallState = s
}

// SetParams setting Params that stay the same during the call*/
func (c *CallSession) SetParams(callID, nodeID, direction, to, from string) {
	c.Lock()
	c.CallID = callID
	c.NodeID = nodeID
	c.Direction = direction
	c.To = to
	c.From = from
	c.Unlock()
}

// SetFrom TODO DESCRIPTION
func (c *CallSession) SetFrom(from string) {
	c.Lock()
	c.From = from
	c.Unlock()
}

// SetTo TODO DESCRIPTION
func (c *CallSession) SetTo(to string) {
	c.Lock()
	c.To = to
	c.Unlock()
}

// SetAction TODO DESCRIPTION
func (c *CallSession) SetActive(active bool) {
	c.Lock()
	c.Active = active
	c.Unlock()
}

// SetAction TODO DESCRIPTION
func (c *CallSession) GetActive() bool {
	c.RLock()
	a := c.Active
	c.RUnlock()

	return a
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

// Actions TODO DESCRIPTION
type Actions struct {
	sync.RWMutex
	m map[string]string
}

// CallParams TODO DESCRIPTION
type CallParams struct {
	TagID      string
	CallID     string
	NodeID     string
	Direction  string
	ToNumber   string
	FromNumber string
	CallState  CallState
	EndReason  string
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
