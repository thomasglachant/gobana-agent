package core

import (
	"fmt"
	baseLog "log"
	"runtime/debug"
)

var Logger = &LoggerStruct{
	DebugEnabled: false,
}

type LogInterface interface {
	Infof(prefix string, format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Criticalf(prefix string, format string, v ...interface{})
	Noticef(prefix string, format string, v ...interface{})
	Debugf(prefix string, format string, v ...interface{})
}

// LoggerStruct represents a logger.
type LoggerStruct struct {
	DebugEnabled bool
}

// Infof print an information log.
func (logger *LoggerStruct) Infof(prefix, format string, v ...interface{}) {
	str := fmt.Sprintf("%s.INFO: %s", prefix, fmt.Sprintf(format, v...))
	baseLog.Print(str)
}

// Errorf print an error log.
func (logger *LoggerStruct) Errorf(prefix, format string, v ...interface{}) {
	str := fmt.Sprintf("%s.ERROR: %s", prefix, fmt.Sprintf(format, v...))
	baseLog.Print(str)
}

// Criticalf print an critical log and exit app.
func (logger *LoggerStruct) Criticalf(prefix, format string, v ...interface{}) {
	str := fmt.Sprintf("%s.CRITICAL: %s", prefix, fmt.Sprintf(format, v...))
	if logger.DebugEnabled {
		debug.PrintStack()
	}
	baseLog.Fatal(str)
}

// Noticef print an notice log.
func (logger *LoggerStruct) Noticef(prefix, format string, v ...interface{}) {
	str := fmt.Sprintf("%s.NOTICE: %s", prefix, fmt.Sprintf(format, v...))
	baseLog.Print(str)
}

// Debugf print an debug log (only if DebugEnable is set to true).
func (logger *LoggerStruct) Debugf(prefix, format string, v ...interface{}) {
	if logger.DebugEnabled {
		str := fmt.Sprintf("%s.DEBUG: %s", prefix, fmt.Sprintf(format, v...))
		baseLog.Print(str)
	}
}
