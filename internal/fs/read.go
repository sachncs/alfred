package fs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alfred/alfred/internal/tool"
)

// ReadInput is the JSON schema for the ReadTool input.
type ReadInput struct {
	Path       string `json:"path"`
	StartLine  int    `json:"startLine,omitempty"`
	EndLine    int    `json:"endLine,omitempty"`
	MaxBytes   int    `json:"maxBytes,omitempty"`
}

// ReadOutput is the structured payload returned by ReadTool.
type ReadOutput struct {
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	MIMEType string `json:"mimeType"`
	Lines    int    `json:"lines"`
	Truncated bool  `json:"truncated"`
}

// ReadTool reads a file from the active workspace and returns its
// contents (truncated to MaxBytes if set) plus metadata.
type ReadTool struct {
	FileSystemTool
}

// NewReadTool returns a fresh ReadTool.
func NewReadTool() *ReadTool {
	return &ReadTool{}
}

// Name implements tool.Tool.
func (*ReadTool) Name() string { return "fs_read" }

// Description implements tool.Tool.
func (*ReadTool) Description() string {
	return "Read a UTF-8 text file from the workspace. Returns content with size and MIME type metadata."
}

// Schema implements tool.Tool.
func (*ReadTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path":     {"type": "string", "description": "Workspace-relative or absolute path to the file"},
			"startLine": {"type": "integer", "minimum": 0, "description": "First line to read (0-indexed, inclusive). Omit for start of file."},
			"endLine":   {"type": "integer", "minimum": 0, "description": "Last line to read (exclusive). Omit for end of file."},
			"maxBytes":  {"type": "integer", "minimum": 1, "description": "Truncate content to this many bytes if exceeded."}
		},
		"required": ["path"]
	}`)
}

// Execute implements tool.Tool.
func (t *ReadTool) Execute(ctx context.Context, raw json.RawMessage, tc *tool.Context) (*tool.Result, error) {
	var in ReadInput
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	if in.Path == "" {
		return tool.FailureMsg("path is required"), nil
	}

	root := ""
	if tc != nil {
		root = tc.WorkspaceRoot
	}
	abs, err := t.ResolvePath(in.Path, root)
	if err != nil {
		return tool.Failure(err), nil
	}

	info, err := os.Stat(abs)
	if err != nil {
		return tool.Failure(err), nil
	}
	if info.IsDir() {
		return tool.FailureMsg(fmt.Sprintf("%q is a directory", in.Path)), nil
	}

	// Honour context cancellation before reading.
	if err := ctx.Err(); err != nil {
		return tool.Failure(err), nil
	}

	data, err := os.ReadFile(abs)
	if err != nil {
		return tool.Failure(err), nil
	}

	truncated := false
	if in.MaxBytes > 0 && len(data) > in.MaxBytes {
		data = data[:in.MaxBytes]
		truncated = true
	}

	content := string(data)
	lines := strings.Count(content, "\n")

	// Line slicing (1-indexed via StartLine/EndLine for human-friendliness).
	if in.StartLine > 0 || in.EndLine > 0 {
		content = sliceLines(content, in.StartLine, in.EndLine)
	}

	mime := detectMIME(abs, data)

	out := ReadOutput{
		Path:      in.Path,
		Size:      info.Size(),
		MIMEType:  mime,
		Lines:     lines,
		Truncated: truncated,
	}
	return tool.SuccessWith(content, out), nil
}

func sliceLines(content string, startLine, endLine int) string {
	if startLine < 0 {
		startLine = 0
	}
	all := strings.Split(content, "\n")
	from := startLine
	if from > len(all) {
		return ""
	}
	to := endLine
	if to <= 0 || to > len(all) {
		to = len(all)
	}
	return strings.Join(all[from:to], "\n")
}

// detectMIME uses the file extension as a fast heuristic. More
// sophisticated sniffing (e.g. magic bytes) lands in Phase 2+.
func detectMIME(path string, data []byte) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return "text/x-go"
	case ".json":
		return "application/json"
	case ".md":
		return "text/markdown"
	case ".txt", "":
		return "text/plain"
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".ts":
		return "application/typescript"
	case ".py":
		return "text/x-python"
	case ".yaml", ".yml":
		return "application/yaml"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".pdf":
		return "application/pdf"
	default:
		// Look at the first 512 bytes to detect binary.
		probe := data
		if len(probe) > 512 {
			probe = probe[:512]
		}
		for _, b := range probe {
			if b == 0 {
				return "application/octet-stream"
			}
		}
		return "text/plain"
	}
}
