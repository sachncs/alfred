package runtime

import (
	"sync"
	"time"
)

// TodoItem is a single task within a thread.
type TodoItem struct {
	ID        string    `json:"id"`
	ThreadID  string    `json:"threadId"`
	Title     string    `json:"title"`
	Status    string    `json:"status"` // pending, in_progress, done
	Priority  int       `json:"priority"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TodoStore manages per-thread todo items.
type TodoStore struct {
	mu     sync.Mutex
	todos  []TodoItem
	nextID int
}

// NewTodoStore creates a store.
func NewTodoStore() *TodoStore {
	return &TodoStore{}
}

// Add creates a new todo for a thread.
func (s *TodoStore) Add(threadID, title string, priority int) TodoItem {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	now := time.Now().UTC()
	t := TodoItem{
		ID:        string(rune('a' - 1 + s.nextID)),
		ThreadID:  threadID,
		Title:     title,
		Status:    "pending",
		Priority:  priority,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.todos = append(s.todos, t)
	return t
}

// List returns all todos for a thread.
func (s *TodoStore) List(threadID string) []TodoItem {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []TodoItem
	for _, t := range s.todos {
		if t.ThreadID == threadID {
			out = append(out, t)
		}
	}
	return out
}

// UpdateStatus sets the status of a todo.
func (s *TodoStore) UpdateStatus(todoID, status string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.todos {
		if s.todos[i].ID == todoID {
			s.todos[i].Status = status
			s.todos[i].UpdatedAt = time.Now().UTC()
			return true
		}
	}
	return false
}

// Delete removes a todo.
func (s *TodoStore) Delete(todoID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.todos {
		if s.todos[i].ID == todoID {
			s.todos = append(s.todos[:i], s.todos[i+1:]...)
			return true
		}
	}
	return false
}
