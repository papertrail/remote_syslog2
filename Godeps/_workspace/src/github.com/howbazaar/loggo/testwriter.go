package loggo

import (
	"path"
	"time"
)

// TestLogValues represents a single logging call.
type TestLogValues struct {
	Level     Level
	Module    string
	Filename  string
	Line      int
	Timestamp time.Time
	Message   string
}

// TestWriter is a useful Writer for testing purposes.  Each component of the
// logging message is stored in the Log array.
type TestWriter struct {
	Log []TestLogValues
}

// Write saves the params as members in the TestLogValues struct appended to the Log array.
func (writer *TestWriter) Write(level Level, module, filename string, line int, timestamp time.Time, message string) {
	if writer.Log == nil {
		writer.Log = []TestLogValues{}
	}
	writer.Log = append(writer.Log,
		TestLogValues{level, module, path.Base(filename), line, timestamp, message})
}

// Clear removes any saved log messages.
func (writer *TestWriter) Clear() {
	writer.Log = []TestLogValues{}
}
