package signalwire

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	assert "github.com/stretchr/testify/assert"
)

func TestRelay(t *testing.T) {
	t.Run(
		"RelayPhoneDial",
		func(t *testing.T) {
			relay := new(RelaySession)

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			Imock := NewMockIBlade(mockCtrl)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			/*expected result of BladeExecute*/
			jsonText := `{
			"requester_nodeid": "a6046c1f-0507-4195-8901-f5d405a0a05e",
			"responder_nodeid": "661980d5-d646-4ee3-b224-8631d03ccbe2",
			"result": {
				"code": "200",
				"message": "Call started",
				"node_id": "09ef7c73-6ee1-4f58-a4c4-c34caa689bfa",
				"call_id": "012c57d7-947f-437d-9b53-1dabf89fc1f6"
				}
			}`
			bytes := []byte(jsonText)
			var ReplyBladeExecuteDecode ReplyBladeExecute
			err := json.Unmarshal(bytes, &ReplyBladeExecuteDecode)
			if err != nil {
				t.Fatalf("err: %v", err)
			}

			Imock.EXPECT().BladeExecute(ctx, gomock.Any(), gomock.Any()).Return(&ReplyBladeExecuteDecode, nil)
			blade := &BladeSession{I: Imock}

			err = blade.EventCalling.Cache.InitCache(CacheExpiry*time.Second, CacheCleaning*time.Second)
			assert.Nil(t, err, "must be able to initialize cache")

			relay.Blade = blade
			call := new(CallSession)
			err = relay.RelayPhoneDial(ctx, call, "+12222222", "+13333333", 10)

			assert.NotEqual(t, len(call.TagID), 0, "tag must be set")
			assert.NotEqual(t, call.Timeout, 0, "timeout must be set")

			assert.Nil(t, err, "RelayPhoneDial return error must be nil")
		},
	)
	t.Run(
		"RelayPhoneConnect",
		func(t *testing.T) {
			relay := new(RelaySession)

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			Imock := NewMockIBlade(mockCtrl)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			/*expected result of BladeExecute*/
			jsonText := `{
				"requester_nodeid": "bd55c94f-f8f6-4410-a656-96d21357460c",
				"responder_nodeid": "661980d5-d646-4ee3-b224-8631d03ccbe2",
				"result": {
					"code": "200",
					"message": "Connecting call"
				}
			}`
			bytes := []byte(jsonText)
			var ReplyBladeExecuteDecode ReplyBladeExecute
			err := json.Unmarshal(bytes, &ReplyBladeExecuteDecode)
			if err != nil {
				t.Fatalf("err: %v", err)
			}

			Imock.EXPECT().BladeExecute(ctx, gomock.Any(), gomock.Any()).Return(&ReplyBladeExecuteDecode, nil)
			blade := &BladeSession{I: Imock}
			err = blade.EventCalling.Cache.InitCache(CacheExpiry*time.Second, CacheCleaning*time.Second)
			assert.Nil(t, err, "must be able to initialize cache")

			relay.Blade = blade
			call := new(CallSession)

			call.SetParams("57f0e4d1-82a8-4af1-b74b-e1841b88a339", "61b04307-09b6-43d9-8702-0fbd364eaef0", "+13333333", "+12222222", "go-test", CallOutbound)
			err = relay.RelayPhoneConnect(ctx, call, "+12222222", "+13333333")

			assert.Nil(t, err, "RelayPhoneConnect return error must be nil")
		},
	)
}
