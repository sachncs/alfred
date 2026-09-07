package log

import (
	"bytes"
	"strings"
	"testing"
)

func TestRedactSkKey(t *testing.T) {
	input := "api key: sk-proj123456789012345678901234567890"
	got := Redact(input)
	if strings.Contains(got, "sk-proj") {
		t.Errorf("sk key not redacted: %s", got)
	}
	if !strings.Contains(got, "[REDACTED]") {
		t.Errorf("expected [REDACTED] in output: %s", got)
	}
}

func TestRedactBearerToken(t *testing.T) {
	input := "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
	got := Redact(input)
	if strings.Contains(got, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9") {
		t.Errorf("bearer token not redacted: %s", got)
	}
	if !strings.Contains(got, "[REDACTED]") {
		t.Errorf("expected [REDACTED] in output: %s", got)
	}
}

func TestRedactKeyEquals(t *testing.T) {
	input := "url?key=abc123secret456"
	got := Redact(input)
	if strings.Contains(got, "abc123secret456") {
		t.Errorf("key= value not redacted: %s", got)
	}
}

func TestRedactTokenAssignment(t *testing.T) {
	input := `token: "my-secret-token-value"`
	got := Redact(input)
	if strings.Contains(got, "my-secret-token-value") {
		t.Errorf("token assignment not redacted: %s", got)
	}
}

func TestRedactNormalTextUntouched(t *testing.T) {
	input := "hello world, this is a normal log message"
	got := Redact(input)
	if got != input {
		t.Errorf("normal text was modified: got %q, want %q", got, input)
	}
}

func TestRedactingWriter(t *testing.T) {
	var buf bytes.Buffer
	w := NewRedactingWriter(&buf)
	msg := "request with key=secret123\n"
	_, err := w.Write([]byte(msg))
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}
	got := buf.String()
	if strings.Contains(got, "secret123") {
		t.Errorf("RedactingWriter did not redact: %s", got)
	}
}

func TestLoggerPrintf(t *testing.T) {
	var buf bytes.Buffer
	l := NewWithWriter(&buf)
	l.Printf("port %d", 8899)
	got := buf.String()
	if !strings.Contains(got, "port 8899") {
		t.Errorf("log output missing message: %s", got)
	}
}
