package contract_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/sachncs/alfred/internal/contract"
)

func TestThreadJSONRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	in := contract.Thread{
		ID:        "thr-1",
		Title:     "smoke",
		Status:    contract.ThreadStatusIdle,
		CreatedAt: now,
		UpdatedAt: now,
		Metadata:  map[string]any{"k": "v"},
	}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out contract.Thread
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.ID != in.ID || out.Title != in.Title || out.Status != in.Status {
		t.Fatalf("round-trip mismatch: got %+v", out)
	}
}

func TestTurnItemDiscriminator(t *testing.T) {
	text := "hello"
	item := contract.TurnItem{
		ID:   "it-1",
		Kind: contract.ItemKindAssistantText,
		Text: &text,
	}
	b, _ := json.Marshal(item)
	if !strings.Contains(string(b), `"kind":"assistant_text"`) {
		t.Fatalf("kind discriminator missing: %s", b)
	}
}

func TestErrorEnvelope(t *testing.T) {
	e := contract.ErrorfWith(contract.CodeValidation, "bad", map[string]any{"field": "title"})
	b, _ := json.Marshal(e)
	s := string(b)
	if !strings.Contains(s, `"code":"validation"`) {
		t.Fatalf("code missing: %s", s)
	}
	if !strings.Contains(s, `"field":"title"`) {
		t.Fatalf("details missing: %s", s)
	}
}

func TestHealthResponse(t *testing.T) {
	h := contract.HealthResponse{Status: "ok", Version: "0.1.0", Timestamp: time.Now()}
	b, err := json.Marshal(h)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(b), `"status":"ok"`) {
		t.Fatalf("status missing: %s", b)
	}
}

func TestErrorf(t *testing.T) {
	e := contract.Errorf(contract.CodeNotFound, "thread not found")
	b, _ := json.Marshal(e)
	s := string(b)
	if !strings.Contains(s, `"code":"not_found"`) {
		t.Fatalf("code missing: %s", s)
	}
	if !strings.Contains(s, `"message":"thread not found"`) {
		t.Fatalf("message missing: %s", s)
	}
	if strings.Contains(s, `"details"`) {
		t.Fatalf("details should be omitted for Errorf: %s", s)
	}
}

func TestErrorCodes(t *testing.T) {
	codes := []string{
		contract.CodeUnauthorized,
		contract.CodeForbidden,
		contract.CodeNotFound,
		contract.CodeValidation,
		contract.CodeTurnInProgress,
		contract.CodeUnavailable,
		contract.CodeInternal,
	}
	seen := make(map[string]bool)
	for _, c := range codes {
		if c == "" {
			t.Fatal("empty error code")
		}
		if seen[c] {
			t.Fatalf("duplicate error code: %s", c)
		}
		seen[c] = true
	}
}
