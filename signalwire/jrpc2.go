package signalwire

import (
	"io/ioutil"
	"log"
)

// Jsonrpc2Logger is a wrapper for sourcegraph jsonrpc2 logger
type Jsonrpc2Logger struct{}

//revive:disable:unused-receiver

// Printf is a logger for Jsonrpc2 library
func (l *Jsonrpc2Logger) Printf(format string, args ...interface{}) {
	Log.Trace(format, args)
}

//revive:enable:unused-receiver

/*
  We need to silence default logger, for some strange reason
  sourcegraph/jsonrpc2 uses Log to print debug message without
  any way of disabling this behavior.
  https://github.com/sourcegraph/jsonrpc2/blob/master/jsonrpc2.go#L579
*/

func init() { // nolint: gochecknoinits
	log.SetOutput(ioutil.Discard)
}
