package capability

import (
	"context"
	"encoding/json"
	"errors"
)

// Resource represents a managed external resource (file handle, surface, etc.).
type Resource interface {
	// Release frees the resource. After Release, TakeResource returns ErrReleased.
	Release() error
	// IsReleased reports whether the resource has been freed.
	IsReleased() bool
}

// ResourceCapability extends Capability with resource lifecycle management.
type ResourceCapability interface {
	Capability
	// Bind attaches a resource handle. Returns error if already bound.
	Bind(ctx context.Context, handle any) error
	// Unbind releases the current resource. No-op if not bound.
	Unbind() error
	// TakeResource returns the current resource, or ErrNotBound if none.
	TakeResource(ctx context.Context) (Resource, error)
	// IsBound reports whether a resource is currently bound.
	IsBound() bool
}

// ErrNotBound is returned when TakeResource is called without a bound resource.
var ErrNotBound = errors.New("capability: not bound")

// ErrAlreadyBound is returned when Bind is called on an already-bound capability.
var ErrAlreadyBound = errors.New("capability: already bound")

// BaseResource provides common resource lifecycle for embedding.
type BaseResource struct {
	resource Resource
}

// Bind attaches a resource.
func (b *BaseResource) Bind(_ context.Context, handle any) error {
	if b.resource != nil {
		return ErrAlreadyBound
	}
	r, ok := handle.(Resource)
	if !ok {
		// ponytail: accept any handle, wrap in genericResource
		r = &genericResource{value: handle}
	}
	b.resource = r
	return nil
}

// Unbind releases the resource.
func (b *BaseResource) Unbind() error {
	if b.resource == nil {
		return nil
	}
	err := b.resource.Release()
	b.resource = nil
	return err
}

// TakeResource returns the bound resource.
func (b *BaseResource) TakeResource(_ context.Context) (Resource, error) {
	if b.resource == nil {
		return nil, ErrNotBound
	}
	return b.resource, nil
}

// IsBound reports whether a resource is bound.
func (b *BaseResource) IsBound() bool { return b.resource != nil }

// genericResource wraps an arbitrary value as a Resource.
type genericResource struct {
	value    any
	released bool
}

func (g *genericResource) Release() error   { g.released = true; return nil }
func (g *genericResource) IsReleased() bool { return g.released }

// Ensure BaseResource satisfies Resource.
var _ Resource = (*genericResource)(nil)

// DispatchResult is the wire format for capability dispatch responses.
type DispatchResult struct {
	Output json.RawMessage `json:"output,omitempty"`
	Error  string          `json:"error,omitempty"`
}
