package capability

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

// ErrUnknownCapability is returned by Broker.Dispatch when no capability
// is registered under the requested ID.
var ErrUnknownCapability = errors.New("capability: unknown id")

// ErrUnavailable is returned when the capability exists but reports
// IsAvailable == false.
var ErrUnavailable = errors.New("capability: unavailable")

// Broker routes capability invocations to registered capabilities.
//
// Broker is safe for concurrent use.
type Broker struct {
	mu sync.RWMutex
	m  map[ID]Capability
}

// NewBroker constructs an empty Broker.
func NewBroker() *Broker {
	return &Broker{m: map[ID]Capability{}}
}

// Register adds (or replaces) a capability.
func (b *Broker) Register(c Capability) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.m[c.ID()] = c
}

// Unregister removes a capability. Returns true if it was present.
func (b *Broker) Unregister(id ID) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	_, ok := b.m[id]
	if ok {
		delete(b.m, id)
	}
	return ok
}

// IDs returns the list of registered capability IDs (defensive copy).
func (b *Broker) IDs() []ID {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]ID, 0, len(b.m))
	for id := range b.m {
		out = append(out, id)
	}
	return out
}

// Get returns the capability registered under id, or nil.
func (b *Broker) Get(id ID) Capability {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.m[id]
}

// Dispatch routes a call to the registered capability. Returns
// ErrUnknownCapability if the id is not registered, ErrUnavailable if
// the capability exists but reports unavailable.
func (b *Broker) Dispatch(ctx context.Context, id ID, input json.RawMessage) (json.RawMessage, error) {
	b.mu.RLock()
	c, ok := b.m[id]
	b.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownCapability, id)
	}
	if !c.IsAvailable(ctx) {
		return nil, fmt.Errorf("%w: %s", ErrUnavailable, id)
	}
	return c.Invoke(ctx, input)
}
