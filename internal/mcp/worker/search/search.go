// Package searchworker implements the search MCP worker.
package searchworker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alfred/alfred/internal/mcp/worker"
	"github.com/alfred/alfred/internal/search/providers"
	"github.com/alfred/alfred/internal/tool"
)

// SearchInput is the input schema for the research_search tool.
type SearchInput struct {
	Query      string `json:"query"`
	MaxResults int    `json:"maxResults,omitempty"`
	Sources    string `json:"sources,omitempty"` // comma-separated: arxiv,biorxiv,...
}

// NewSearchWorker creates a search MCP worker with all providers.
func NewSearchWorker() *worker.WorkerServer {
	svc := providers.NewResearchService(
		&providers.DuckDuckGoProvider{},
		&providers.ArxivProvider{},
		&providers.BiorxivProvider{},
		&providers.EuropePMCProvider{},
		&providers.SemanticScholarProvider{},
		&providers.TavilyProvider{},
	)

	searchTool := &searchTool{service: svc}
	return worker.NewWorkerServer("search-worker", []tool.Tool{searchTool}).WithHandler(
		func(ctx context.Context, name string, input json.RawMessage, tc *tool.Context) (*tool.Result, error) {
			if name == "research_search" {
				return searchTool.Execute(ctx, input, tc)
			}
			return tool.FailureMsg(fmt.Sprintf("unknown tool: %s", name)), nil
		},
	)
}

type searchTool struct {
	service *providers.ResearchService
}

func (s *searchTool) Name() string { return "research_search" }

func (s *searchTool) Description() string {
	return "Search academic papers and web results from multiple sources (arXiv, bioRxiv, Europe PMC, Semantic Scholar, DuckDuckGo, Tavily)."
}

func (s *searchTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"query": {"type": "string", "description": "Search query"},
			"maxResults": {"type": "integer", "description": "Max results per source (default 5)"},
			"sources": {"type": "string", "description": "Comma-separated source filter (e.g. arxiv,biorxiv)"}
		},
		"required": ["query"]
	}`)
}

func (s *searchTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in SearchInput
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	if in.Query == "" {
		return tool.FailureMsg("query is required"), nil
	}
	if in.MaxResults <= 0 {
		in.MaxResults = 5
	}

	result, err := s.service.Search(ctx, providers.SearchQuery{
		Query:      in.Query,
		MaxResults: in.MaxResults,
	})
	if err != nil {
		return tool.Failure(err), nil
	}

	b, _ := json.Marshal(result)
	return tool.SuccessWith(
		fmt.Sprintf("Found %d results from %d sources", result.TotalHits, len(result.SourcesUsed)),
		json.RawMessage(b),
	), nil
}
