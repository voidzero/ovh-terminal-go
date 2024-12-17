// internal/logger/logger.go
package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// LogLevel represents logging severity
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

var levelStrings = map[LogLevel]string{
	DebugLevel: "DEBUG",
	InfoLevel:  "INFO",
	WarnLevel:  "WARN",
	ErrorLevel: "ERROR",
}

var levelFromString = map[string]LogLevel{
	"debug": DebugLevel,
	"info":  InfoLevel,
	"warn":  WarnLevel,
	"error": ErrorLevel,
}

// Logger wraps the logging functionality
type Logger struct {
	level      LogLevel
	fileLogger *log.Logger
	console    bool
	fields     map[string]interface{}
}

var Log *Logger

func init() {
	Log = NewLogger()
}

// LoggerOption defines a function that configures a Logger
type LoggerOption func(*Logger)

// WithFields adds default fields to every log entry
func WithFields(fields map[string]interface{}) LoggerOption {
	return func(l *Logger) {
		for k, v := range fields {
			l.fields[k] = v
		}
	}
}

// WithConsole enables console logging
func WithConsole(enabled bool) LoggerOption {
	return func(l *Logger) {
		l.console = enabled
	}
}

// NewLogger creates a new logger instance with options
func NewLogger(opts ...LoggerOption) *Logger {
	l := &Logger{
		level:  InfoLevel,
		fields: make(map[string]interface{}),
	}

	for _, opt := range opts {
		opt(l)
	}

	return l
}

// Configure sets up the logger based on configuration
func (l *Logger) Configure(level string, logFile string, useConsole bool) error {
	// Set log level
	if lvl, ok := levelFromString[level]; ok {
		l.level = lvl
	} else {
		return fmt.Errorf("invalid log level: %s", level)
	}

	// Setup file logging if specified
	if logFile != "" && logFile != "none" {
		if err := os.MkdirAll(filepath.Dir(logFile), 0o755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}

		file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}

		l.fileLogger = log.New(file, "", 0) // We'll format the prefix ourselves
	}

	l.console = useConsole
	return nil
}

// With creates a new logger with additional fields
func (l *Logger) With(fields map[string]interface{}) *Logger {
	newLogger := &Logger{
		level:      l.level,
		fileLogger: l.fileLogger,
		console:    l.console,
		fields:     make(map[string]interface{}),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new fields
	for k, v := range fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// formatMessage creates a structured log message
func (l *Logger) formatMessage(level, msg string, keyvals ...interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")

	// Start with timestamp and level
	result := fmt.Sprintf("[%s] %-5s: %s", timestamp, level, msg)

	// Add default fields
	for k, v := range l.fields {
		result += fmt.Sprintf(" %v=%v", k, v)
	}

	// Add additional key-value pairs
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			result += fmt.Sprintf(" %v=%v", keyvals[i], keyvals[i+1])
		}
	}

	return result
}

// Log methods
func (l *Logger) log(level LogLevel, msg string, keyvals ...interface{}) {
	if l.level <= level {
		logLine := l.formatMessage(levelStrings[level], msg, keyvals...)

		if l.fileLogger != nil {
			l.fileLogger.Println(logLine)
		}

		if l.console {
			fmt.Fprintln(os.Stderr, logLine)
		}
	}
}

func (l *Logger) Debug(msg string, keyvals ...interface{}) {
	l.log(DebugLevel, msg, keyvals...)
}

func (l *Logger) Info(msg string, keyvals ...interface{}) {
	l.log(InfoLevel, msg, keyvals...)
}

func (l *Logger) Warn(msg string, keyvals ...interface{}) {
	l.log(WarnLevel, msg, keyvals...)
}

func (l *Logger) Error(msg string, keyvals ...interface{}) {
	l.log(ErrorLevel, msg, keyvals...)
}

