package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// LogLevel represents logging severity
type LogLevel int

var Log *Logger

func init() {
	Log = NewLogger()
}

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// Logger wraps the logging functionality
type Logger struct {
	level      LogLevel
	fileLogger *log.Logger
	console    bool
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	return &Logger{
		level:   InfoLevel,
		console: true,
	}
}

// Configure sets up the logger based on configuration
func (l *Logger) Configure(level string, logFile string, useConsole bool) error {
	// Set log level
	switch level {
	case "debug":
		l.level = DebugLevel
	case "info":
		l.level = InfoLevel
	case "warn":
		l.level = WarnLevel
	case "error":
		l.level = ErrorLevel
	default:
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

		l.fileLogger = log.New(file, "", log.LstdFlags)
	}

	l.console = useConsole
	return nil
}

// Log methods
func (l *Logger) Debug(msg string, keyvals ...interface{}) {
	if l.level <= DebugLevel {
		l.log("DEBUG", msg, keyvals...)
	}
}

func (l *Logger) Info(msg string, keyvals ...interface{}) {
	if l.level <= InfoLevel {
		l.log("INFO", msg, keyvals...)
	}
}

func (l *Logger) Warn(msg string, keyvals ...interface{}) {
	if l.level <= WarnLevel {
		l.log("WARN", msg, keyvals...)
	}
}

func (l *Logger) Error(msg string, keyvals ...interface{}) {
	if l.level <= ErrorLevel {
		l.log("ERROR", msg, keyvals...)
	}
}

func (l *Logger) log(level, msg string, keyvals ...interface{}) {
	// Format key-value pairs
	var kvStr string
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			kvStr += fmt.Sprintf(" %v=%v", keyvals[i], keyvals[i+1])
		}
	}

	logLine := fmt.Sprintf("%s: %s%s", level, msg, kvStr)

	// Write to file if configured
	if l.fileLogger != nil {
		l.fileLogger.Println(logLine)
	}

	// Write to console if enabled
	if l.console {
		fmt.Fprintln(os.Stderr, logLine)
	}
}

