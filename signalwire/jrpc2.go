package signalwire

// Jsonrpc2Logger is a wrapper for sourcegraph jsonrpc2 logger
type Jsonrpc2Logger struct{}

//revive:disable:unused-receiver

// Printf is a logger for Jsonrpc2 library
func (l *Jsonrpc2Logger) Printf(format string, args ...interface{}) {
	Log.Trace(format, args...)
}

//revive:enable:unused-receiver
