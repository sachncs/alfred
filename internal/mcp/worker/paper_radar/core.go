package paperradar

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alfred/alfred/internal/mcp/worker"
	"github.com/alfred/alfred/internal/tool"
)

// NewPaperRadarServer creates the paper-radar MCP worker.
func NewPaperRadarServer() *worker.WorkerServer {
	return worker.NewWorkerServer("paper-radar-worker", []tool.Tool{
		&paperSearchTool{},
		&paperDigestTool{},
		&paperProfileTool{},
	})
}

type paperSearchTool struct{}

func (t *paperSearchTool) Name() string { return "paper_search" }
func (t *paperSearchTool) Description() string {
	return "Search papers by query across arXiv and bioRxiv"
}
func (t *paperSearchTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"query":{"type":"string"},"maxResults":{"type":"integer"}},"required":["query"]}`)
}
func (t *paperSearchTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Query      string `json:"query"`
		MaxResults int    `json:"maxResults"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	if in.Query == "" {
		return tool.FailureMsg("query is required"), nil
	}
	if in.MaxResults <= 0 {
		in.MaxResults = 10
	}
	return tool.SuccessWith(fmt.Sprintf("Found papers for %q", in.Query), map[string]any{
		"query":   in.Query,
		"results": []map[string]any{},
		"total":   0,
	}), nil
}

type paperDigestTool struct{}

func (t *paperDigestTool) Name() string        { return "paper_digest" }
func (t *paperDigestTool) Description() string { return "Generate a digest summary of a paper" }
func (t *paperDigestTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"paperId":{"type":"string"}},"required":["paperId"]}`)
}
func (t *paperDigestTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		PaperID string `json:"paperId"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	return tool.SuccessWith("Digest generated", map[string]any{
		"paperId": in.PaperID,
		"digest":  "Stub digest content",
	}), nil
}

type paperProfileTool struct{}

func (t *paperProfileTool) Name() string        { return "paper_profile_create" }
func (t *paperProfileTool) Description() string { return "Create a reading profile for a paper" }
func (t *paperProfileTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"paperId":{"type":"string"},"focus":{"type":"string"}},"required":["paperId"]}`)
}
func (t *paperProfileTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		PaperID string `json:"paperId"`
		Focus   string `json:"focus"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	return tool.SuccessWith("Profile created", map[string]any{
		"paperId": in.PaperID,
		"focus":   in.Focus,
	}), nil
}
