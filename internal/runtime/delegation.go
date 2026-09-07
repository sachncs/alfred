package runtime

import (
	"sync"
)

// AgentHandle identifies a sub-agent.
type AgentHandle struct {
	ID   string
	Name string
}

// Delegator manages sub-agent delegation.
type Delegator struct {
	mu     sync.Mutex
	nextID int
	agents map[string]AgentHandle
}

// NewDelegator creates a Delegator.
func NewDelegator() *Delegator {
	return &Delegator{agents: make(map[string]AgentHandle)}
}

// Delegate spawns a sub-agent and returns its handle.
func (d *Delegator) Delegate(name string) AgentHandle {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.nextID++
	id := string(rune('0' + d.nextID))
	handle := AgentHandle{ID: id, Name: name}
	d.agents[id] = handle
	return handle
}

// Get returns the agent handle if it exists.
func (d *Delegator) Get(id string) (AgentHandle, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	h, ok := d.agents[id]
	return h, ok
}

// List returns all known sub-agents.
func (d *Delegator) List() []AgentHandle {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := make([]AgentHandle, 0, len(d.agents))
	for _, h := range d.agents {
		out = append(out, h)
	}
	return out
}
