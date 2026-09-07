package providers

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// DuckDuckGoProvider scrapes DuckDuckGo HTML results.
type DuckDuckGoProvider struct{}

func (d *DuckDuckGoProvider) Name() string { return "duckduckgo" }

func (d *DuckDuckGoProvider) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	u := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", encodeQuery(query))
	body, err := doGet(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("duckduckgo: %w", err)
	}
	return parseDDGHTML(string(body), limit), nil
}

var (
	ddgResultRe  = regexp.MustCompile(`<a[^>]+class="result__a"[^>]*href="([^"]*)"[^>]*>(.*?)</a>`)
	ddgSnippetRe = regexp.MustCompile(`<a[^>]+class="result__snippet"[^>]*>(.*?)</a>`)
)

func parseDDGHTML(html string, limit int) []Result {
	var results []Result
	matches := ddgResultRe.FindAllStringSubmatch(html, -1)
	snippets := ddgSnippetRe.FindAllStringSubmatch(html, -1)

	for i, m := range matches {
		if limit > 0 && len(results) >= limit {
			break
		}
		title := stripTags(m[2])
		if title == "" {
			continue
		}
		snippet := ""
		if i < len(snippets) {
			snippet = stripTags(snippets[i][1])
		}
		results = append(results, Result{
			Title:   title,
			URL:     m[1],
			Snippet: snippet,
			Source:  "duckduckgo",
		})
	}
	return results
}

func stripTags(s string) string {
	s = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	return s
}
