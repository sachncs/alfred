package workspacetabular

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/sachncs/alfred/internal/mcp/worker"
	"github.com/sachncs/alfred/internal/tool"
)

// ColumnInfo describes a column in the tabular data.
type ColumnInfo struct {
	Name     string  `json:"name"`
	Type     string  `json:"type"` // "number", "string", "mixed"
	NonEmpty int     `json:"nonEmpty"`
	Min      *string `json:"min,omitempty"`
	Max      *string `json:"max,omitempty"`
}

// TabularPreview is the structured output of a tabular preview.
type TabularPreview struct {
	Path      string       `json:"path"`
	Columns   []ColumnInfo `json:"columns"`
	TotalRows int          `json:"totalRows"`
	Preview   [][]string   `json:"preview"`
	Delimiter string       `json:"delimiter"`
}

func NewTabularServer() *worker.WorkerServer {
	return worker.NewWorkerServer("workspace-tabular-worker", []tool.Tool{
		&tabularPreviewTool{},
	})
}

type tabularPreviewTool struct{}

func (t *tabularPreviewTool) Name() string        { return "tabular_preview" }
func (t *tabularPreviewTool) Description() string { return "Preview tabular data (CSV/TSV/Parquet)" }
func (t *tabularPreviewTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"},"maxRows":{"type":"integer"}},"required":["path"]}`)
}
func (t *tabularPreviewTool) Execute(_ context.Context, raw json.RawMessage, tc *tool.Context) (*tool.Result, error) {
	var in struct {
		Path    string `json:"path"`
		MaxRows int    `json:"maxRows"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	if in.Path == "" {
		return tool.FailureMsg("path is required"), nil
	}
	if in.MaxRows <= 0 {
		in.MaxRows = 50
	}

	// Resolve path
	abs := in.Path
	if tc != nil && tc.WorkspaceRoot != "" && !strings.HasPrefix(in.Path, "/") {
		abs = tc.WorkspaceRoot + "/" + in.Path
	}

	data, err := os.ReadFile(abs)
	if err != nil {
		return tool.Failure(err), nil
	}

	delim := detectDelimiter(data)
	preview, columns, totalRows, err := parseCSV(bytes.NewReader(data), delim, in.MaxRows)
	if err != nil {
		return tool.FailureMsg(fmt.Sprintf("parse error: %v", err)), nil
	}

	result := TabularPreview{
		Path:      in.Path,
		Columns:   columns,
		TotalRows: totalRows,
		Preview:   preview,
		Delimiter: string(delim),
	}

	return tool.SuccessWith(fmt.Sprintf("%d rows, %d columns", totalRows, len(columns)), result), nil
}

// detectDelimiter checks the first few lines to choose comma vs tab.
func detectDelimiter(data []byte) rune {
	firstLine := data
	if i := bytes.IndexByte(data, '\n'); i > 0 {
		firstLine = data[:i]
	}
	tabs := bytes.Count(firstLine, []byte("\t"))
	commas := bytes.Count(firstLine, []byte(","))
	if tabs > commas {
		return '\t'
	}
	return ','
}

// parseCSV reads CSV data and returns preview rows, column info, and total row count.
func parseCSV(r io.Reader, delim rune, maxPreviewRows int) ([][]string, []ColumnInfo, int, error) {
	cr := csv.NewReader(r)
	cr.Comma = delim
	cr.TrimLeadingSpace = true

	// Read header
	header, err := cr.Read()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("read header: %w", err)
	}

	// Initialize column info
	columns := make([]ColumnInfo, len(header))
	for i, h := range header {
		columns[i] = ColumnInfo{Name: h, Type: "string"}
	}

	// Track column types and stats
	colValues := make([][]string, len(header))
	preview := make([][]string, 0, maxPreviewRows)
	totalRows := 0

	for {
		row, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, totalRows, fmt.Errorf("read row %d: %w", totalRows+1, err)
		}
		totalRows++

		// Collect for preview
		if totalRows <= maxPreviewRows {
			preview = append(preview, row)
		}

		// Collect values for column analysis
		for i, val := range row {
			if i < len(colValues) {
				colValues[i] = append(colValues[i], val)
			}
		}
	}

	// Analyze column types
	for i := range columns {
		if i < len(colValues) {
			columns[i] = analyzeColumn(columns[i].Name, colValues[i])
		}
	}

	return preview, columns, totalRows, nil
}

// analyzeColumn determines the type and statistics for a column.
func analyzeColumn(name string, values []string) ColumnInfo {
	col := ColumnInfo{Name: name, NonEmpty: 0}
	allNumeric := true
	allInt := true
	var minVal, maxVal string
	first := true

	for _, v := range values {
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			continue
		}
		col.NonEmpty++

		if _, err := strconv.ParseFloat(trimmed, 64); err != nil {
			allNumeric = false
			allInt = false
		} else if _, err := strconv.Atoi(trimmed); err != nil {
			allInt = false
		}

		if first {
			minVal = trimmed
			maxVal = trimmed
			first = false
		} else {
			if trimmed < minVal {
				minVal = trimmed
			}
			if trimmed > maxVal {
				maxVal = trimmed
			}
		}
	}

	switch {
	case allInt:
		col.Type = "integer"
	case allNumeric:
		col.Type = "number"
	default:
		col.Type = "string"
	}

	if col.NonEmpty > 0 {
		col.Min = &minVal
		col.Max = &maxVal
	}

	return col
}
