package signalwire

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	. "github.com/golang/mock/gomock"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	jsonrpc2 "github.com/sourcegraph/jsonrpc2"
	assert "github.com/stretchr/testify/assert"
)

var upgrader = websocket.Upgrader{}

func wsEcho(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	defer c.Close()

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			break
		}

		if err = c.WriteMessage(mt, message); err != nil {
			break
		}
	}
}

type JEnvelope struct {
	Jsonrpc string
	ID      string
	Method  string
	Params  json.RawMessage
}

func TestBlade(t *testing.T) {
	t.Run(
		"BladeInit",
		func(t *testing.T) {
			Logger = logrus.New()

			mockCtrl := NewController(t)
			defer mockCtrl.Finish()
			Imock := NewMockIBlade(mockCtrl)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			u := url.URL{
				Scheme: "wss",
				Host:   "test.addr",
				Path:   "/",
			}

			s := httptest.NewServer(http.HandlerFunc(wsEcho))
			defer s.Close()

			l := "ws" + strings.TrimPrefix(s.URL, "http")
			wsconn, resp, err := websocket.DefaultDialer.Dial(l, nil)
			if err != nil {
				t.Fatalf("%v", err)
			}

			defer resp.Body.Close()
			defer wsconn.Close()

			// setup fake function that "opens" a WSS connection
			Imock.EXPECT().BladeWSOpenConn(ctx, u).Return(wsconn, nil)
			blade := &BladeSession{I: Imock} // use the mock interface with a real Blade Session

			// call the real BladeInit()
			err = blade.BladeInit(ctx, "test.addr")
			assert.Nil(t, err, "should not be an error from BladeInit")

			if blade.LastError != nil {
				t.Errorf("Err: %v", blade.LastError)
			}

			assert.Nil(t, blade.LastError)
			t.Logf("test.SessionID: %v\n", blade.SessionID)
			assert.NotEqual(t, len(blade.SessionID), 0, "SessionID must exist")
		},
	)

	t.Run(
		"BladeSetup",
		func(t *testing.T) {
			Logger = logrus.New()

			mockCtrl := NewController(t)
			defer mockCtrl.Finish()
			Imock := NewMockIBlade(mockCtrl)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			/*
				blade := NewBladeSession()
				assert.NotNil(t, blade)
				blade.I = Imock*/

			/*expected result of BladeExecute*/
			jsonText := `{
				"requester_nodeid": "7ce2755d-c280-4e19-9ec9-494ce74575d0",
				"responder_nodeid": "aefe545e-4f72-4b9e-bbed-4805fd211ca8",
				"result": {
					 "protocol": "signalwire_00691cf54f879ebb40ef6a0005ad9b93b888cb69eae0b7d1e1a5d45c8cb80c53_917b9402-2986-418a-bbcc-2704f6d9d27f_7c42d161-fe47-4891-8707-8eb25b928719"
				}
			}`

			bytes := []byte(jsonText)
			var ReplySetupDecode ReplyResultSetup

			err := json.Unmarshal(bytes, &ReplySetupDecode)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			t.Logf("ReplySetupDecode: %v\n", ReplySetupDecode)

			Imock.EXPECT().BladeExecute(ctx, Any(), Any()).Return(&ReplySetupDecode, nil)
			blade := &BladeSession{I: Imock}

			// call the real BladeSetup()
			err = blade.BladeSetup(ctx)
			assert.Nil(t, err, "should not be an error from BladeSetup")

			if blade.LastError != nil {
				t.Errorf("Err: %v", blade.LastError)
			}

			assert.Nil(t, blade.LastError)
			assert.NotEqual(t, len(blade.Protocol), 0, "Protocol must exist")
		},
	)

	t.Run(
		"BladeSignalwireReceive",
		func(t *testing.T) {
			Logger = logrus.New()

			mockCtrl := NewController(t)
			defer mockCtrl.Finish()
			Imock := NewMockIBlade(mockCtrl)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			/*expected result of BladeExecute*/
			jsonText := `{
				"requester_nodeid": "ddf2d3c1-4c0c-49af-a400-be4a44bf7c50",
				"responder_nodeid": "476b6fc0-2fae-4a3f-986d-1568aee1deca",
				"result": {
				  "code": "200",
				  "message": "Receiving all inbound related to the requested relay contexts and available scopes"
				}
			}`

			bytes := []byte(jsonText)
			var ReplyBladeExecuteDecode ReplyBladeExecute

			err := json.Unmarshal(bytes, &ReplyBladeExecuteDecode)
			if err != nil {
				t.Fatalf("err: %v", err)
			}

			Imock.EXPECT().BladeExecute(ctx, Any(), Any()).Return(&ReplyBladeExecuteDecode, nil)
			blade := &BladeSession{I: Imock}

			err = blade.BladeSignalwireReceive(ctx, []string{"test"})
			assert.Nil(t, err, "should not be an error from BladeSignalwireReceive")
		},
	)
	t.Run(
		"handleBladeBroadcast",
		func(t *testing.T) {
			Logger = logrus.New()

			mockCtrl := NewController(t)
			defer mockCtrl.Finish()
			Imock := NewMockIBlade(mockCtrl)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			/* this would be coming from jsonrpc2 lib*/
			jsonText := `{
				"jsonrpc": "2.0",
				"id": "936e7b1d-2259-414d-a30a-2d1eb0c86680",
				"method": "blade.broadcast",
				"params": {
					"broadcaster_nodeid": "661980d5-d646-4ee3-b224-8631d03ccbe2",
					"protocol": "signalwire_00691cf54f879ebb40ef6a0005ad9b93b888cb69eae0b7d1e1a5d45c8cb80c53_96009415-e9a3-4038-b621-096ed586a332_7c42d161-fe47-4891-8707-8eb25b928719",
					"channel": "notifications",
					"event": "queuing.relay.events",
					"params": {
						"event_type": "calling.call.receive",
						"timestamp": 1568920048.8261311,
						"project_id": "7c42d161-fe47-4891-8707-8eb25b928719",
						"space_id": "f6e0ee46-4bd4-4856-99bb-0f3bc3d3e787",
						"params": {
						"call_state": "created",
						"context": "test-golang",
						"device": {
							"type": "phone",
							"params": {
							"from_number": "+13254805374",
							"to_number": "+16672286294"
							}
						},
						"direction": "inbound",
						"call_id": "7bebef58-e3c0-4dc7-a2c7-a8c2ffc152fc",
						"node_id": "61b04307-09b6-43d9-8702-0fbd364eaef0"
					},
					"context": "test-golang",
					"event_channel": "signalwire_00691cf54f879ebb40ef6a0005ad9b93b888cb69eae0b7d1e1a5d45c8cb80c53_96009415-e9a3-4038-b621-096ed586a332_7c42d161-fe47-4891-8707-8eb25b928719"
					}
				}
			}`

			bytes := []byte(jsonText)

			var req jsonrpc2.Request

			var jraw JEnvelope
			err := json.Unmarshal(bytes, &jraw)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			req.Params = &jraw.Params

			blade := &BladeSession{I: Imock}

			//			call := new(CallSession)
			Imock.EXPECT().handleInboundCall(ctx, "7bebef58-e3c0-4dc7-a2c7-a8c2ffc152fc").Return(false)

			var I IEventCalling = EventCallingNew()
			calling := &EventCalling{I: I}
			/*point back to blade*/
			calling.blade = blade
			/* point back to Calling obj*/
			calling.I = calling

			if err = calling.Cache.InitCache(CacheExpiry*time.Second, CacheCleaning*time.Second); err != nil {
				t.Fatalf("failed to initialize cache")
			}

			/*copy Calling on blade obj*/
			blade.EventCalling = *calling
			err = blade.handleBladeBroadcast(ctx, &req)
			assert.Nil(t, err, "handleBladeBroadcast should not return error")

			var call *CallSession
			call, err = blade.EventCalling.getCall(ctx, "", "7bebef58-e3c0-4dc7-a2c7-a8c2ffc152fc")
			if err != nil {
				t.Fatalf("failed to get call")
			}
			assert.Equal(t, "7bebef58-e3c0-4dc7-a2c7-a8c2ffc152fc", call.CallID, "CallId does not match")
			assert.Equal(t, "61b04307-09b6-43d9-8702-0fbd364eaef0", call.NodeID, "NodeID does not match")
			assert.Equal(t, "inbound", call.Direction, "Direction does not match")
		},
	)
}
