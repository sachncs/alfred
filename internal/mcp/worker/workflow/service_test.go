package workflow

import (
	"context"
	"encoding/json"
	"testing"

	alfredtool "github.com/sachncs/alfred/internal/tool"
)

func TestWorkflowListTool(t *testing.T) {
	store := newWorkflowStore()
	tl := &workflowListTool{store: store}

	t.Run("empty list", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})
}

func TestWorkflowRunTool(t *testing.T) {
	store := newWorkflowStore()
	tl := &workflowRunTool{store: store}

	t.Run("missing workflowId", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure")
		}
	})

	t.Run("unknown workflow", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"workflowId":"unknown"}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure")
		}
	})

	t.Run("valid run", func(t *testing.T) {
		store.register(&WorkflowDef{ID: "wf-1", Name: "Test", Steps: []WorkflowStep{{ID: "s1", Name: "Step 1"}}})
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"workflowId":"wf-1"}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`not json`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure")
		}
	})
}

func TestWorkflowStatusTool(t *testing.T) {
	store := newWorkflowStore()
	tl := &workflowStatusTool{store: store}

	t.Run("missing runId", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure")
		}
	})

	t.Run("unknown run", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"runId":"unknown"}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure")
		}
	})

	t.Run("known run", func(t *testing.T) {
		store.register(&WorkflowDef{ID: "wf-1", Name: "Test", Steps: []WorkflowStep{{ID: "s1"}}})
		run, _ := store.createRun("wf-1")
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"runId":"`+run.ID+`"}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})
}

func TestWorkflowValidateTool(t *testing.T) {
	store := newWorkflowStore()
	tl := &workflowValidateTool{store: store}

	t.Run("valid workflow", func(t *testing.T) {
		wf := `{"id":"wf-test","name":"Test","steps":[{"id":"s1","name":"Step1","tool":"t1"}]}`
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"workflow":`+wf+`}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})

	t.Run("workflow with cycle", func(t *testing.T) {
		wf := `{"id":"wf-cyc","name":"Cyclic","steps":[{"id":"s1","name":"A","tool":"t1","dependsOn":["s2"]},{"id":"s2","name":"B","tool":"t2","dependsOn":["s1"]}]}`
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"workflow":`+wf+`}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})

	t.Run("workflow with bad dep", func(t *testing.T) {
		wf := `{"id":"wf-bad","name":"Bad","steps":[{"id":"s1","name":"A","tool":"t1","dependsOn":["nonexistent"]}]}`
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"workflow":`+wf+`}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`not json`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure")
		}
	})
}

func TestValidateWorkflow(t *testing.T) {
	t.Run("valid DAG", func(t *testing.T) {
		wf := &WorkflowDef{Steps: []WorkflowStep{
			{ID: "s1", DependsOn: []string{}},
			{ID: "s2", DependsOn: []string{"s1"}},
		}}
		errs := validateWorkflow(wf)
		if len(errs) != 0 {
			t.Fatalf("expected no errors, got %v", errs)
		}
	})

	t.Run("cycle", func(t *testing.T) {
		wf := &WorkflowDef{Steps: []WorkflowStep{
			{ID: "s1", DependsOn: []string{"s2"}},
			{ID: "s2", DependsOn: []string{"s1"}},
		}}
		errs := validateWorkflow(wf)
		if len(errs) == 0 {
			t.Fatal("expected cycle error")
		}
	})

	t.Run("missing dep", func(t *testing.T) {
		wf := &WorkflowDef{Steps: []WorkflowStep{
			{ID: "s1", DependsOn: []string{"ghost"}},
		}}
		errs := validateWorkflow(wf)
		if len(errs) == 0 {
			t.Fatal("expected missing dep error")
		}
	})
}

func TestWorkflowStore(t *testing.T) {
	store := newWorkflowStore()

	t.Run("register and get", func(t *testing.T) {
		store.register(&WorkflowDef{ID: "wf-1", Name: "Test"})
		got, ok := store.get("wf-1")
		if !ok || got.Name != "Test" {
			t.Fatal("expected to find workflow")
		}
	})

	t.Run("list", func(t *testing.T) {
		wfs := store.list()
		if len(wfs) == 0 {
			t.Fatal("expected workflows")
		}
	})

	t.Run("create run", func(t *testing.T) {
		run, ok := store.createRun("wf-1")
		if !ok || run.Status != RunRunning {
			t.Fatal("expected running run")
		}
	})

	t.Run("create run for missing wf", func(t *testing.T) {
		_, ok := store.createRun("nope")
		if ok {
			t.Fatal("expected failure")
		}
	})

	t.Run("complete run", func(t *testing.T) {
		run, _ := store.createRun("wf-1")
		ok := store.completeRun(run.ID, RunComplete)
		if !ok {
			t.Fatal("expected completion to succeed")
		}
		got, _ := store.getRun(run.ID)
		if got.Status != RunComplete {
			t.Fatal("expected completed status")
		}
	})

	t.Run("complete missing run", func(t *testing.T) {
		ok := store.completeRun("nope", RunComplete)
		if ok {
			t.Fatal("expected failure")
		}
	})
}

func TestToolInterface(t *testing.T) {
	var _ alfredtool.Tool = &workflowListTool{}
	var _ alfredtool.Tool = &workflowRunTool{}
	var _ alfredtool.Tool = &workflowStatusTool{}
	var _ alfredtool.Tool = &workflowValidateTool{}
}
