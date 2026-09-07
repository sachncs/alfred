package projectdag

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/alfred/alfred/internal/mcp/worker"
	"github.com/alfred/alfred/internal/tool"
)

// ProjectNode represents a node in the project graph.
type ProjectNode struct {
	ID          string    `json:"id"`
	ThreadID    string    `json:"threadId"`
	Claim       string    `json:"claim"`
	EvidenceIDs []string  `json:"evidenceIds,omitempty"`
	Status      string    `json:"status"` // draft, active, reviewed, archived
	ReviewedAt  time.Time `json:"reviewedAt,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

// ProjectEdge connects two project nodes.
type ProjectEdge struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Relation string `json:"relation"`
}

// ProjectGraph holds the compiled evidence graph for a project.
type ProjectGraph struct {
	mu    sync.RWMutex
	nodes map[string]*ProjectNode
	edges map[string]*ProjectEdge
	seq   int
}

func newProjectGraph() *ProjectGraph {
	return &ProjectGraph{
		nodes: make(map[string]*ProjectNode),
		edges: make(map[string]*ProjectEdge),
	}
}

func (g *ProjectGraph) addNode(threadID, claim string, evidenceIDs []string) *ProjectNode {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.seq++
	id := fmt.Sprintf("proj-%03d", g.seq)
	node := &ProjectNode{
		ID:          id,
		ThreadID:    threadID,
		Claim:       claim,
		EvidenceIDs: evidenceIDs,
		Status:      "draft",
		CreatedAt:   time.Now(),
	}
	g.nodes[id] = node
	return node
}

func (g *ProjectGraph) getNode(id string) (*ProjectNode, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	n, ok := g.nodes[id]
	return n, ok
}

func (g *ProjectGraph) addEdge(from, to, relation string) (*ProjectEdge, bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, ok := g.nodes[from]; !ok {
		return nil, false
	}
	if _, ok := g.nodes[to]; !ok {
		return nil, false
	}
	key := from + "->" + to
	if _, exists := g.edges[key]; exists {
		return nil, false
	}
	edge := &ProjectEdge{From: from, To: to, Relation: relation}
	g.edges[key] = edge
	return edge, true
}

func (g *ProjectGraph) reviewNode(id string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	n, ok := g.nodes[id]
	if !ok {
		return false
	}
	n.Status = "reviewed"
	n.ReviewedAt = time.Now()
	return true
}

func (g *ProjectGraph) snapshot() ([]*ProjectNode, []*ProjectEdge) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	nodes := make([]*ProjectNode, 0, len(g.nodes))
	for _, n := range g.nodes {
		nodes = append(nodes, n)
	}
	edges := make([]*ProjectEdge, 0, len(g.edges))
	for _, e := range g.edges {
		edges = append(edges, e)
	}
	return nodes, edges
}

// NewProjectDAGServer creates the project-dag MCP worker.
func NewProjectDAGServer() *worker.WorkerServer {
	graph := newProjectGraph()
	return worker.NewWorkerServer("project-dag-worker", []tool.Tool{
		&projectUpdateTool{graph: graph},
		&projectGraphTool{graph: graph},
		&projectReviewTool{graph: graph},
	})
}

type projectUpdateTool struct {
	graph *ProjectGraph
}

func (t *projectUpdateTool) Name() string { return "project_update" }
func (t *projectUpdateTool) Description() string {
	return "Compile thread evidence into a project graph"
}
func (t *projectUpdateTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"threadId":{"type":"string"},"claim":{"type":"string"},"evidenceIds":{"type":"array","items":{"type":"string"}}},"required":["threadId","claim"]}`)
}
func (t *projectUpdateTool) Execute(_ context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		ThreadID    string   `json:"threadId"`
		Claim       string   `json:"claim"`
		EvidenceIDs []string `json:"evidenceIds"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	if in.ThreadID == "" {
		return tool.FailureMsg("threadId is required"), nil
	}
	if in.Claim == "" {
		return tool.FailureMsg("claim is required"), nil
	}

	node := t.graph.addNode(in.ThreadID, in.Claim, in.EvidenceIDs)
	nodes, edges := t.graph.snapshot()
	return tool.SuccessWith("Project graph updated", map[string]any{
		"node":       node,
		"totalNodes": len(nodes),
		"totalEdges": len(edges),
	}), nil
}

type projectGraphTool struct {
	graph *ProjectGraph
}

func (t *projectGraphTool) Name() string        { return "project_graph" }
func (t *projectGraphTool) Description() string { return "Return the current project graph" }
func (t *projectGraphTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{}}`)
}
func (t *projectGraphTool) Execute(_ context.Context, _ json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	nodes, edges := t.graph.snapshot()
	return tool.SuccessWith("Project graph", map[string]any{
		"nodes": nodes,
		"edges": edges,
	}), nil
}

type projectReviewTool struct {
	graph *ProjectGraph
}

func (t *projectReviewTool) Name() string        { return "project_review" }
func (t *projectReviewTool) Description() string { return "Request human review of a project node" }
func (t *projectReviewTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"nodeId":{"type":"string"},"reason":{"type":"string"}},"required":["nodeId"]}`)
}
func (t *projectReviewTool) Execute(_ context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		NodeID string `json:"nodeId"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	if in.NodeID == "" {
		return tool.FailureMsg("nodeId is required"), nil
	}

	if !t.graph.reviewNode(in.NodeID) {
		return tool.FailureMsg(fmt.Sprintf("node %q not found", in.NodeID)), nil
	}

	node, _ := t.graph.getNode(in.NodeID)
	return tool.SuccessWith("Review requested", node), nil
}
