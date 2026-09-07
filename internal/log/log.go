// Package log provides a thin logging wrapper that redacts secrets.
package log

import (
	"io"
	"log"
	"os"
)

// RedactingLogger wraps the stdlib logger with a redacting writer.
type RedactingLogger struct {
	logger *log.Logger
}

// New creates a RedactingLogger that writes to os.Stderr with redaction.
func New() *RedactingLogger {
	return &RedactingLogger{
		logger: log.New(NewRedactingWriter(os.Stderr), "", log.LstdFlags|log.Lmicroseconds),
	}
}

// NewWithWriter creates a RedactingLogger writing to the given writer.
func NewWithWriter(w io.Writer) *RedactingLogger {
	return &RedactingLogger{
		logger: log.New(NewRedactingWriter(w), "", log.LstdFlags|log.Lmicroseconds),
	}
}

// Printf logs a formatted message.
func (l *RedactingLogger) Printf(format string, v ...any) {
	l.logger.Printf(format, v...)
}

// Println logs a message.
func (l *RedactingLogger) Println(v ...any) {
	l.logger.Println(v...)
}

// Print logs a message (no newline).
func (l *RedactingLogger) Print(v ...any) {
	l.logger.Print(v...)
}

// Writer returns the underlying io.Writer (redacting).
func (l *RedactingLogger) Writer() io.Writer {
	return l.logger.Writer()
}
