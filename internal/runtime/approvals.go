package runtime

import (
	"fmt"
	"sync"
	"time"
)

// ApprovalRequest is a pending tool approval.
type ApprovalRequest struct {
	ID        string         `json:"id"`
	ThreadID  string         `json:"threadId"`
	ToolName  string         `json:"toolName"`
	Reason    string         `json:"reason,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
	Resolved  bool           `json:"resolved"`
	Approved  bool           `json:"approved"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// ApprovalStore is an in-process store for tool approval requests.
type ApprovalStore struct {
	mu      sync.RWMutex
	counter int
	pending map[string]*ApprovalRequest
}

// NewApprovalStore creates an empty approval store.
func NewApprovalStore() *ApprovalStore {
	return &ApprovalStore{pending: make(map[string]*ApprovalRequest)}
}

// Pending creates a new approval request and returns it.
func (s *ApprovalStore) Pending(threadID, toolName, reason string, meta map[string]any) *ApprovalRequest {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counter++
	req := &ApprovalRequest{
		ID:        fmt.Sprintf("appr-%d", s.counter),
		ThreadID:  threadID,
		ToolName:  toolName,
		Reason:    reason,
		CreatedAt: time.Now().UTC(),
		Resolved:  false,
		Metadata:  meta,
	}
	s.pending[req.ID] = req
	return req
}

// Resolve marks an approval as resolved.
func (s *ApprovalStore) Resolve(id string, approved bool) (*ApprovalRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	req, ok := s.pending[id]
	if !ok {
		return nil, fmt.Errorf("approval %q not found", id)
	}
	if req.Resolved {
		return nil, fmt.Errorf("approval %q already resolved", id)
	}
	req.Resolved = true
	req.Approved = approved
	return req, nil
}

// Get returns a single approval by ID.
func (s *ApprovalStore) Get(id string) (*ApprovalRequest, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	req, ok := s.pending[id]
	return req, ok
}

// List returns all pending approvals.
func (s *ApprovalStore) List() []*ApprovalRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*ApprovalRequest, 0, len(s.pending))
	for _, r := range s.pending {
		out = append(out, r)
	}
	return out
}
