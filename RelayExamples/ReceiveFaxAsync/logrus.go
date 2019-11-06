package main

import (
	"fmt"
	"path"
	"runtime"

	"github.com/signalwire/signalwire-golang/signalwire"
	"github.com/sirupsen/logrus"
)

func init() {
	signalwire.Log = CreateNewLogrusLogger()
}

// LogrusLogger is a package level logger using logrus
type LogrusLogger struct {
	Log *logrus.Logger
}

// CreateNewLogrusLogger creates new Logrus logger
func CreateNewLogrusLogger() *LogrusLogger {
	l := new(LogrusLogger)

	l.Log = logrus.New()

	l.Log.SetFormatter(
		&logrus.TextFormatter{
			DisableColors: false,
			FullTimestamp: true,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				pc := make([]uintptr, 20)
				n := runtime.Callers(1, pc)

				if n > 0 {
					pc = pc[:n]

					frames := runtime.CallersFrames(pc)

					var next bool

					for {
						frame, more := frames.Next()

						if next {
							return fmt.Sprintf("%s()", frame.Function), fmt.Sprintf("%s:%d", path.Base(frame.File), frame.Line)
						}

						if f.PC == frame.PC {
							next = true
						}

						if !more {
							break
						}
					}
				}

				return "", ""
			},
		},
	)

	l.Log.SetReportCaller(true)

	return l
}

// Trace is a trace level logger
func (l *LogrusLogger) Trace(format string, args ...interface{}) {
	if l == nil {
		panic("logger undefined")
	}

	l.Log.Tracef(format, args...)
}

// Debug is a debug level logger
func (l *LogrusLogger) Debug(format string, args ...interface{}) {
	if l == nil {
		panic("logger undefined")
	}

	l.Log.Debugf(format, args...)
}

// Info is a info level logger
func (l *LogrusLogger) Info(format string, args ...interface{}) {
	if l == nil {
		panic("logger undefined")
	}

	l.Log.Infof(format, args...)
}

// Warn is a warn level logger
func (l *LogrusLogger) Warn(format string, args ...interface{}) {
	if l == nil {
		panic("logger undefined")
	}

	l.Log.Warnf(format, args...)
}

// Error is a error level logger
func (l *LogrusLogger) Error(format string, args ...interface{}) {
	if l == nil {
		panic("logger undefined")
	}

	l.Log.Errorf(format, args...)
}

// Fatal is a fatal level logger
func (l *LogrusLogger) Fatal(format string, args ...interface{}) {
	if l == nil {
		panic("logger undefined")
	}

	l.Log.Fatalf(format, args...)
}

// Panic is a panic level logger
func (l *LogrusLogger) Panic(format string, args ...interface{}) {
	if l == nil {
		panic("logger undefined")
	}

	l.Log.Panicf(format, args...)
}

// SetLevel defines maximum level of a logger output
func (l *LogrusLogger) SetLevel(level int) {
	if l == nil {
		panic("logger undefined")
	}

	var logrusLevel logrus.Level

	switch level {
	case signalwire.TraceLevelLog:
		logrusLevel = logrus.TraceLevel
	case signalwire.DebugLevelLog:
		logrusLevel = logrus.DebugLevel
	case signalwire.InfoLevelLog:
		logrusLevel = logrus.InfoLevel
	case signalwire.WarnLevelLog:
		logrusLevel = logrus.WarnLevel
	case signalwire.ErrorLevelLog:
		logrusLevel = logrus.ErrorLevel
	case signalwire.FatalLevelLog:
		logrusLevel = logrus.FatalLevel
	case signalwire.PanicLevelLog:
		logrusLevel = logrus.PanicLevel
	}

	l.Log.SetLevel(logrusLevel)
}
