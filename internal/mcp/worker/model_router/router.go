// Package worker provides the model-router MCP worker.
package modelrouter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	internallog "github.com/sachncs/alfred/internal/log"
	"github.com/sachncs/alfred/internal/mcp/worker"
	"github.com/sachncs/alfred/internal/tool"
)

var errLog = internallog.New()

// NewModelRouterWorker creates a model-router MCP worker.
// ponytail: proxies to upstream API based on model prefix.
func NewModelRouterWorker() *worker.WorkerServer {
	router := &modelRouter{
		client: &http.Client{Timeout: 120 * time.Second},
	}
	proxyTool := &proxyTool{router: router}
	return worker.NewWorkerServer("model-router", []tool.Tool{proxyTool}).WithHandler(
		func(ctx context.Context, name string, input json.RawMessage, tc *tool.Context) (*tool.Result, error) {
			if name == "model_proxy" {
				return proxyTool.Execute(ctx, input, tc)
			}
			return tool.FailureMsg(fmt.Sprintf("unknown tool: %s", name)), nil
		},
	)
}

type modelRouter struct {
	client *http.Client
}

type proxyTool struct {
	router *modelRouter
}

func (p *proxyTool) Name() string { return "model_proxy" }

func (p *proxyTool) Description() string {
	return "Proxy a request to an upstream LLM API (OpenAI, Anthropic, etc.). Routes based on model name prefix."
}

func (p *proxyTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"model": {"type": "string", "description": "Model name (e.g. gpt-4, claude-3-opus)"},
			"messages": {"type": "array", "description": "Chat messages"},
			"maxTokens": {"type": "integer", "description": "Max tokens"}
		},
		"required": ["model", "messages"]
	}`)
}

func (p *proxyTool) Execute(ctx context.Context, raw json.RawMessage, _ *tool.Context) (*tool.Result, error) {
	var in proxyInput
	if err := json.Unmarshal(raw, &in); err != nil {
		return tool.FailureMsg(fmt.Sprintf("invalid input: %v", err)), nil
	}

	// Route based on model prefix
	if strings.HasPrefix(in.Model, "claude") {
		return p.proxyAnthropic(ctx, in)
	}
	return p.proxyOpenAI(ctx, in)
}

type proxyInput struct {
	Model     string          `json:"model"`
	Messages  json.RawMessage `json:"messages"`
	MaxTokens int             `json:"maxTokens"`
}

func (p *proxyTool) proxyOpenAI(ctx context.Context, in proxyInput) (*tool.Result, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return tool.FailureMsg("OPENAI_API_KEY not set"), nil
	}

	body, _ := json.Marshal(map[string]any{
		"model":      in.Model,
		"messages":   json.RawMessage(in.Messages),
		"max_tokens": in.MaxTokens,
	})

	req, _ := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := p.router.client.Do(req)
	if err != nil {
		return tool.Failure(err), nil
	}
	defer resp.Body.Close() //nolint:errcheck

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		errLog.Printf("model_router: openai %d body=%s", resp.StatusCode, internallog.Redact(string(respBody)))
		return tool.FailureMsg(fmt.Sprintf("openai %d: %s", resp.StatusCode, truncateSnippet(respBody))), nil
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	_ = json.Unmarshal(respBody, &result)

	content := ""
	if len(result.Choices) > 0 {
		content = result.Choices[0].Message.Content
	}
	return tool.SuccessWith(content, json.RawMessage(respBody)), nil
}

func (p *proxyTool) proxyAnthropic(ctx context.Context, in proxyInput) (*tool.Result, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return tool.FailureMsg("ANTHROPIC_API_KEY not set"), nil
	}

	maxTokens := in.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 1024
	}

	body, _ := json.Marshal(map[string]any{
		"model":      in.Model,
		"messages":   json.RawMessage(in.Messages),
		"max_tokens": maxTokens,
	})

	req, _ := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.router.client.Do(req)
	if err != nil {
		return tool.Failure(err), nil
	}
	defer resp.Body.Close() //nolint:errcheck

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		errLog.Printf("model_router: anthropic %d body=%s", resp.StatusCode, internallog.Redact(string(respBody)))
		return tool.FailureMsg(fmt.Sprintf("anthropic %d: %s", resp.StatusCode, truncateSnippet(respBody))), nil
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	_ = json.Unmarshal(respBody, &result)

	content := ""
	if len(result.Content) > 0 {
		content = result.Content[0].Text
	}
	return tool.SuccessWith(content, json.RawMessage(respBody)), nil
}

// truncateSnippet returns the first 256 bytes of body, appending a
// truncation marker if the body was longer. Used so error messages do not
// amplify leaked upstream content into the model context.
func truncateSnippet(body []byte) string {
	const limit = 256
	if len(body) <= limit {
		return string(body)
	}
	return string(body[:limit]) + "... [truncated]"
}
