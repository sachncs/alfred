// Package capability defines the Capability interface and the Broker
// that routes calls to registered capabilities.
package capability

import (
	"context"
	"encoding/json"
)

// ID is the stable, machine-readable identifier of a capability.
type ID string

// Capability is the base interface every concrete capability must satisfy.
//
// Implementations may be anything from a simple function wrapper to a
// complex file-system or surface-inspection resource.
type Capability interface {
	// ID returns the stable identifier used to register and dispatch.
	ID() ID

	// IsAvailable reports whether the capability can be invoked right
	// now (resources loaded, deps ready, etc.). Used by the renderer
	// to grey out unavailable actions.
	IsAvailable(ctx context.Context) bool

	// Invoke runs the capability with the given JSON input. Returns the
	// JSON-serializable result. Errors are wrapped with %w.
	Invoke(ctx context.Context, input json.RawMessage) (json.RawMessage, error)
}

// Function wraps a plain function as a Capability for simple cases.
type Function struct {
	id         ID
	available  func(ctx context.Context) bool
	invoke     func(ctx context.Context, input json.RawMessage) (json.RawMessage, error)
}

// NewFunction constructs a Function capability.
func NewFunction(id ID, invoke func(ctx context.Context, input json.RawMessage) (json.RawMessage, error)) *Function {
	return &Function{id: id, available: func(_ context.Context) bool { return true }, invoke: invoke}
}

// WithAvailability attaches an availability predicate.
func (f *Function) WithAvailability(fn func(ctx context.Context) bool) *Function {
	f.available = fn
	return f
}

// ID implements Capability.
func (f *Function) ID() ID { return f.id }

// IsAvailable implements Capability.
func (f *Function) IsAvailable(ctx context.Context) bool {
	if f.available == nil {
		return true
	}
	return f.available(ctx)
}

// Invoke implements Capability.
func (f *Function) Invoke(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	return f.invoke(ctx, input)
}
