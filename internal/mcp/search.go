package mcp

import (
	"math"
	"strings"
)

// SearchResult is a scored tool match.
type SearchResult struct {
	Name  string  `json:"name"`
	Score float64 `json:"score"`
}

// BM25Search ranks tools by BM25 score against a query.
// ponytail: simplified BM25 with k1=1.5, b=0.75.
func BM25Search(query string, tools []ToolDoc, topK int) []SearchResult {
	if len(tools) == 0 || query == "" {
		return nil
	}

	queryTerms := tokenize(query)
	if len(queryTerms) == 0 {
		return nil
	}

	// Build document frequency
	df := make(map[string]int)
	docs := make([][]string, len(tools))
	for i, t := range tools {
		terms := tokenize(t.Name + " " + t.Description + " " + t.Tags)
		docs[i] = terms
		seen := make(map[string]bool)
		for _, term := range terms {
			if !seen[term] {
				df[term]++
				seen[term] = true
			}
		}
	}

	n := float64(len(tools))
	avgdl := 0.0
	for _, d := range docs {
		avgdl += float64(len(d))
	}
	if avgdl > 0 {
		avgdl /= n
	}

	const (
		k1 = 1.5
		b  = 0.75
	)

	results := make([]SearchResult, 0, len(tools))
	for i, terms := range docs {
		score := 0.0
		dl := float64(len(terms))
		tf := make(map[string]int)
		for _, t := range terms {
			tf[t]++
		}
		for _, qt := range queryTerms {
			dfi := float64(df[qt])
			if dfi == 0 {
				continue
			}
			idf := math.Log((n-dfi+0.5)/(dfi+0.5) + 1)
			tfVal := float64(tf[qt])
			num := tfVal * (k1 + 1)
			den := tfVal + k1*(1-b+b*dl/avgdl)
			score += idf * num / den
		}
		if score > 0 {
			results = append(results, SearchResult{Name: tools[i].Name, Score: score})
		}
	}

	// Sort by score descending
	for i := 1; i < len(results); i++ {
		for j := i; j > 0 && results[j].Score > results[j-1].Score; j-- {
			results[j], results[j-1] = results[j-1], results[j]
		}
	}

	if topK > 0 && len(results) > topK {
		results = results[:topK]
	}
	return results
}

// ToolDoc is a searchable tool document.
type ToolDoc struct {
	Name        string
	Description string
	Tags        string
}

func tokenize(s string) []string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		return ' '
	}, s)
	return strings.Fields(s)
}
