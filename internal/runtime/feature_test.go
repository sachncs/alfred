package runtime

import (
	"testing"
)

func TestCompactorNoOp(t *testing.T) {
	c := NewCompactor(nil, 40, 20)
	turns := make([]struct{ ID string }, 10)
	result := c.MaybeCompact(nil)
	if result != nil {
		t.Error("expected nil for nil input")
	}
	_ = turns
}

func TestCompactorCompactsWhenOver(t *testing.T) {
	client := &stubClient{responses: []responseSeq{{text: "summary"}}}
	c := NewCompactor(client, 40, 3)

	type miniTurn struct{ ID string }
	turns := make([]miniTurn, 5)
	_ = turns

	if !c.ShouldCompact(10) {
		t.Error("expected ShouldCompact(10) = true")
	}
	if c.ShouldCompact(2) {
		t.Error("expected ShouldCompact(2) = false")
	}
}

func TestTokenBudgetUnlimited(t *testing.T) {
	b := NewTokenBudget(0, 0)
	if b.InputRemaining() != -1 {
		t.Error("expected unlimited input")
	}
	if b.OutputRemaining() != -1 {
		t.Error("expected unlimited output")
	}
	b.AddInput(100)
	b.AddOutput(50)
	if b.OverInput() {
		t.Error("should not be over input")
	}
}

func TestTokenBudgetLimited(t *testing.T) {
	b := NewTokenBudget(100, 50)
	b.AddInput(80)
	if b.InputRemaining() != 20 {
		t.Errorf("remaining = %d, want 20", b.InputRemaining())
	}
	b.AddInput(30)
	if !b.OverInput() {
		t.Error("should be over input")
	}
}

func TestEstimateTokens(t *testing.T) {
	n := EstimateTokens("hello world")
	if n != 3 {
		t.Errorf("tokens = %d, want 3", n)
	}
}

func TestBuildPromptUnlimited(t *testing.T) {
	history := []string{"hello", "world"}
	result := BuildPrompt(history, 0)
	if result != "hello\n\nworld" {
		t.Errorf("prompt = %q", result)
	}
}

func TestBuildPromptLimited(t *testing.T) {
	history := []string{"a short msg", "a much longer message that exceeds"}
	result := BuildPrompt(history, 100) // generous limit
	if len(result) == 0 {
		t.Error("expected non-empty prompt")
	}
}

func TestHistoryPruner(t *testing.T) {
	p := NewHistoryPruner(3)
	// Don't need real turns — test the logic
	if p.ShouldPrune(5) != true {
		t.Error("should prune")
	}
	if p.ShouldPrune(2) != false {
		t.Error("should not prune")
	}
}

func TestSteeringQueue(t *testing.T) {
	q := NewSteeringQueue()
	q.Push("low", 1)
	q.Push("high", 10)
	if q.Len() != 2 {
		t.Errorf("len = %d, want 2", q.Len())
	}
	ins := q.Pop()
	if ins == nil || ins.Text != "high" {
		t.Errorf("expected high priority, got %v", ins)
	}
	if q.Len() != 1 {
		t.Errorf("len after pop = %d, want 1", q.Len())
	}
	ins = q.Pop()
	if ins == nil || ins.Text != "low" {
		t.Errorf("expected low priority, got %v", ins)
	}
	if q.Pop() != nil {
		t.Error("expected nil from empty queue")
	}
}

func TestDelegator(t *testing.T) {
	d := NewDelegator()
	h := d.Delegate("test-agent")
	if h.Name != "test-agent" {
		t.Errorf("name = %q, want test-agent", h.Name)
	}
	list := d.List()
	if len(list) != 1 {
		t.Errorf("list len = %d, want 1", len(list))
	}
}

func TestUsageTracker(t *testing.T) {
	tr := NewUsageTracker()
	tr.Record("gpt-4", 100, 50)
	tr.Record("gpt-4", 200, 100)
	in, out := tr.Totals()
	if in != 300 || out != 150 {
		t.Errorf("totals = %d/%d, want 300/150", in, out)
	}
	recent := tr.Recent(1)
	if len(recent) != 1 {
		t.Errorf("recent = %d, want 1", len(recent))
	}
}

func TestToolBudget(t *testing.T) {
	b := NewToolBudget(map[string]int{"bash": 3})
	b.StartTurn("t1")
	if !b.Allow("bash") {
		t.Error("should allow")
	}
	b.Record("bash")
	b.Record("bash")
	b.Record("bash")
	if b.Allow("bash") {
		t.Error("should not allow after 3")
	}
	if b.Remaining("bash") != 0 {
		t.Errorf("remaining = %d, want 0", b.Remaining("bash"))
	}
	// Unknown tool is unlimited.
	if !b.Allow("echo") {
		t.Error("echo should be unlimited")
	}
}

func TestPromptCache(t *testing.T) {
	c := NewPromptCache()
	c.Set("k1", "v1")
	if c.Get("k1") != "v1" {
		t.Error("cache miss")
	}
	c.Invalidate("k1")
	if c.Get("k1") != "" {
		t.Error("expected empty after invalidate")
	}
}

func TestGoalStore(t *testing.T) {
	s := NewGoalStore()
	g := s.AddGoal("Test Goal", "desc")
	if g.Title != "Test Goal" {
		t.Errorf("title = %q", g.Title)
	}
	goals := s.ListGoals()
	if len(goals) != 1 {
		t.Errorf("goals = %d, want 1", len(goals))
	}

	todo := s.AddTodo(g.ID, "do thing", 5)
	if todo.Title != "do thing" {
		t.Errorf("todo title = %q", todo.Title)
	}
	todos := s.ListTodos(g.ID)
	if len(todos) != 1 {
		t.Errorf("todos = %d, want 1", len(todos))
	}

	s.UpdateTodoStatus(todo.ID, "done")
	updated := s.ListTodos(g.ID)
	if updated[0].Status != "done" {
		t.Errorf("status = %q, want done", updated[0].Status)
	}

	s.CompleteGoal(g.ID)
	completed := s.ListGoals()
	if completed[0].Status != "completed" {
		t.Errorf("goal status = %q, want completed", completed[0].Status)
	}
}
