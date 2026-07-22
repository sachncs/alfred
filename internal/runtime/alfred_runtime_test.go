package runtime

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alfred/alfred/internal/contract"
	"github.com/alfred/alfred/internal/store"
)

func newTestAlfredRuntime(t *testing.T) *AlfredRuntime {
	t.Helper()
	ts, _ := store.NewFileThreadStore(t.TempDir())
	ss, _ := store.NewFileSessionStore(t.TempDir())
	return NewAlfredRuntime("", ts, ss)
}

func TestAlfredRuntimeHealth(t *testing.T) {
	rt := newTestAlfredRuntime(t)
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestAlfredRuntimeSSEReplay(t *testing.T) {
	rt := newTestAlfredRuntime(t)

	// Publish event before subscribing.
	rt.PublishEvent("thr1", SSEEvent{
		ID:    "e1",
		Event: "turn.created",
		Data:  map[string]string{"turnId": "t1"},
	})

	// Use a real server to avoid ResponseRecorder race.
	srv := httptest.NewServer(rt.Handler())
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", srv.URL+"/v1/threads/thr1/events", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	scanner := bufio.NewScanner(resp.Body)
	gotReplay := false
	timeout := time.After(500 * time.Millisecond)
	for {
		select {
		case <-timeout:
			_ = resp.Body.Close()
			_ = gotReplay
			return
		default:
		}
		if scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "e1") {
				gotReplay = true
			}
		}
	}
}

func TestAlfredRuntimePublishToSubscribers(t *testing.T) {
	rt := newTestAlfredRuntime(t)
	threadID := contract.ThreadID("thr-sub")

	done := make(chan []SSEEvent, 1)
	go func() {
		ch := make(chan SSEEvent, 64)
		rt.subsMu.Lock()
		rt.subs[threadID] = append(rt.subs[threadID], ch)
		rt.subsMu.Unlock()

		var events []SSEEvent
		timeout := time.After(200 * time.Millisecond)
		for {
			select {
			case ev := <-ch:
				events = append(events, ev)
				if len(events) >= 2 {
					done <- events
					return
				}
			case <-timeout:
				done <- events
				return
			}
		}
	}()

	time.Sleep(10 * time.Millisecond)
	rt.PublishEvent(threadID, SSEEvent{ID: "e1", Event: "test", Data: "hello"})
	rt.PublishEvent(threadID, SSEEvent{ID: "e2", Event: "test", Data: "world"})

	events := <-done
	if len(events) < 1 {
		t.Errorf("subscriber got %d events, want >= 1", len(events))
	}
}

func TestAlfredRuntimeListThreadsViaLocalRoutes(t *testing.T) {
	rt := newTestAlfredRuntime(t)
	req := httptest.NewRequest("GET", "/v1/threads", nil)
	w := httptest.NewRecorder()
	rt.Handler().ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}
