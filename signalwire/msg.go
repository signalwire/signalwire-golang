package signalwire

import (
	"context"
	"sync"
	"time"
)

// MsgState keeps the state of a text message (sms)
type MsgState int

// Msg state constants
const (
	MsgQueued MsgState = iota
	MsgInitiated
	MsgSent
	MsgDelivered
	MsgUndelivered
	MsgFailed
	MsgReceived
)

func (s MsgState) String() string {
	return [...]string{"Queued", "Initiated", "Sent", "Delivered", "Undelivered", "Failed", "Received"}[s]
}

// Message Direction (in or out)
type MsgDirection int

// Call state constants
const (
	MsgInbound MsgDirection = iota
	MsgOutbound
)

func (s MsgDirection) String() string {
	return [...]string{"Inbound", "Outbound"}[s]
}

// MsgSession internal representation of a text message
type MsgSession struct {
	MsgParams MsgParams
	Blade     *BladeSession
	sync.RWMutex
	I            IMsg
	ReplyMsgID   chan string
	MsgStateChan chan MsgState
}

type MsgParams struct {
	MsgID     string
	Context   string
	NodeID    string
	Direction MsgDirection
	To        string
	From      string
	MsgState  MsgState
	Segments  uint     // number of segments of the message
	Body      string   // body of the message
	Tags      []string // optional client data this call is tagged with
	Media     []string // an array of URLs to send in the message
	Reason    string   // reason: present on undelivered and failed.
}

// IMsg Msg Interface
type IMsg interface {
	MsgInit(ctx context.Context)
	MsgCleanup(ctx context.Context)
	UpdateCallState(s CallState)
	WaitMsgState(ctx context.Context, want MsgState) bool
}

// MsgInit creates the channels communicating msg states
func (m *MsgSession) MsgInit(_ context.Context) {
	m.MsgStateChan = make(chan MsgState, EventQueue)
	m.ReplyMsgID = make(chan string)
}

// MsgCleanup close the channels, etc
func (m *MsgSession) MsgCleanup(_ context.Context) {
	close(m.MsgStateChan)
	close(m.ReplyMsgID)

	m = nil /*let the garbage collector clean it up */
}

// UpdateMsgState TODO DESCRIPTION
func (m *MsgSession) UpdateMsgState(s MsgState) {
	m.Lock()
	m.MsgParams.MsgState = s
	m.Unlock()
}

// SetParams TODO DESCRIPTION
func (m *MsgSession) SetParams(msgID, to, from, signalwireContext string, direction MsgDirection) {
	m.Lock()
	m.MsgParams.MsgID = msgID
	m.MsgParams.Direction = direction
	m.MsgParams.To = to
	m.MsgParams.From = from
	m.MsgParams.Context = signalwireContext
	m.Unlock()
}

// SetFrom TODO DESCRIPTION
func (m *MsgSession) SetFrom(from string) {
	m.Lock()
	m.MsgParams.From = from
	m.Unlock()
}

// GetFrom TODO DESCRIPTION
func (m *MsgSession) GetFrom() string {
	m.RLock()
	from := m.MsgParams.From
	m.RUnlock()

	return from
}

// SetTo TODO DESCRIPTION
func (m *MsgSession) SetTo(to string) {
	m.Lock()
	m.MsgParams.To = to
	m.Unlock()
}

// GetTo TODO DESCRIPTION
func (m *MsgSession) GetTo() string {
	m.RLock()
	to := m.MsgParams.To
	m.RUnlock()

	return to
}

func (m *MsgSession) waitMsgStateInternal(_ context.Context, want MsgState) bool {
	var ret bool

	var out bool

	for {
		select {
		case state := <-m.MsgStateChan:
			switch state {
			case want:
				out = true
				ret = true
			case MsgUndelivered:
				fallthrough
			case MsgFailed:
				out = true
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

// SetFailureReason TODO DESCRIPTION
func (m *MsgSession) SetFailureReason(reason string) {
	m.Lock()
	m.MsgParams.Reason = reason
	m.Unlock()
}

// GetFailureReason TODO DESCRIPTION
func (m *MsgSession) GetFailureReason() string {
	m.RLock()
	r := m.MsgParams.Reason
	m.RUnlock()

	return r
}

// GetState TODO DESCRIPTION
func (m *MsgSession) GetState() MsgState {
	m.RLock()
	s := m.MsgParams.MsgState
	m.RUnlock()

	return s
}

// GetSegments TODO DESCRIPTION
func (m *MsgSession) GetSegments() uint {
	m.RLock()
	s := m.MsgParams.Segments
	m.RUnlock()

	return s
}

// SetContext TODO DESCRIPTION
func (m *MsgSession) SetContext(c string) {
	m.Lock()
	m.MsgParams.Context = c
	m.Unlock()
}

// SetBody TODO DESCRIPTION
func (m *MsgSession) SetBody(body string) {
	m.Lock()
	m.MsgParams.Body = body
	m.Unlock()
}

// SetMsgID TODO DESCRIPTION
func (m *MsgSession) SetMsgID(msgID string) {
	m.Lock()
	m.MsgParams.MsgID = msgID
	m.Unlock()
}
