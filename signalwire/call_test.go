package signalwire

import (
	"context"
	"sync"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func TestCall(t *testing.T) {
	t.Run(
		"CallSessionGeneric",
		func(t *testing.T) {
			call := new(CallSession)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			call.CallInit(ctx)
			/* zero-value of a channel is nil, CallInit() creates new channels */
			assert.NotNil(t, call.CallStateChan, "channel must exist")
			assert.NotNil(t, call.CallConnectStateChan, "channel must exist")
			call.SetParams("01748c6a-d5e5-4f6b-bf56-f803f2e9ae33", "2097e3a2-42eb-4b07-b75f-e9f4aecb54c8", "+123456", "+198765", "go-test", CallOutbound)
			assert.NotEqual(t, len(call.CallID), 0, "callID must be set")
			assert.NotEqual(t, len(call.NodeID), 0, "nodeID must be set")
			assert.Equal(t, call.Direction, CallOutbound, "call direction must be set")
			call.UpdateCallState(Answered)
			assert.Equal(t, call.CallState, Answered, "call must be in Answered state")
			call.UpdateCallConnectState(CallConnectConnected)
			assert.Equal(t, call.CallConnectState, CallConnectConnected, "call must be in Answered state")
			var ret bool
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				ret = call.WaitCallStateInternal(ctx, Ending) // buffered channel now
			}()
			// non-blocking
			select {
			case call.CallStateChan <- Ending:
			default:
				t.Errorf("cannot write to channel")
			}
			wg.Wait()
			assert.Equal(t, ret, true, "call should have received Hangup state")
			ret = false
			var wg2 sync.WaitGroup
			wg2.Add(1)
			go func() {
				defer wg2.Done()
				ret = call.WaitCallConnectState(ctx, CallConnectDisconnected)
			}()
			select {
			case call.CallConnectStateChan <- CallConnectDisconnected:
			default:
				t.Errorf("cannot write to channel")
			}
			wg2.Wait()
			assert.Equal(t, ret, true, "call should have received Disconnected state")
			call.CallCleanup(ctx)
		},
	)
}
