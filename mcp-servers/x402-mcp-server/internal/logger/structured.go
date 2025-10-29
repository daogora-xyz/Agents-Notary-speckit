package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// Level represents log severity
type Level string

const (
	DEBUG Level = "DEBUG"
	INFO  Level = "INFO"
	WARN  Level = "WARN"
	ERROR Level = "ERROR"
)

// Logger provides structured JSON logging
type Logger struct {
	level  Level
	output io.Writer
}

// New creates a new structured logger
func New(level Level, output io.Writer) *Logger {
	if output == nil {
		output = os.Stdout
	}
	return &Logger{
		level:  level,
		output: output,
	}
}

// Entry represents a structured log entry
type Entry struct {
	Level   string                 `json:"level"`
	Time    string                 `json:"ts"`
	Message string                 `json:"msg"`
	Fields  map[string]interface{} `json:",inline,omitempty"`
}

// shouldLog determines if a message at the given level should be logged
func (l *Logger) shouldLog(level Level) bool {
	levels := map[Level]int{
		DEBUG: 0,
		INFO:  1,
		WARN:  2,
		ERROR: 3,
	}
	return levels[level] >= levels[l.level]
}

// log writes a structured log entry
func (l *Logger) log(level Level, msg string, fields map[string]interface{}) {
	if !l.shouldLog(level) {
		return
	}

	entry := Entry{
		Level:   string(level),
		Time:    time.Now().UTC().Format(time.RFC3339),
		Message: msg,
		Fields:  fields,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		// Fallback to plain text if JSON marshaling fails
		fmt.Fprintf(l.output, "[%s] %s: %s (marshal error: %v)\n",
			entry.Time, entry.Level, entry.Message, err)
		return
	}

	fmt.Fprintln(l.output, string(data))
}

// Debug logs a debug message with optional fields
func (l *Logger) Debug(msg string, fields map[string]interface{}) {
	l.log(DEBUG, msg, fields)
}

// Info logs an info message with optional fields
func (l *Logger) Info(msg string, fields map[string]interface{}) {
	l.log(INFO, msg, fields)
}

// Warn logs a warning message with optional fields
func (l *Logger) Warn(msg string, fields map[string]interface{}) {
	l.log(WARN, msg, fields)
}

// Error logs an error message with optional fields
func (l *Logger) Error(msg string, fields map[string]interface{}) {
	l.log(ERROR, msg, fields)
}

// WithFields returns a new logger with additional context fields
func (l *Logger) WithFields(fields map[string]interface{}) *ContextLogger {
	return &ContextLogger{
		logger: l,
		fields: fields,
	}
}

// ContextLogger wraps Logger with persistent context fields
type ContextLogger struct {
	logger *Logger
	fields map[string]interface{}
}

// mergeFields combines context fields with additional fields
func (cl *ContextLogger) mergeFields(additional map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range cl.fields {
		merged[k] = v
	}
	for k, v := range additional {
		merged[k] = v
	}
	return merged
}

// Debug logs with context fields
func (cl *ContextLogger) Debug(msg string, fields map[string]interface{}) {
	cl.logger.Debug(msg, cl.mergeFields(fields))
}

// Info logs with context fields
func (cl *ContextLogger) Info(msg string, fields map[string]interface{}) {
	cl.logger.Info(msg, cl.mergeFields(fields))
}

// Warn logs with context fields
func (cl *ContextLogger) Warn(msg string, fields map[string]interface{}) {
	cl.logger.Warn(msg, cl.mergeFields(fields))
}

// Error logs with context fields
func (cl *ContextLogger) Error(msg string, fields map[string]interface{}) {
	cl.logger.Error(msg, cl.mergeFields(fields))
}
