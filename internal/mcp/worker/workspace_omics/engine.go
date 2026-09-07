package workspaceomics

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/alfred/alfred/internal/mcp/worker"
	"github.com/alfred/alfred/internal/tool"
)

// OmicsData holds parsed Matrix Market data.
type OmicsData struct {
	Path     string      `json:"path"`
	Format   string      `json:"format"`
	Object   string      `json:"object"`
	Rows     int         `json:"rows"`
	Cols     int         `json:"cols"`
	NonZero  int         `json:"nonZero"`
	ValueMin float64     `json:"valueMin"`
	ValueMax float64     `json:"valueMax"`
	Preview  [][]float64 `json:"preview"`
}

func NewOmicsServer() *worker.WorkerServer {
	return worker.NewWorkerServer("workspace-omics-worker", []tool.Tool{
		&omicsPreviewTool{},
		&omicsSelectDatasetTool{},
	})
}

type omicsPreviewTool struct{}

func (t *omicsPreviewTool) Name() string        { return "omics_preview" }
func (t *omicsPreviewTool) Description() string { return "Preview omics data (Matrix Market)" }
func (t *omicsPreviewTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}`)
}
func (t *omicsPreviewTool) Execute(_ context.Context, raw json.RawMessage, tc *tool.Context) (*tool.Result, error) {
	var in struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	if in.Path == "" {
		return tool.FailureMsg("path is required"), nil
	}

	abs := in.Path
	if tc != nil && tc.WorkspaceRoot != "" && !strings.HasPrefix(in.Path, "/") {
		abs = tc.WorkspaceRoot + "/" + in.Path
	}

	data, err := parseMatrixMarket(abs)
	if err != nil {
		return tool.Failure(err), nil
	}

	return tool.SuccessWith(fmt.Sprintf("%dx%d, %d non-zeros", data.Rows, data.Cols, data.NonZero), data), nil
}

type omicsSelectDatasetTool struct{}

func (t *omicsSelectDatasetTool) Name() string        { return "omics_select_dataset" }
func (t *omicsSelectDatasetTool) Description() string { return "Select a dataset from omics data" }
func (t *omicsSelectDatasetTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"},"datasetId":{"type":"string"}},"required":["path","datasetId"]}`)
}
func (t *omicsSelectDatasetTool) Execute(_ context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Path      string `json:"path"`
		DatasetID string `json:"datasetId"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	if in.Path == "" || in.DatasetID == "" {
		return tool.FailureMsg("path and datasetId are required"), nil
	}
	return tool.SuccessWith("Dataset selected", map[string]any{
		"path":      in.Path,
		"datasetId": in.DatasetID,
	}), nil
}

// parseMatrixMarket reads a Matrix Market file.
func parseMatrixMarket(path string) (*OmicsData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	result := &OmicsData{
		Path:     path,
		ValueMin: math.MaxFloat64,
		ValueMax: -math.MaxFloat64,
	}

	scanner := bufio.NewScanner(f)
	lineNum := 0
	var preview [][]float64

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineNum++

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "%") {
			continue
		}

		// First non-comment line is the header: %%MatrixMarket matrix coordinate real general
		if lineNum == 1 || (result.Format == "" && strings.HasPrefix(line, "%%")) {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				result.Object = parts[1] // "matrix"
			}
			if len(parts) >= 4 {
				result.Format = parts[3] // "real", "complex", etc.
			}
			continue
		}

		// Dimensions line: M N K
		if result.Rows == 0 {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				result.Rows, _ = strconv.Atoi(parts[0])
				result.Cols, _ = strconv.Atoi(parts[1])
				if len(parts) >= 3 {
					result.NonZero, _ = strconv.Atoi(parts[2])
				}
			}
			continue
		}

		// Data lines: i j value
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			val, err := strconv.ParseFloat(parts[2], 64)
			if err == nil {
				if val < result.ValueMin {
					result.ValueMin = val
				}
				if val > result.ValueMax {
					result.ValueMax = val
				}
				if len(preview) < 50 {
					preview = append(preview, []float64{
						float64(len(preview) + 1), // row
						val,                       // simplified
					})
				}
			}
		}
	}

	result.Preview = preview
	return result, scanner.Err()
}
