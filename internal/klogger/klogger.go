package klogger

import (
	"fmt"
	"runtime"
	"time"

	"gitlab.com/nunet/device-management-service/telemetry"
)

// LogLevel represents different log levels
type LogLevel int

const (
	// LogLevelInfo represents the log level for informational messages
	LogLevelInfo LogLevel = iota
	// LogLevelWarning represents the log level for warning messages
	LogLevelWarning
	// LogLevelError represents the log level for error messages
	LogLevelError
)

// CustomLogger represents the custom logger struct
type CustomLogger struct {
	logLevel LogLevel
}

// GlobalLogger is a global instance of CustomLogger
var Logger *CustomLogger

// NewCustomLogger creates a new instance of CustomLogger with the specified log level
func NewCustomLogger(logLevel LogLevel) *CustomLogger {
	return &CustomLogger{logLevel: logLevel}
}

// log prints the log message with the current time, log level, file name, and line number
func (cl *CustomLogger) log(level string, message string) {
	// Use the runtime.Caller function to get the caller's information
	_, file, line, ok := runtime.Caller(2) // Adjust the stack depth accordingly
	if !ok {
		file = "???"
		line = 0
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("[%s] %s: %s (%s:%d)\n", timestamp, level, message, file, line)
}

// Info logs an informational message
func (cl *CustomLogger) Info(message string) {
	cl.log("INFO", message)
	telemetry.DmsLoggs(message, "INFO")
}

// Warning logs a warning message
func (cl *CustomLogger) Warning(message string) {
	cl.log("WARNING", message)
	telemetry.DmsLoggs(message, "Warning")
}

// Error logs an error message
func (cl *CustomLogger) Error(message string) {
	cl.log("ERROR", message)
	telemetry.DmsLoggs(message, "Error")
}

// InitializeLogger initializes the global logger with the specified log level
func InitializeLogger(logLevel LogLevel) {
	Logger = NewCustomLogger(logLevel)
}
