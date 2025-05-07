package common

import (
	"encoding/json"
	"time"
)

// LogLevel represents the severity of a log entry
type LogLevel string

const (
	DEBUG   LogLevel = "DEBUG"
	INFO    LogLevel = "INFO"
	WARNING LogLevel = "WARNING"
	ERROR   LogLevel = "ERROR"
	FATAL   LogLevel = "FATAL"
)

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Level     LogLevel          `json:"level"`
	Service   string            `json:"service"`
	Message   string            `json:"message"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// NewLogEntry creates a new log entry with the current timestamp
func NewLogEntry(level LogLevel, service, message string, metadata map[string]string) LogEntry {
	return LogEntry{
		Timestamp: time.Now().UTC(),
		Level:     level,
		Service:   service,
		Message:   message,
		Metadata:  metadata,
	}
}

// ToJSON serializes the log entry to JSON
func (l LogEntry) ToJSON() ([]byte, error) {
	return json.Marshal(l)
}

// FromJSON deserializes a log entry from JSON
func FromJSON(data []byte) (LogEntry, error) {
	var log LogEntry
	err := json.Unmarshal(data, &log)
	return log, err
} 