package providers

import (
	"context"
	"encoding/json"
	"fmt"
)

// BiorxivProvider queries the bioRxiv API.
type BiorxivProvider struct{}

func (b *BiorxivProvider) Name() string { return "biorxiv" }

func (b *BiorxivProvider) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	if limit <= 0 {
		limit = 10
	}
	u := fmt.Sprintf("https://api.biorxiv.org/details/biorxiv/%s/%d", encodeQuery(query), limit)
	body, err := doGet(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("biorxiv: %w", err)
	}
	var resp struct {
		Collection []struct {
			Title    string `json:"title"`
			Doi      string `json:"doi"`
			Abstract string `json:"abstract"`
			Category string `json:"category"`
			Date     string `json:"date"`
			Authors  string `json:"authors"`
		} `json:"collection"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("biorxiv: parse: %w", err)
	}
	var results []Result
	for _, p := range resp.Collection {
		results = append(results, Result{
			Title:   p.Title,
			URL:     fmt.Sprintf("https://doi.org/%s", p.Doi),
			Snippet: p.Abstract,
			Source:  "biorxiv",
			Date:    p.Date,
			Authors: p.Authors,
		})
	}
	return results, nil
}
