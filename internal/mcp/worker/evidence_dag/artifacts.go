package evidencedag

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sachncs/alfred/internal/mcp/worker"
	"github.com/sachncs/alfred/internal/tool"
)

// Assessment represents the A0/A1/A2 evidence quality level.
type Assessment string

const (
	AssessmentNone Assessment = "none"
	AssessmentA0   Assessment = "A0" // Raw claim, no corroboration
	AssessmentA1   Assessment = "A1" // Single corroborated source
	AssessmentA2   Assessment = "A2" // Multi-source consensus
)

// EvidenceNode is a typed node in the evidence DAG.
type EvidenceNode struct {
	ID         string     `json:"id"`
	Claim      string     `json:"claim"`
	Source     string     `json:"source,omitempty"`
	Confidence float64    `json:"confidence"`
	Level      Assessment `json:"level"`
	CreatedAt  time.Time  `json:"createdAt"`
}

// EvidenceEdge connects two evidence nodes.
type EvidenceEdge struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Relation string `json:"relation"`
}

// EvidenceGraph is the in-memory DAG.
type EvidenceGraph struct {
	mu    sync.RWMutex
	nodes map[string]*EvidenceNode
	edges map[string]*EvidenceEdge // keyed by "from->to"
	seq   int
}

func newEvidenceGraph() *EvidenceGraph {
	return &EvidenceGraph{
		nodes: make(map[string]*EvidenceNode),
		edges: make(map[string]*EvidenceEdge),
	}
}

func (g *EvidenceGraph) addNode(claim, source string, confidence float64) *EvidenceNode {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.seq++
	id := fmt.Sprintf("ev-%03d", g.seq)
	level := assessLevel(confidence, source)
	node := &EvidenceNode{
		ID:         id,
		Claim:      claim,
		Source:     source,
		Confidence: confidence,
		Level:      level,
		CreatedAt:  time.Now(),
	}
	g.nodes[id] = node
	return node
}

func (g *EvidenceGraph) getNode(id string) (*EvidenceNode, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	n, ok := g.nodes[id]
	return n, ok
}

func (g *EvidenceGraph) addEdge(from, to, relation string) (*EvidenceEdge, bool) {
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
	edge := &EvidenceEdge{From: from, To: to, Relation: relation}
	g.edges[key] = edge
	return edge, true
}

func (g *EvidenceGraph) snapshot() ([]*EvidenceNode, []*EvidenceEdge) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	nodes := make([]*EvidenceNode, 0, len(g.nodes))
	for _, n := range g.nodes {
		nodes = append(nodes, n)
	}
	edges := make([]*EvidenceEdge, 0, len(g.edges))
	for _, e := range g.edges {
		edges = append(edges, e)
	}
	return nodes, edges
}

func (g *EvidenceGraph) audit() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var findings []string

	// Check for orphan nodes (no in/out edges)
	hasEdge := make(map[string]bool)
	for _, e := range g.edges {
		hasEdge[e.From] = true
		hasEdge[e.To] = true
	}
	for id, n := range g.nodes {
		if !hasEdge[id] {
			findings = append(findings, fmt.Sprintf("orphan node %s: %s", id, n.Claim))
		}
	}

	// Check for cycles via DFS
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	for id := range g.nodes {
		if !visited[id] {
			if g.hasCycleDFS(id, visited, recStack) {
				findings = append(findings, fmt.Sprintf("cycle detected involving node %s", id))
			}
		}
	}

	// Check for low-confidence nodes
	for _, n := range g.nodes {
		if n.Confidence < 0.3 {
			findings = append(findings, fmt.Sprintf("low confidence node %s: %.2f", n.ID, n.Confidence))
		}
	}

	if len(findings) == 0 {
		findings = append(findings, "no issues found")
	}
	return findings
}

func (g *EvidenceGraph) hasCycleDFS(node string, visited, recStack map[string]bool) bool {
	visited[node] = true
	recStack[node] = true
	for _, e := range g.edges {
		if e.From == node {
			if !visited[e.To] {
				if g.hasCycleDFS(e.To, visited, recStack) {
					return true
				}
			} else if recStack[e.To] {
				return true
			}
		}
	}
	recStack[node] = false
	return false
}

// assessLevel determines the A0/A1/A2 level based on confidence and source.
func assessLevel(confidence float64, source string) Assessment {
	if confidence >= 0.8 && source != "" {
		return AssessmentA2
	}
	if confidence >= 0.5 {
		return AssessmentA1
	}
	return AssessmentA0
}

// NewEvidenceDAGServer creates the evidence-dag MCP worker.
func NewEvidenceDAGServer() *worker.WorkerServer {
	graph := newEvidenceGraph()
	return worker.NewWorkerServer("evidence-dag-worker", []tool.Tool{
		&evidenceUpdateTool{graph: graph},
		&evidenceSnapshotTool{graph: graph},
		&evidenceAuditTool{graph: graph},
	})
}

type evidenceUpdateTool struct {
	graph *EvidenceGraph
}

func (t *evidenceUpdateTool) Name() string        { return "evidence_update" }
func (t *evidenceUpdateTool) Description() string { return "Update the evidence graph with new claims" }
func (t *evidenceUpdateTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"claim":{"type":"string"},"source":{"type":"string"},"confidence":{"type":"number"},"parentId":{"type":"string"},"relation":{"type":"string"}},"required":["claim"]}`)
}
func (t *evidenceUpdateTool) Execute(_ context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Claim      string  `json:"claim"`
		Source     string  `json:"source"`
		Confidence float64 `json:"confidence"`
		ParentID   string  `json:"parentId"`
		Relation   string  `json:"relation"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}

	node := t.graph.addNode(in.Claim, in.Source, in.Confidence)

	// If a parent is specified, add an edge
	if in.ParentID != "" {
		rel := in.Relation
		if rel == "" {
			rel = "supports"
		}
		// Edge creation may fail if parent doesn't exist or edge is duplicate;
		// we still return the node regardless since the node itself was added.
		_, _ = t.graph.addEdge(in.ParentID, node.ID, rel)
	}

	return tool.SuccessWith("Evidence updated", node), nil
}

type evidenceSnapshotTool struct {
	graph *EvidenceGraph
}

func (t *evidenceSnapshotTool) Name() string { return "evidence_snapshot" }
func (t *evidenceSnapshotTool) Description() string {
	return "Snapshot the current evidence graph state"
}
func (t *evidenceSnapshotTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{}}`)
}
func (t *evidenceSnapshotTool) Execute(_ context.Context, _ json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	nodes, edges := t.graph.snapshot()
	snapshotID := fmt.Sprintf("snap-%d", time.Now().UnixMilli())
	return tool.SuccessWith("Snapshot captured", map[string]any{
		"snapshotId": snapshotID,
		"nodes":      nodes,
		"edges":      edges,
		"nodeCount":  len(nodes),
		"edgeCount":  len(edges),
	}), nil
}

type evidenceAuditTool struct {
	graph *EvidenceGraph
}

func (t *evidenceAuditTool) Name() string        { return "evidence_audit" }
func (t *evidenceAuditTool) Description() string { return "Audit the evidence chain for consistency" }
func (t *evidenceAuditTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"depth":{"type":"integer"}}}`)
}
func (t *evidenceAuditTool) Execute(_ context.Context, _ json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	findings := t.graph.audit()
	status := "clean"
	for _, f := range findings {
		if f != "no issues found" {
			status = "issues_found"
			break
		}
	}
	return tool.SuccessWith("Audit complete", map[string]any{
		"findings": findings,
		"status":   status,
	}), nil
}
