# search

`mcp.search` — federated web / paper search.

## Purpose

Query multiple upstream providers (DuckDuckGo, arxiv, biorxiv, Europe PMC,
Semantic Scholar, Tavily) and return a unified `Result` list.

## Tools

- `research_search` — search a query string across configured providers

## Input

```json
{
  "query": "transformer scaling laws",
  "maxResults": 10,
  "sources": "arxiv,biorxiv"
}
```

| Field | Type | Required | Notes |
| --- | --- | --- | --- |
| `query` | string | yes | free-form query |
| `maxResults` | int | no | per-provider cap (default 10) |
| `sources` | string | no | comma-separated list of provider names |

## Output

A list of `Result` objects:

```json
[
  {
    "title": "Scaling Laws for Neural Language Models",
    "url": "https://arxiv.org/abs/2001.08361",
    "snippet": "We study empirical scaling laws...",
    "source": "arxiv",
    "date": "2020-01-23",
    "authors": "J. Kaplan, et al."
  }
]
```

## Example

```bash
curl -X POST http://127.0.0.1:8899/v1/capabilities/mcp.search \
  -H 'Content-Type: application/json' \
  -d '{"tool":"research_search","input":{"query":"scaling laws","maxResults":5}}'
```

## Configuration

Provider credentials are pulled from environment variables (e.g.
`TAVILY_API_KEY`); absent keys silently disable that provider.
