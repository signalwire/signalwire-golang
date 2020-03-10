package signalwire

import (
	"context"
	"sync"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	assert "github.com/stretchr/testify/assert"
)

func TestConsumer(t *testing.T) {
	t.Run(
		"ConsumerConnectDisconnect",
		func(t *testing.T) {
			consumer := new(Consumer)
			assert.NotNil(t, consumer, "Consumer must not be nil")

			signalwireContexts := []string{"Context1"}
			consumer.Setup("ProjectID", "TokenID", signalwireContexts)

			var wg sync.WaitGroup
			wg.Add(2)

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			Imock := NewMockIClientSession(mockCtrl)
			c := &ClientSession{I: Imock}
			consumer.Client = c
			//			t := time.NewTimer(1)

			Imock.EXPECT().connectInternal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Do(func(ctx context.Context, cancel context.CancelFunc, runWG *sync.WaitGroup, t *time.Timer) {
				/*pretend we connected successfully */
				consumer.Client.Operational <- struct{}{}
				runWG.Done()
			})
			Imock.EXPECT().disconnectInternal().Return(nil).Times(1)

			timer := time.NewTimer(1 * time.Second)
			var err error
			go func() {
				if err = consumer.Run(); err != nil {
					Log.Error("Error occurred while starting Signalwire Consumer\n")
				}
				t.Log("consumer.Run() stopped\n")
				wg.Done()
			}()

			go func() {
				<-timer.C
				/* this will call Disconnect() which is mocked*/
				if errStop := consumer.Stop(); errStop != nil {
					Log.Error("Error occurred while stopping Signalwire Consumer: %v\n", errStop)
				}

				wg.Done()
			}()

			wg.Wait()

			assert.Nil(t, err, "Consumer/Run() return error must be nil")
		},
	)
}
