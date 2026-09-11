package multiagent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sachncs/alfred/internal/mcp/worker"
	"github.com/sachncs/alfred/internal/tool"
)

// TaskStatus represents the lifecycle state of a delegated task.
type TaskStatus string

const (
	StatusQueued   TaskStatus = "queued"
	StatusRunning  TaskStatus = "running"
	StatusComplete TaskStatus = "completed"
	StatusFailed   TaskStatus = "failed"
)

// ChildTask represents a delegated sub-agent task.
type ChildTask struct {
	ID        string     `json:"id"`
	Task      string     `json:"task"`
	AgentType string     `json:"agentType"`
	Status    TaskStatus `json:"status"`
	Result    string     `json:"result,omitempty"`
	Error     string     `json:"error,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// TaskStore is a thread-safe in-memory store for delegated tasks.
type TaskStore struct {
	mu    sync.RWMutex
	tasks map[string]*ChildTask
	seq   int
}

func newTaskStore() *TaskStore {
	return &TaskStore{tasks: make(map[string]*ChildTask)}
}

func (s *TaskStore) create(task, agentType string) *ChildTask {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	id := fmt.Sprintf("delegate-%03d", s.seq)
	now := time.Now()
	t := &ChildTask{
		ID:        id,
		Task:      task,
		AgentType: agentType,
		Status:    StatusQueued,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.tasks[id] = t
	return t
}

func (s *TaskStore) get(id string) (*ChildTask, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	return t, ok
}

func (s *TaskStore) update(id string, status TaskStatus, result, errMsg string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tasks[id]
	if !ok {
		return false
	}
	t.Status = status
	t.Result = result
	t.Error = errMsg
	t.UpdatedAt = time.Now()
	return true
}

func (s *TaskStore) list() []*ChildTask {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*ChildTask, 0, len(s.tasks))
	for _, t := range s.tasks {
		out = append(out, t)
	}
	return out
}

// NewMultiAgentServer creates the multi-agent MCP worker.
func NewMultiAgentServer() *worker.WorkerServer {
	store := newTaskStore()
	return worker.NewWorkerServer("multi-agent-worker", []tool.Tool{
		&delegateTaskTool{store: store},
		&taskStatusTool{store: store},
	})
}

type delegateTaskTool struct {
	store *TaskStore
}

func (t *delegateTaskTool) Name() string { return "delegate_task" }
func (t *delegateTaskTool) Description() string {
	return "Delegate a task to a sub-agent for parallel execution"
}
func (t *delegateTaskTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"task":{"type":"string"},"agentType":{"type":"string"}},"required":["task"]}`)
}
func (t *delegateTaskTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Task      string `json:"task"`
		AgentType string `json:"agentType"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	if in.Task == "" {
		return tool.FailureMsg("task is required"), nil
	}
	if in.AgentType == "" {
		in.AgentType = "general"
	}

	child := t.store.create(in.Task, in.AgentType)

	// Spawn goroutine to execute the task
	go t.runTask(child)

	return tool.SuccessWith(fmt.Sprintf("Task %s delegated", child.ID), map[string]any{
		"taskId":    child.ID,
		"task":      in.Task,
		"agentType": in.AgentType,
		"status":    child.Status,
	}), nil
}

func (t *delegateTaskTool) runTask(child *ChildTask) {
	t.store.update(child.ID, StatusRunning, "", "")

	// Simulate task execution with the task description as result.
	// In production this would spawn a real agent with a tool context.
	result := fmt.Sprintf("Completed: %s", child.Task)
	t.store.update(child.ID, StatusComplete, result, "")
}

type taskStatusTool struct {
	store *TaskStore
}

func (t *taskStatusTool) Name() string { return "task_status" }
func (t *taskStatusTool) Description() string {
	return "Check the status of a delegated task"
}
func (t *taskStatusTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"taskId":{"type":"string"}},"required":["taskId"]}`)
}
func (t *taskStatusTool) Execute(_ context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		TaskID string `json:"taskId"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}

	child, ok := t.store.get(in.TaskID)
	if !ok {
		return tool.FailureMsg(fmt.Sprintf("task %q not found", in.TaskID)), nil
	}

	return tool.SuccessWith(fmt.Sprintf("Task %s: %s", child.ID, child.Status), child), nil
}
