package projectdag

import (
	"context"
	"encoding/json"
	"testing"

	alfredtool "github.com/sachncs/alfred/internal/tool"
)

func TestProjectUpdateTool(t *testing.T) {
	graph := newProjectGraph()
	tl := &projectUpdateTool{graph: graph}

	t.Run("missing threadId", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"claim":"test"}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure")
		}
	})

	t.Run("missing claim", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"threadId":"t1"}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure")
		}
	})

	t.Run("valid update", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"threadId":"t1","claim":"X causes Y","evidenceIds":["ev1"]}`), nil)
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

func TestProjectGraphTool(t *testing.T) {
	graph := newProjectGraph()
	tl := &projectGraphTool{graph: graph}

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
		graph.addNode("t1", "claim1", nil)
		res, err := tl.Execute(context.Background(), nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})
}

func TestProjectReviewTool(t *testing.T) {
	graph := newProjectGraph()
	tl := &projectReviewTool{graph: graph}

	t.Run("missing nodeId", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure")
		}
	})

	t.Run("unknown node", func(t *testing.T) {
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"nodeId":"unknown"}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.OK {
			t.Fatal("expected failure")
		}
	})

	t.Run("review known node", func(t *testing.T) {
		node := graph.addNode("t1", "claim", nil)
		res, err := tl.Execute(context.Background(), json.RawMessage(`{"nodeId":"`+node.ID+`","reason":"looks good"}`), nil)
		if err != nil {
			t.Fatal(err)
		}
		if !res.OK {
			t.Fatalf("expected success, got error: %s", res.Error)
		}
	})
}

func TestProjectGraph(t *testing.T) {
	graph := newProjectGraph()

	t.Run("add and get node", func(t *testing.T) {
		node := graph.addNode("t1", "claim", []string{"ev1"})
		got, ok := graph.getNode(node.ID)
		if !ok || got.Claim != "claim" {
			t.Fatal("expected to find node")
		}
	})

	t.Run("add edge", func(t *testing.T) {
		n1 := graph.addNode("t1", "a", nil)
		n2 := graph.addNode("t1", "b", nil)
		edge, ok := graph.addEdge(n1.ID, n2.ID, "depends_on")
		if !ok || edge.Relation != "depends_on" {
			t.Fatal("expected edge")
		}
	})

	t.Run("add edge to nonexistent node", func(t *testing.T) {
		n1 := graph.addNode("t1", "c", nil)
		_, ok := graph.addEdge(n1.ID, "nonexistent", "depends_on")
		if ok {
			t.Fatal("expected failure")
		}
	})

	t.Run("review node", func(t *testing.T) {
		node := graph.addNode("t1", "review me", nil)
		ok := graph.reviewNode(node.ID)
		if !ok {
			t.Fatal("expected review to succeed")
		}
		got, _ := graph.getNode(node.ID)
		if got.Status != "reviewed" {
			t.Fatalf("expected reviewed, got %s", got.Status)
		}
	})

	t.Run("review nonexistent", func(t *testing.T) {
		ok := graph.reviewNode("nope")
		if ok {
			t.Fatal("expected failure")
		}
	})

	t.Run("snapshot", func(t *testing.T) {
		nodes, edges := graph.snapshot()
		if len(nodes) == 0 {
			t.Fatal("expected nodes")
		}
		_ = edges
	})
}

func TestToolInterface(t *testing.T) {
	var _ alfredtool.Tool = &projectUpdateTool{}
	var _ alfredtool.Tool = &projectGraphTool{}
	var _ alfredtool.Tool = &projectReviewTool{}
}
