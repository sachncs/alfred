package log

import (
	"io"
	"regexp"
)

// Patterns that look like secrets.
var secretPatterns = []*regexp.Regexp{
	// OpenAI / generic API keys: sk-... (32+ chars)
	regexp.MustCompile(`(sk-[A-Za-z0-9]{20,})`),
	// Bearer tokens: Bearer <token>
	regexp.MustCompile(`(Bearer\s+[A-Za-z0-9\-._~+/]+=*)`),
	// key= query params
	regexp.MustCompile(`(key=[A-Za-z0-9\-._~+/]+=*)`),
	// Generic "token" assignments: token=... or token: "..."
	regexp.MustCompile(`((?:token|secret|password|api_?key)\s*[:=]\s*)[^\s,;}{]+`),
}

// RedactingWriter wraps an io.Writer and replaces secret-looking
// substrings with "[REDACTED]" before forwarding.
type RedactingWriter struct {
	w io.Writer
}

// NewRedactingWriter wraps w with secret redaction.
func NewRedactingWriter(w io.Writer) *RedactingWriter {
	return &RedactingWriter{w: w}
}

// Write redacts secrets then forwards to the underlying writer.
func (rw *RedactingWriter) Write(p []byte) (int, error) {
	redacted := Redact(string(p))
	return rw.w.Write([]byte(redacted))
}

// Redact replaces secret-looking substrings in s with "[REDACTED]".
func Redact(s string) string {
	result := s
	for _, re := range secretPatterns {
		result = re.ReplaceAllStringFunc(result, func(match string) string {
			// For patterns with a capture group prefix (key=, token:, etc.),
			// preserve the prefix and redact only the value.
			for _, sub := range re.FindStringSubmatch(result) {
				if sub != "" && sub != match {
					// There's a capture group — find the prefix length.
					idx := len(sub)
					if idx < len(match) {
						return sub + "[REDACTED]"
					}
				}
			}
			return "[REDACTED]"
		})
	}
	return result
}
