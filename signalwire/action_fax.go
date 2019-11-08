package signalwire

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
)

// FaxEventType type of a Faxing (Send/Receive) event
type FaxEventType int

// Call state constants
const (
	FaxError FaxEventType = iota
	FaxPage
	FaxFinished
)

func (s FaxEventType) String() string {
	return [...]string{"Error", "Page", Finished}[s]
}

// FaxDirection direction of a fax Action (send/receive)
type FaxDirection int

// Call state constants
const (
	FaxSend FaxDirection = iota
	FaxReceive
)

func (s FaxDirection) String() string {
	return [...]string{"send", "receive"}[s]
}

// FaxResult TODO DESCRIPTION
type FaxResult struct {
	Identity       string
	RemoteIdentity string
	Document       string
	Direction      FaxDirection
	Pages          uint16
	Successful     bool
	Event          json.RawMessage
}

// FaxAction TODO DESCRIPTION
type FaxAction struct {
	CallObj   *CallObj
	ControlID string
	Completed bool
	Result    FaxResult
	Payload   *json.RawMessage
	eventType FaxEventType
	err       error
	done      chan bool
	sync.RWMutex
}

// IFaxAction unit-tests only
type IFaxAction interface {
	faxAsyncStop() error
	Stop()
	GetCompleted() bool
	GetResult() FaxResult
}

// generic name for error event
const (
	StrError = "error"
)

// FaxParamsInternal TODO DESCRIPTION
type FaxParamsInternal struct {
	doc        string
	id         string
	headerInfo string
}

// ReceiveFax TODO DESCRIPTION
func (callobj *CallObj) ReceiveFax() (*FaxResult, error) {
	a := new(FaxAction)

	if callobj.Calling == nil {
		return &a.Result, errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return &a.Result, errors.New("nil Relay object")
	}

	ctrlID, _ := GenUUIDv4()
	err := callobj.Calling.Relay.RelayReceiveFax(callobj.Calling.Ctx, callobj.call, &ctrlID, nil)

	if err != nil {
		return &a.Result, err
	}

	callobj.callbacksRunFax(callobj.Calling.Ctx, ctrlID, a, true)

	return &a.Result, nil
}

// SendFax TODO DESCRIPTION
func (callobj *CallObj) SendFax(doc, id, headerInfo string) (*FaxResult, error) {
	a := new(FaxAction)

	if callobj.Calling == nil {
		return &a.Result, errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return &a.Result, errors.New("nil Relay object")
	}

	ctrlID, _ := GenUUIDv4()

	var fax FaxParamsInternal

	fax.doc = doc
	fax.id = id
	fax.headerInfo = headerInfo

	err := callobj.Calling.Relay.RelaySendFax(callobj.Calling.Ctx, callobj.call, &ctrlID, &fax, nil)

	if err != nil {
		return &a.Result, err
	}

	callobj.callbacksRunFax(callobj.Calling.Ctx, ctrlID, a, true)

	return &a.Result, nil
}

// SendFaxStop TODO DESCRIPTION
func (callobj *CallObj) SendFaxStop(ctrlID *string) error {
	if callobj.Calling == nil {
		return errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return errors.New("nil Relay object")
	}

	return callobj.Calling.Relay.RelaySendFaxStop(callobj.Calling.Ctx, callobj.call, ctrlID, nil)
}

func (callobj *CallObj) callbacksRunFax(ctx context.Context, ctrlID string, res *FaxAction, norunCB bool) {
	for {
		var out bool

		select {
		case faxevent := <-callobj.call.CallFaxChan:
			if faxevent == FaxFinished {
				out = true
			}

			switch faxevent {
			case FaxFinished:
				res.Lock()

				res.eventType = faxevent
				res.Result.Successful = true
				res.Completed = true

				res.Unlock()

				Log.Debug("Fax finished. ctrlID: %s res [%p] Completed [%v] Successful [%v]\n", ctrlID, res, res.Completed, res.Result.Successful)

				if callobj.OnFaxFinished != nil && !norunCB {
					callobj.OnFaxFinished(res)
				}
			case FaxPage:
				res.Lock()

				res.eventType = faxevent

				res.Unlock()

				Log.Debug("Page event. ctrlID: %s\n", ctrlID)

				if callobj.OnFaxPage != nil && !norunCB {
					callobj.OnFaxPage(res)
				}
			case FaxError:
				Log.Debug("Fax error. ctrlID: %s\n", ctrlID)

				res.Lock()

				res.Completed = true
				res.eventType = faxevent

				res.Unlock()

				if callobj.OnFaxError != nil && !norunCB {
					callobj.OnFaxError(res)
				}
			default:
				Log.Debug("Unknown state. ctrlID: %s\n", ctrlID)
			}
		case fax := <-callobj.call.CallFaxEventChan:
			Log.Debug("go params: %v\n", fax)

			switch fax.EventType {
			case "page":
				page := fax.Params
				if page["number"].(float64) > 0 {
					res.Lock()
					res.Result.Pages = uint16(page["number"].(float64))
					res.Unlock()
				}

				if len(page["direction"].(string)) > 0 {
					switch page["direction"].(string) {
					case "send":
						res.Lock()
						res.Result.Direction = FaxSend
						res.Unlock()
					case "receive":
						res.Lock()
						res.Result.Direction = FaxReceive
						res.Unlock()
					}
				}
			case StrError:
				evError := fax.Params

				res.Lock()

				res.err = errors.New(evError["description"].(string))

				res.Unlock()
			case "finished":
				evFinished := fax.Params

				var ok bool

				res.Lock()

				res.Result.Document, ok = evFinished["document"].(string)
				if !ok {
					res.err = errors.New("type assertion failed")
					return
				}

				if len(evFinished["identity"].(string)) > 0 {
					res.Result.Identity, ok = evFinished["identity"].(string)
					if !ok {
						res.err = errors.New("type assertion failed")
						return
					}
				}

				if len(evFinished["remote_identity"].(string)) > 0 {
					res.Result.RemoteIdentity, ok = evFinished["remote_identity"].(string)
					if !ok {
						res.err = errors.New("type assertion failed")
						return
					}
				}

				res.Result.Pages = uint16(evFinished["pages"].(float64))

				res.Unlock()
			}

			callobj.call.CallFaxReadyChan <- struct{}{}
		case rawEvent := <-callobj.call.CallFaxRawEventChan:
			res.Lock()
			res.Result.Event = *rawEvent
			res.Unlock()

			callobj.call.CallFaxReadyChan <- struct{}{}
		case <-callobj.call.Hangup:
			out = true
		case <-ctx.Done():
			out = true
		}

		if out {
			res.done <- res.Result.Successful
			break
		}
	}
}

// ReceiveFaxAsync TODO DESCRIPTION
func (callobj *CallObj) ReceiveFaxAsync() (*FaxAction, error) {
	res := new(FaxAction)

	if callobj.Calling == nil {
		return res, errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return res, errors.New("nil Relay object")
	}

	res.CallObj = callobj
	done := make(chan struct{}, 1)

	go func() {
		go func() {
			res.done = make(chan bool, 2)
			// wait to get control ID (buffered channel)
			ctrlID := <-callobj.call.CallFaxControlID

			callobj.callbacksRunFax(callobj.Calling.Ctx, ctrlID, res, false)
		}()

		newCtrlID, _ := GenUUIDv4()
		res.Lock()
		res.ControlID = newCtrlID
		res.Unlock()

		err := callobj.Calling.Relay.RelayReceiveFax(callobj.Calling.Ctx, callobj.call, &newCtrlID, &res.Payload)

		if err != nil {
			res.Lock()

			res.err = err

			res.Completed = true

			res.Unlock()
		}
		done <- struct{}{}
	}()

	<-done

	return res, res.err
}

// SendFaxAsync TODO DESCRIPTION
func (callobj *CallObj) SendFaxAsync(doc, id, headerInfo string) (*FaxAction, error) {
	res := new(FaxAction)

	if callobj.Calling == nil {
		return res, errors.New("nil Calling object")
	}

	if callobj.Calling.Relay == nil {
		return res, errors.New("nil Relay object")
	}

	res.CallObj = callobj
	done := make(chan struct{}, 1)

	go func() {
		go func() {
			res.done = make(chan bool, 2)
			// wait to get control ID (buffered channel)
			ctrlID := <-callobj.call.CallFaxControlID

			callobj.callbacksRunFax(callobj.Calling.Ctx, ctrlID, res, false)
		}()

		newCtrlID, _ := GenUUIDv4()
		res.Lock()
		res.ControlID = newCtrlID
		res.Unlock()

		var fax FaxParamsInternal

		fax.doc = doc
		fax.id = id
		fax.headerInfo = headerInfo
		err := callobj.Calling.Relay.RelaySendFax(callobj.Calling.Ctx, callobj.call, &newCtrlID, &fax, &res.Payload)

		if err != nil {
			res.Lock()

			res.err = err

			res.Completed = true

			res.Unlock()
		}
		done <- struct{}{}
	}()

	<-done

	return res, nil
}

// ctrlIDCopy TODO DESCRIPTION
func (action *FaxAction) ctrlIDCopy() (string, error) {
	action.RLock()

	if len(action.ControlID) == 0 {
		action.RUnlock()
		return "", errors.New("no controlID")
	}

	c := action.ControlID

	action.RUnlock()

	return c, nil
}

// sendfaxAsyncStop TODO DESCRIPTION
func (action *FaxAction) faxAsyncStop() error {
	if action.CallObj.Calling == nil {
		return errors.New("nil Calling object")
	}

	if action.CallObj.Calling.Relay == nil {
		return errors.New("nil Relay object")
	}

	c, err := action.ctrlIDCopy()
	if err != nil {
		return err
	}

	call := action.CallObj.call

	return action.CallObj.Calling.Relay.RelaySendFaxStop(action.CallObj.Calling.Ctx, call, &c, &action.Payload)
}

// Stop TODO DESCRIPTION
func (action *FaxAction) Stop() StopResult {
	res := new(StopResult)
	action.err = action.faxAsyncStop()

	if action.err == nil {
		waitStop(res, action.done)
	}

	return *res
}

// GetCompleted TODO DESCRIPTION
func (action *FaxAction) GetCompleted() bool {
	action.RLock()

	ret := action.Completed

	action.RUnlock()

	return ret
}

// GetResult TODO DESCRIPTION
func (action *FaxAction) GetResult() FaxResult {
	action.RLock()

	ret := action.Result

	action.RUnlock()

	return ret
}

// GetSuccessful TODO DESCRIPTION
func (action *FaxAction) GetSuccessful() bool {
	action.RLock()

	ret := action.Result.Successful

	action.RUnlock()

	return ret
}

// GetDocument TODO DESCRIPTION
func (action *FaxAction) GetDocument() string {
	action.RLock()

	ret := action.Result.Document

	action.RUnlock()

	return ret
}

// GetPages TODO DESCRIPTION
func (action *FaxAction) GetPages() uint16 {
	action.RLock()

	ret := action.Result.Pages

	action.RUnlock()

	return ret
}

// GetIdentity TODO DESCRIPTION
func (action *FaxAction) GetIdentity() string {
	action.RLock()

	ret := action.Result.Identity

	action.RUnlock()

	return ret
}

// GetRemoteIdentity  TODO DESCRIPTION
func (action *FaxAction) GetRemoteIdentity() string {
	action.RLock()

	ret := action.Result.RemoteIdentity

	action.RUnlock()

	return ret
}

// GetPayload TODO DESCRIPTION
func (action *FaxAction) GetPayload() *json.RawMessage {
	action.RLock()

	ret := action.Payload

	action.RUnlock()

	return ret
}

// GetEvent TODO DESCRIPTION
func (action *FaxAction) GetEvent() *json.RawMessage {
	action.RLock()

	ret := &action.Result.Event

	action.RUnlock()

	return ret
}
