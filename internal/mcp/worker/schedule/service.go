package schedule

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sachncs/alfred/internal/mcp/worker"
	"github.com/sachncs/alfred/internal/tool"
)

// TaskStatus represents the lifecycle state of a scheduled task.
type TaskStatus string

const (
	TaskActive TaskStatus = "active"
	TaskPaused TaskStatus = "paused"
	TaskDone   TaskStatus = "completed"
)

// ScheduledTask represents a task in the scheduler.
type ScheduledTask struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Schedule  string     `json:"schedule"`
	Kind      string     `json:"kind"`
	Status    TaskStatus `json:"status"`
	CreatedAt time.Time  `json:"createdAt"`
}

// TaskRun records a single execution of a scheduled task.
type TaskRun struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"taskId"`
	Status    string    `json:"status"`
	StartedAt time.Time `json:"startedAt"`
	EndedAt   time.Time `json:"endedAt"`
	Output    string    `json:"output,omitempty"`
}

// TaskStore is an in-memory store for scheduled tasks and their run history.
type TaskStore struct {
	mu     sync.RWMutex
	tasks  map[string]*ScheduledTask
	runs   map[string][]*TaskRun // taskID -> runs
	seq    int
	runSeq int
}

func newTaskStore() *TaskStore {
	return &TaskStore{
		tasks: make(map[string]*ScheduledTask),
		runs:  make(map[string][]*TaskRun),
	}
}

func (s *TaskStore) create(name, schedule, kind string) *ScheduledTask {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	id := fmt.Sprintf("sched-%03d", s.seq)
	task := &ScheduledTask{
		ID:        id,
		Name:      name,
		Schedule:  schedule,
		Kind:      kind,
		Status:    TaskActive,
		CreatedAt: time.Now(),
	}
	s.tasks[id] = task
	return task
}

func (s *TaskStore) get(id string) (*ScheduledTask, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	return t, ok
}

func (s *TaskStore) list() []*ScheduledTask {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*ScheduledTask, 0, len(s.tasks))
	for _, t := range s.tasks {
		out = append(out, t)
	}
	return out
}

func (s *TaskStore) addRun(taskID, status, output string) *TaskRun {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.runSeq++
	run := &TaskRun{
		ID:        fmt.Sprintf("run-%03d", s.runSeq),
		TaskID:    taskID,
		Status:    status,
		StartedAt: time.Now(),
		EndedAt:   time.Now(),
		Output:    output,
	}
	s.runs[taskID] = append(s.runs[taskID], run)
	return run
}

func (s *TaskStore) getRuns(taskID string) []*TaskRun {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.runs[taskID]
}

// NewScheduleServer creates the schedule MCP worker.
func NewScheduleServer() *worker.WorkerServer {
	store := newTaskStore()
	return worker.NewWorkerServer("schedule-worker", []tool.Tool{
		&scheduleCreateTool{store: store},
		&scheduleRunTool{store: store},
		&scheduleAuditTool{store: store},
	})
}

type scheduleCreateTool struct {
	store *TaskStore
}

func (t *scheduleCreateTool) Name() string        { return "schedule_create" }
func (t *scheduleCreateTool) Description() string { return "Create a scheduled task" }
func (t *scheduleCreateTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"name":{"type":"string"},"schedule":{"type":"string"},"kind":{"type":"string"}},"required":["name","schedule"]}`)
}
func (t *scheduleCreateTool) Execute(_ context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Name     string `json:"name"`
		Schedule string `json:"schedule"`
		Kind     string `json:"kind"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	if in.Name == "" {
		return tool.FailureMsg("name is required"), nil
	}
	if in.Schedule == "" {
		return tool.FailureMsg("schedule is required"), nil
	}
	if in.Kind == "" {
		in.Kind = "once"
	}

	task := t.store.create(in.Name, in.Schedule, in.Kind)
	return tool.SuccessWith("Task created", task), nil
}

type scheduleRunTool struct {
	store *TaskStore
}

func (t *scheduleRunTool) Name() string        { return "schedule_run" }
func (t *scheduleRunTool) Description() string { return "Run a scheduled task immediately" }
func (t *scheduleRunTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"taskId":{"type":"string"}},"required":["taskId"]}`)
}
func (t *scheduleRunTool) Execute(_ context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		TaskID string `json:"taskId"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	if in.TaskID == "" {
		return tool.FailureMsg("taskId is required"), nil
	}

	task, ok := t.store.get(in.TaskID)
	if !ok {
		return tool.FailureMsg(fmt.Sprintf("task %q not found", in.TaskID)), nil
	}

	run := t.store.addRun(task.ID, "completed", fmt.Sprintf("Executed: %s", task.Name))
	return tool.SuccessWith("Task triggered", map[string]any{
		"taskId": task.ID,
		"runId":  run.ID,
		"status": "completed",
	}), nil
}

type scheduleAuditTool struct {
	store *TaskStore
}

func (t *scheduleAuditTool) Name() string        { return "schedule_audit" }
func (t *scheduleAuditTool) Description() string { return "Audit task run history" }
func (t *scheduleAuditTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"taskId":{"type":"string"}}}`)
}
func (t *scheduleAuditTool) Execute(_ context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		TaskID string `json:"taskId"`
	}
	_ = json.Unmarshal(raw, &in)

	if in.TaskID != "" {
		runs := t.store.getRuns(in.TaskID)
		return tool.SuccessWith(fmt.Sprintf("%d runs for task %s", len(runs), in.TaskID), map[string]any{
			"taskId": in.TaskID,
			"runs":   runs,
		}), nil
	}

	// Return all tasks with their run counts
	tasks := t.store.list()
	summary := make([]map[string]any, 0, len(tasks))
	for _, task := range tasks {
		runs := t.store.getRuns(task.ID)
		summary = append(summary, map[string]any{
			"taskId":   task.ID,
			"name":     task.Name,
			"status":   task.Status,
			"runCount": len(runs),
		})
	}
	return tool.SuccessWith("Audit complete", map[string]any{
		"tasks": summary,
	}), nil
}
