# model_router

`mcp.model_router` — proxy to upstream LLM APIs (OpenAI / Anthropic).

## Purpose

Route chat completion requests to the right upstream based on the model name
prefix (`gpt-*` → OpenAI, `claude-*` → Anthropic).

## Tools

- `model_proxy` — forward a chat completion request to the upstream

## Input

```json
{
  "model": "gpt-4o",
  "messages": [
    {"role": "system", "content": "You are a research assistant."},
    {"role": "user", "content": "Summarize arxiv:2401.00001"}
  ],
  "maxTokens": 1024
}
```

| Field | Type | Required | Notes |
| --- | --- | --- | --- |
| `model` | string | yes | upstream model name |
| `messages` | array | yes | OpenAI chat format |
| `maxTokens` | int | no | default 1024 |

## Output

```json
{
  "content": "...",
  "raw": { "...full upstream response..." }
}
```

## Configuration

| Env var | Required for |
| --- | --- |
| `OPENAI_API_KEY` | OpenAI models (`gpt-*`) |
| `ANTHROPIC_API_KEY` | Anthropic models (`claude-*`) |

## Notes

Failure messages are truncated to 256 bytes and run through the redacting
logger so leaked secrets cannot amplify into the model context.
