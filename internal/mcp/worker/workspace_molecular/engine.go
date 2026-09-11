package workspacemolecular

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sachncs/alfred/internal/mcp/worker"
	"github.com/sachncs/alfred/internal/tool"
)

// PDBInfo holds parsed PDB file information.
type PDBInfo struct {
	Path          string       `json:"path"`
	Title         string       `json:"title,omitempty"`
	Method        string       `json:"method,omitempty"`
	Resolution    float64      `json:"resolution,omitempty"`
	Chains        []ChainInfo  `json:"chains"`
	TotalAtoms    int          `json:"totalAtoms"`
	TotalResidues int          `json:"totalResidues"`
	BoundingBox   *BoundingBox `json:"boundingBox,omitempty"`
}

// ChainInfo describes a protein chain.
type ChainInfo struct {
	ID           string `json:"id"`
	ResidueCount int    `json:"residueCount"`
	AtomCount    int    `json:"atomCount"`
}

// BoundingBox is the spatial extent of the structure.
type BoundingBox struct {
	MinX, MinY, MinZ float64
	MaxX, MaxY, MaxZ float64
}

func NewMolecularServer() *worker.WorkerServer {
	return worker.NewWorkerServer("workspace-molecular-worker", []tool.Tool{
		&molecularPreviewTool{},
		&molecularUpdateWorkbenchTool{},
	})
}

type molecularPreviewTool struct{}

func (t *molecularPreviewTool) Name() string        { return "molecular_preview" }
func (t *molecularPreviewTool) Description() string { return "Preview a PDB molecular structure" }
func (t *molecularPreviewTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}`)
}
func (t *molecularPreviewTool) Execute(_ context.Context, raw json.RawMessage, tc *tool.Context) (*tool.Result, error) {
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

	info, err := parsePDB(abs)
	if err != nil {
		return tool.Failure(err), nil
	}

	return tool.SuccessWith(fmt.Sprintf("%d atoms, %d residues", info.TotalAtoms, info.TotalResidues), info), nil
}

type molecularUpdateWorkbenchTool struct{}

func (t *molecularUpdateWorkbenchTool) Name() string { return "molecular_update_workbench" }
func (t *molecularUpdateWorkbenchTool) Description() string {
	return "Update the molecular workbench"
}
func (t *molecularUpdateWorkbenchTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"},"mutations":{"type":"array","items":{"type":"object"}}},"required":["path"]}`)
}
func (t *molecularUpdateWorkbenchTool) Execute(_ context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in struct {
		Path      string `json:"path"`
		Mutations any    `json:"mutations"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}
	if in.Path == "" {
		return tool.FailureMsg("path is required"), nil
	}
	return tool.SuccessWith("Workbench updated", map[string]any{
		"path":   in.Path,
		"status": "updated",
	}), nil
}

// parsePDB reads a PDB file and extracts structural information.
func parsePDB(path string) (*PDBInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	info := &PDBInfo{
		Path: path,
		BoundingBox: &BoundingBox{
			MinX: 1e9, MinY: 1e9, MinZ: 1e9,
			MaxX: -1e9, MaxY: -1e9, MaxZ: -1e9,
		},
	}

	chainMap := make(map[string]*chainData)
	residueSet := make(map[string]bool)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 6 {
			continue
		}
		record := strings.TrimSpace(line[:6])

		switch {
		case record == "TITLE":
			if len(line) > 10 {
				info.Title += strings.TrimSpace(line[10:]) + " "
			}
		case record == "EXPDTA":
			if len(line) > 10 {
				info.Method = strings.TrimSpace(line[10:])
			}
		case record == "REMARK" && len(line) > 11 && strings.TrimSpace(line[6:11]) == "2":
			parts := strings.Fields(line)
			for i, p := range parts {
				if p == "RESOLUTION." && i+1 < len(parts) {
					if v, err := strconv.ParseFloat(parts[i+1], 64); err == nil {
						info.Resolution = v
					}
				}
			}
		case record == "ATOM" || record == "HETATM":
			info.TotalAtoms++
			chainID := " "
			if len(line) > 21 {
				chainID = string(line[21])
			}
			resNum := 0
			if len(line) > 26 {
				resStr := strings.TrimSpace(line[22:26])
				resNum, _ = strconv.Atoi(resStr)
			}
			if len(line) > 54 {
				x := parseCoord(line[31:38])
				y := parseCoord(line[39:46])
				z := parseCoord(line[47:54])
				if x < info.BoundingBox.MinX {
					info.BoundingBox.MinX = x
				}
				if x > info.BoundingBox.MaxX {
					info.BoundingBox.MaxX = x
				}
				if y < info.BoundingBox.MinY {
					info.BoundingBox.MinY = y
				}
				if y > info.BoundingBox.MaxY {
					info.BoundingBox.MaxY = y
				}
				if z < info.BoundingBox.MinZ {
					info.BoundingBox.MinZ = z
				}
				if z > info.BoundingBox.MaxZ {
					info.BoundingBox.MaxZ = z
				}
			}

			cd, ok := chainMap[chainID]
			if !ok {
				cd = &chainData{id: chainID}
				chainMap[chainID] = cd
			}
			cd.atomCount++
			resKey := fmt.Sprintf("%s:%d", chainID, resNum)
			if !residueSet[resKey] {
				residueSet[resKey] = true
				cd.residueCount++
				info.TotalResidues++
			}
		}
	}

	chains := make([]ChainInfo, 0, len(chainMap))
	for _, cd := range chainMap {
		chains = append(chains, ChainInfo{
			ID:           cd.id,
			ResidueCount: cd.residueCount,
			AtomCount:    cd.atomCount,
		})
	}
	info.Chains = chains

	return info, scanner.Err()
}

type chainData struct {
	id           string
	residueCount int
	atomCount    int
}

func parseCoord(s string) float64 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return v
}
