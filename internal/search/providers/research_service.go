package providers

import (
	"context"
	"sort"
	"sync"
)

// ResearchService orchestrates multiple search providers.
type ResearchService struct {
	providers []Provider
}

// NewResearchService creates a service with the given providers.
func NewResearchService(providers ...Provider) *ResearchService {
	return &ResearchService{providers: providers}
}

// SearchQuery is the input to the research service.
type SearchQuery struct {
	Query      string `json:"query"`
	MaxResults int    `json:"maxResults,omitempty"` // per provider
}

// ResearchResult is the aggregated output.
type ResearchResult struct {
	Query       string   `json:"query"`
	Results     []Result `json:"results"`
	TotalHits   int      `json:"totalHits"`
	SourcesUsed []string `json:"sourcesUsed"`
}

// Search queries all providers concurrently, merges and ranks results.
func (s *ResearchService) Search(ctx context.Context, q SearchQuery) (*ResearchResult, error) {
	limit := q.MaxResults
	if limit <= 0 {
		limit = 5
	}

	type providerResult struct {
		source  string
		results []Result
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	var all []providerResult
	var sources []string

	for _, p := range s.providers {
		wg.Add(1)
		go func(p Provider) {
			defer wg.Done()
			results, err := p.Search(ctx, q.Query, limit)
			mu.Lock()
			defer mu.Unlock()
			if err == nil && len(results) > 0 {
				all = append(all, providerResult{source: p.Name(), results: results})
				sources = append(sources, p.Name())
			}
		}(p)
	}
	wg.Wait()

	// Merge and sort by source priority (academic first)
	var merged []Result
	for _, pr := range all {
		merged = append(merged, pr.results...)
	}
	sort.Slice(merged, func(i, j int) bool {
		ai := academicPriority(merged[i].Source)
		aj := academicPriority(merged[j].Source)
		if ai != aj {
			return ai < aj
		}
		return merged[i].Source < merged[j].Source
	})

	return &ResearchResult{
		Query:       q.Query,
		Results:     merged,
		TotalHits:   len(merged),
		SourcesUsed: sources,
	}, nil
}

// academicPriority ranks sources — lower = more academic.
func academicPriority(source string) int {
	switch source {
	case "arxiv":
		return 0
	case "biorxiv":
		return 1
	case "europe_pmc":
		return 2
	case "semantic_scholar":
		return 3
	case "tavily":
		return 4
	case "duckduckgo":
		return 5
	default:
		return 10
	}
}
