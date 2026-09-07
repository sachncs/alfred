package tool

import "encoding/json"

// Result is the envelope returned by Tool.Execute.
//
// A Result is never nil when Execute returns a nil error. The Content
// field is human-readable text suitable for inclusion in a model message.
// Structured data lives in Structured (json.RawMessage so the model can
// parse it) for tools that produce typed output.
type Result struct {
	// OK indicates success or failure. A Result with OK=false should also
	// carry Error.
	OK bool `json:"ok"`

	// Content is the human-readable text representation of the result.
	Content string `json:"content"`

	// Structured is the typed result payload, if the tool produces one.
	Structured json.RawMessage `json:"structured,omitempty"`

	// Error is the human-readable error message when OK=false.
	Error string `json:"error,omitempty"`

	// Metadata is free-form additional context (line counts, byte counts,
	// detected MIME types, etc.). The runtime may surface this in traces.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Success returns a successful Result with the given content.
func Success(content string) *Result {
	return &Result{OK: true, Content: content}
}

// SuccessWith returns a successful Result with content and structured data.
func SuccessWith(content string, structured any) *Result {
	r := &Result{OK: true, Content: content}
	if structured != nil {
		if b, err := json.Marshal(structured); err == nil {
			r.Structured = b
		}
	}
	return r
}

// Failure returns a failed Result with the given error message.
func Failure(err error) *Result {
	r := &Result{OK: false}
	if err != nil {
		r.Error = err.Error()
	}
	return r
}

// FailureMsg returns a failed Result with a plain error message.
func FailureMsg(msg string) *Result {
	return &Result{OK: false, Error: msg}
}

// WithMetadata returns a copy of the Result with the given metadata added.
func (r *Result) WithMetadata(kv map[string]any) *Result {
	clone := *r
	if clone.Metadata == nil {
		clone.Metadata = map[string]any{}
	}
	for k, v := range kv {
		clone.Metadata[k] = v
	}
	return &clone
}
