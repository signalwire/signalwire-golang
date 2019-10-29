package signalwire

import (
	"context"
	"errors"
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
	OnMessageStateChange func(*MsgObj)
	OnMessageQueued      func(*MsgObj)
	OnInitiated          func(*MsgObj)
	OnMessageSent        func(*MsgObj)
	OnMessageDelivered   func(*MsgObj)
	OnMessageUndelivered func(*MsgObj)
	OnMessageFailed      func(*MsgObj)
}

type IMsgObj interface {
}

// IMessaging object visible to the end user
type IMessaging interface {
	Send(context, fromNumber, toNumber, text string) SendResult
	NewMessage() *MsgObj
	SendMsg(m *MsgObj) SendResult
}

// SendResult TODO DESCRIPTION
type SendResult struct {
	Successful bool
	Msg        *MsgObj
	I          IMessaging
	err        error
}

// MsgObjNew TODO DESCRIPTION
func MsgObjNew() *MsgObj {
	return &MsgObj{}
}

// NewMessage  TODO DESCRIPTION
func (messaging *Messaging) NewMessage(context, from, to, text string) *MsgObj {
	newmsg := new(MsgSession)

	var I IMsgObj = MsgObjNew()

	m := &MsgObj{I: I}
	m.msg = newmsg
	m.Messaging = messaging
	m.msg.SetFrom(from)
	m.msg.SetTo(to)
	m.msg.SetContext(context)
	m.msg.SetBody(text)

	return m
}

// Send TODO DESCRIPTION
func (messaging *Messaging) Send(fromNumber, toNumber, signalwireContext, msgBody string) SendResult {
	res := new(SendResult)

	if messaging.Relay == nil {
		return *res
	}

	if messaging.Ctx == nil {
		return *res
	}

	newmsg := new(MsgSession)
	newmsg.MsgInit(messaging.Ctx)

	var err error

	var msgID string

	msgID, err = messaging.Relay.RelaySendMessage(messaging.Ctx, newmsg, fromNumber, toNumber, signalwireContext, msgBody)
	if err != nil {
		Log.Error("RelaySendMessage: %v", err)
		res.err = err

		return *res
	}

	newmsg.SetParams(msgID, fromNumber, toNumber, signalwireContext, MsgOutbound)

	var I IMsgObj = MsgObjNew()

	m := &MsgObj{I: I}
	m.msg = newmsg
	m.Messaging = messaging

	if ret := newmsg.WaitMsgStateInternal(messaging.Ctx, MsgDelivered); !ret {
		Log.Debug("did not get Delivered state\n")

		res.Msg = m

		return *res
	}

	res.Msg = m
	res.Successful = true

	return *res
}

// SendMsg TODO DESCRIPTION
func (messaging *Messaging) SendMsg(mObj *MsgObj) SendResult {
	res := new(SendResult)

	if messaging.Relay == nil {
		return *res
	}

	if messaging.Ctx == nil {
		return *res
	}

	mObj.msg.MsgInit(messaging.Ctx)

	var err error

	var msgID string

	done := make(chan struct{})

	go func() {
		if ret := mObj.msg.WaitMsgStateInternal(messaging.Ctx, MsgDelivered); !ret {
			res.err = errors.New("did not get Delivered state")
			Log.Error("%v\n", res.err)

			res.Msg = mObj

			res.Successful = false
		}

		done <- struct{}{}
	}()

	msgID, err = messaging.Relay.RelaySendMessage(messaging.Ctx, mObj.msg, mObj.msg.MsgParams.From, mObj.msg.MsgParams.To, mObj.msg.MsgParams.Context, mObj.msg.MsgParams.Body)
	if err != nil {
		Log.Error("RelaySendMessage: %v", err)
		res.err = err

		return *res
	}

	mObj.msg.SetMsgID(msgID)
	res.Msg = mObj
	res.Successful = true

	<-done

	return *res
}

// GetSuccessful TODO DESCRIPTION
func (resultSend *SendResult) GetSuccessful() bool {
	return resultSend.Successful
}

// GetReason TODO DESCRIPTION
func (resultSend *SendResult) GetReason() string {
	return resultSend.Msg.msg.GetFailureReason()
}
