package paperradar

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/alfred/alfred/internal/mcp/worker"
	"github.com/alfred/alfred/internal/search/providers"
	"github.com/alfred/alfred/internal/tool"
)

// paperStore is a thread-safe in-memory store for paper metadata.
type paperStore struct {
	mu       sync.RWMutex
	papers   map[string]*Paper
	profiles map[string]*ReadingProfile
}

func newPaperStore() *paperStore {
	return &paperStore{
		papers:   make(map[string]*Paper),
		profiles: make(map[string]*ReadingProfile),
	}
}

func (s *paperStore) add(p *Paper) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.papers[p.ID] = p
}

func (s *paperStore) get(id string) (*Paper, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.papers[id]
	return p, ok
}

func (s *paperStore) upsertProfile(p *ReadingProfile) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.profiles[p.PaperID] = p
}

func (s *paperStore) getProfile(id string) (*ReadingProfile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.profiles[id]
	return p, ok
}

// Paper represents a metadata record for a discovered paper.
type Paper struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Authors   []string  `json:"authors,omitempty"`
	Snippet   string    `json:"snippet,omitempty"`
	URL       string    `json:"url,omitempty"`
	Source    string    `json:"source,omitempty"`
	Date      string    `json:"date,omitempty"`
	Digest    string    `json:"digest,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// ReadingProfile tracks how a paper is being read.
type ReadingProfile struct {
	PaperID   string    `json:"paperId"`
	Focus     string    `json:"focus,omitempty"`
	Status    string    `json:"status"`
	Notes     string    `json:"notes,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// NewPaperRadarServer creates the paper-radar MCP worker with real search.
func NewPaperRadarServer() *worker.WorkerServer {
	store := newPaperStore()
	svc := providers.NewResearchService(
		&providers.ArxivProvider{},
		&providers.BiorxivProvider{},
	)
	return worker.NewWorkerServer("paper-radar-worker", []tool.Tool{
		&paperSearchTool{svc: svc, store: store},
		&paperDigestTool{store: store},
		&paperProfileTool{store: store},
	})
}

type paperSearchTool struct {
	svc   *providers.ResearchService
	store *paperStore
}

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

	result, err := t.svc.Search(ctx, providers.SearchQuery{
		Query:      in.Query,
		MaxResults: in.MaxResults,
	})
	if err != nil {
		return tool.Failure(err), nil
	}

	papers := make([]*Paper, 0, len(result.Results))
	for _, r := range result.Results {
		id := paperID(r.URL)
		p := &Paper{
			ID:        id,
			Title:     r.Title,
			Authors:   strings.Split(r.Authors, ", "),
			Snippet:   r.Snippet,
			URL:       r.URL,
			Source:    r.Source,
			Date:      r.Date,
			CreatedAt: time.Now(),
		}
		t.store.add(p)
		papers = append(papers, p)
	}

	return tool.SuccessWith(fmt.Sprintf("Found %d papers for %q", len(papers), in.Query), map[string]any{
		"query":       in.Query,
		"papers":      papers,
		"total":       len(papers),
		"sourcesUsed": result.SourcesUsed,
	}), nil
}

type paperDigestTool struct {
	store *paperStore
}

func (t *paperDigestTool) Name() string        { return "paper_digest" }
func (t *paperDigestTool) Description() string { return "Generate a digest summary of a paper" }
func (t *paperDigestTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"paperId":{"type":"string"}},"required":["paperId"]}`)
}
func (t *paperDigestTool) Execute(_ context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		PaperID string `json:"paperId"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}

	paper, ok := t.store.get(in.PaperID)
	if !ok {
		return tool.FailureMsg(fmt.Sprintf("paper %q not found; run paper_search first", in.PaperID)), nil
	}

	digest := generateDigest(paper)
	paper.Digest = digest

	return tool.SuccessWith("Digest generated", map[string]any{
		"paperId": in.PaperID,
		"title":   paper.Title,
		"digest":  digest,
	}), nil
}

type paperProfileTool struct {
	store *paperStore
}

func (t *paperProfileTool) Name() string        { return "paper_profile_create" }
func (t *paperProfileTool) Description() string { return "Create a reading profile for a paper" }
func (t *paperProfileTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"paperId":{"type":"string"},"focus":{"type":"string"}},"required":["paperId"]}`)
}
func (t *paperProfileTool) Execute(_ context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		PaperID string `json:"paperId"`
		Focus   string `json:"focus"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}

	_, ok := t.store.get(in.PaperID)
	if !ok {
		return tool.FailureMsg(fmt.Sprintf("paper %q not found; run paper_search first", in.PaperID)), nil
	}

	profile := &ReadingProfile{
		PaperID:   in.PaperID,
		Focus:     in.Focus,
		Status:    "active",
		CreatedAt: time.Now(),
	}
	t.store.upsertProfile(profile)

	return tool.SuccessWith("Profile created", map[string]any{
		"paperId": in.PaperID,
		"focus":   in.Focus,
		"status":  "active",
	}), nil
}

// generateDigest builds a text digest from paper metadata.
func generateDigest(p *Paper) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Title: %s\n", p.Title)
	if len(p.Authors) > 0 {
		fmt.Fprintf(&b, "Authors: %s\n", strings.Join(p.Authors, ", "))
	}
	if p.Source != "" {
		fmt.Fprintf(&b, "Source: %s\n", p.Source)
	}
	if p.Date != "" {
		fmt.Fprintf(&b, "Date: %s\n", p.Date)
	}
	if p.Snippet != "" {
		fmt.Fprintf(&b, "\nAbstract:\n%s\n", p.Snippet)
	}
	return b.String()
}

// paperID derives a stable ID from a URL.
func paperID(url string) string {
	if i := strings.LastIndex(url, "/"); i >= 0 {
		return url[i+1:]
	}
	return url
}
