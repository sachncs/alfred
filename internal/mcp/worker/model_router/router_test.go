package modelrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/sachncs/alfred/internal/tool"
)

func TestTruncateSnippet(t *testing.T) {
	body := bytes.Repeat([]byte("a"), 1024)
	got := truncateSnippet(body)
	if len(got) >= len(body) {
		t.Fatalf("expected truncation, got len=%d body=%d", len(got), len(body))
	}
	if !strings.Contains(got, "[truncated]") {
		t.Fatalf("missing truncation marker: %q", got)
	}
}

func TestTruncateSnippetShort(t *testing.T) {
	got := truncateSnippet([]byte("hello"))
	if got != "hello" {
		t.Fatalf("expected pass-through, got %q", got)
	}
}

func TestProxyOpenAIAuthMissing(t *testing.T) {
	_ = os.Unsetenv("OPENAI_API_KEY")
	p := &proxyTool{router: &modelRouter{client: http.DefaultClient}}
	res, err := p.proxyOpenAI(context.Background(), proxyInput{Model: "gpt-4o", Messages: json.RawMessage(`[]`)})
	if err != nil {
		t.Fatal(err)
	}
	if res.OK {
		t.Fatal("expected failure when key missing")
	}
	if !strings.Contains(res.Error, "OPENAI_API_KEY") {
		t.Fatalf("unexpected error: %s", res.Error)
	}
}

func TestProxyAnthropicAuthMissing(t *testing.T) {
	_ = os.Unsetenv("ANTHROPIC_API_KEY")
	p := &proxyTool{router: &modelRouter{client: http.DefaultClient}}
	res, err := p.proxyAnthropic(context.Background(), proxyInput{Model: "claude-3-opus", Messages: json.RawMessage(`[]`)})
	if err != nil {
		t.Fatal(err)
	}
	if res.OK {
		t.Fatal("expected failure when key missing")
	}
	if !strings.Contains(res.Error, "ANTHROPIC_API_KEY") {
		t.Fatalf("unexpected error: %s", res.Error)
	}
}

var _ = tool.FailureMsg
