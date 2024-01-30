package logger

import (
	"fmt"
	"os"
	"sync"
	"time"
)

const (
	// Log levels
	InfoLevel  = "INFO"
	WarnLevel  = "WARN"
	ErrorLevel = "ERROR"

	// Log file
	logFileName = "app.log"
)

var (
	once     sync.Once
	instance *Logger
)

// Logger represents the singleton logger instance.
type Logger struct {
	file *os.File
}

// initializeLogger initializes the logger instance.
func initializeLogger() {
	file, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
	}

	instance = &Logger{file: file}
}

// GetLogger returns the singleton logger instance.
func GetLogger() *Logger {
	once.Do(initializeLogger)
	return instance
}

// log prints the log message to both console and file.
func (l *Logger) log(level, message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] %s: %s\n", timestamp, level, message)

	// Print to console with colorizing
	switch level {
	case InfoLevel:
		fmt.Printf("\033[1;34m%s\033[0m", logEntry) // Blue
	case WarnLevel:
		fmt.Printf("\033[1;33m%s\033[0m", logEntry) // Yellow
	case ErrorLevel:
		fmt.Printf("\033[1;31m%s\033[0m", logEntry) // Red
	default:
		fmt.Print(logEntry)
	}

	// Write to log file
	if l.file != nil {
		_, _ = l.file.WriteString(logEntry)
	}
}

// Info logs an informational message.
func (l *Logger) Info(message string) {
	l.log(InfoLevel, message)
}

// Warn logs a warning message.
func (l *Logger) Warn(message string) {
	l.log(WarnLevel, message)
}

// Error logs an error message.
func (l *Logger) Error(message string) {
	l.log(ErrorLevel, message)
}

// CloseFile closes the log file.
func (l *Logger) CloseFile() {
	if l.file != nil {
		_ = l.file.Close()
	}
}
