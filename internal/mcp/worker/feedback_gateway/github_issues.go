package feedbackgateway

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/alfred/alfred/internal/mcp/worker"
	"github.com/alfred/alfred/internal/tool"
)

// IssueStatus represents the lifecycle state of a GitHub Issue.
type IssueStatus string

const (
	StatusPending   IssueStatus = "pending"
	StatusCreated   IssueStatus = "created"
	StatusDuplicate IssueStatus = "duplicate"
)

// Issue represents a GitHub Issue payload.
type Issue struct {
	ID             string      `json:"id"`
	Title          string      `json:"title"`
	Body           string      `json:"body"`
	Labels         []string    `json:"labels,omitempty"`
	Status         IssueStatus `json:"status"`
	IdempotencyKey string      `json:"idempotencyKey"`
	CreatedAt      time.Time   `json:"createdAt"`
}

// IssueStore tracks submitted issues for idempotency.
type IssueStore struct {
	mu     sync.RWMutex
	issues map[string]*Issue
	keyMap map[string]string // idempotencyKey -> issueID
	seq    int
}

func newIssueStore() *IssueStore {
	return &IssueStore{
		issues: make(map[string]*Issue),
		keyMap: make(map[string]string),
	}
}

func (s *IssueStore) submit(title, body string, labels []string) *Issue {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := idempotencyKey(title, body)
	if existingID, ok := s.keyMap[key]; ok {
		existing := s.issues[existingID]
		existing.Status = StatusDuplicate
		return existing
	}

	s.seq++
	id := fmt.Sprintf("fb-%03d", s.seq)
	issue := &Issue{
		ID:             id,
		Title:          title,
		Body:           body,
		Labels:         labels,
		Status:         StatusCreated,
		IdempotencyKey: key,
		CreatedAt:      time.Now(),
	}
	s.issues[id] = issue
	s.keyMap[key] = id
	return issue
}

func (s *IssueStore) get(id string) (*Issue, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	i, ok := s.issues[id]
	return i, ok
}

func (s *IssueStore) findByKey(key string) (*Issue, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.keyMap[key]
	if !ok {
		return nil, false
	}
	return s.issues[id], true
}

// idempotencyKey generates a deterministic key from title+body.
func idempotencyKey(title, body string) string {
	h := sha256.Sum256([]byte(title + "\n" + body))
	return fmt.Sprintf("%x", h[:8])
}

// NewFeedbackGatewayServer creates the feedback-gateway MCP worker.
func NewFeedbackGatewayServer() *worker.WorkerServer {
	store := newIssueStore()
	return worker.NewWorkerServer("feedback-gateway-worker", []tool.Tool{
		&feedbackSubmitTool{store: store},
		&feedbackStatusTool{store: store},
	})
}

type feedbackSubmitTool struct {
	store *IssueStore
}

func (t *feedbackSubmitTool) Name() string { return "feedback_submit" }
func (t *feedbackSubmitTool) Description() string {
	return "Submit feedback as an idempotent GitHub Issue"
}
func (t *feedbackSubmitTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"title":{"type":"string"},"body":{"type":"string"},"labels":{"type":"array","items":{"type":"string"}}},"required":["title","body"]}`)
}
func (t *feedbackSubmitTool) Execute(_ context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Title  string   `json:"title"`
		Body   string   `json:"body"`
		Labels []string `json:"labels"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	if in.Title == "" {
		return tool.FailureMsg("title is required"), nil
	}
	if in.Body == "" {
		return tool.FailureMsg("body is required"), nil
	}

	issue := t.store.submit(in.Title, in.Body, in.Labels)

	return tool.SuccessWith(fmt.Sprintf("Issue %s: %s", issue.ID, issue.Status), map[string]any{
		"issueId": issue.ID,
		"title":   issue.Title,
		"status":  issue.Status,
		"labels":  issue.Labels,
		"payload": map[string]any{
			"title":  issue.Title,
			"body":   issue.Body,
			"labels": issue.Labels,
		},
	}), nil
}

type feedbackStatusTool struct {
	store *IssueStore
}

func (t *feedbackStatusTool) Name() string        { return "feedback_status" }
func (t *feedbackStatusTool) Description() string { return "Check feedback submission status" }
func (t *feedbackStatusTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"issueId":{"type":"string"}},"required":["issueId"]}`)
}
func (t *feedbackStatusTool) Execute(_ context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		IssueID string `json:"issueId"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}

	issue, ok := t.store.get(in.IssueID)
	if !ok {
		return tool.FailureMsg(fmt.Sprintf("issue %q not found", in.IssueID)), nil
	}

	return tool.SuccessWith(fmt.Sprintf("Issue %s: %s", issue.ID, issue.Status), issue), nil
}
