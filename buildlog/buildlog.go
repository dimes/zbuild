// Package buildlog contains a logger for builds
package buildlog

import (
	"fmt"
	"os"
)

// LogLevel is an interface that determines what should be logged at various log levels
type LogLevel interface {
	Priority() int
	Format() string
}

var (
	// Debug logs at the debug level
	Debug LogLevel = &logLevel{0, "[DEBUG] %s\n"}

	// Info logs at the info level
	Info LogLevel = &logLevel{1, "%s\n"}

	// Warning logs at the warning level
	Warning LogLevel = &logLevel{2, "[WARNING] %s\n"}

	// Error logs at the error level
	Error LogLevel = &logLevel{3, "[ERROR] %s\n"}

	// Fatal logs at the error level. It has the largest possible priority
	Fatal LogLevel = &logLevel{int(^uint(0) >> 1), "[FATAL] %s\n"}

	currentLevel = Info
)

type logLevel struct {
	priority int
	format   string
}

func (l *logLevel) Priority() int {
	return l.priority
}

func (l *logLevel) Format() string {
	return l.format
}

// SetLogLevel sets the log level
func SetLogLevel(level LogLevel) {
	currentLevel = level
}

// Debugf logs at the debug level
func Debugf(format string, a ...interface{}) {
	logAtLevel(Debug, format, a...)
}

// Infof logs at the info level
func Infof(format string, a ...interface{}) {
	logAtLevel(Info, format, a...)
}

// Warningf logs at the warning level
func Warningf(format string, a ...interface{}) {
	logAtLevel(Warning, format, a...)
}

// Errorf logs at the error level
func Errorf(format string, a ...interface{}) {
	logAtLevel(Error, format, a...)
}

// Fatalf logs at the fatal level. This also causes the program to exit
func Fatalf(format string, a ...interface{}) {
	logAtLevel(Fatal, format, a...)
	os.Exit(1)
}

func logAtLevel(level LogLevel, format string, a ...interface{}) {
	if currentLevel.Priority() > level.Priority() {
		return
	}

	formattedMessage := fmt.Sprintf(format, a...)
	fmt.Fprintf(os.Stdout, level.Format(), formattedMessage)
}
