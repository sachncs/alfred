package runtime

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alfred/alfred/internal/contract"
	"github.com/alfred/alfred/internal/store"
)

func TestHTTPRuntimeHealth(t *testing.T) {
	ts, _ := store.NewFileThreadStore(t.TempDir())
	ss, _ := store.NewFileSessionStore(t.TempDir())
	rt := NewHTTPRuntime("", ts, ss)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
	var resp contract.HealthResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp.Status != "ok" {
		t.Errorf("status = %q, want ok", resp.Status)
	}
}

func TestHTTPRuntimeHealthz(t *testing.T) {
	ts, _ := store.NewFileThreadStore(t.TempDir())
	ss, _ := store.NewFileSessionStore(t.TempDir())
	rt := NewHTTPRuntime("", ts, ss)

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestBearerAuthValid(t *testing.T) {
	ts, _ := store.NewFileThreadStore(t.TempDir())
	ss, _ := store.NewFileSessionStore(t.TempDir())
	rt := NewHTTPRuntime("secret-token", ts, ss)

	req := httptest.NewRequest("GET", "/health", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestBearerAuthInvalid(t *testing.T) {
	ts, _ := store.NewFileThreadStore(t.TempDir())
	ss, _ := store.NewFileSessionStore(t.TempDir())
	rt := NewHTTPRuntime("secret-token", ts, ss)

	req := httptest.NewRequest("GET", "/health", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)

	if w.Code != 401 {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestBearerAuthMissing(t *testing.T) {
	ts, _ := store.NewFileThreadStore(t.TempDir())
	ss, _ := store.NewFileSessionStore(t.TempDir())
	rt := NewHTTPRuntime("secret-token", ts, ss)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)

	if w.Code != 401 {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestBearerAuthDisabledWhenEmpty(t *testing.T) {
	ts, _ := store.NewFileThreadStore(t.TempDir())
	ss, _ := store.NewFileSessionStore(t.TempDir())
	rt := NewHTTPRuntime("", ts, ss)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200 (auth disabled)", w.Code)
	}
}

func TestHTTPRuntimeStoreAccessors(t *testing.T) {
	ts, _ := store.NewFileThreadStore(t.TempDir())
	ss, _ := store.NewFileSessionStore(t.TempDir())
	rt := NewHTTPRuntime("", ts, ss)

	if rt.ThreadStore() == nil {
		t.Error("ThreadStore() returned nil")
	}
	if rt.SessionStore() == nil {
		t.Error("SessionStore() returned nil")
	}
	if rt.State().Status != "idle" {
		t.Errorf("initial status = %q, want idle", rt.State().Status)
	}
}

func TestHTTPRuntimeStartReturnsOnCancel(t *testing.T) {
	ts, _ := store.NewFileThreadStore(t.TempDir())
	ss, _ := store.NewFileSessionStore(t.TempDir())
	rt := NewHTTPRuntime("", ts, ss)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- rt.Start(ctx, 18234)
	}()
	time.AfterFunc(50*time.Millisecond, cancel)
	err := <-done
	// Server should stop after context cancel.
	_ = err
}

func TestBearerAuthMalformedHeader(t *testing.T) {
	ts, _ := store.NewFileThreadStore(t.TempDir())
	ss, _ := store.NewFileSessionStore(t.TempDir())
	rt := NewHTTPRuntime("token", ts, ss)

	req := httptest.NewRequest("GET", "/health", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)

	if w.Code != 401 {
		t.Errorf("status = %d, want 401 for Basic auth", w.Code)
	}
}

// State returns the runtime state.
func (r *HTTPRuntime) State() *RuntimeState {
	return r.state
}
