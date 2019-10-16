package signalwire

// Available log levels
const (
	TraceLevelLog int = iota + 1
	DebugLevelLog
	InfoLevelLog
	WarnLevelLog
	ErrorLevelLog
	FatalLevelLog
	PanicLevelLog
)

// LoggerWrapper defines custom logger interface
type LoggerWrapper interface {
	Trace(format string, args ...interface{})
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	Fatal(format string, args ...interface{})
	Panic(format string, args ...interface{})

	SetLevel(level int)
}
