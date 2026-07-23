package web_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/alfred/alfred/internal/contract"
	"github.com/alfred/alfred/web"
)

type stubThreads struct {
	threads []contract.Thread
}

func (s *stubThreads) List(_ int, _ string) ([]contract.Thread, string, error) {
	return s.threads, "", nil
}
func (s *stubThreads) Get(id contract.ThreadID) (*contract.Thread, error) {
	for _, t := range s.threads {
		if t.ID == id {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("not found")
}
func (s *stubThreads) Create(t *contract.Thread) error {
	s.threads = append(s.threads, *t)
	return nil
}

type stubSession struct{}

func (s *stubSession) Read(_ contract.ThreadID, _ int) ([]contract.TurnItem, error) {
	return nil, nil
}

func TestHandleChat(t *testing.T) {
	h := web.NewHandlers()
	h.Threads = &stubThreads{threads: []contract.Thread{
		{ID: "thr-1", Title: "Test Thread"},
	}}
	h.Session = &stubSession{}

	mux := http.NewServeMux()
	web.Register(mux, h, "")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))

	if w.Code != 200 {
		t.Fatalf("GET / = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Fatalf("Content-Type = %q", ct)
	}
}

func TestHandleDesign(t *testing.T) {
	h := web.NewHandlers()

	mux := http.NewServeMux()
	web.Register(mux, h, "")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/design", nil))

	if w.Code != 200 {
		t.Fatalf("GET /design = %d, want 200", w.Code)
	}
}

func TestHandleSidebar(t *testing.T) {
	h := web.NewHandlers()
	h.Threads = &stubThreads{threads: []contract.Thread{
		{ID: "thr-1", Title: "Alpha"},
		{ID: "thr-2", Title: "Beta"},
	}}

	mux := http.NewServeMux()
	web.Register(mux, h, "")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/ui/sidebar", nil))

	if w.Code != 200 {
		t.Fatalf("GET /ui/sidebar = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !contains(body, "Alpha") || !contains(body, "Beta") {
		t.Fatalf("sidebar missing threads: %s", body)
	}
}

func TestHandleCreateThread(t *testing.T) {
	threads := &stubThreads{}
	h := web.NewHandlers()
	h.Threads = threads
	h.Create = threads

	mux := http.NewServeMux()
	web.Register(mux, h, "")

	form := url.Values{}
	form.Set("title", "My Thread")
	req := httptest.NewRequest("POST", "/ui/threads", nil)
	req.Form = form
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("POST /ui/threads = %d, want 200", w.Code)
	}
	if len(threads.threads) != 1 {
		t.Fatalf("expected 1 thread, got %d", len(threads.threads))
	}
	if threads.threads[0].Title != "My Thread" {
		t.Fatalf("thread title = %q, want %q", threads.threads[0].Title, "My Thread")
	}
}

func TestHandleTimelineEmpty(t *testing.T) {
	h := web.NewHandlers()
	h.Session = &stubSession{}

	mux := http.NewServeMux()
	web.Register(mux, h, "")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/ui/threads/thr-1/timeline", nil))

	if w.Code != 200 {
		t.Fatalf("GET /ui/threads/thr-1/timeline = %d, want 200", w.Code)
	}
}

func TestSubmitTurnEmpty(t *testing.T) {
	h := web.NewHandlers()
	h.Thread = &stubThreads{}

	mux := http.NewServeMux()
	web.Register(mux, h, "")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/ui/threads/thr-1/turns", nil))

	if w.Code != 400 {
		t.Fatalf("POST /ui/threads/thr-1/turns empty = %d, want 400", w.Code)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
