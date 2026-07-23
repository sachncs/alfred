package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/alfred/alfred/internal/mcp/worker"
	"github.com/alfred/alfred/internal/tool"
)

// RunStatus represents the lifecycle state of a workflow run.
type RunStatus string

const (
	RunRunning  RunStatus = "running"
	RunComplete RunStatus = "completed"
	RunFailed   RunStatus = "failed"
)

// WorkflowStep is a single step in a workflow definition.
type WorkflowStep struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Tool      string   `json:"tool"`
	DependsOn []string `json:"dependsOn,omitempty"`
}

// WorkflowDef is a workflow definition with steps.
type WorkflowDef struct {
	ID    string         `json:"id"`
	Name  string         `json:"name"`
	Steps []WorkflowStep `json:"steps"`
}

// WorkflowRun tracks a running workflow.
type WorkflowRun struct {
	ID         string               `json:"id"`
	WorkflowID string               `json:"workflowId"`
	Status     RunStatus            `json:"status"`
	StepStates map[string]RunStatus `json:"stepStates"`
	CreatedAt  time.Time            `json:"createdAt"`
	UpdatedAt  time.Time            `json:"updatedAt"`
}

// WorkflowStore holds workflow definitions and run history.
type WorkflowStore struct {
	mu        sync.RWMutex
	workflows map[string]*WorkflowDef
	runs      map[string]*WorkflowRun
	seq       int
	runSeq    int
}

func newWorkflowStore() *WorkflowStore {
	return &WorkflowStore{
		workflows: make(map[string]*WorkflowDef),
		runs:      make(map[string]*WorkflowRun),
	}
}

func (s *WorkflowStore) register(wf *WorkflowDef) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if wf.ID == "" {
		s.seq++
		wf.ID = fmt.Sprintf("wf-%03d", s.seq)
	}
	s.workflows[wf.ID] = wf
}

func (s *WorkflowStore) get(id string) (*WorkflowDef, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	wf, ok := s.workflows[id]
	return wf, ok
}

func (s *WorkflowStore) list() []*WorkflowDef {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*WorkflowDef, 0, len(s.workflows))
	for _, wf := range s.workflows {
		out = append(out, wf)
	}
	return out
}

func (s *WorkflowStore) createRun(workflowID string) (*WorkflowRun, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	wf, ok := s.workflows[workflowID]
	if !ok {
		return nil, false
	}
	s.runSeq++
	run := &WorkflowRun{
		ID:         fmt.Sprintf("run-%03d", s.runSeq),
		WorkflowID: workflowID,
		Status:     RunRunning,
		StepStates: make(map[string]RunStatus),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	for _, step := range wf.Steps {
		run.StepStates[step.ID] = RunRunning
	}
	s.runs[run.ID] = run
	return run, true
}

func (s *WorkflowStore) getRun(id string) (*WorkflowRun, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	run, ok := s.runs[id]
	return run, ok
}

func (s *WorkflowStore) completeRun(id string, status RunStatus) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	run, ok := s.runs[id]
	if !ok {
		return false
	}
	run.Status = status
	run.UpdatedAt = time.Now()
	// Mark all steps as the same status
	for k := range run.StepStates {
		run.StepStates[k] = status
	}
	return true
}

// validateWorkflow checks a workflow definition for cycles and missing deps.
func validateWorkflow(wf *WorkflowDef) []string {
	var errs []string
	stepIDs := make(map[string]bool)
	for _, s := range wf.Steps {
		stepIDs[s.ID] = true
	}
	for _, s := range wf.Steps {
		for _, dep := range s.DependsOn {
			if !stepIDs[dep] {
				errs = append(errs, fmt.Sprintf("step %s depends on unknown step %s", s.ID, dep))
			}
		}
	}
	// Cycle detection via topological sort
	inDegree := make(map[string]int)
	adj := make(map[string][]string)
	for _, s := range wf.Steps {
		if _, ok := inDegree[s.ID]; !ok {
			inDegree[s.ID] = 0
		}
		for _, dep := range s.DependsOn {
			adj[dep] = append(adj[dep], s.ID)
			inDegree[s.ID]++
		}
	}
	var queue []string
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}
	visited := 0
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		visited++
		for _, next := range adj[curr] {
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}
	if visited < len(wf.Steps) {
		errs = append(errs, "workflow contains a cycle")
	}
	return errs
}

// NewWorkflowServer creates the workflow MCP worker.
func NewWorkflowServer() *worker.WorkerServer {
	store := newWorkflowStore()
	return worker.NewWorkerServer("workflow-worker", []tool.Tool{
		&workflowListTool{store: store},
		&workflowRunTool{store: store},
		&workflowStatusTool{store: store},
		&workflowValidateTool{store: store},
	})
}

type workflowListTool struct {
	store *WorkflowStore
}

func (t *workflowListTool) Name() string        { return "workflow_list" }
func (t *workflowListTool) Description() string { return "List available workflows" }
func (t *workflowListTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{}}`)
}
func (t *workflowListTool) Execute(_ context.Context, _ json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	wfs := t.store.list()
	return tool.SuccessWith(fmt.Sprintf("%d workflows", len(wfs)), map[string]any{
		"workflows": wfs,
	}), nil
}

type workflowRunTool struct {
	store *WorkflowStore
}

func (t *workflowRunTool) Name() string        { return "workflow_run" }
func (t *workflowRunTool) Description() string { return "Run a workflow" }
func (t *workflowRunTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"workflowId":{"type":"string"},"inputs":{"type":"object"}},"required":["workflowId"]}`)
}
func (t *workflowRunTool) Execute(_ context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		WorkflowID string `json:"workflowId"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	if in.WorkflowID == "" {
		return tool.FailureMsg("workflowId is required"), nil
	}

	run, ok := t.store.createRun(in.WorkflowID)
	if !ok {
		return tool.FailureMsg(fmt.Sprintf("workflow %q not found", in.WorkflowID)), nil
	}

	return tool.SuccessWith("Workflow started", run), nil
}

type workflowStatusTool struct {
	store *WorkflowStore
}

func (t *workflowStatusTool) Name() string        { return "workflow_status" }
func (t *workflowStatusTool) Description() string { return "Get workflow run status" }
func (t *workflowStatusTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"runId":{"type":"string"}},"required":["runId"]}`)
}
func (t *workflowStatusTool) Execute(_ context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		RunID string `json:"runId"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}

	run, ok := t.store.getRun(in.RunID)
	if !ok {
		return tool.FailureMsg(fmt.Sprintf("run %q not found", in.RunID)), nil
	}

	return tool.SuccessWith(fmt.Sprintf("Run %s: %s", run.ID, run.Status), run), nil
}

type workflowValidateTool struct {
	store *WorkflowStore
}

func (t *workflowValidateTool) Name() string        { return "workflow_validate" }
func (t *workflowValidateTool) Description() string { return "Validate a workflow definition" }
func (t *workflowValidateTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"workflow":{"type":"object"}},"required":["workflow"]}`)
}
func (t *workflowValidateTool) Execute(_ context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Workflow WorkflowDef `json:"workflow"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}

	errs := validateWorkflow(&in.Workflow)
	valid := len(errs) == 0

	// Register the workflow if valid
	if valid {
		t.store.register(&in.Workflow)
	}

	return tool.SuccessWith(fmt.Sprintf("Valid: %v", valid), map[string]any{
		"valid":      valid,
		"errors":     errs,
		"workflowId": in.Workflow.ID,
	}), nil
}
