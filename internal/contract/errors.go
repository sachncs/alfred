package contract

// Error is the wire envelope for any error returned over HTTP or IPC.
// Code is a stable machine-readable string (e.g. "thread_not_found").
// Message is human-readable and may change.
// Details is optional structured context.
type Error struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// Standard error codes used across the runtime.
const (
	CodeUnauthorized   = "unauthorized"
	CodeForbidden      = "forbidden"
	CodeNotFound       = "not_found"
	CodeValidation     = "validation"
	CodeTurnInProgress = "turn_in_progress"
	CodeUnavailable    = "unavailable"
	CodeInternal       = "internal"
)

// Errorf constructs an Error with a code and message.
func Errorf(code, message string) Error {
	return Error{Code: code, Message: message}
}

// ErrorfWith constructs an Error with code, message, and details.
func ErrorfWith(code, message string, details map[string]any) Error {
	return Error{Code: code, Message: message, Details: details}
}
