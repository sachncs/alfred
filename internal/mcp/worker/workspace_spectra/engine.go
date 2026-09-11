package workspacespectra

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/sachncs/alfred/internal/mcp/worker"
	"github.com/sachncs/alfred/internal/tool"
)

// SpectraData holds parsed JCAMP-DX data.
type SpectraData struct {
	Path     string            `json:"path"`
	Metadata map[string]string `json:"metadata"`
	Points   int               `json:"points"`
	XRange   [2]float64        `json:"xRange"`
	YRange   [2]float64        `json:"yRange"`
	Preview  [][2]float64      `json:"preview"`
}

func NewSpectraServer() *worker.WorkerServer {
	return worker.NewWorkerServer("workspace-spectra-worker", []tool.Tool{
		&spectraPreviewTool{},
	})
}

type spectraPreviewTool struct{}

func (t *spectraPreviewTool) Name() string        { return "spectra_preview" }
func (t *spectraPreviewTool) Description() string { return "Preview spectra data (JCAMP-DX)" }
func (t *spectraPreviewTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}`)
}
func (t *spectraPreviewTool) Execute(_ context.Context, raw json.RawMessage, tc *tool.Context) (*tool.Result, error) {
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

	data, err := parseJCAMPDX(abs)
	if err != nil {
		return tool.Failure(err), nil
	}

	return tool.SuccessWith(fmt.Sprintf("%d data points", data.Points), data), nil
}

// parseJCAMPDX reads a JCAMP-DX file and extracts metadata and data points.
func parseJCAMPDX(path string) (*SpectraData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	result := &SpectraData{
		Path:     path,
		Metadata: make(map[string]string),
		XRange:   [2]float64{math.MaxFloat64, -math.MaxFloat64},
		YRange:   [2]float64{math.MaxFloat64, -math.MaxFloat64},
	}

	scanner := bufio.NewScanner(f)
	inDataBlock := false
	var points [][2]float64

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "##END") {
			continue
		}

		// Parse metadata lines (##KEY=VALUE)
		if strings.HasPrefix(line, "##") {
			inDataBlock = false
			if idx := strings.Index(line, "="); idx > 0 {
				key := strings.TrimSpace(line[2:idx])
				val := strings.TrimSpace(line[idx+1:])
				result.Metadata[key] = val
			}
			continue
		}

		// Detect data block start (lines that look like XY pairs)
		if !inDataBlock && len(line) > 0 && (line[0] >= '0' && line[0] <= '9' || line[0] == '-' || line[0] == '.') {
			inDataBlock = true
		}

		// Parse data points
		if inDataBlock {
			xy := parseXYLine(line)
			if xy != nil {
				points = append(points, *xy)
				x, y := xy[0], xy[1]
				if x < result.XRange[0] {
					result.XRange[0] = x
				}
				if x > result.XRange[1] {
					result.XRange[1] = x
				}
				if y < result.YRange[0] {
					result.YRange[0] = y
				}
				if y > result.YRange[1] {
					result.YRange[1] = y
				}
			}
		}
	}

	result.Points = len(points)
	if len(points) > 50 {
		result.Preview = points[:50]
	} else {
		result.Preview = points
	}

	return result, scanner.Err()
}

// parseXYLine tries to parse a line as "x y" or "x,y".
func parseXYLine(line string) *[2]float64 {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	// Try space-separated
	parts := strings.Fields(line)
	if len(parts) >= 2 {
		x, err1 := strconv.ParseFloat(parts[0], 64)
		y, err2 := strconv.ParseFloat(parts[1], 64)
		if err1 == nil && err2 == nil {
			return &[2]float64{x, y}
		}
	}

	// Try comma-separated
	parts = strings.Split(line, ",")
	if len(parts) >= 2 {
		x, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		y, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err1 == nil && err2 == nil {
			return &[2]float64{x, y}
		}
	}

	return nil
}
