package providers

import (
	"context"
	"encoding/xml"
	"fmt"
	"strings"
)

// ArxivProvider queries the arXiv API.
type ArxivProvider struct{}

func (a *ArxivProvider) Name() string { return "arxiv" }

func (a *ArxivProvider) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	if limit <= 0 {
		limit = 10
	}
	u := fmt.Sprintf("http://export.arxiv.org/api/query?search_query=all:%s&max_results=%d&sortBy=relevance", encodeQuery(query), limit)
	body, err := doGet(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("arxiv: %w", err)
	}
	return parseArxivAtom(string(body)), nil
}

type arxivFeed struct {
	XMLName xml.Name     `xml:"feed"`
	Entries []arxivEntry `xml:"entry"`
}

type arxivEntry struct {
	Title     string `xml:"title"`
	Summary   string `xml:"summary"`
	ID        string `xml:"id"`
	Published string `xml:"published"`
	Authors   []struct {
		Name string `xml:"name"`
	} `xml:"author"`
}

func parseArxivAtom(xmlStr string) []Result {
	var feed arxivFeed
	if err := xml.Unmarshal([]byte(xmlStr), &feed); err != nil {
		return nil
	}
	var results []Result
	for _, e := range feed.Entries {
		authors := make([]string, 0, len(e.Authors))
		for _, a := range e.Authors {
			authors = append(authors, a.Name)
		}
		results = append(results, Result{
			Title:   strings.TrimSpace(e.Title),
			URL:     e.ID,
			Snippet: strings.TrimSpace(e.Summary),
			Source:  "arxiv",
			Date:    e.Published,
			Authors: strings.Join(authors, ", "),
		})
	}
	return results
}
