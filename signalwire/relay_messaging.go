package signalwire

import (
	"context"
	"errors"
)

// RelaySendMessage send a text message (sms)
func (relay *RelaySession) RelaySendMessage(ctx context.Context, fromNumber, toNumber, context, msgBody string) (string, error) {
	var err error

	if relay == nil {
		return "", errors.New("empty relay object")
	}

	if relay.Blade == nil {
		return "", errors.New("blade server object not defined")
	}

	msg := new(MsgSession)
	msg.MsgInit(ctx)

	v := ParamsBladeExecuteStruct{
		Protocol: relay.Blade.Protocol,
		Method:   "messaging.send",
		Params: ParamsMessagingSend{
			ToNumber:   toNumber,
			FromNumber: fromNumber,
			Context:    context,
			Body:       msgBody,
		},
	}

	var ReplyBladeExecuteDecode ReplyBladeExecuteSendMsg

	reply, err := relay.Blade.I.BladeExecute(ctx, &v, &ReplyBladeExecuteDecode)
	if err != nil {
		return "", err
	}

	r, ok := reply.(*ReplyBladeExecuteSendMsg)
	if !ok {
		return "", errors.New("type assertion failed")
	}

	Log.Debug("reply ReplyBladeExecuteDecode: %v\n", r)

	if r.Result.Code != okCode {
		return "", errors.New(r.Result.Message)
	}

	if len(r.Result.MsgID) == 0 {
		return "", errors.New("invalid message ID")
	}

	if err := relay.Blade.EventMessaging.Cache.SetMsgCache(r.Result.MsgID, msg); err != nil {
		return "", errors.New("cannot cache msg")
	}

	return r.Result.MsgID, nil
}
