# write_assist

`mcp.write_assist` — writing-mode helpers (markdown conversion, inline completion).

## Purpose

Provide lightweight utilities the chat agent can call while drafting content.

## Tools

- `markdown_to_latex` — convert a Markdown string to LaTeX
- `inline_completion` — produce an inline completion given a context window

## Examples

```bash
# Convert markdown to LaTeX
curl -X POST http://127.0.0.1:8899/v1/capabilities/mcp.write_assist \
  -H 'Content-Type: application/json' \
  -d '{"tool":"markdown_to_latex","input":{"markdown":"# Hello"}}'

# Inline completion
curl -X POST http://127.0.0.1:8899/v1/capabilities/mcp.write_assist \
  -H 'Content-Type: application/json' \
  -d '{"tool":"inline_completion","input":{"prefix":"def add(a,b):"}}'
```
