package mcp_test

import (
	"testing"

	"github.com/alfred/alfred/internal/mcp"
)

func TestBM25Search(t *testing.T) {
	tools := []mcp.ToolDoc{
		{Name: "read_file", Description: "Read a file from disk", Tags: "file read"},
		{Name: "write_file", Description: "Write content to a file", Tags: "file write"},
		{Name: "grep", Description: "Search file contents with regex", Tags: "search regex"},
		{Name: "research_search", Description: "Search academic papers and web", Tags: "research search papers"},
	}

	results := mcp.BM25Search("research papers", tools, 2)
	if len(results) == 0 {
		t.Fatal("expected results")
	}
	if results[0].Name != "research_search" {
		t.Errorf("top result = %s, want research_search", results[0].Name)
	}
}

func TestBM25SearchEmpty(t *testing.T) {
	results := mcp.BM25Search("", nil, 10)
	if len(results) != 0 {
		t.Errorf("expected empty, got %d", len(results))
	}
}

func TestBM25SearchTopK(t *testing.T) {
	tools := []mcp.ToolDoc{
		{Name: "a", Description: "alpha beta"},
		{Name: "b", Description: "beta gamma"},
		{Name: "c", Description: "alpha gamma delta"},
	}
	results := mcp.BM25Search("alpha", tools, 1)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}
