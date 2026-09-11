package providers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDoGetReadsFullBody(t *testing.T) {
	body := strings.Repeat("abcdefghij", 20000) // 200 KiB > old 64 KiB buffer
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	got, err := doGet(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("doGet: %v", err)
	}
	if len(got) != len(body) {
		t.Errorf("got %d bytes, want %d", len(got), len(body))
	}
}

func TestDoPostReadsFullBody(t *testing.T) {
	body := strings.Repeat("xyz", 50000) // 150 KiB
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	got, err := doPost(context.Background(), srv.URL, "application/json", []byte(`{}`))
	if err != nil {
		t.Fatalf("doPost: %v", err)
	}
	if len(got) != len(body) {
		t.Errorf("got %d bytes, want %d", len(got), len(body))
	}
}
