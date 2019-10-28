package signalwire

import (
	"context"
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
	RelayConnect(ctx context.Context, call *CallSession, ringback *[]RingbackStruct, devices *[][]DeviceStruct) error
	RelayCallAnswer(ctx context.Context, call *CallSession) error
	RelayPlayTTS(ctx context.Context, call *CallSession, ctrlID string, text, language, gender string) error
	RelayPlayRingtone(ctx context.Context, call *CallSession, ctrlID string, name string, duration float64) error
	RelayPlaySilence(ctx context.Context, call *CallSession, ctrlID string, duration float64) error
	RelayPlay(ctx context.Context, call *CallSession, controlID string, play []PlayStruct) error
	RelayPlayVolume(ctx context.Context, call *CallSession, ctrlID *string, vol float64) error
	RelayPlayResume(ctx context.Context, call *CallSession, ctrlID *string) error
	RelayPlayPause(ctx context.Context, call *CallSession, ctrlID *string) error
	RelayPlayStop(ctx context.Context, call *CallSession, ctrlID *string) error
	RelayDetectDigit(ctx context.Context, call *CallSession, controlID string, digits string) error
	RelayDetectFax(ctx context.Context, call *CallSession, controlID string, faxtone string) error
	RelayDetectMachine(ctx context.Context, call *CallSession, controlID string, det *DetectMachineParams) error
	RelayDetect(ctx context.Context, call *CallSession, controlID string, detect DetectStruct) error
	RelayDetectStop(ctx context.Context, call *CallSession, ctrlID *string) error
	RelaySendFax(ctx context.Context, call *CallSession, ctrlID *string, doc, id, headerInfo string) error
	RelayReceiveFax(ctx context.Context, call *CallSession, ctrlID *string) error
	RelaySendFaxStop(ctx context.Context, call *CallSession, ctrlID *string) error
	RelayReceiveFaxStop(ctx context.Context, call *CallSession, ctrlID *string) error
	RelayTapAudio(ctx context.Context, call *CallSession, ctrlID, direction string, device *TapDevice) (TapDevice, error)
	RelayTap(ctx context.Context, call *CallSession, controlID string, tap TapStruct, device *TapDevice) (TapDevice, error)
	RelayTapStop(ctx context.Context, call *CallSession, ctrlID *string) error
	RelaySendDigits(ctx context.Context, call *CallSession, controlID, digits string) error
	RelayPlayAndCollect(ctx context.Context, call *CallSession, controlID string, playlist *[]PlayStruct, collect *CollectStruct) error
	RelayPlayAndCollectVolume(ctx context.Context, call *CallSession, ctrlID *string, vol float64) error
	RelayPlayAndCollectStop(ctx context.Context, call *CallSession, ctrlID *string) error
}
