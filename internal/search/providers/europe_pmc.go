package providers

import (
	"context"
	"encoding/json"
	"fmt"
)

// EuropePMCProvider queries the Europe PMC REST API.
type EuropePMCProvider struct{}

func (e *EuropePMCProvider) Name() string { return "europe_pmc" }

func (e *EuropePMCProvider) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	if limit <= 0 {
		limit = 10
	}
	u := fmt.Sprintf("https://www.ebi.ac.uk/europepmc/webservices/rest/search?query=%s&resultType=core&pageSize=%d&format=json", encodeQuery(query), limit)
	body, err := doGet(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("europe_pmc: %w", err)
	}
	var resp struct {
		ResultList struct {
			Result []struct {
				Title        string `json:"title"`
				Doi          string `json:"doi"`
				AbstractText string `json:"abstractText"`
				FirstPDate   string `json:"firstPublicationDate"`
				AuthorString string `json:"authorString"`
			} `json:"result"`
		} `json:"resultList"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("europe_pmc: parse: %w", err)
	}
	var results []Result
	for _, p := range resp.ResultList.Result {
		url := fmt.Sprintf("https://doi.org/%s", p.Doi)
		if p.Doi == "" {
			url = fmt.Sprintf("https://europepmc.org/search?query=%s", p.Title)
		}
		results = append(results, Result{
			Title:   p.Title,
			URL:     url,
			Snippet: p.AbstractText,
			Source:  "europe_pmc",
			Date:    p.FirstPDate,
			Authors: p.AuthorString,
		})
	}
	return results, nil
}
