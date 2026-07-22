package runtime

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alfred/alfred/internal/contract"
	"github.com/alfred/alfred/internal/store"
)

func newLocalWithStores(t *testing.T) *LocalRuntime {
	t.Helper()
	ts, _ := store.NewFileThreadStore(t.TempDir())
	ss, _ := store.NewFileSessionStore(t.TempDir())
	return NewLocalRuntime("", ts, ss)
}

func TestMemoryRoundTrip(t *testing.T) {
	rt := newLocalWithStores(t)

	// Set.
	body, _ := json.Marshal(map[string]any{"key": "theme", "value": "dark"})
	req := httptest.NewRequest("POST", "/v1/memory/user", bytes.NewReader(body))
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("set: %d", w.Code)
	}

	// List.
	req = httptest.NewRequest("GET", "/v1/memory/user", nil)
	w = httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("list: %d", w.Code)
	}
	var resp struct {
		Entries []MemoryEntry `json:"entries"`
	}
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if len(resp.Entries) != 1 || resp.Entries[0].Key != "theme" {
		t.Errorf("entries = %+v", resp.Entries)
	}

	// Delete.
	req = httptest.NewRequest("DELETE", "/v1/memory/user/theme", nil)
	w = httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	if w.Code != 204 {
		t.Fatalf("delete: %d", w.Code)
	}

	// List again, should be empty.
	req = httptest.NewRequest("GET", "/v1/memory/user", nil)
	w = httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if len(resp.Entries) != 0 {
		t.Errorf("after delete: entries = %+v", resp.Entries)
	}
}

func TestMemoryInvalidScope(t *testing.T) {
	rt := newLocalWithStores(t)
	req := httptest.NewRequest("GET", "/v1/memory/bogus", nil)
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	if w.Code != 400 {
		t.Errorf("invalid scope: %d, want 400", w.Code)
	}
}

func TestSkillsRoute(t *testing.T) {
	rt := newLocalWithStores(t)
	rt.SkillsLoader().SetWorkspaceRoot(t.TempDir())
	req := httptest.NewRequest("GET", "/v1/skills", nil)
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("status = %d", w.Code)
	}
	var resp struct {
		Skills []SkillInfo `json:"skills"`
	}
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp.Skills == nil {
		t.Error("skills should be empty slice, not nil")
	}
}

func TestUsageRoute(t *testing.T) {
	rt := newLocalWithStores(t)
	rt.UsageTracker().Record("gpt-4", 100, 50)
	req := httptest.NewRequest("GET", "/v1/usage", nil)
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"input":100`) || !strings.Contains(body, `"output":50`) {
		t.Errorf("body = %s", body)
	}
}

func TestApprovalLifecycle(t *testing.T) {
	rt := newLocalWithStores(t)
	req := rt.ApprovalStore().Pending("thr1", "bash", "needs approval", nil)

	// List.
	listReq := httptest.NewRequest("GET", "/v1/approvals", nil)
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, listReq)
	if w.Code != 200 {
		t.Fatalf("list: %d", w.Code)
	}

	// Resolve.
	body, _ := json.Marshal(map[string]bool{"approved": true})
	resReq := httptest.NewRequest("POST", "/v1/approvals/"+req.ID+"/resolve", bytes.NewReader(body))
	w = httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, resReq)
	if w.Code != 200 {
		t.Errorf("resolve: %d", w.Code)
	}

	// Resolve again should 404 (already resolved).
	resReq = httptest.NewRequest("POST", "/v1/approvals/"+req.ID+"/resolve", bytes.NewReader(body))
	w = httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, resReq)
	if w.Code != 404 {
		t.Errorf("double-resolve: %d, want 404", w.Code)
	}
}

func TestUserInputResolve(t *testing.T) {
	rt := newLocalWithStores(t)
	req := rt.userInputs.Pending("thr1", "what color?", nil)

	body, _ := json.Marshal(map[string]string{"value": "blue"})
	resReq := httptest.NewRequest("POST", "/v1/user-inputs/"+req.ID+"/resolve", bytes.NewReader(body))
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, resReq)
	if w.Code != 200 {
		t.Fatalf("resolve: %d", w.Code)
	}

	got, ok := rt.userInputs.GetValue(req.ID)
	if !ok || got != "blue" {
		t.Errorf("value = %q, ok=%v", got, ok)
	}
}

func TestAttachmentUpload(t *testing.T) {
	rt := newLocalWithStores(t)
	attDir := t.TempDir()
	if err := rt.AttachmentStore().SetDir(attDir); err != nil {
		t.Fatal(err)
	}

	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	_ = w.WriteField("threadId", "thr1")
	fw, _ := w.CreateFormFile("file", "hello.txt")
	_, _ = fw.Write([]byte("hello world"))
	_ = w.Close()

	req := httptest.NewRequest("POST", "/v1/attachments", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(rec, req)
	if rec.Code != 201 {
		t.Fatalf("upload: %d, body=%s", rec.Code, rec.Body.String())
	}
	var att Attachment
	_ = json.NewDecoder(rec.Body).Decode(&att)
	if att.Size != 11 {
		t.Errorf("size = %d", att.Size)
	}
	if att.Name != "hello.txt" {
		t.Errorf("name = %q", att.Name)
	}
}

func TestAttachmentOversize(t *testing.T) {
	rt := newLocalWithStores(t)
	_ = rt.AttachmentStore().SetDir(t.TempDir())

	// Build multipart with huge file.
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	fw, _ := w.CreateFormFile("file", "big.bin")
	// Write just over the limit.
	chunk := bytes.Repeat([]byte("a"), 1024*1024)
	for i := 0; i <= 50; i++ {
		_, _ = fw.Write(chunk)
	}
	_ = w.Close()

	req := httptest.NewRequest("POST", "/v1/attachments", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	rt.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("oversize: %d, want 413", rec.Code)
	}
}

var _ = contract.TurnStatusRunning
