package capability

import (
	"context"
	"fmt"
	"sync"
)

// Binding tracks a capability-to-resource binding lifecycle.
type Binding struct {
	mu        sync.RWMutex
	bindings  map[ID]Resource
	capBroker *Broker
}

// NewBinding creates a Binding manager.
func NewBinding(b *Broker) *Binding {
	return &Binding{bindings: make(map[ID]Resource), capBroker: b}
}

// Bind binds a resource to a capability. The capability must be a ResourceCapability.
func (b *Binding) Bind(ctx context.Context, id ID, handle any) error {
	c := b.capBroker.Get(id)
	if c == nil {
		return ErrUnknownCapability
	}
	rc, ok := c.(ResourceCapability)
	if !ok {
		return fmt.Errorf("capability %s does not support resource binding", id)
	}
	if err := rc.Bind(ctx, handle); err != nil {
		return err
	}
	r, _ := rc.TakeResource(ctx)
	b.mu.Lock()
	b.bindings[id] = r
	b.mu.Unlock()
	return nil
}

// Unbind releases the resource bound to a capability.
func (b *Binding) Unbind(id ID) error {
	c := b.capBroker.Get(id)
	if c == nil {
		return ErrUnknownCapability
	}
	rc, ok := c.(ResourceCapability)
	if !ok {
		return nil
	}
	if err := rc.Unbind(); err != nil {
		return err
	}
	b.mu.Lock()
	delete(b.bindings, id)
	b.mu.Unlock()
	return nil
}

// IsBound reports whether a capability has a bound resource.
func (b *Binding) IsBound(id ID) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	_, ok := b.bindings[id]
	return ok
}

// ReleaseAll releases all bound resources.
func (b *Binding) ReleaseAll() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for id, r := range b.bindings {
		_ = r.Release()
		delete(b.bindings, id)
	}
}
