package capability

import (
	"context"
	"sync"
)

// DiscoveryEvent is emitted when a capability is registered or unregistered.
type DiscoveryEvent struct {
	Kind string // "registered" or "unregistered"
	ID   ID
}

// Observer watches capability lifecycle events.
type Observer struct {
	mu  sync.Mutex
	chs []chan DiscoveryEvent
}

// NewObserver creates an Observer.
func NewObserver() *Observer { return &Observer{} }

// Subscribe returns a channel that receives discovery events. Caller must call Unsubscribe.
func (o *Observer) Subscribe() <-chan DiscoveryEvent {
	ch := make(chan DiscoveryEvent, 16)
	o.mu.Lock()
	o.chs = append(o.chs, ch)
	o.mu.Unlock()
	return ch
}

// Unsubscribe removes a previously subscribed channel.
func (o *Observer) Unsubscribe(ch <-chan DiscoveryEvent) {
	o.mu.Lock()
	defer o.mu.Unlock()
	for i, c := range o.chs {
		if (<-chan DiscoveryEvent)(c) == ch {
			close(c)
			o.chs = append(o.chs[:i], o.chs[i+1:]...)
			return
		}
	}
}

// Notify sends an event to all subscribers. Non-blocking — drops if subscriber is full.
func (o *Observer) Notify(ev DiscoveryEvent) {
	o.mu.Lock()
	defer o.mu.Unlock()
	for _, ch := range o.chs {
		select {
		case ch <- ev:
		default:
		}
	}
}

// Discover lists all capabilities with their availability status.
type Discovery struct {
	broker *Broker
}

// NewDiscovery creates a Discovery backed by the given broker.
func NewDiscovery(b *Broker) *Discovery { return &Discovery{broker: b} }

// Capabilities returns all registered capabilities with their status.
func (d *Discovery) Capabilities(ctx context.Context) []CapabilityInfo {
	ids := d.broker.IDs()
	out := make([]CapabilityInfo, 0, len(ids))
	for _, id := range ids {
		c := d.broker.Get(id)
		out = append(out, CapabilityInfo{
			ID:        string(id),
			Available: c.IsAvailable(ctx),
		})
	}
	return out
}

// CapabilityInfo is the wire representation of a capability.
type CapabilityInfo struct {
	ID        string `json:"id"`
	Available bool   `json:"available"`
}
