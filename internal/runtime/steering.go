package runtime

import (
	"sync"
)

// Instruction is a pending steering instruction.
type Instruction struct {
	ID       string
	Text     string
	Priority int
}

// SteeringQueue manages pending steering instructions for the agent.
type SteeringQueue struct {
	mu           sync.Mutex
	instructions []Instruction
	counter      int
}

// NewSteeringQueue creates a new queue.
func NewSteeringQueue() *SteeringQueue {
	return &SteeringQueue{}
}

// Push adds an instruction to the queue.
func (q *SteeringQueue) Push(text string, priority int) string {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.counter++
	id := string(rune('A' - 1 + q.counter)) // ponytail: simple A,B,C... IDs
	q.instructions = append(q.instructions, Instruction{
		ID: id, Text: text, Priority: priority,
	})
	return id
}

// Pop removes and returns the highest-priority instruction, or nil if empty.
func (q *SteeringQueue) Pop() *Instruction {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.instructions) == 0 {
		return nil
	}
	best := 0
	for i := 1; i < len(q.instructions); i++ {
		if q.instructions[i].Priority > q.instructions[best].Priority {
			best = i
		}
	}
	ins := q.instructions[best]
	q.instructions = append(q.instructions[:best], q.instructions[best+1:]...)
	return &ins
}

// Len returns the number of pending instructions.
func (q *SteeringQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.instructions)
}
