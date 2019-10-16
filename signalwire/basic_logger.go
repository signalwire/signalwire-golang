package signalwire

import (
	"io/ioutil"
	"log"
	"os"
)

// BasicLogger is a package level logger
type BasicLogger struct {
	TraceLevel *log.Logger
	DebugLevel *log.Logger
	InfoLevel  *log.Logger
	WarnLevel  *log.Logger
	ErrorLevel *log.Logger
	FatalLevel *log.Logger
	PanicLevel *log.Logger
}

// CreateNewBasicLogger creates new logger
func CreateNewBasicLogger() *BasicLogger {
	return &BasicLogger{
		TraceLevel: log.New(
			os.Stderr,
			"TRACE: ",
			log.Ldate|log.Ltime),
		DebugLevel: log.New(
			os.Stderr,
			"DEBUG: ",
			log.Ldate|log.Ltime),
		InfoLevel: log.New(
			os.Stderr,
			"INFO: ",
			log.Ldate|log.Ltime),
		WarnLevel: log.New(
			os.Stderr,
			"WARN: ",
			log.Ldate|log.Ltime),
		ErrorLevel: log.New(
			os.Stderr,
			"ERROR: ",
			log.Ldate|log.Ltime),
		FatalLevel: log.New(
			os.Stderr,
			"FATAL: ",
			log.Ldate|log.Ltime),
		PanicLevel: log.New(
			os.Stderr,
			"PANIC: ",
			log.Ldate|log.Ltime),
	}
}

// Trace is a trace level logger
func (l *BasicLogger) Trace(format string, args ...interface{}) {
	if l == nil {
		panic("logger undefined")
	}

	if l.TraceLevel == nil {
		panic("trace logger undefined")
	}

	l.TraceLevel.Printf(format, args...)
}

// Debug is a debug level logger
func (l *BasicLogger) Debug(format string, args ...interface{}) {
	if l == nil {
		panic("logger undefined")
	}

	if l.DebugLevel == nil {
		panic("debug logger undefined")
	}

	l.DebugLevel.Printf(format, args...)
}

// Info is a info level logger
func (l *BasicLogger) Info(format string, args ...interface{}) {
	if l == nil {
		panic("logger undefined")
	}

	if l.InfoLevel == nil {
		panic("info logger undefined")
	}

	l.InfoLevel.Printf(format, args...)
}

// Warn is a warn level logger
func (l *BasicLogger) Warn(format string, args ...interface{}) {
	if l == nil {
		panic("logger undefined")
	}

	if l.InfoLevel == nil {
		panic("warn logger undefined")
	}

	l.WarnLevel.Printf(format, args...)
}

// Error is a error level logger
func (l *BasicLogger) Error(format string, args ...interface{}) {
	if l == nil {
		panic("logger undefined")
	}

	if l.ErrorLevel == nil {
		panic("error logger undefined")
	}

	l.ErrorLevel.Printf(format, args...)
}

// Fatal is a fatal level logger
func (l *BasicLogger) Fatal(format string, args ...interface{}) {
	if l == nil {
		panic("logger undefined")
	}

	if l.FatalLevel == nil {
		panic("fatal logger undefined")
	}

	l.FatalLevel.Fatalf(format, args...)
}

// Panic is a panic level logger
func (l *BasicLogger) Panic(format string, args ...interface{}) {
	if l == nil {
		panic("logger undefined")
	}

	if l.PanicLevel == nil {
		panic("panic logger undefined")
	}

	l.PanicLevel.Fatalf(format, args...)
}

// SetLevel defines maximum level of a logger output
func (l *BasicLogger) SetLevel(level int) {
	if l == nil {
		panic("logger undefined")
	}

	if l.TraceLevel == nil {
		panic("trace logger undefined")
	}

	if l.DebugLevel == nil {
		panic("debug logger undefined")
	}

	if l.InfoLevel == nil {
		panic("info logger undefined")
	}

	if l.WarnLevel == nil {
		panic("warn logger undefined")
	}

	if l.ErrorLevel == nil {
		panic("error logger undefined")
	}

	if l.FatalLevel == nil {
		panic("fatal logger undefined")
	}

	if l.PanicLevel == nil {
		panic("panic logger undefined")
	}

	if level < TraceLevelLog {
		l.PanicLevel.SetOutput(ioutil.Discard)
	}

	if level < DebugLevelLog {
		l.DebugLevel.SetOutput(ioutil.Discard)
	}

	if level < InfoLevelLog {
		l.InfoLevel.SetOutput(ioutil.Discard)
	}

	if level < WarnLevelLog {
		l.WarnLevel.SetOutput(ioutil.Discard)
	}

	if level < ErrorLevelLog {
		l.ErrorLevel.SetOutput(ioutil.Discard)
	}

	if level < FatalLevelLog {
		l.FatalLevel.SetOutput(ioutil.Discard)
	}

	if level < PanicLevelLog {
		l.FatalLevel.SetOutput(ioutil.Discard)
	}
}
