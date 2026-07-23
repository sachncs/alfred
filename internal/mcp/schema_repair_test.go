package mcp_test

import (
	"testing"

	"github.com/alfred/alfred/internal/mcp"
)

func TestRepairSchemaFences(t *testing.T) {
	input := "```json\n{\"name\":\"test\"}\n```"
	fixed, errs := mcp.RepairSchema(input)
	if fixed != `{"name":"test"}` {
		t.Errorf("got %q", fixed)
	}
	if len(errs) == 0 {
		t.Error("expected repair notes")
	}
}

func TestRepairSchemaTrailingComma(t *testing.T) {
	input := `{"a": 1, "b": 2,}`
	fixed, _ := mcp.RepairSchema(input)
	if fixed != `{"a": 1, "b": 2}` {
		t.Errorf("got %q", fixed)
	}
}

func TestRepairSchemaValid(t *testing.T) {
	input := `{"valid": true}`
	fixed, errs := mcp.RepairSchema(input)
	if fixed != input {
		t.Errorf("modified valid input: %q", fixed)
	}
	// No repair notes for valid input
	for _, e := range errs {
		if e != "could not repair schema" {
			t.Errorf("unexpected note: %s", e)
		}
	}
}
