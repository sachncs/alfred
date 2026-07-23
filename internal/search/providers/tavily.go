package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

// TavilyProvider queries the Tavily Search API.
type TavilyProvider struct{}

func (t *TavilyProvider) Name() string { return "tavily" }

func (t *TavilyProvider) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	apiKey := os.Getenv("TAVILY_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("tavily: TAVILY_API_KEY not set")
	}
	if limit <= 0 {
		limit = 10
	}
	payload, _ := json.Marshal(map[string]any{
		"api_key":      apiKey,
		"query":        query,
		"max_results":  limit,
		"search_depth": "basic",
	})
	body, err := doPost(ctx, "https://api.tavily.com/search", "application/json", payload)
	if err != nil {
		return nil, fmt.Errorf("tavily: %w", err)
	}
	var resp struct {
		Results []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Content string `json:"content"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("tavily: parse: %w", err)
	}
	var results []Result
	for _, r := range resp.Results {
		results = append(results, Result{
			Title:   r.Title,
			URL:     r.URL,
			Snippet: r.Content,
			Source:  "tavily",
		})
	}
	return results, nil
}
