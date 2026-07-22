package runtime

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/alfred/alfred/internal/contract"
	"github.com/alfred/alfred/internal/store"
)

func newTestLocalRuntime(t *testing.T) *LocalRuntime {
	t.Helper()
	ts, _ := store.NewFileThreadStore(t.TempDir())
	ss, _ := store.NewFileSessionStore(t.TempDir())
	return NewLocalRuntime("", ts, ss)
}

func TestLocalRuntimeListThreadsEmpty(t *testing.T) {
	rt := newTestLocalRuntime(t)
	req := httptest.NewRequest("GET", "/v1/threads", nil)
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
	var resp contract.ListThreadsResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if len(resp.Threads) != 0 {
		t.Errorf("threads = %d, want 0", len(resp.Threads))
	}
}

func TestLocalRuntimeCreateAndGetThread(t *testing.T) {
	rt := newTestLocalRuntime(t)

	// Create
	body, _ := json.Marshal(contract.CreateThreadRequest{Title: "test"})
	req := httptest.NewRequest("POST", "/v1/threads", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)

	if w.Code != 201 {
		t.Fatalf("create status = %d, want 201", w.Code)
	}
	var createResp contract.CreateThreadResponse
	json.NewDecoder(w.Body).Decode(&createResp)
	if createResp.Thread.Title != "test" {
		t.Errorf("title = %q, want test", createResp.Thread.Title)
	}

	// Get
	req = httptest.NewRequest("GET", "/v1/threads/"+string(createResp.Thread.ID), nil)
	w = httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("get status = %d, want 200", w.Code)
	}
}

func TestLocalRuntimeDeleteThread(t *testing.T) {
	rt := newTestLocalRuntime(t)

	// Create
	body, _ := json.Marshal(contract.CreateThreadRequest{Title: "to-delete"})
	req := httptest.NewRequest("POST", "/v1/threads", bytes.NewReader(body))
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	var createResp contract.CreateThreadResponse
	json.NewDecoder(w.Body).Decode(&createResp)

	// Delete
	req = httptest.NewRequest("DELETE", "/v1/threads/"+string(createResp.Thread.ID), nil)
	w = httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)

	if w.Code != 204 {
		t.Errorf("delete status = %d, want 204", w.Code)
	}
}

func TestLocalRuntimeGetSession(t *testing.T) {
	rt := newTestLocalRuntime(t)
	req := httptest.NewRequest("GET", "/v1/sessions/nonexistent", nil)
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestLocalRuntimeWorkspaceStatus(t *testing.T) {
	rt := newTestLocalRuntime(t)
	req := httptest.NewRequest("GET", "/v1/workspace/status", nil)
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestLocalRuntimeToolDiagnostics(t *testing.T) {
	rt := newTestLocalRuntime(t)
	req := httptest.NewRequest("GET", "/v1/tool-diagnostics", nil)
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestLocalRuntimeUsage(t *testing.T) {
	rt := newTestLocalRuntime(t)
	req := httptest.NewRequest("GET", "/v1/usage", nil)
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestLocalRuntimeMemory(t *testing.T) {
	rt := newTestLocalRuntime(t)

	// List
	req := httptest.NewRequest("GET", "/v1/memory/user", nil)
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("list status = %d, want 200", w.Code)
	}

	// Set
	body, _ := json.Marshal(map[string]any{"key": "theme", "value": "dark"})
	req = httptest.NewRequest("POST", "/v1/memory/user", bytes.NewReader(body))
	w = httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("set status = %d, want 200", w.Code)
	}
}

func TestLocalRuntimeSkills(t *testing.T) {
	rt := newTestLocalRuntime(t)
	req := httptest.NewRequest("GET", "/v1/skills", nil)
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}
