package evidencedag

import (
	"context"
	"encoding/json"
	"testing"

	alfredtool "github.com/sachncs/alfred/internal/tool"
)

func TestEvidenceUpdateTool(t *testing.T) {
	graph := newEvidenceGraph()
	tl := &evidenceUpdateTool{graph: graph}

	t.Run("add node", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"claim":"X causes Y","source":"paper1","confidence":0.9}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})

	t.Run("add with parent", func(t *testing.T) {
		n1 := graph.addNode("parent claim", "src", 0.7)
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"claim":"child claim","confidence":0.6,"parentId":"`+n1.ID+`","relation":"contradicts"}`), nil)
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

func TestEvidenceSnapshotTool(t *testing.T) {
	graph := newEvidenceGraph()
	tl := &evidenceSnapshotTool{graph: graph}

	t.Run("empty graph", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})

	t.Run("graph with nodes", func(t *testing.T) {
		graph.addNode("claim1", "src", 0.5)
		graph.addNode("claim2", "src", 0.8)
		graph.addEdge("ev-001", "ev-002", "supports")
		res, err := tl.Execute(context.Background(), nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})
}

func TestEvidenceAuditTool(t *testing.T) {
	graph := newEvidenceGraph()
	tl := &evidenceAuditTool{graph: graph}

	t.Run("clean graph", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})

	t.Run("graph with orphans", func(t *testing.T) {
		graph.addNode("orphan claim", "", 0.2)
		res, err := tl.Execute(context.Background(), nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})
}

func TestEvidenceGraph(t *testing.T) {
	graph := newEvidenceGraph()

	t.Run("add and get node", func(t *testing.T) {
		node := graph.addNode("claim", "src", 0.7)
		got, ok := graph.getNode(node.ID)
		if !ok || got.Claim != "claim" {
			t.Fatal("expected to find node")
		}
	})

	t.Run("add edge", func(t *testing.T) {
		n1 := graph.addNode("a", "src", 0.5)
		n2 := graph.addNode("b", "src", 0.6)
		edge, ok := graph.addEdge(n1.ID, n2.ID, "supports")
		if !ok || edge.Relation != "supports" {
			t.Fatal("expected edge to be created")
		}
	})

	t.Run("add edge to nonexistent node", func(t *testing.T) {
		n1 := graph.addNode("c", "src", 0.5)
		_, ok := graph.addEdge(n1.ID, "nonexistent", "supports")
		if ok {
			t.Fatal("expected edge creation to fail")
		}
	})

	t.Run("duplicate edge", func(t *testing.T) {
		n1 := graph.addNode("d", "src", 0.5)
		n2 := graph.addNode("e", "src", 0.6)
		graph.addEdge(n1.ID, n2.ID, "supports")
		_, ok := graph.addEdge(n1.ID, n2.ID, "supports")
		if ok {
			t.Fatal("expected duplicate edge to fail")
		}
	})

	t.Run("snapshot", func(t *testing.T) {
		nodes, edges := graph.snapshot()
		if len(nodes) == 0 {
			t.Fatal("expected nodes")
		}
		_ = edges
	})

	t.Run("audit with cycle", func(t *testing.T) {
		g := newEvidenceGraph()
		n1 := g.addNode("x", "src", 0.5)
		n2 := g.addNode("y", "src", 0.5)
		g.addEdge(n1.ID, n2.ID, "supports")
		g.addEdge(n2.ID, n1.ID, "supports")
		findings := g.audit()
		hasCycle := false
		for _, f := range findings {
			if f != "no issues found" {
				hasCycle = true
			}
		}
		if !hasCycle {
			t.Fatal("expected cycle detection")
		}
	})
}

func TestAssessLevel(t *testing.T) {
	tests := []struct {
		confidence float64
		source     string
		want       Assessment
	}{
		{0.9, "paper1", AssessmentA2},
		{0.6, "", AssessmentA1},
		{0.2, "", AssessmentA0},
	}
	for _, tt := range tests {
		got := assessLevel(tt.confidence, tt.source)
		if got != tt.want {
			t.Errorf("assessLevel(%v, %q) = %s, want %s", tt.confidence, tt.source, got, tt.want)
		}
	}
}

func TestToolInterface(t *testing.T) {
	var _ alfredtool.Tool = &evidenceUpdateTool{}
	var _ alfredtool.Tool = &evidenceSnapshotTool{}
	var _ alfredtool.Tool = &evidenceAuditTool{}
}
