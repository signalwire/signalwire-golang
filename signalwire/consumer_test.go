package signalwire

import (
	"context"
	"sync"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	assert "github.com/stretchr/testify/assert"
)

func TestConsumer(t *testing.T) {
	t.Run(
		"ConsumerConnectDisconnect",
		func(t *testing.T) {
			Logger = logrus.New()
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

			Imock.EXPECT().Connect(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Do(func(ctx context.Context, cancel context.CancelFunc, runWG *sync.WaitGroup) {
				runWG.Done()
			})
			Imock.EXPECT().Disconnect().Return(nil).Times(1)

			timer := time.NewTimer(1 * time.Second)
			var err error
			go func() {
				if err = consumer.Run(); err != nil {
					log.Errorf("Error occurred while starting Signalwire Consumer")
				}
				t.Log("consumer.Run() stopped\n")
				wg.Done()
			}()

			go func() {
				<-timer.C
				/*pretend we connected successfully */
				consumer.Client.Operational <- struct{}{}

				/* this will call Disconnect() which is mocked*/
				if errStop := consumer.Stop(); errStop != nil {
					log.Errorf("Error occurred while stopping Signalwire Consumer: %v\n", errStop)
				}

				wg.Done()
			}()

			wg.Wait()

			assert.Nil(t, err, "Consumer/Run() return error must be nil")
		},
	)
}
