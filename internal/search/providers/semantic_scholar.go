package providers

import (
	"context"
	"encoding/json"
	"fmt"
)

// SemanticScholarProvider queries the Semantic Scholar Graph API.
type SemanticScholarProvider struct{}

func (s *SemanticScholarProvider) Name() string { return "semantic_scholar" }

func (s *SemanticScholarProvider) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	if limit <= 0 {
		limit = 10
	}
	u := fmt.Sprintf("https://api.semanticscholar.org/graph/v1/paper/search?query=%s&limit=%d&fields=title,abstract,url,authors,year", encodeQuery(query), limit)
	body, err := doGet(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("semantic_scholar: %w", err)
	}
	var resp struct {
		Data []struct {
			Title    string `json:"title"`
			Abstract string `json:"abstract"`
			URL      string `json:"url"`
			Year     int    `json:"year"`
			Authors  []struct {
				Name string `json:"name"`
			} `json:"authors"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("semantic_scholar: parse: %w", err)
	}
	var results []Result
	for _, p := range resp.Data {
		authors := make([]string, 0, len(p.Authors))
		for _, a := range p.Authors {
			authors = append(authors, a.Name)
		}
		results = append(results, Result{
			Title:   p.Title,
			URL:     p.URL,
			Snippet: p.Abstract,
			Source:  "semantic_scholar",
			Date:    fmt.Sprintf("%d", p.Year),
			Authors: joinStrings(authors, ", "),
		})
	}
	return results, nil
}

func joinStrings(ss []string, sep string) string {
	out := ""
	for i, s := range ss {
		if i > 0 {
			out += sep
		}
		out += s
	}
	return out
}
