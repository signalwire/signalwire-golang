package signalwire

import (
	"context"
	"errors"
	"sync"
)

// FaxType: type of a Faxing (Send/Receive) event
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

// FaxDirection: direction of a fax Action (send/receive)
type FaxDirection int

// Call state constants
const (
	FaxSend FaxDirection = iota
	FaxReceive
)

func (s FaxDirection) String() string {
	return [...]string{"send", "receive"}[s]
}

type FaxResult struct {
	Identity       string
	RemoteIdentity string
	Document       string
	Direction      FaxDirection
	Pages          uint16
	Successful     bool
}

// FaxAction TODO DESCRIPTION
type FaxAction struct {
	CallObj   *CallObj
	ControlID string
	Completed bool
	Result    FaxResult
	eventType FaxEventType
	err       error
	sync.RWMutex
}

// IFaxAction unit-tests only
type IFaxAction interface {
	faxAsyncStop() error
	Stop()
	GetCompleted() bool
	GetResult() FaxResult
}

const (
	StrError = "error"
)

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
	err := callobj.Calling.Relay.RelayReceiveFax(callobj.Calling.Ctx, callobj.call, &ctrlID)

	if err != nil {
		return &a.Result, err
	}

	callobj.callbacksRunFax(callobj.Calling.Ctx, ctrlID, a)

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
	err := callobj.Calling.Relay.RelaySendFax(callobj.Calling.Ctx, callobj.call, &ctrlID, doc, id, headerInfo)

	if err != nil {
		return &a.Result, err
	}

	callobj.callbacksRunFax(callobj.Calling.Ctx, ctrlID, a)

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

	return callobj.Calling.Relay.RelaySendFaxStop(callobj.Calling.Ctx, callobj.call, ctrlID)
}

func (callobj *CallObj) callbacksRunFax(_ context.Context, ctrlID string, res *FaxAction) {
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

				if callobj.OnFaxFinished != nil {
					callobj.OnFaxFinished(res)
				}
			case FaxPage:
				res.Lock()

				res.eventType = faxevent

				res.Unlock()

				Log.Debug("Page event. ctrlID: %s\n", ctrlID)

				if callobj.OnFaxPage != nil {
					callobj.OnFaxPage(res)
				}
			case FaxError:
				Log.Debug("Fax error. ctrlID: %s\n", ctrlID)

				res.Lock()

				res.Completed = true
				res.eventType = faxevent

				res.Unlock()

				if callobj.OnFaxError != nil {
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
		case <-callobj.call.Hangup:
			out = true
		}

		if out {
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

	go func() {
		go func() {
			// wait to get control ID (buffered channel)
			ctrlID := <-callobj.call.CallFaxControlID

			callobj.callbacksRunFax(callobj.Calling.Ctx, ctrlID, res)
		}()

		newCtrlID, _ := GenUUIDv4()
		res.Lock()
		res.ControlID = newCtrlID
		res.Unlock()

		err := callobj.Calling.Relay.RelayReceiveFax(callobj.Calling.Ctx, callobj.call, &newCtrlID)

		if err != nil {
			res.Lock()

			res.err = err

			res.Completed = true

			res.Unlock()
		}
	}()

	return res, nil
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

	go func() {
		go func() {
			// wait to get control ID (buffered channel)
			ctrlID := <-callobj.call.CallFaxControlID

			callobj.callbacksRunFax(callobj.Calling.Ctx, ctrlID, res)
		}()

		newCtrlID, _ := GenUUIDv4()
		res.Lock()
		res.ControlID = newCtrlID
		res.Unlock()

		err := callobj.Calling.Relay.RelaySendFax(callobj.Calling.Ctx, callobj.call, &newCtrlID, doc, id, headerInfo)

		if err != nil {
			res.Lock()

			res.err = err

			res.Completed = true

			res.Unlock()
		}
	}()

	return res, nil
}

// ctrlIDCopy TODO DESCRIPTION
func (sendfaxaction *FaxAction) ctrlIDCopy() (string, error) {
	sendfaxaction.RLock()

	if len(sendfaxaction.ControlID) == 0 {
		sendfaxaction.RUnlock()
		return "", errors.New("no controlID")
	}

	c := sendfaxaction.ControlID

	sendfaxaction.RUnlock()

	return c, nil
}

// sendfaxAsyncStop TODO DESCRIPTION
func (sendfaxaction *FaxAction) faxAsyncStop() error {
	if sendfaxaction.CallObj.Calling == nil {
		return errors.New("nil Calling object")
	}

	if sendfaxaction.CallObj.Calling.Relay == nil {
		return errors.New("nil Relay object")
	}

	c, err := sendfaxaction.ctrlIDCopy()
	if err != nil {
		return err
	}

	call := sendfaxaction.CallObj.call

	return sendfaxaction.CallObj.Calling.Relay.RelaySendFaxStop(sendfaxaction.CallObj.Calling.Ctx, call, &c)
}

// Stop TODO DESCRIPTION
func (sendfaxaction *FaxAction) Stop() {
	sendfaxaction.err = sendfaxaction.faxAsyncStop()
}

// GetCompleted TODO DESCRIPTION
func (sendfaxaction *FaxAction) GetCompleted() bool {
	sendfaxaction.RLock()

	ret := sendfaxaction.Completed

	sendfaxaction.RUnlock()

	return ret
}

// GetResult TODO DESCRIPTION
func (sendfaxaction *FaxAction) GetResult() FaxResult {
	sendfaxaction.RLock()

	ret := sendfaxaction.Result

	sendfaxaction.RUnlock()

	return ret
}

// GetSuccessful TODO DESCRIPTION
func (sendfaxaction *FaxAction) GetSuccessful() bool {
	sendfaxaction.RLock()

	ret := sendfaxaction.Result.Successful

	sendfaxaction.RUnlock()

	return ret
}

// GetDocument TODO DESCRIPTION
func (sendfaxaction *FaxAction) GetDocument() string {
	sendfaxaction.RLock()

	ret := sendfaxaction.Result.Document

	sendfaxaction.RUnlock()

	return ret
}

// GetPages TODO DESCRIPTION
func (sendfaxaction *FaxAction) GetPages() uint16 {
	sendfaxaction.RLock()

	ret := sendfaxaction.Result.Pages

	sendfaxaction.RUnlock()

	return ret
}

// GetIdentity TODO DESCRIPTION
func (sendfaxaction *FaxAction) GetIdentity() string {
	sendfaxaction.RLock()

	ret := sendfaxaction.Result.Identity

	sendfaxaction.RUnlock()

	return ret
}

// GetRemoteIdentity  TODO DESCRIPTION
func (sendfaxaction *FaxAction) GetRemoteIdentity() string {
	sendfaxaction.RLock()

	ret := sendfaxaction.Result.RemoteIdentity

	sendfaxaction.RUnlock()

	return ret
}
