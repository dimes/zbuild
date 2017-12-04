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
	File() *os.File
}

var (
	// Output has the highest priority and is inteded to be used for chaining commands
	Output LogLevel = &logLevel{int(^uint(0) >> 1), "%s", os.Stdout}

	// Debug logs at the debug level
	Debug LogLevel = &logLevel{0, "[DEBUG] %s\n", os.Stderr}

	// Info logs at the info level
	Info LogLevel = &logLevel{1, "%s\n", os.Stderr}

	// Warning logs at the warning level
	Warning LogLevel = &logLevel{2, "[WARNING] %s\n", os.Stderr}

	// Error logs at the error level
	Error LogLevel = &logLevel{3, "[ERROR] %s\n", os.Stderr}

	// Fatal logs at the error level. It has the largest possible priority
	Fatal LogLevel = &logLevel{int(^uint(0) >> 1), "[FATAL] %s\n", os.Stderr}

	currentLevel = Info
)

type logLevel struct {
	priority int
	format   string
	file     *os.File
}

func (l *logLevel) Priority() int {
	return l.priority
}

func (l *logLevel) Format() string {
	return l.format
}

func (l *logLevel) File() *os.File {
	return l.file
}

// SetLogLevel sets the log level
func SetLogLevel(level LogLevel) {
	currentLevel = level
}

// Outputf prints the output directly
func Outputf(format string, a ...interface{}) {
	logAtLevel(Output, format, a...)
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
	fmt.Fprintf(level.File(), level.Format(), formattedMessage)
}
