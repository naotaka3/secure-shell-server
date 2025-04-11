package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// Logger provides logging functionality.
type Logger struct {
	logger *log.Logger
}

// New creates a new logger.
func New() *Logger {
	return &Logger{
		logger: log.New(os.Stderr, "", log.LstdFlags),
	}
}

// NewWithWriter creates a new logger with a specific writer.
func NewWithWriter(w io.Writer) *Logger {
	return &Logger{
		logger: log.New(w, "", log.LstdFlags),
	}
}

// LogCommandAttempt logs an attempted command execution.
func (l *Logger) LogCommandAttempt(cmd string, args []string, allowed bool) {
	status := "ALLOWED"
	if !allowed {
		status = "BLOCKED"
	}

	timestamp := time.Now().Format(time.RFC3339)
	l.logger.Printf("%s [%s] Command: %s %v\n", timestamp, status, cmd, args)
}

// LogErrorf logs an error with formatted message.
func (l *Logger) LogErrorf(format string, args ...interface{}) {
	timestamp := time.Now().Format(time.RFC3339)
	message := fmt.Sprintf(format, args...)
	l.logger.Printf("%s [ERROR] %s\n", timestamp, message)
}

// LogError logs an error message.
func (l *Logger) LogError(message string) {
	timestamp := time.Now().Format(time.RFC3339)
	l.logger.Printf("%s [ERROR] %s\n", timestamp, message)
}

// LogInfof logs an informational message with formatting.
func (l *Logger) LogInfof(format string, args ...interface{}) {
	timestamp := time.Now().Format(time.RFC3339)
	message := fmt.Sprintf(format, args...)
	l.logger.Printf("%s [INFO] %s\n", timestamp, message)
}

// LogInfo logs an informational message.
func (l *Logger) LogInfo(message string) {
	timestamp := time.Now().Format(time.RFC3339)
	l.logger.Printf("%s [INFO] %s\n", timestamp, message)
}
