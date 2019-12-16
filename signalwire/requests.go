package signalwire

import (
	"encoding/json"
)

// DevicePhoneParams TODO DESCRIPTION
type DevicePhoneParams struct {
	ToNumber   string `json:"to_number"`
	FromNumber string `json:"from_number"`
	Timeout    uint   `json:"timeout"`
}

// DeviceAgoraParams TODO DESCRIPTION
type DeviceAgoraParams struct {
	To      string `json:"to"`
	From    string `json:"from"`
	Appid   string `json:"appid"`
	Channel string `json:"channel"`
}

// DeviceStruct TODO DESCRIPTION
type DeviceStruct struct {
	Type string `json:"type"`
	// todo: make Params interface{}
	Params DevicePhoneParams `json:"params"`
}

// ParamsCallingBeginStruct TODO DESCRIPTION
type ParamsCallingBeginStruct struct {
	Device DeviceStruct `json:"device"`
	Tag    string       `json:"tag"`
}

// ParamsSignalwireReceive TODO DESCRIPTION
type ParamsSignalwireReceive struct {
	Contexts []string `json:"contexts"`
}

// RingbackRingtoneParams TODO DESCRIPTION
type RingbackRingtoneParams PlayRingtoneParams

// RingbackSilenceParams TODO DESCRIPTION
type RingbackSilenceParams PlaySilenceParams

// RingbackTTSParams TODO DESCRIPTION
type RingbackTTSParams PlayTTSParams

// RingbackAudioParams TODO DESCRIPTION
type RingbackAudioParams PlayAudioParams

// RingbackStruct TODO DESCRIPTION
type RingbackStruct struct {
	Type   string      `json:"type"`
	Params interface{} `json:"params"`
}

// ParamsCallConnectStruct TODO DESCRIPTION
type ParamsCallConnectStruct struct {
	Ringback *[]RingbackStruct `json:"ringback,omitempty"`
	Devices  *[][]DeviceStruct `json:"devices"`
	NodeID   string            `json:"node_id"`
	CallID   string            `json:"call_id"`
	Tag      string            `json:"tag,omitempty"`
}

// ParamsCommandStruct TODO DESCRIPTION
type ParamsCommandStruct interface{}

// ParamsBladeExecuteStruct TODO DESCRIPTION
type ParamsBladeExecuteStruct struct {
	Protocol string              `json:"protocol"`
	Method   string              `json:"method"`
	Params   ParamsCommandStruct `json:"params"`
}

// ParamsSubscriptionStruct TODO DESCRIPTION
type ParamsSubscriptionStruct struct {
	Command  string   `json:"command"`
	Protocol string   `json:"protocol"`
	Channels []string `json:"channels"`
}

// ParamsSignalwireSetupStruct TODO DESCRIPTION
type ParamsSignalwireSetupStruct struct {
}

// ParamsSignalwireStruct TODO DESCRIPTION
type ParamsSignalwireStruct struct {
	Params ParamsSignalwireSetupStruct `json:"params"`

	Protocol string `json:"protocol"`
	Method   string `json:"method"`
}

// BladeVersionStruct TODO DESCRIPTION
type BladeVersionStruct struct {
	Major    int `json:"major"`
	Minor    int `json:"minor"`
	Revision int `json:"revision"`
}

// ParamsConnectStruct TODO DESCRIPTION
type ParamsConnectStruct struct {
	Version        BladeVersionStruct `json:"version"`
	SessionID      string             `json:"session_id"`
	Authentication AuthStruct         `json:"authentication"`
	Agent          string             `json:"agent"`
}

// AuthStruct TODO DESCRIPTION
type AuthStruct struct {
	Project string `json:"project"`
	Token   string `json:"token"`
}

// ReqBladeConnect TODO DESCRIPTION
type ReqBladeConnect struct {
	Method string              `json:"method"`
	Params ParamsConnectStruct `json:"params"`
}

// ParamsAuthStruct TODO DESCRIPTION
type ParamsAuthStruct struct {
	RequesterNodeID string `json:"requester_node_id"`
	ResponderNodeID string `json:"responder_node_id"`
	OriginalID      string `json:"original_id"`
	NodeID          string `json:"node_id"`
	ConnectionID    string `json:"connection_id"`
}

// ReqBladeAuthenticate TODO DESCRIPTION
type ReqBladeAuthenticate struct {
	Method string           `json:"method"`
	Params ParamsAuthStruct `json:"params"`
}

// ReqBladeSetup TODO DESCRIPTION
type ReqBladeSetup struct {
	Method string                 `json:"method"`
	Params ParamsSignalwireStruct `json:"params"`
}

// ErrorStruct is RPC error object
type ErrorStruct struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ReplyError TODO DESCRIPTION
type ReplyError struct {
	Error ErrorStruct `json:"error"`
}

// ReplyAuthStruct TODO DESCRIPTION
type ReplyAuthStruct struct {
	Project   string   `json:"project"`
	ExpiresAt string   `json:"expires_at"`
	Scopes    []string `json:"scopes"`
	Signature string   `json:"signature"`
}

// ReplyResultConnect TODO DESCRIPTION
type ReplyResultConnect struct {
	SessionRestored      bool            `json:"session_restored"`
	SessionID            string          `json:"sessionid"`
	NodeID               string          `json:"node_id"`
	MasterNodeID         string          `json:"master_nodeid"`
	Authorization        ReplyAuthStruct `json:"authorization"`
	Routes               []string        `json:"routes"`
	Protocols            []string        `json:"protocols"`
	Subscriptions        []string        `json:"subscriptions"`
	Authorities          []string        `json:"authorities"`
	Authorizations       []string        `json:"authorizations"`
	Accesses             []string        `json:"accesses"`
	ProtocolsUncertified []string        `json:"protocols_uncertified"`
}

// ReplyBladeConnect TODO DESCRIPTION
type ReplyBladeConnect struct {
	Result ReplyResultConnect `json:"result"`
}

// ReplyResultResultSetup TODO DESCRIPTION
type ReplyResultResultSetup struct {
	Protocol string `json:"protocol"`
}

// ReplyResultSetup TODO DESCRIPTION
type ReplyResultSetup struct {
	RequesterNodeID string                 `json:"requester_node_id"`
	ResponderNodeID string                 `json:"responder_node_id"`
	Result          ReplyResultResultSetup `json:"result"`
}

// ReplyResultSubscription TODO DESCRIPTION
type ReplyResultSubscription struct {
	Protocol          string   `json:"protocol"`
	Command           string   `json:"command"`
	SubscribeChannels []string `json:"subscribe_channels"`
}

// ReplyBladeSetup TODO DESCRIPTION
type ReplyBladeSetup struct {
	Result ReplyResultSetup
}

// ReplyBladeExecuteResult TODO DESCRIPTION
type ReplyBladeExecuteResult struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ReplyBladeExecuteResultTap TODO DESCRIPTION
type ReplyBladeExecuteResultTap struct {
	Code         string    `json:"code"`
	Message      string    `json:"message"`
	CallID       string    `json:"call_id"`
	ControlID    string    `json:"control_id"`
	SourceDevice TapDevice `json:"source_device"`
}

// ReplyBladeExecuteResultSendMsg TODO DESCRIPTION
type ReplyBladeExecuteResultSendMsg struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	MsgID   string `json:"message_id"`
}

// ReplyBladeExecute TODO DESCRIPTION
type ReplyBladeExecute struct {
	RequesterNodeID string                  `json:"requester_node_id"`
	ResponderNodeID string                  `json:"responder_node_id"`
	Result          ReplyBladeExecuteResult `json:"result"`
}

// ReplyBladeExecuteTap TODO DESCRIPTION
type ReplyBladeExecuteTap struct {
	RequesterNodeID string                     `json:"requester_node_id"`
	ResponderNodeID string                     `json:"responder_node_id"`
	Result          ReplyBladeExecuteResultTap `json:"result"`
}

// ReplyBladeExecuteSendMsg TODO DESCRIPTION
type ReplyBladeExecuteSendMsg struct {
	RequesterNodeID string                         `json:"requester_node_id"`
	ResponderNodeID string                         `json:"responder_node_id"`
	Result          ReplyBladeExecuteResultSendMsg `json:"result"`
}

// PeerStruct  TODO DESCRIPTION
type PeerStruct struct {
	CallID string `json:"call_id"`
	NodeID string `json:"node_id"`
}

// PeerDeviceStruct  TODO DESCRIPTION
type PeerDeviceStruct struct {
	CallID string       `json:"call_id"`
	NodeID string       `json:"node_id"`
	Device DeviceStruct `json:"device"`
}

// ParamsEventCallingCallConnect  TODO DESCRIPTION
type ParamsEventCallingCallConnect struct {
	ConnectState string           `json:"connect_state"`
	CallID       string           `json:"call_id"`
	NodeID       string           `json:"node_id"`
	TagID        string           `json:"tag"`
	Peer         PeerDeviceStruct `json:"peer"`
}

// ParamsEventCallingCallState TODO DESCRIPTION
type ParamsEventCallingCallState struct {
	CallState string       `json:"call_state"`
	Direction string       `json:"direction"`
	Device    DeviceStruct `json:"device"`
	EndReason string       `json:"end_reason"`
	CallID    string       `json:"call_id"`
	NodeID    string       `json:"node_id"`
	TagID     string       `json:"tag"`
}

// ParamsEventCallingCallReceive TODO DESCRIPTION
type ParamsEventCallingCallReceive struct {
	CallState string       `json:"call_state"`
	Context   string       `json:"context"`
	Direction string       `json:"direction"`
	Device    DeviceStruct `json:"device"`
	CallID    string       `json:"call_id"`
	NodeID    string       `json:"node_id"`
	TagID     string       `json:"tag"`
}

// ParamsGenericAction TODO DESCRIPTION
type ParamsGenericAction struct {
	CallID    string `json:"call_id"`
	NodeID    string `json:"node_id"`
	ControlID string `json:"control_id"`
}

// ParamsEventCallingCallPlay TODO DESCRIPTION
type ParamsEventCallingCallPlay struct {
	PlayState string `json:"state"`
	CallID    string `json:"call_id"`
	NodeID    string `json:"node_id"`
	ControlID string `json:"control_id"`
}

// AudioStruct TODO DESCRIPTION
type AudioStruct struct {
	Format    string `json:"format,omitempty"`
	Direction string `json:"direction,omitempty"`
	Stereo    bool   `json:"stereo,omitempty"`
}

// ParamsRecord TODO DESCRIPTION
type ParamsRecord struct {
	Audio AudioStruct `json:"audio"`
}

// ParamsEventCallingCallRecord TODO DESCRIPTION
type ParamsEventCallingCallRecord struct {
	CallID      string       `json:"call_id"`
	NodeID      string       `json:"node_id"`
	ControlID   string       `json:"control_id"`
	TagID       string       `json:"tag"`
	Params      ParamsRecord `json:"params"`
	RecordState string       `json:"state"`
	Duration    uint         `json:"duration"`
	URL         string       `json:"url"`
	Size        uint         `json:"size"`
}

// ParamsEventDetect TODO DESCRIPTION
type ParamsEventDetect struct {
	Event string `json:"event"`
}

// DetectEventStruct TODO DESCRIPTION
type DetectEventStruct struct {
	Type   string            `json:"type"`
	Params ParamsEventDetect `json:"params"`
}

// ParamsEventCallingCallDetect TODO DESCRIPTION
type ParamsEventCallingCallDetect struct {
	CallID    string            `json:"call_id"`
	NodeID    string            `json:"node_id"`
	ControlID string            `json:"control_id"`
	Detect    DetectEventStruct `json:"detect"`
}

// FaxTypeParamsPage TODO DESCRIPTION
type FaxTypeParamsPage struct {
	Direction string `json:"direction"`
	Number    uint16 `json:"number"`
}

// FaxTypeParamsFinished  TODO DESCRIPTION
type FaxTypeParamsFinished struct {
	Direction      string `json:"direction"`
	Identity       string `json:"identity"`
	RemoteIdentity string `json:"remote_identity"`
	Document       string `json:"document"`
	Pages          uint16 `json:"pages"`
	Success        bool   `json:"success"`
	Result         uint16 `json:"result"`
	ResultText     string `json:"result_text"`
	Format         string `json:"format"`
}

// FaxTypeParamsError TODO DESCRIPTION
type FaxTypeParamsError struct {
	Description string `json:"description"`
}

// FaxEventStruct TODO DESCRIPTION
type FaxEventStruct struct {
	EventType string                 `json:"type"`
	Params    map[string]interface{} `json:"params"`
}

// ParamsEventCallingFax TODO DESCRIPTION
type ParamsEventCallingFax struct {
	CallID    string         `json:"call_id"`
	NodeID    string         `json:"node_id"`
	ControlID string         `json:"control_id"`
	Fax       FaxEventStruct `json:"fax"`
}

// ParamsQueueingRelayEvents TODO DESCRIPTION
type ParamsQueueingRelayEvents struct {
	EventType    string          `json:"event_type,omitempty"`
	EventChannel string          `json:"event_channel,omitempty"`
	Timestamp    float64         `json:"timestamp,omitempty"`
	Project      string          `json:"project_id,omitempty"`
	Space        string          `json:"space_id,omitempty"`
	Context      string          `json:"context,omitempty"`
	Message      json.RawMessage `json:"message,omitempty"`
	Params       interface{}     `json:"params"`
}

// NotifParamsBladeBroadcast TODO DESCRIPTION
type NotifParamsBladeBroadcast struct {
	BroadcasterNodeID string                    `json:"broadcaster_nodeid"`
	Protocol          string                    `json:"protocol"`
	Channel           string                    `json:"channel"`
	Event             string                    `json:"event"`
	Params            ParamsQueueingRelayEvents `json:"params"`
}

// ParamsNetcastEvent TODO DESCRIPTION
type ParamsNetcastEvent struct {
	Protocol string `json:"protocol"`
}

// NotifParamsBladeNetcast TODO DESCRIPTION
type NotifParamsBladeNetcast struct {
	NetcasterNodeID string             `json:"netcaster_nodeid"`
	Command         string             `json:"command"`
	Params          ParamsNetcastEvent `json:"params"`
}

// ParamsCallEndStruct TODO DESCRIPTION
type ParamsCallEndStruct struct {
	CallID string `json:"call_id"`
	NodeID string `json:"node_id"`
	Reason string `json:"reason"`
}

// ParamsDisconnect - empty
type ParamsDisconnect struct{}

// ReplyResultDisconnect - empty
type ReplyResultDisconnect struct{}

// ParamsCallAnswer TODO DESCRIPTION
type ParamsCallAnswer struct {
	CallID string `json:"call_id"`
	NodeID string `json:"node_id"`
}

// PlayAudioParams TODO DESCRIPTION
type PlayAudioParams struct {
	URL string `json:"url"`
}

// PlayTTSParams TODO DESCRIPTION
type PlayTTSParams struct {
	Text     string `json:"text"`
	Language string `json:"language,omitempty"`
	Gender   string `json:"gender,omitempty"`
}

// PlaySilenceParams TODO DESCRIPTION
type PlaySilenceParams struct {
	Duration float64 `json:"duration"`
}

// PlayRingtoneParams TODO DESCRIPTION
type PlayRingtoneParams struct {
	Name     string  `json:"name"`
	Duration float64 `json:"duration"`
}

// PlayParams TODO DESCRIPTION
type PlayParams interface{}

// PlayStruct TODO DESCRIPTION
type PlayStruct struct {
	Type   string     `json:"type"`
	Params PlayParams `json:"params"`
}

// ParamsCallPlay TODO DESCRIPTION
type ParamsCallPlay struct {
	CallID    string       `json:"call_id"`
	NodeID    string       `json:"node_id"`
	ControlID string       `json:"control_id"`
	Volume    float64      `json:"volume,omitempty"`
	Play      []PlayStruct `json:"play"`
}

// ParamsCallPlayStop TODO DESCRIPTION
type ParamsCallPlayStop ParamsGenericAction

// ParamsCallPlayPause TODO DESCRIPTION
type ParamsCallPlayPause ParamsGenericAction

// ParamsCallPlayResume TODO DESCRIPTION
type ParamsCallPlayResume ParamsGenericAction

// ParamsCallPlayVolume TODO DESCRIPTION
type ParamsCallPlayVolume struct {
	CallID    string  `json:"call_id"`
	NodeID    string  `json:"node_id"`
	ControlID string  `json:"control_id"`
	Volume    float64 `json:"volume,omitempty"`
}

// RecordParams TODO DESCRIPTION
type RecordParams struct {
	Format            string `json:"format,omitempty"`
	Direction         string `json:"direction,omitempty"`
	Terminators       string `json:"terminators,omitempty"`
	InitialTimeout    uint16 `json:"initial_timeout,omitempty"`
	EndSilenceTimeout uint16 `json:"end_silence_timeout,omitempty"`
	Beep              bool   `json:"beep,omitempty"`
	Stereo            bool   `json:"stereo,omitempty"`
}

// RecordStruct TODO DESCRIPTION
type RecordStruct struct {
	Audio RecordParams `json:"audio"`
}

// ParamsCallRecord TODO DESCRIPTION
type ParamsCallRecord struct {
	CallID    string       `json:"call_id"`
	NodeID    string       `json:"node_id"`
	ControlID string       `json:"control_id"`
	Record    RecordStruct `json:"record"`
}

// ParamsCallRecordStop TODO DESCRIPTION
type ParamsCallRecordStop ParamsGenericAction

// DetectMachineParamsInternal TODO DESCRIPTION
type DetectMachineParamsInternal struct {
	InitialTimeout        float64 `json:"initial_timeout,omitempty"`
	EndSilenceTimeout     float64 `json:"end_silence_timeout,omitempty"`
	MachineVoiceThreshold float64 `json:"machine_voice_threshold,omitempty"`
	MachineWordsThreshold float64 `json:"machine_words_threshold,omitempty"`
}

// DetectFaxParamsInternal TODO DESCRIPTION
type DetectFaxParamsInternal struct {
	Tone string `json:"tone,omitempty"`
}

// DetectDigitParamsInternal TODO DESCRIPTION
type DetectDigitParamsInternal struct {
	Digits string `json:"digits,omitempty"`
}

// DetectStruct TODO DESCRIPTION
type DetectStruct struct {
	Type   string      `json:"type"`
	Params interface{} `json:"params"`
}

// ParamsCallDetect TODO DESCRIPTION
type ParamsCallDetect struct {
	CallID    string       `json:"call_id"`
	NodeID    string       `json:"node_id"`
	ControlID string       `json:"control_id"`
	Detect    DetectStruct `json:"detect"`
	Timeout   float64      `json:"timeout,omitempty"`
}

// ParamsCallDetectStop TODO DESCRIPTION
type ParamsCallDetectStop ParamsGenericAction

// ParamsSendFax  TODO DESCRIPTION
type ParamsSendFax struct {
	CallID     string `json:"call_id"`
	NodeID     string `json:"node_id"`
	ControlID  string `json:"control_id"`
	Document   string `json:"document"`
	Identity   string `json:"identity,omitempty"`
	HeaderInfo string `json:"header_info,omitempty"`
}

// ParamsFaxStop TODO DESCRIPTION
type ParamsFaxStop ParamsGenericAction

// TapAudioParams TODO DESCRIPTION
type TapAudioParams struct {
	Direction string `json:"direction"`
}

// TapStruct TODO DESCRIPTION
type TapStruct struct {
	Type   string         `json:"type"`
	Params TapAudioParams `json:"params"`
}

// TapDeviceParams TODO DESCRIPTION
type TapDeviceParams struct {
	Addr  string `json:"addr,omitempty"`
	Codec string `json:"codec,omitempty"`
	Port  uint16 `json:"port,omitempty"`
	Ptime uint8  `json:"ptime,omitempty"`
	Rate  uint   `json:"rate,omitempty"`
	URI   string `json:"uri,omitempty"`
}

// TapDevice TODO DESCRIPTION
type TapDevice struct {
	Type   string          `json:"type"`
	Params TapDeviceParams `json:"params"`
}

// ParamsCallTap TODO DESCRIPTION
type ParamsCallTap struct {
	CallID    string    `json:"call_id"`
	NodeID    string    `json:"node_id"`
	ControlID string    `json:"control_id"`
	Tap       TapStruct `json:"tap"`
	Device    TapDevice `json:"device"`
}

// ParamsCallTapStop TODO DESCRIPTION
type ParamsCallTapStop ParamsGenericAction

// ParamsEventCallingCallTap TODO DESCRIPTION
type ParamsEventCallingCallTap struct {
	TapState  string    `json:"state"`
	CallID    string    `json:"call_id"`
	NodeID    string    `json:"node_id"`
	ControlID string    `json:"control_id"`
	Tap       TapStruct `json:"tap"`
	Device    TapDevice `json:"device"`
}

// ParamsEventCallingCallSendDigits TODO DESCRIPTION
type ParamsEventCallingCallSendDigits struct {
	SendDigitsState string `json:"state"`
	CallID          string `json:"call_id"`
	NodeID          string `json:"node_id"`
	ControlID       string `json:"control_id"`
}

// ParamsCallSendDigits TODO DESCRIPTION
type ParamsCallSendDigits struct {
	CallID    string `json:"call_id"`
	NodeID    string `json:"node_id"`
	ControlID string `json:"control_id"`
	Digits    string `json:"digits"`
}

// ResultCollectDigitParams TODO DESCRIPTION
type ResultCollectDigitParams struct {
	Digits     string `json:"digits"`
	Terminator string `json:"terminator"`
}

// ResultCollectSpeechParams TODO DESCRIPTION
type ResultCollectSpeechParams struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
}

// ResultCollect TODO DESCRIPTION
type ResultCollect struct {
	Type   string                 `json:"type"`
	Params map[string]interface{} `json:"params,omitempty"`
}

// ParamsEventCallingCallPlayAndCollect TODO DESCRIPTION
type ParamsEventCallingCallPlayAndCollect struct {
	CallID    string        `json:"call_id"`
	NodeID    string        `json:"node_id"`
	ControlID string        `json:"control_id"`
	Final     bool          `json:"final,omitempty"`
	Result    ResultCollect `json:"result"`
}

// CollectDigits TODO DESCRIPTION
type CollectDigits struct {
	Terminators  string `json:"terminators,omitempty"`
	Max          uint16 `json:"max"`
	DigitTimeout uint16 `json:"digit_timeout,omitempty"`
}

// CollectSpeech TODO DESCRIPTION
type CollectSpeech struct {
	EndSilenceTimeout uint16   `json:"end_silence_timeout,omitempty"`
	SpeechTimeout     uint16   `json:"speech_timeout,omitempty"`
	Language          string   `json:"language,omitempty"`
	Hints             []string `json:"hints,omitempty"`
}

// CollectStruct TODO DESCRIPTION
type CollectStruct struct {
	Digits         *CollectDigits `json:"digits,omitempty"`
	Speech         *CollectSpeech `json:"speech,omitempty"`
	InitialTimeout uint16         `json:"initial_timeout,omitempty"`
	PartialResults bool           `json:"partial_results,omitempty"`
}

// ParamsCallPlayAndCollect TODO DESCRIPTION
type ParamsCallPlayAndCollect struct {
	CallID    string        `json:"call_id"`
	NodeID    string        `json:"node_id"`
	ControlID string        `json:"control_id"`
	Volume    float64       `json:"volume,omitempty"`
	Play      []PlayStruct  `json:"play"`
	Collect   CollectStruct `json:"collect"`
}

// ParamsMessagingSend TODO DESCRIPTION
type ParamsMessagingSend struct {
	ToNumber   string   `json:"to_number"`
	FromNumber string   `json:"from_number"`
	Context    string   `json:"context"`
	Body       string   `json:"body"`
	Tags       []string `json:"tags,omitempty"`
	Region     string   `json:"region,omitempty"`
	Media      []string `json:"media,omitempty"`
}

// ParamsEventMessagingState TODO DESCRIPTION
type ParamsEventMessagingState struct {
	ToNumber     string   `json:"to_number"`
	FromNumber   string   `json:"from_number"`
	Direction    string   `json:"direction"`
	Context      string   `json:"context"`
	Body         string   `json:"body"`
	Tags         []string `json:"tags"`
	Media        []string `json:"media"`
	Segments     uint     `json:"segments"`
	MessageState string   `json:"message_state"`
	MsgID        string   `json:"message_id"`
}

// ParamsEventTaskingTask TODO DESCRIPTION
type ParamsEventTaskingTask struct {
	SpaceID      string      `json:"space_id"`
	ProjectID    string      `json:"project_id"`
	Context      string      `json:"context"`
	Message      interface{} `json:"message"`
	Timestamp    float64     `json:"timestamp"`
	EventChannel string      `json:"event_channel"`
}
