package runtime

import (
	"sync"
	"time"
)

// Goal tracks a high-level objective.
type Goal struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status"` // active, completed, abandoned
	CreatedAt   time.Time `json:"createdAt"`
}

// Todo tracks a single task within a goal.
type Todo struct {
	ID        string    `json:"id"`
	GoalID    string    `json:"goalId"`
	Title     string    `json:"title"`
	Status    string    `json:"status"` // pending, in_progress, done
	Priority  int       `json:"priority"`
	CreatedAt time.Time `json:"createdAt"`
}

// GoalStore manages goals and todos in memory.
type GoalStore struct {
	mu     sync.Mutex
	goals  []Goal
	todos  []Todo
	nextID int
}

// NewGoalStore creates a store.
func NewGoalStore() *GoalStore {
	return &GoalStore{}
}

// AddGoal creates a new goal.
func (s *GoalStore) AddGoal(title, description string) Goal {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	g := Goal{
		ID:          string(rune('0' + s.nextID)),
		Title:       title,
		Description: description,
		Status:      "active",
		CreatedAt:   time.Now().UTC(),
	}
	s.goals = append(s.goals, g)
	return g
}

// ListGoals returns all goals.
func (s *GoalStore) ListGoals() []Goal {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Goal, len(s.goals))
	copy(out, s.goals)
	return out
}

// AddTodo creates a new todo under a goal.
func (s *GoalStore) AddTodo(goalID, title string, priority int) Todo {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	t := Todo{
		ID:        string(rune('0' + s.nextID)),
		GoalID:    goalID,
		Title:     title,
		Status:    "pending",
		Priority:  priority,
		CreatedAt: time.Now().UTC(),
	}
	s.todos = append(s.todos, t)
	return t
}

// ListTodos returns all todos for a goal.
func (s *GoalStore) ListTodos(goalID string) []Todo {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []Todo
	for _, t := range s.todos {
		if t.GoalID == goalID {
			out = append(out, t)
		}
	}
	return out
}

// UpdateTodoStatus updates a todo's status.
func (s *GoalStore) UpdateTodoStatus(todoID, status string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.todos {
		if s.todos[i].ID == todoID {
			s.todos[i].Status = status
			return true
		}
	}
	return false
}

// CompleteGoal marks a goal as completed.
func (s *GoalStore) CompleteGoal(goalID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.goals {
		if s.goals[i].ID == goalID {
			s.goals[i].Status = "completed"
			return true
		}
	}
	return false
}
