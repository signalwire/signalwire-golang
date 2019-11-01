package signalwire

import (
	"context"
	"sync"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	assert "github.com/stretchr/testify/assert"
)

func TestActionRecord(t *testing.T) {
	t.Run(
		"RecordAsync",
		func(t *testing.T) {
			consumer := new(Consumer)
			assert.NotNil(t, consumer, "Consumer must not be nil")

			signalwireContexts := []string{"Context1"}
			consumer.Setup("ProjectID", "TokenID", signalwireContexts)

			ctx, cancel := context.WithCancel(context.Background())

			var wg sync.WaitGroup
			wg.Add(2)

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			Imock := NewMockIClientSession(mockCtrl)
			IRelayMock := NewMockIRelay(mockCtrl)
			relay := &RelaySession{I: IRelayMock}
			c := &ClientSession{
				I:      Imock,
				Relay:  *relay,
				Ctx:    ctx,
				Cancel: cancel,
				Calling: Calling{
					Relay: relay,
					Ctx:   ctx,
				},
			}
			consumer.Client = c

			consumer.Ready = func(consumer *Consumer) {
				fromNumber := "+1324444444"
				toNumber := "+1325555555"

				newcall := consumer.Client.Calling.NewCall(fromNumber, toNumber)
				newcall.call.CallInit(ctx) // this is called in RelayPhoneDial which is mocked
				go func() {
					// fake Answer state in buffered channel
					newcall.call.CallStateChan <- Answered
				}()

				resultDial := consumer.Client.Calling.Dial(newcall)

				assert.True(t, resultDial.Successful)

				var rec RecordParams

				rec.Beep = true
				rec.Format = "wav"
				rec.Stereo = false
				rec.Direction = RecordDirectionBoth.String()
				rec.InitialTimeout = 10
				rec.EndSilenceTimeout = 3
				rec.Terminators = "#*"

				done := make(chan struct{})
				go func() {
					controlID := "abbf0931-3ca5-4a91-8e29-450a815b0ade"
					newcall.call.CallRecordChans[controlID] = make(chan RecordState, EventQueue)
					newcall.call.CallRecordEventChans[controlID] = make(chan ParamsEventCallingCallRecord, EventQueue)
					newcall.call.CallRecordReadyChans[controlID] = make(chan struct{})
					// fake CtrlID in buffered channel
					newcall.call.CallRecordControlIDs <- controlID
					// fake 'finished' state
					newcall.call.CallRecordChans[controlID] <- RecordFinished
					time.Sleep(200 * time.Millisecond) // the go routines invoked by RecordAudioAsync() should have time to finish.
					done <- struct{}{}
				}()

				recordAction, err := resultDial.Call.RecordAudioAsync(&rec)
				if err != nil {
					Log.Error("Error occurred while trying to record audio\n")
				}
				<-done
				assert.True(t, recordAction.GetSuccessful())
				assert.True(t, recordAction.GetCompleted())
				assert.Equal(t, RecordFinished, recordAction.GetState())
			}

			Imock.EXPECT().connect(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Do(func(ctx context.Context, cancel context.CancelFunc, runWG *sync.WaitGroup) {
				runWG.Done()
			})
			IRelayMock.EXPECT().RelayPhoneDial(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			IRelayMock.EXPECT().RelayRecordAudio(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

			Imock.EXPECT().disconnect().Return(nil).Times(1)

			var err error
			go func() {
				if err = consumer.Run(); err != nil {
					Log.Error("Error occurred while starting Signalwire Consumer\n")
				}
				t.Log("consumer.Run() stopped\n")
				wg.Done()
			}()

			go func() {
				time.Sleep(1 * time.Second)
				/*pretend we connected successfully */
				consumer.Client.Operational <- struct{}{}

				/* make a fake call */

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
