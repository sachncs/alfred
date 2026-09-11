package runtime

import (
	"fmt"
	"net/http"

	"github.com/sachncs/alfred/internal/contract"
)

// UserInputRequest is a pending mid-turn user input question.
type UserInputRequest struct {
	ID        string         `json:"id"`
	ThreadID  string         `json:"threadId"`
	Prompt    string         `json:"prompt"`
	Resolved  bool           `json:"resolved"`
	Value     string         `json:"value,omitempty"`
	CreatedAt int64          `json:"createdAt"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// UserInputStore tracks pending mid-turn user input requests.
type UserInputStore struct {
	*ApprovalStore
	values map[string]string
}

// NewUserInputStore creates a user-input store.
func NewUserInputStore() *UserInputStore {
	return &UserInputStore{
		ApprovalStore: NewApprovalStore(),
		values:        make(map[string]string),
	}
}

// Pending creates a new user input request.
func (s *UserInputStore) Pending(threadID, prompt string, meta map[string]any) *ApprovalRequest {
	req := s.ApprovalStore.Pending(threadID, "user_input", prompt, meta)
	s.values[req.ID] = ""
	return req
}

// Resolve stores the user's value and marks resolved.
func (s *UserInputStore) ResolveWithValue(id, value string) (*ApprovalRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	req, ok := s.pending[id]
	if !ok {
		return nil, errNotFoundApproval(id)
	}
	if req.Resolved {
		return nil, errAlreadyResolved(id)
	}
	req.Resolved = true
	req.Approved = true
	s.values[id] = value
	return req, nil
}

// GetValue returns the resolved value for an id.
func (s *UserInputStore) GetValue(id string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.values[id]
	return v, ok
}

func errNotFoundApproval(id string) error { return fmt.Errorf("user input %q not found", id) }
func errAlreadyResolved(id string) error  { return fmt.Errorf("user input %q already resolved", id) }

func (r *LocalRuntime) handleResolveUserInput(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	var in struct {
		Value string `json:"value"`
	}
	if err := RouteDecode(req, &in); err != nil {
		RouteError(w, contract.CodeValidation, err.Error(), 400)
		return
	}
	resolved, err := r.userInputs.ResolveWithValue(id, in.Value)
	if err != nil {
		RouteError(w, contract.CodeNotFound, err.Error(), 404)
		return
	}
	RouteJSON(w, map[string]any{"resolved": true, "value": in.Value, "approval": resolved}, 200)
}
