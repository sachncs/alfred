// Package search provides search tools (Grep, Find).
package search

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/alfred/alfred/internal/tool"
)

// GrepInput is the JSON schema for GrepTool input.
type GrepInput struct {
	Pattern string `json:"pattern"`
	Path    string `json:"path,omitempty"`
	Glob    string `json:"glob,omitempty"`
}

// GrepMatch is a single match result.
type GrepMatch struct {
	File string `json:"file"`
	Line int    `json:"line"`
	Text string `json:"text"`
}

// GrepOutput is the structured payload returned by GrepTool.
type GrepOutput struct {
	Matches []GrepMatch `json:"matches"`
	Count   int         `json:"count"`
}

// GrepTool performs regex search over files, similar to rg.
type GrepTool struct{}

// NewGrepTool returns a fresh GrepTool.
func NewGrepTool() *GrepTool { return &GrepTool{} }

// Name implements tool.Tool.
func (*GrepTool) Name() string { return "search_grep" }

// Description implements tool.Tool.
func (*GrepTool) Description() string {
	return "Search file contents using regex patterns. Supports glob filters."
}

// Schema implements tool.Tool.
func (*GrepTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"pattern": {"type": "string", "description": "Regex pattern to search for"},
			"path":    {"type": "string", "description": "Directory to search in (default: workspace root)"},
			"glob":    {"type": "string", "description": "Glob filter for file names (e.g. *.go)"}
		},
		"required": ["pattern"]
	}`)
}

// Execute implements tool.Tool.
func (t *GrepTool) Execute(ctx context.Context, raw json.RawMessage, tc *tool.Context) (*tool.Result, error) {
	var in GrepInput
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	if in.Pattern == "" {
		return tool.FailureMsg("pattern is required"), nil
	}

	re, err := regexp.Compile(in.Pattern)
	if err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid regex: %v", err)), nil
	}

	root := "."
	if tc != nil && tc.WorkspaceRoot != "" {
		root = tc.WorkspaceRoot
	}
	if in.Path != "" {
		root = in.Path
	}

	var matches []GrepMatch
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		// Skip .git directory.
		if strings.Contains(path, "/.git/") || strings.HasSuffix(path, "/.git") {
			return nil
		}
		// Apply glob filter.
		if in.Glob != "" {
			matched, _ := filepath.Match(in.Glob, filepath.Base(path))
			if !matched {
				return nil
			}
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return grepFile(path, re, &matches)
	})
	if err != nil {
		return tool.Failure(err), nil
	}

	if len(matches) == 0 {
		return tool.SuccessWith("no matches found", GrepOutput{Count: 0}), nil
	}

	out := GrepOutput{Matches: matches, Count: len(matches)}
	return tool.SuccessWith(fmt.Sprintf("%d matches found", len(matches)), out), nil
}

func grepFile(path string, re *regexp.Regexp, matches *[]GrepMatch) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if re.MatchString(line) {
			*matches = append(*matches, GrepMatch{
				File: path,
				Line: lineNum,
				Text: line,
			})
		}
	}
	return scanner.Err()
}
