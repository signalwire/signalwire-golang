package signalwire

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	jsonrpc2 "github.com/sourcegraph/jsonrpc2"
	ws "github.com/sourcegraph/jsonrpc2/websocket"
)

const (
	okCode = "200"
)

// BladeAuth holds auth data for the WS connection
type BladeAuth struct {
	ProjectID string
	TokenID   string
}

// SessionState TODO DESCRIPTION
type SessionState int

// Blade Session Statuses
const (
	BladeOffline SessionState = 1 + iota
	BladeConnecting
	BladeConnected
	BladeSetup
	BladeSubscribed
	BladeRunning
	BladeClosing
	BladeClosed
	BladeShutdown
)

func (s SessionState) String() string {
	return [...]string{"Offline", "Connecting", "Connected", "Setup",
		"Subscribed", "Running", "Closing", "Closed", "Shutdown"}[s]
}

// BladeSession cache Session information
type BladeSession struct {
	SessionID            string
	Protocol             string
	SpaceID              string
	LastError            error
	SessionState         SessionState
	Certified            bool
	SignalwireChannels   []string
	SignalwireContexts   []string
	bladeAuth            BladeAuth
	Ctx                  context.Context
	conn                 *jsonrpc2.Conn
	jOpts                []jsonrpc2.CallOption
	DisconnectChan       chan struct{}
	Calls                [MaxSimCalls]CallSession
	BladeHandlerIncoming ReqHandler
	I                    IBlade
	Inbound              chan string
	Netcast              chan string
	InboundDone          chan struct{}
	InboundMsg           chan string
	InboundMsgDone       chan struct{}
	EventCalling         EventCalling
	EventMessaging       EventMessaging
	EventTasking         EventTasking
}

// IBlade TODO DESCRIPTION
type IBlade interface {
	GetConnection() (*jsonrpc2.Conn, error)
	BladeCleanup() error
	BladeWSOpenConn(ctx context.Context, u url.URL) (*websocket.Conn, error)
	BladeInit(ctx context.Context, addr string) error
	BladeConnect(ctx context.Context, bladeAuth *BladeAuth) error
	BladeSetup(ctx context.Context) error
	BladeAddSubscription(ctx context.Context, signalwireChannels []string) error
	BladeExecute(ctx context.Context, v interface{}, res interface{}) (interface{}, error)
	BladeSignalwireReceive(ctx context.Context, signalwireContexts []string) error
	BladeWaitDisconnect(ctx context.Context)
	BladeDisconnect(ctx context.Context) error
	BladeWaitInboundCall(ctx context.Context) (*CallSession, error)

	handleBladeBroadcast(ctx context.Context, req *jsonrpc2.Request) error
	handleBladeNetcast(ctx context.Context, req *jsonrpc2.Request) error
	handleBladeDisconnect(ctx context.Context, c *jsonrpc2.Request) error
	handleInboundCall(ctx context.Context, callID string) bool
	handleInboundMessage(ctx context.Context, callID string) bool
	eventNotif(ctx context.Context, broadcast NotifParamsBladeBroadcast, rawEvent *json.RawMessage) error
}

// ISessionControl TODO DESCRIPTION
type ISessionControl interface {
	addBlade(c *jsonrpc2.Conn, b *BladeSession)
	getBlade(c *jsonrpc2.Conn) *BladeSession
	removeBlade(c *jsonrpc2.Conn)
}

// BladeSessionControl Control Pool of Sessions
type BladeSessionControl struct {
	sync.RWMutex
	m map[*jsonrpc2.Conn]*BladeSession
}

// NewBladeSessionControl creates new Control Pool of Sessions object
func NewBladeSessionControl() *BladeSessionControl {
	obj := new(BladeSessionControl)
	obj.m = make(map[*jsonrpc2.Conn]*BladeSession)

	return obj
}

// NewBladeSession creates new cache Session information object
func NewBladeSession() *BladeSession {
	obj := new(BladeSession)

	obj.SignalwireChannels = make([]string, 0)
	obj.SignalwireContexts = make([]string, 0)

	obj.jOpts = make([]jsonrpc2.CallOption, 0)

	return obj
}

// GetConnection returns pointer to jsonrpc2.Conn object
func (blade *BladeSession) GetConnection() (*jsonrpc2.Conn, error) {
	if blade == nil {
		return nil, errors.New("empty blade session object")
	}

	return blade.conn, nil
}

// BladeCleanup TODO DESCRIPTION
func (blade *BladeSession) BladeCleanup() error {
	if blade == nil {
		return errors.New("empty blade session object")
	}

	if blade.conn == nil {
		return errors.New("invalid connection")
	}

	GlobalBladeSessionControl.removeBlade(blade.conn)

	blade.SessionState = BladeClosed

	return blade.conn.Close()
}

// BladeWSOpenConn TODO DESCRIPTION
func (blade *BladeSession) BladeWSOpenConn(ctx context.Context, u url.URL) (*websocket.Conn, error) { // nolint: interfacer
	if blade == nil {
		return nil, errors.New("empty blade session object")
	}

	Log.Debug("connecting to %s\n", u.String())

	c, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// BladeNew TODO DESCRIPTION
func BladeNew() *BladeSession {
	return &BladeSession{}
}

// BladeInit TODO DESCRIPTION
func (blade *BladeSession) BladeInit(ctx context.Context, addr string) error {
	var err error

	if blade == nil {
		return errors.New("empty blade session object")
	}

	blade.Ctx = ctx
	u := url.URL{
		Scheme: "wss",
		Host:   addr,
		Path:   "/",
	}

	c, err := blade.I.BladeWSOpenConn(ctx, u)
	if err != nil {
		blade.LastError = err

		return err
	}

	if c == nil {
		return errors.New("cannot open websocket connection")
	}

	stream := ws.NewObjectStream(c)
	l := new(Jsonrpc2Logger)

	blade.conn = jsonrpc2.NewConn(
		ctx, stream,
		blade.BladeHandlerIncoming,
		jsonrpc2.LogMessages(l),
	)

	if blade.conn == nil {
		return errors.New("failed to initialize jsonrpc2 (invalid connection)")
	}

	reqID, _ := GenUUIDv4()

	id := jsonrpc2.ID{
		Str:      reqID,
		IsString: true,
	}

	blade.jOpts = append(blade.jOpts, jsonrpc2.PickID(id))

	blade.SessionID, err = GenUUIDv4()
	if err != nil {
		return err
	}

	GlobalBladeSessionControl.addBlade(blade.conn, blade)

	var I IEventCalling = EventCallingNew()

	/*the circular references are for unit-testing.
	Work with interfaces so we can generate a mock with mockgen. */

	calling := &EventCalling{I: I}
	calling.blade = blade

	calling.I = calling
	blade.EventCalling = *calling

	if err = calling.Cache.InitCache(CacheExpiry*time.Second, CacheCleaning*time.Second); err != nil {
		return errors.New("failed to initialize cache")
	}

	blade.EventCalling = *calling
	blade.EventCalling.blade = blade

	var J IEventMessaging = EventMessagingNew()

	messaging := &EventMessaging{I: J}
	messaging.blade = blade
	messaging.I = messaging
	blade.EventMessaging = *messaging

	if err = messaging.Cache.InitCache(CacheExpiry*time.Second, CacheCleaning*time.Second); err != nil {
		return errors.New("failed to initialize cache")
	}

	blade.EventMessaging = *messaging
	blade.EventMessaging.blade = blade

	var K IEventTasking = EventTaskingNew()

	tasking := &EventTasking{I: K}
	tasking.blade = blade
	tasking.I = tasking
	blade.EventTasking = *tasking

	if err = tasking.Cache.InitCache(CacheExpiry*time.Second, CacheCleaning*time.Second); err != nil {
		return errors.New("failed to initialize cache")
	}

	blade.EventTasking = *tasking

	blade.EventTasking.blade = blade

	blade.Netcast = make(chan string)
	blade.DisconnectChan = make(chan struct{})

	return nil
}

// BladeConnect TODO DESCRIPTION
func (blade *BladeSession) BladeConnect(ctx context.Context, bladeAuth *BladeAuth) error {
	if blade == nil {
		return errors.New("empty blade session object")
	}

	if blade.conn == nil {
		return errors.New("invalid connection")
	}

	var (
		projID string
		tokID  string
	)

	if bladeAuth == nil {
		projID = blade.bladeAuth.ProjectID
		tokID = blade.bladeAuth.TokenID
	} else {
		projID = bladeAuth.ProjectID
		tokID = bladeAuth.TokenID
	}

	if len(projID) == 0 || len(tokID) == 0 {
		err := fmt.Errorf("no auth")
		blade.LastError = err

		return err
	}

	blade.SessionState = BladeConnecting

	var ReplyConnectDecode ReplyResultConnect

	reqID, _ := GenUUIDv4()

	id := jsonrpc2.ID{
		Str:      reqID,
		IsString: true,
	}

	blade.jOpts = nil
	blade.jOpts = append(blade.jOpts, jsonrpc2.PickID(id))

	if err := blade.conn.Call(
		ctx, "blade.connect",
		ParamsConnectStruct{
			Version: BladeVersionStruct{
				Major:    BladeVersionMajor,
				Minor:    BladeVersionMinor,
				Revision: BladeRevision,
			},
			SessionID: blade.SessionID,
			Authentication: AuthStruct{
				Project: bladeAuth.ProjectID,
				Token:   bladeAuth.TokenID,
			},
		},
		&ReplyConnectDecode, blade.jOpts...,
	); err != nil {
		blade.LastError = err

		return err
	}

	blade.SessionState = BladeConnected

	Log.Debug("reply ReplyBladeConnect: %v\n", ReplyConnectDecode)

	return nil
}

// BladeDisconnect TODO DESCRIPTION
func (blade *BladeSession) BladeDisconnect(ctx context.Context) error {
	if blade == nil {
		return errors.New("empty blade session object")
	}

	blade.SessionState = BladeClosing

	if blade.conn == nil {
		return errors.New("invalid connection")
	}

	var ReplyDisconnectDecode ReplyResultDisconnect

	if err := blade.conn.Call(
		ctx, "blade.disconnect",
		ParamsDisconnect{},
		&ReplyDisconnectDecode, blade.jOpts...,
	); err != nil {
		blade.LastError = err

		return err
	}

	blade.SessionState = BladeShutdown

	Log.Debug("reply ReplyDisconnectDecode: %v\n", ReplyDisconnectDecode)

	return nil
}

// BladeSetup TODO DESCRIPTION
func (blade *BladeSession) BladeSetup(ctx context.Context) error {
	if blade == nil {
		return errors.New("empty blade session object")
	}

	var ReplySetupDecode ReplyResultSetup

	v := ParamsSignalwireStruct{
		Protocol: "signalwire",
		Method:   "setup",
		Params:   ParamsSignalwireSetupStruct{},
	}

	Log.Debug("blade.BladeExecute: %p\n", blade.BladeExecute)

	reply, err := blade.I.BladeExecute(ctx, &v, &ReplySetupDecode)
	if err != nil {
		return err
	}

	r, ok := reply.(*ReplyResultSetup)
	if !ok {
		return errors.New("type assertion failed")
	}

	blade.Protocol = r.Result.Protocol

	Log.Debug("reply ReplySetupDecode: %v\n", ReplySetupDecode)
	Log.Debug("reply Reply r: %v\n", r)

	return nil
}

// BladeAddSubscription TODO DESCRIPTION
func (blade *BladeSession) BladeAddSubscription(ctx context.Context, signalwireChannels []string) error {
	if blade == nil {
		return errors.New("empty blade session object")
	}

	if blade.conn == nil {
		return errors.New("invalid connection")
	}

	var ReplySubscriptionDecode ReplyResultSubscription

	reqID, _ := GenUUIDv4()

	id := jsonrpc2.ID{
		Str:      reqID,
		IsString: true,
	}

	blade.jOpts = nil
	blade.jOpts = append(blade.jOpts, jsonrpc2.PickID(id))

	if err := blade.conn.Call(
		ctx, "blade.subscription",
		ParamsSubscriptionStruct{
			Command:  "add",
			Protocol: blade.Protocol,
			Channels: signalwireChannels,
		},
		&ReplySubscriptionDecode, blade.jOpts...,
	); err != nil {
		blade.LastError = err

		return err
	}

	Log.Debug("reply ReplySubscriptionDecode: %v\n", ReplySubscriptionDecode)

	return nil
}

// BladeExecute TODO DESCRIPTION
func (blade *BladeSession) BladeExecute(ctx context.Context, v interface{}, res interface{}) (interface{}, error) {
	if blade == nil {
		return nil, errors.New("empty blade session object")
	}

	if blade.conn == nil {
		return nil, errors.New("invalid connection")
	}

	reqID, _ := GenUUIDv4()

	id := jsonrpc2.ID{
		Str:      reqID,
		IsString: true,
	}

	blade.jOpts = nil
	blade.jOpts = append(blade.jOpts, jsonrpc2.PickID(id))

	if err := blade.conn.Call(ctx, "blade.execute", v, res, blade.jOpts...); err != nil {
		blade.LastError = err

		return nil, err
	}

	return res, nil
}

// BladeSignalwireReceive TODO DESCRIPTION
func (blade *BladeSession) BladeSignalwireReceive(ctx context.Context, signalwireContexts []string) error {
	if blade == nil {
		return errors.New("empty blade session object")
	}

	var ReplyBladeExecuteDecode ReplyBladeExecute

	v := ParamsBladeExecuteStruct{
		Protocol: blade.Protocol,
		Method:   "signalwire.receive",
		Params: ParamsSignalwireReceive{
			Contexts: signalwireContexts,
		},
	}

	reply, err := blade.I.BladeExecute(ctx, &v, &ReplyBladeExecuteDecode)
	if err != nil {
		return err
	}

	r, ok := reply.(*ReplyBladeExecute)
	if !ok {
		return errors.New("type assertion failed")
	}

	if r.Result.Code != okCode {
		return errors.New(r.Result.Message)
	}

	Log.Debug("r.Result.Message: [%s]\n", r.Result.Message)

	return nil
}

// HandleBladeBroadcast TODO DESCRIPTION
func (blade *BladeSession) handleBladeBroadcast(ctx context.Context, req *jsonrpc2.Request) error {
	if blade == nil {
		return errors.New("empty blade session object")
	}

	if req == nil {
		return errors.New("empty rpc request object")
	}

	var broadcast NotifParamsBladeBroadcast

	if err := json.Unmarshal(*req.Params, &broadcast); err != nil {
		return err
	}

	Log.Debug("broadcast.Event: %v\n", broadcast.Event)
	Log.Debug("broadcast.Channel: %v\n", broadcast.Channel)
	Log.Debug("broadcast.Params.EventType: %v\n", broadcast.Params.EventType)
	Log.Debug("broadcast.Params.Params: %v\n", broadcast.Params.Params)

	return blade.eventNotif(ctx, broadcast, req.Params)
}

// HandleBladeNetcast TODO DESCRIPTION
func (blade *BladeSession) handleBladeNetcast(_ context.Context, req *jsonrpc2.Request) error {
	if blade == nil {
		return errors.New("empty blade session object")
	}

	if req == nil {
		return errors.New("empty rpc request object")
	}

	var netcast NotifParamsBladeNetcast

	if err := json.Unmarshal(*req.Params, &netcast); err != nil {
		return err
	}

	// TODO: not required for uncertified client
	switch netcast.Command {
	case "route.add":
	case "route.remove":
	case "identity.add":
	case "identity.remove":
	case "protocol.add":
		select {
		case blade.Netcast <- netcast.Params.Protocol:
		default:
		}
	case "protocol.remove":
	case "protocol.provider.add":
	case "protocol.provider.remove":
	case "protocol.provider.rank.update":
	case "protocol.provider.data.update":
	case "protocol.method.add":
	case "protocol.method.remove":
	case "protocol.channel.add":
	case "protocol.channel.remove":
	case "subscription.add":
	case "subscription.remove":
	case "authority.add":
	case "authority.remove":
	case "authorization.add":
	case "authorization.remove":
	case "access.add":
	case "access.remove":
	}

	Log.Debug("netcast.Command: %v %p\n", netcast.Command, blade)

	return nil
}

// HandleBladeDisconnect TODO DESCRIPTION
func (blade *BladeSession) handleBladeDisconnect(_ context.Context, c *jsonrpc2.Request) error {
	if blade == nil {
		return errors.New("empty blade session object")
	}

	Log.Debug("handleBladeDisconnect conn [%p] [%p]\n", c, blade)

	if blade.SessionState == BladeConnecting || blade.SessionState == BladeRunning {
		blade.SessionState = BladeClosing
	}

	return blade.I.BladeCleanup()
}

func (ctrl *BladeSessionControl) addBlade(c *jsonrpc2.Conn, b *BladeSession) {
	if ctrl == nil {
		return
	}

	if c == nil {
		return
	}

	ctrl.Lock()

	ctrl.m[c] = b

	ctrl.Unlock()
}

func (ctrl *BladeSessionControl) getBlade(c *jsonrpc2.Conn) *BladeSession {
	if ctrl == nil {
		return nil
	}

	if c == nil {
		return nil
	}

	ctrl.RLock()

	b := ctrl.m[c]

	ctrl.RUnlock()

	return b
}

func (ctrl *BladeSessionControl) removeBlade(c *jsonrpc2.Conn) {
	if ctrl == nil {
		return
	}

	if c == nil {
		return
	}

	ctrl.Lock()
	delete(ctrl.m, c)
	ctrl.Unlock()
}

// ReqHandler TODO DESCRIPTION
type ReqHandler struct {
}

// Handle JSONRPC2.0 reply handler
func (ReqHandler) Handle(ctx context.Context, c *jsonrpc2.Conn, req *jsonrpc2.Request) {
	blade := GlobalBladeSessionControl.getBlade(c)
	if blade == nil {
		Log.Warn("no Blade session for this connection\n")

		return
	}

	switch req.Method {
	case "blade.broadcast":
		Log.Debug("got blade.broadcast conn [%p]\n", c)

		if err := blade.I.handleBladeBroadcast(ctx, req); err != nil {
			Log.Error("HandleBladeBroadcast err: %s\n", err)
		}
	case "blade.netcast":
		Log.Debug("got blade.netcast conn [%p]\n", c)

		if err := blade.I.handleBladeNetcast(ctx, req); err != nil {
			Log.Error("HandleBladeNetcast err: %s\n", err)
		}
	case "blade.disconnect":
		Log.Debug("got blade.disconnect conn [%p]\n", c)

		if err := blade.I.handleBladeDisconnect(ctx, req); err != nil {
			Log.Error("HandleBladeDisconnect err: %s\n", err)
		}
	}

	Log.Debug("%s: %s\n", req.ID, *req.Params)
}

// BladeWaitDisconnect TODO DESCRIPTION
func (blade *BladeSession) BladeWaitDisconnect(_ context.Context) {
	conn, err := blade.GetConnection()
	if err != nil {
		Log.Error("%v\n", err)

		return
	}

	if conn == nil {
		Log.Error("Connection is nil\n")

		return
	}

	select {
	case <-conn.DisconnectNotify(): // remote disconnect
		Log.Debug("got remote disconnect\n")
	case <-blade.DisconnectChan: // local disconnect
		Log.Debug("got local disconnect\n")
	}
}

func (blade *BladeSession) handleInboundCall(_ context.Context, callID string) bool {
	Log.Debug("handleInboundCall callID: %s\n", callID)

	// new inbound call
	select {
	case blade.Inbound <- callID:
		Log.Debug("sent callID to Inbound handler go routine\n")
		return true
	default:
		// channel not open - we're not waiting for inbound calls
		Log.Debug("no new call signal sent\n")
	}

	return false
}

// BladeSetupInbound TODO DESCRIPTION
func (blade *BladeSession) BladeSetupInbound(_ context.Context) {
	blade.Inbound = make(chan string, 1)
	blade.InboundDone = make(chan struct{}, 1)
}

// BladeWaitInboundCall TODO DESCRIPTION
func (blade *BladeSession) BladeWaitInboundCall(ctx context.Context) (*CallSession, error) {
	var (
		callID string
		ret    bool
	)

	for {
		var out bool

		select {
		case callID = <-blade.Inbound:
			out = true
			ret = true
		case <-blade.InboundDone:
			out = true
		}

		if out {
			break
		}
	}

	if !ret {
		// shutdown
		return nil, nil
	}

	call, _ := blade.EventCalling.I.getCall(ctx, "", callID)
	if call == nil {
		return nil, fmt.Errorf("error, nil CallSession")
	}

	return call, nil
}

// eventNotif TODO DESCRIPTION
func (blade *BladeSession) eventNotif(ctx context.Context, broadcast NotifParamsBladeBroadcast, rawEvent *json.RawMessage) error {
	calling := blade.EventCalling
	messaging := blade.EventMessaging
	tasking := blade.EventTasking

	switch broadcast.Event {
	case "queuing.relay.events":
		switch broadcast.Params.EventType {
		case "calling.call.connect":
			if err := calling.onCallingEventConnect(ctx, broadcast, rawEvent); err != nil {
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
			if err := calling.onCallingEventPlay(ctx, broadcast, rawEvent); err != nil {
				return err
			}
		case "calling.call.collect":
			if err := calling.onCallingEventCollect(ctx, broadcast, rawEvent); err != nil {
				return err
			}
		case "calling.call.record":
			if err := calling.onCallingEventRecord(ctx, broadcast, rawEvent); err != nil {
				return err
			}
		case "calling.call.tap":
			if err := calling.onCallingEventTap(ctx, broadcast, rawEvent); err != nil {
				return err
			}
		case "calling.call.detect":
			if err := calling.onCallingEventDetect(ctx, broadcast, rawEvent); err != nil {
				return err
			}
		case "calling.call.fax":
			if err := calling.onCallingEventFax(ctx, broadcast, rawEvent); err != nil {
				return err
			}
		case "calling.call.send_digits":
			if err := calling.onCallingEventSendDigits(ctx, broadcast, rawEvent); err != nil {
				return err
			}
		default:
			Log.Debug("got event_type %s\n", broadcast.Params.EventType)
		}
	case "queuing.relay.messaging":
		switch broadcast.Params.EventType {
		case "messaging.receive":
			// An inbound message has been received.
			// always state "received"
			fallthrough
		case "messaging.state":
			if err := messaging.onMessagingEventState(ctx, broadcast); err != nil {
				return err
			}
		default:
			Log.Debug("got event_type %s\n", broadcast.Params.EventType)
		}
	case "queuing.relay.tasks":
		if err := tasking.onTaskingEvent(ctx, broadcast); err != nil {
			return err
		}
	case "relay":
		Log.Debug("got RELAY event\n")
	default:
		Log.Debug("got event %s . unsupported\n", broadcast.Event)

		return fmt.Errorf("unsupported event")
	}

	return nil
}

func (blade *BladeSession) handleInboundMessage(_ context.Context, msgID string) bool {
	Log.Debug("handleInboundMessage callID: %s\n", msgID)

	// new inbound msg
	select {
	case blade.InboundMsg <- msgID:
		Log.Debug("sent msgID to InboundMsg handler go routine\n")
		return true
	default:
		Log.Debug("no new msg signal sent\n")
	}

	return false
}

// BladeSetupInboundMsg TODO DESCRIPTION
func (blade *BladeSession) BladeSetupInboundMsg(_ context.Context) {
	blade.InboundMsg = make(chan string, 1)
	blade.InboundMsgDone = make(chan struct{}, 1)
}

// BladeWaitInboundMsg TODO DESCRIPTION
func (blade *BladeSession) BladeWaitInboundMsg(ctx context.Context) (*MsgSession, error) {
	var (
		msgID string
		ret   bool
	)

	for {
		var out bool

		select {
		case msgID = <-blade.InboundMsg:
			out = true
			ret = true
		case <-blade.InboundMsgDone:
			out = true
		}

		if out {
			break
		}
	}

	if !ret {
		// shutdown
		return nil, nil
	}

	msg, _ := blade.EventMessaging.I.getMsg(ctx, msgID)
	if msg == nil {
		return nil, fmt.Errorf("error, nil CallSession")
	}

	return msg, nil
}
