package signalwire

import (
	"context"
	"sync"
	"time"
)

// Messaging TODO DESCRIPTION
type Messaging struct {
	Ctx    context.Context
	Cancel context.CancelFunc
	Relay  *RelaySession
}

// MsgObj is the external Message object (as exposed to the user)
type MsgObj struct {
	msg                  *MsgSession
	I                    IMsgObj
	Messaging            *Messaging
	OnMessageStateChange func(*SendResult)
	OnMessageQueued      func(*SendResult)
	OnMessageInitiated   func(*SendResult)
	OnMessageSent        func(*SendResult)
	OnMessageDelivered   func(*SendResult)
	OnMessageUndelivered func(*SendResult)
	OnMessageFailed      func(*SendResult)
}

// IMsgObj TODO DESCRIPTION
type IMsgObj interface {
	SetMedia(media []string)
	SetTags(tags []string)
	SetRegion(params string)
	GetFrom() string
	GetTo() string
	GetBody() string
}

// IMessaging object visible to the end user
type IMessaging interface {
	Send(signalwireContext, fromNumber, toNumber, text string) *SendResult
	NewMessage() *MsgObj
	SendMsg(m *MsgObj) *SendResult
}

// SendResult TODO DESCRIPTION
type SendResult struct {
	I     IMessaging
	Msg   *MsgObj
	state MsgState
	err   error
	sync.RWMutex
	Successful bool
	Completed  bool
}

// MsgObjNew TODO DESCRIPTION
func MsgObjNew() *MsgObj {
	return &MsgObj{}
}

// NewMessage  TODO DESCRIPTION
func (messaging *Messaging) NewMessage(signalwireContext, from, to, text string) *MsgObj {
	newmsg := new(MsgSession)

	var I IMsgObj = MsgObjNew()

	m := &MsgObj{I: I}
	m.msg = newmsg
	m.Messaging = messaging
	m.msg.SetFrom(from)
	m.msg.SetTo(to)
	m.msg.SetContext(signalwireContext)
	m.msg.SetBody(text)

	return m
}

func (msgobj *MsgObj) callbacksRunSend(_ context.Context, res *SendResult) {
	var out bool

	timer := time.NewTimer(BroadcastEventTimeout * time.Second)

	for {
		select {
		case state := <-msgobj.msg.MsgStateChan:
			res.RLock()

			prevstate := res.state

			res.RUnlock()

			switch state {
			case MsgDelivered:
				res.Lock()

				res.state = state
				res.Successful = true
				res.Completed = true

				res.Unlock()

				out = true

				if msgobj.OnMessageDelivered != nil {
					go msgobj.OnMessageDelivered(res)
				}

			case MsgSent:
				res.Lock()

				res.state = state

				res.Unlock()
				timer.Reset(BroadcastEventTimeout * time.Second)

				if msgobj.OnMessageSent != nil {
					go msgobj.OnMessageSent(res)
				}
			case MsgUndelivered:
				res.Lock()

				res.Completed = true
				res.state = state

				res.Unlock()

				out = true

				if msgobj.OnMessageUndelivered != nil {
					go msgobj.OnMessageUndelivered(res)
				}
			case MsgFailed:
				res.Lock()

				res.state = state

				res.Unlock()

				out = true

				if msgobj.OnMessageFailed != nil {
					go msgobj.OnMessageFailed(res)
				}
			case MsgQueued:
				res.Lock()

				res.state = state

				res.Unlock()
				timer.Reset(BroadcastEventTimeout * time.Second)

				if msgobj.OnMessageQueued != nil {
					go msgobj.OnMessageQueued(res)
				}
			case MsgInitiated:
				res.Lock()

				res.state = state

				res.Unlock()
				timer.Reset(BroadcastEventTimeout * time.Second)

				if msgobj.OnMessageInitiated != nil {
					go msgobj.OnMessageInitiated(res)
				}
			default:
				Log.Debug("Unknown state.")
			}

			if prevstate != state && msgobj.OnMessageStateChange != nil {
				go msgobj.OnMessageStateChange(res)
			}
		case <-timer.C:
			out = true
		}

		if out {
			break
		}
	}
}

// Send TODO DESCRIPTION
func (messaging *Messaging) Send(fromNumber, toNumber, signalwireContext, msgBody string) *SendResult {
	res := new(SendResult)

	if messaging.Relay == nil {
		return res
	}

	if messaging.Ctx == nil {
		return res
	}

	newmsg := new(MsgSession)
	newmsg.MsgInit(messaging.Ctx)

	var err error

	var msgID string

	msgID, err = messaging.Relay.RelaySendMessage(messaging.Ctx, newmsg, fromNumber, toNumber, signalwireContext, msgBody)
	if err != nil {
		Log.Error("RelaySendMessage: %v", err)
		res.err = err

		return res
	}

	newmsg.SetParams(msgID, fromNumber, toNumber, signalwireContext, MsgOutbound)

	var I IMsgObj = MsgObjNew()

	m := &MsgObj{I: I}
	m.msg = newmsg
	m.Messaging = messaging

	res.Msg = m

	/*no callbacks*/
	res.Msg.callbacksRunSend(messaging.Ctx, res)

	return res
}

// SendMsg TODO DESCRIPTION
func (messaging *Messaging) SendMsg(mObj *MsgObj) *SendResult {
	res := new(SendResult)

	if messaging.Relay == nil {
		return res
	}

	if messaging.Ctx == nil {
		return res
	}

	mObj.msg.MsgInit(messaging.Ctx)

	var err error

	var msgID string

	done := make(chan struct{})

	go func() {
		mObj.callbacksRunSend(messaging.Ctx, res)

		done <- struct{}{}
	}()

	msgID, err = messaging.Relay.RelaySendMessage(messaging.Ctx, mObj.msg, mObj.msg.MsgParams.From, mObj.msg.MsgParams.To, mObj.msg.MsgParams.Context, mObj.msg.MsgParams.Body)
	if err != nil {
		Log.Error("RelaySendMessage: %v", err)
		res.err = err

		return res
	}

	mObj.msg.SetMsgID(msgID)
	res.Msg = mObj
	res.Successful = true

	<-done

	return res
}

// GetSuccessful TODO DESCRIPTION
func (resultSend *SendResult) GetSuccessful() bool {
	return resultSend.Successful
}

// GetReason TODO DESCRIPTION
func (resultSend *SendResult) GetReason() string {
	return resultSend.Msg.msg.GetFailureReason()
}

// GetMsgID TODO DESCRIPTION
func (resultSend *SendResult) GetMsgID() string {
	return resultSend.Msg.msg.GetMsgID()
}

// SetMedia TODO DESCRIPTION
func (msgobj *MsgObj) SetMedia(media []string) {
	msgobj.msg.SetMedia(media)
}

// SetTags TODO DESCRIPTION
func (msgobj *MsgObj) SetTags(tags []string) {
	msgobj.msg.SetTags(tags)
}

// SetRegion TODO DESCRIPTION
func (msgobj *MsgObj) SetRegion(region string) {
	msgobj.msg.SetRegion(region)
}

// GetFrom TODO DESCRIPTION
func (msgobj *MsgObj) GetFrom() string {
	return msgobj.msg.GetFrom()
}

// GetTo TODO DESCRIPTION
func (msgobj *MsgObj) GetTo() string {
	return msgobj.msg.GetTo()
}

// GetBody TODO DESCRIPTION
func (msgobj *MsgObj) GetBody() string {
	return msgobj.msg.GetBody()
}
