package runtime

import (
	"strconv"
	"testing"

	"github.com/alfred/alfred/internal/contract"
)

func TestCompactorNoOp(t *testing.T) {
	c := NewCompactor(40, 20)
	result := c.MaybeCompact(nil)
	if result != nil {
		t.Error("expected nil for nil input")
	}
}

func TestCompactorCompactsWhenOver(t *testing.T) {
	c := NewCompactor(40, 3)

	if !c.ShouldCompact(10) {
		t.Error("expected ShouldCompact(10) = true")
	}
	if c.ShouldCompact(2) {
		t.Error("expected ShouldCompact(2) = false")
	}
}

func TestCompactorNormal(t *testing.T) {
	c := NewCompactor(40, 3)
	turns := make([]contract.Turn, 5)
	for i := range turns {
		turns[i] = contract.Turn{
			ID:     contract.TurnID(string(rune('a' + i))),
			Status: contract.TurnStatusCompleted,
		}
	}
	compacted := c.Compact(turns, CompactionNormal)
	if len(compacted) != 4 { // 1 summary + 3 recent
		t.Errorf("normal compact: got %d turns, want 4", len(compacted))
	}
	if compacted[0].Items[0].Kind != contract.ItemKindCompaction {
		t.Error("first turn should be compaction summary")
	}
}

func TestCompactorAggressive(t *testing.T) {
	c := NewCompactor(40, 6)
	turns := make([]contract.Turn, 10)
	compacted := c.Compact(turns, CompactionAggressive)
	if len(compacted) > 5 {
		t.Errorf("aggressive compact: got %d, want <= 5", len(compacted))
	}
}

func TestCompactorForce(t *testing.T) {
	c := NewCompactor(40, 3)
	turns := make([]contract.Turn, 10)
	compacted := c.Compact(turns, CompactionForce)
	if len(compacted) != 3 { // 1 summary + 2 recent
		t.Errorf("force compact: got %d, want 3", len(compacted))
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
	result := BuildPrompt(history, 100)
	if len(result) == 0 {
		t.Error("expected non-empty prompt")
	}
}

func TestHistoryPruner(t *testing.T) {
	p := NewHistoryPruner(3)
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

func TestCacheTelemetry(t *testing.T) {
	ct := NewCacheTelemetry()
	ct.RecordHit("prompt")
	ct.RecordHit("prompt")
	ct.RecordMiss("prompt")
	hits, misses := ct.Stats("prompt")
	if hits != 2 || misses != 1 {
		t.Errorf("hits=%d misses=%d, want 2/1", hits, misses)
	}
	ct.Reset()
	hits, _ = ct.Stats("prompt")
	if hits != 0 {
		t.Errorf("hits after reset = %d, want 0", hits)
	}
}

func TestStripANSI(t *testing.T) {
	input := "\x1b[31mred text\x1b[0m"
	got := StripANSI(input)
	if got != "red text" {
		t.Errorf("StripANSI = %q, want %q", got, "red text")
	}
}

func TestHistoryHygiene(t *testing.T) {
	h := NewHistoryHygiene()
	text := "hello\x1b[32m world\x1b[0m"
	clean := h.CleanText(text)
	if clean != "hello world" {
		t.Errorf("CleanText = %q, want %q", clean, "hello world")
	}
}

func TestSummaryCount(t *testing.T) {
	if SummaryCount(10, 6) != 4 {
		t.Error("wrong count")
	}
	if SummaryCount(3, 5) != 0 {
		t.Error("should not go negative")
	}
}

func TestStrconvItoa(t *testing.T) {
	tests := []struct {
		in   int
		want string
	}{
		{0, "0"}, {1, "1"}, {42, "42"}, {123, "123"},
	}
	for _, tt := range tests {
		if got := strconv.Itoa(tt.in); got != tt.want {
			t.Errorf("strconv.Itoa(%d) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
