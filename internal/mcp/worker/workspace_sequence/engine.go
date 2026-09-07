package workspacesequence

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/alfred/alfred/internal/mcp/worker"
	"github.com/alfred/alfred/internal/tool"
)

// SequenceRecord represents one sequence from a FASTA file.
type SequenceRecord struct {
	ID          string  `json:"id"`
	Description string  `json:"description,omitempty"`
	Length      int     `json:"length"`
	Type        string  `json:"type"` // "DNA", "RNA", "protein"
	GCContent   float64 `json:"gcContent,omitempty"`
	First100    string  `json:"first100,omitempty"`
}

func NewSequenceServer() *worker.WorkerServer {
	return worker.NewWorkerServer("workspace-sequence-worker", []tool.Tool{
		&sequencePreviewTool{},
	})
}

type sequencePreviewTool struct{}

func (t *sequencePreviewTool) Name() string { return "sequence_preview" }
func (t *sequencePreviewTool) Description() string {
	return "Preview a sequence record (FASTA/GenBank)"
}
func (t *sequencePreviewTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}`)
}
func (t *sequencePreviewTool) Execute(_ context.Context, raw json.RawMessage, tc *tool.Context) (*tool.Result, error) {
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

	sequences, err := parseFasta(abs)
	if err != nil {
		return tool.Failure(err), nil
	}

	return tool.SuccessWith(fmt.Sprintf("%d sequences", len(sequences)), map[string]any{
		"path":      in.Path,
		"sequences": sequences,
		"count":     len(sequences),
	}), nil
}

// parseFasta reads a FASTA file and returns sequence records.
func parseFasta(path string) ([]SequenceRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var sequences []SequenceRecord
	var current *SequenceRecord
	var seqBuilder strings.Builder

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r\n")
		if strings.HasPrefix(line, ">") {
			// Save previous sequence
			if current != nil {
				current.Length = seqBuilder.Len()
				seq := seqBuilder.String()
				current.Type = detectType(seq)
				if current.Type == "DNA" || current.Type == "RNA" {
					current.GCContent = gcContent(seq)
				}
				if current.Length > 100 {
					current.First100 = seq[:100]
				} else {
					current.First100 = seq
				}
				sequences = append(sequences, *current)
			}
			// Parse header
			header := line[1:]
			parts := strings.SplitN(header, " ", 2)
			id := parts[0]
			desc := ""
			if len(parts) > 1 {
				desc = parts[1]
			}
			current = &SequenceRecord{ID: id, Description: desc}
			seqBuilder.Reset()
		} else if current != nil {
			seqBuilder.WriteString(strings.ToUpper(strings.TrimSpace(line)))
		}
	}

	// Don't forget the last sequence
	if current != nil {
		current.Length = seqBuilder.Len()
		seq := seqBuilder.String()
		current.Type = detectType(seq)
		if current.Type == "DNA" || current.Type == "RNA" {
			current.GCContent = gcContent(seq)
		}
		if current.Length > 100 {
			current.First100 = seq[:100]
		} else {
			current.First100 = seq
		}
		sequences = append(sequences, *current)
	}

	return sequences, scanner.Err()
}

// detectType determines if a sequence is DNA, RNA, or protein.
func detectType(seq string) string {
	if seq == "" {
		return "unknown"
	}
	hasU := false
	hasT := false
	ambiguous := 0
	for _, r := range seq {
		switch r {
		case 'U':
			hasU = true
		case 'T':
			hasT = true
		case 'A', 'C', 'G':
			// nucleotide, continue
		default:
			if !unicode.IsLetter(r) {
				continue
			}
			ambiguous++
		}
	}
	if hasU && !hasT {
		return "RNA"
	}
	if hasT && !hasU {
		return "DNA"
	}
	// If neither T nor U, and mostly standard bases, assume DNA
	if !hasT && !hasU && ambiguous == 0 {
		return "DNA"
	}
	return "protein"
}

// gcContent computes the GC percentage of a nucleotide sequence.
func gcContent(seq string) float64 {
	if seq == "" {
		return 0
	}
	var gc, total int
	for _, r := range seq {
		if r == 'G' || r == 'C' || r == 'g' || r == 'c' {
			gc++
		}
		if unicode.IsLetter(r) {
			total++
		}
	}
	if total == 0 {
		return 0
	}
	return float64(gc) / float64(total) * 100
}
