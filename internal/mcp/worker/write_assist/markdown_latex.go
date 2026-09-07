package writeassist

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alfred/alfred/internal/mcp/worker"
	"github.com/alfred/alfred/internal/tool"
)

// NewWriteAssistWorker creates a write-assist MCP worker.
func NewWriteAssistWorker() *worker.WorkerServer {
	markdownTool := &markdownLatexTool{}
	completionTool := &inlineCompletionTool{}
	return worker.NewWorkerServer("write-assist", []tool.Tool{markdownTool, completionTool}).WithHandler(
		func(ctx context.Context, name string, input json.RawMessage, tc *tool.Context) (*tool.Result, error) {
			switch name {
			case "markdown_to_latex":
				return markdownTool.Execute(ctx, input, tc)
			case "inline_completion":
				return completionTool.Execute(ctx, input, tc)
			default:
				return tool.FailureMsg(fmt.Sprintf("unknown tool: %s", name)), nil
			}
		},
	)
}

type markdownLatexTool struct{}

func (m *markdownLatexTool) Name() string        { return "markdown_to_latex" }
func (m *markdownLatexTool) Description() string { return "Convert Markdown text to LaTeX format." }
func (m *markdownLatexTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"markdown": {"type": "string", "description": "Markdown text to convert"},
			"template": {"type": "string", "description": "LaTeX template (default: article)"}
		},
		"required": ["markdown"]
	}`)
}

func (m *markdownLatexTool) Execute(_ context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Markdown string `json:"markdown"`
		Template string `json:"template"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	if in.Markdown == "" {
		return tool.FailureMsg("markdown is required"), nil
	}

	// ponytail: basic markdown→latex conversion
	latex := markdownToLatex(in.Markdown)
	return tool.SuccessWith("Converted to LaTeX", map[string]string{"latex": latex}), nil
}

func markdownToLatex(md string) string {
	// Simple conversion — headings, bold, italic, code
	out := md
	// Headers
	for i := 6; i >= 1; i-- {
		prefix := ""
		for j := 0; j < i; j++ {
			prefix += "#"
		}
		out = replaceAll(out, prefix+" ", "\\section*{")
		// Close braces at end of line — ponytail: naive
	}
	// Bold: **text** → \textbf{text}
	out = replaceAll(out, "**", "\\textbf{")
	// Italic: *text* → \textit{text}
	out = replaceAll(out, "*", "\\textit{")
	// Code: `text` → \texttt{text}
	out = replaceAll(out, "`", "\\texttt{")
	return out
}

func replaceAll(s, old, new string) string {
	for {
		i := indexOf(s, old)
		if i < 0 {
			return s
		}
		s = s[:i] + new + s[i+len(old):]
	}
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

type inlineCompletionTool struct{}

func (c *inlineCompletionTool) Name() string { return "inline_completion" }
func (c *inlineCompletionTool) Description() string {
	return "Complete code at a given cursor position."
}
func (c *inlineCompletionTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"prefix": {"type": "string", "description": "Code before cursor"},
			"suffix": {"type": "string", "description": "Code after cursor"},
			"language": {"type": "string", "description": "Programming language"}
		},
		"required": ["prefix"]
	}`)
}

func (c *inlineCompletionTool) Execute(_ context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Prefix   string `json:"prefix"`
		Suffix   string `json:"suffix"`
		Language string `json:"language"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	if in.Prefix == "" {
		return tool.FailureMsg("prefix is required"), nil
	}
	// ponytail: stub completion — real LLM integration later
	return tool.SuccessWith("", map[string]string{"completion": ""}), nil
}
