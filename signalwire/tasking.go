package signalwire

import (
	"context"
	"time"
)

// Tasking TODO DESCRIPTION
type Tasking struct {
	Ctx      context.Context
	Cancel   context.CancelFunc
	Relay    *RelaySession
	Consumer *Consumer
	TaskChan chan ParamsEventTaskingTask
}

// ITasking object visible to the end user
type ITasking interface {
	Deliver(signalwireContext string, params interface{}) error
}

func (tasking *Tasking) callbacksRunDeliver(_ context.Context) {
	timer := time.NewTimer(BroadcastEventTimeout * time.Second)

	consumer := tasking.Consumer
	select {
	case params := <-tasking.TaskChan:
		if consumer.OnTask != nil {
			go consumer.OnTask(consumer, params)
		}

	case <-timer.C:
	}
}

// Deliver TODO DESCRIPTION
func (tasking *Tasking) Deliver(signalwireContext string, b []byte) bool {
	if tasking.Relay == nil {
		return false
	}

	if tasking.Ctx == nil {
		return false
	}

	if err := tasking.Relay.RelayTaskDeliver(tasking.Ctx, TaskingEndpoint, tasking.Consumer.Project, tasking.Consumer.Token, signalwireContext, b); err != nil {
		Log.Error("RelayTaskDeliver: %v", err)
		return false
	}

	tasking.callbacksRunDeliver(tasking.Ctx)

	return true
}
