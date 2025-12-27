// Package log provides a simple internal logging library with no external dependencies.
// It supports multiple log levels and structured logging with key-value pairs.
package log

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Level represents the severity level of a log message.
type Level int

const (
	// DebugLevel is for detailed debugging information.
	DebugLevel Level = iota
	// InfoLevel is for general informational messages.
	InfoLevel
	// WarnLevel is for warning messages about potential issues.
	WarnLevel
	// ErrorLevel is for error messages.
	ErrorLevel
)

// String returns the string representation of a log level.
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger represents a logger instance.
type Logger struct {
	mu     sync.Mutex
	out    io.Writer
	level  Level
	prefix string
}

// global logger instance
var std = &Logger{
	out:   os.Stderr,
	level: WarnLevel, // Default to Warn level
}

// New creates a new Logger instance.
func New(out io.Writer, level Level) *Logger {
	return &Logger{
		out:   out,
		level: level,
	}
}

// SetOutput sets the output destination for the logger.
func SetOutput(w io.Writer) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.out = w
}

// SetLevel sets the minimum log level for the logger.
func SetLevel(level Level) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.level = level
}

// GetLevel returns the current log level.
func GetLevel() Level {
	std.mu.Lock()
	defer std.mu.Unlock()
	return std.level
}

// SetPrefix sets a prefix for all log messages.
func SetPrefix(prefix string) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.prefix = prefix
}

// log is the internal logging function.
func (l *Logger) log(level Level, msg string, fields map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	
	// Build the log message
	var output string
	if l.prefix != "" {
		output = fmt.Sprintf("[%s] %s [%s] %s", timestamp, l.prefix, level.String(), msg)
	} else {
		output = fmt.Sprintf("[%s] [%s] %s", timestamp, level.String(), msg)
	}

	// Add structured fields
	if len(fields) > 0 {
		for k, v := range fields {
			output += fmt.Sprintf(" %s=%v", k, v)
		}
	}

	fmt.Fprintln(l.out, output)
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string) {
	l.log(DebugLevel, msg, nil)
}

// Debugf logs a formatted debug message.
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(DebugLevel, fmt.Sprintf(format, args...), nil)
}

// Info logs an info message.
func (l *Logger) Info(msg string) {
	l.log(InfoLevel, msg, nil)
}

// Infof logs a formatted info message.
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(InfoLevel, fmt.Sprintf(format, args...), nil)
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string) {
	l.log(WarnLevel, msg, nil)
}

// Warnf logs a formatted warning message.
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(WarnLevel, fmt.Sprintf(format, args...), nil)
}

// Error logs an error message.
func (l *Logger) Error(msg string) {
	l.log(ErrorLevel, msg, nil)
}

// Errorf logs a formatted error message.
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(ErrorLevel, fmt.Sprintf(format, args...), nil)
}

// WithFields logs a message with structured key-value fields.
func (l *Logger) WithFields(level Level, msg string, fields map[string]interface{}) {
	l.log(level, msg, fields)
}

// Global logging functions that use the standard logger

// Debug logs a debug message using the standard logger.
func Debug(msg string) {
	std.log(DebugLevel, msg, nil)
}

// Debugf logs a formatted debug message using the standard logger.
func Debugf(format string, args ...interface{}) {
	std.log(DebugLevel, fmt.Sprintf(format, args...), nil)
}

// Info logs an info message using the standard logger.
func Info(msg string) {
	std.log(InfoLevel, msg, nil)
}

// Infof logs a formatted info message using the standard logger.
func Infof(format string, args ...interface{}) {
	std.log(InfoLevel, fmt.Sprintf(format, args...), nil)
}

// Warn logs a warning message using the standard logger.
func Warn(msg string) {
	std.log(WarnLevel, msg, nil)
}

// Warnf logs a formatted warning message using the standard logger.
func Warnf(format string, args ...interface{}) {
	std.log(WarnLevel, fmt.Sprintf(format, args...), nil)
}

// Error logs an error message using the standard logger.
func Error(msg string) {
	std.log(ErrorLevel, msg, nil)
}

// Errorf logs a formatted error message using the standard logger.
func Errorf(format string, args ...interface{}) {
	std.log(ErrorLevel, fmt.Sprintf(format, args...), nil)
}

// WithFields logs a message with structured key-value fields using the standard logger.
func WithFields(level Level, msg string, fields map[string]interface{}) {
	std.log(level, msg, fields)
}
