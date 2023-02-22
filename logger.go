package vmdiff

import (
	"io"
	"log"
	"os"
)

// Logger is implemented by any logging system that is used for standard logs.
type Logger interface {
	Errorf(string, ...interface{})
	Infof(string, ...interface{})
	Debugf(string, ...interface{})
	Tracef(string, ...interface{})
}

type logLevel int

const (
	ERROR logLevel = iota
	INFO
	DEBUG
	TRACE
)

type defaultLogger struct {
	*log.Logger
	level logLevel
}

func getDefaultLogger(level logLevel) *defaultLogger {
	return &defaultLogger{
		Logger: log.New(os.Stderr, "vmdiff: ", log.LstdFlags|log.Lmsgprefix),
		level:  level,
	}
}

func (l *defaultLogger) Errorf(f string, v ...interface{}) {
	if l.level >= ERROR {
		l.Printf("ERROR: "+f, v...)
	}
}

func (l *defaultLogger) Infof(f string, v ...interface{}) {
	if l.level >= INFO {
		l.Printf("INFO: "+f, v...)
	}
}

func (l *defaultLogger) Debugf(f string, v ...interface{}) {
	if l.level >= DEBUG {
		l.Printf("DEBUG: "+f, v...)
	}
}

func (l *defaultLogger) Tracef(f string, v ...interface{}) {
	if l.level >= TRACE {
		l.Printf("TRACE: "+f, v...)
	}
}

var logger Logger = getDefaultLogger(INFO)

// Sets the Logger to the given logger
// The default logger uses the log package from the Go standard library
// and write to stderr at INFO level
func SetLogger(l Logger) {
	logger = l
}

// Disable the logging completely
func DisableLogging() {
	l, ok := logger.(*defaultLogger)
	if ok {
		l.SetOutput(io.Discard)
	}
}

// Set the default logger log level.
// Default level is INFO
func SetDefaultLogLevel(l logLevel) {
	logger = getDefaultLogger(l)
}
