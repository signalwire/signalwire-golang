package signalwire

import (
	"context"
	"encoding/json"
)

// RelaySession TODO DESCRIPTION
type RelaySession struct {
	Blade *BladeSession
	I     IRelay
}

// IRelay TODO DESCRIPTION
type IRelay interface {
	/*calling*/
	RelayPhoneDial(ctx context.Context, call *CallSession, fromNumber string, toNumber string, timeout uint, payload **json.RawMessage) error
	RelayPhoneConnect(ctx context.Context, call *CallSession, fromNumber string, toNumber string, payload **json.RawMessage) error
	RelayCallEnd(ctx context.Context, call *CallSession, payload **json.RawMessage) error
	RelayStop(ctx context.Context) error
	RelayOnInboundAnswer(ctx context.Context) (*CallSession, error)
	RelayPlayAudio(ctx context.Context, call *CallSession, ctrlID string, url string, payload **json.RawMessage) error
	RelayRecordAudio(ctx context.Context, call *CallSession, ctrlID string, rec *RecordParams, payload **json.RawMessage) error
	RelayRecordAudioStop(ctx context.Context, call *CallSession, ctrlID *string, payload **json.RawMessage) error
	RelayConnect(ctx context.Context, call *CallSession, ringback *[]RingbackStruct, devices *[][]DeviceStruct, payload **json.RawMessage) error
	RelayCallAnswer(ctx context.Context, call *CallSession, payload **json.RawMessage) error
	RelayPlayTTS(ctx context.Context, call *CallSession, ctrlID string, tts *TTSParamsInternal, payload **json.RawMessage) error
	RelayPlayRingtone(ctx context.Context, call *CallSession, ctrlID string, name string, duration float64, payload **json.RawMessage) error
	RelayPlaySilence(ctx context.Context, call *CallSession, ctrlID string, duration float64, payload **json.RawMessage) error
	RelayPlay(ctx context.Context, call *CallSession, controlID string, play []PlayStruct, payload **json.RawMessage) error
	RelayPlayVolume(ctx context.Context, call *CallSession, ctrlID *string, vol float64, payload **json.RawMessage) error
	RelayPlayResume(ctx context.Context, call *CallSession, ctrlID *string, payload **json.RawMessage) error
	RelayPlayPause(ctx context.Context, call *CallSession, ctrlID *string, payload **json.RawMessage) error
	RelayPlayStop(ctx context.Context, call *CallSession, ctrlID *string, payload **json.RawMessage) error
	RelayDetectDigit(ctx context.Context, call *CallSession, controlID string, digits string, timeout float64, payload **json.RawMessage) error
	RelayDetectFax(ctx context.Context, call *CallSession, controlID string, faxtone string, timeout float64, payload **json.RawMessage) error
	RelayDetectMachine(ctx context.Context, call *CallSession, controlID string, det *DetectMachineParamsInternal, timeout float64, payload **json.RawMessage) error
	RelayDetect(ctx context.Context, call *CallSession, controlID string, detect DetectStruct, timeout float64, payload **json.RawMessage) error
	RelayDetectStop(ctx context.Context, call *CallSession, ctrlID *string, payload **json.RawMessage) error
	RelaySendFax(ctx context.Context, call *CallSession, ctrlID *string, fax *FaxParamsInternal, payload **json.RawMessage) error
	RelayReceiveFax(ctx context.Context, call *CallSession, ctrlID *string, payload **json.RawMessage) error
	RelaySendFaxStop(ctx context.Context, call *CallSession, ctrlID *string, payload **json.RawMessage) error
	RelayReceiveFaxStop(ctx context.Context, call *CallSession, ctrlID *string, payload **json.RawMessage) error
	RelayTapAudio(ctx context.Context, call *CallSession, ctrlID, direction string, device *TapDevice, payload **json.RawMessage) (TapDevice, error)
	RelayTap(ctx context.Context, call *CallSession, controlID string, tap TapStruct, device *TapDevice, payload **json.RawMessage) (TapDevice, error)
	RelayTapStop(ctx context.Context, call *CallSession, ctrlID *string, payload **json.RawMessage) error
	RelaySendDigits(ctx context.Context, call *CallSession, controlID, digits string, payload **json.RawMessage) error
	RelayPlayAndCollect(ctx context.Context, call *CallSession, controlID string, playlist *[]PlayStruct, collect *CollectStruct, payload **json.RawMessage) error
	RelayPlayAndCollectVolume(ctx context.Context, call *CallSession, ctrlID *string, vol float64, payload **json.RawMessage) error
	RelayPlayAndCollectStop(ctx context.Context, call *CallSession, ctrlID *string, payload **json.RawMessage) error
	/*messaging*/
	RelaySendMessage(ctx context.Context, msg *MsgSession, fromNumber, toNumber, signalwireContext, msgBody string) (string, error)
	/*tasking*/
	RelayTaskDeliver(context.Context, string, string, string, string, []byte) error
}

// RelayNew TODO DESCRIPTION
func RelayNew() *RelaySession {
	return &RelaySession{}
}
