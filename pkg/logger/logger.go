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
	file   *os.File
}

// New creates a new logger with no output.
// Logs will be discarded unless a writer is provided.
func New() *Logger {
	return &Logger{
		logger: log.New(io.Discard, "", log.LstdFlags),
	}
}

// NewWithPath creates a new logger that writes to the specified file path.
// If the path is empty, logs are discarded.
func NewWithPath(path string) (*Logger, error) {
	if path == "" {
		return New(), nil
	}

	// Open log file (create if not exists, append mode)
	const filePermission = 0o644 // Read-write for owner, read-only for others
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, filePermission)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &Logger{
		logger: log.New(file, "", log.LstdFlags),
		file:   file,
	}, nil
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

// Close closes the logger's file if it exists.
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
