# paper_radar

`mcp.paper_radar` — track recent papers across arxiv / biorxiv / PubMed.

## Purpose

Build reading profiles, search the recent-paper firehose, and produce digests.

## Tools

- `paper_search` — search recent papers
- `paper_digest` — summarize a paper
- `paper_profile_create` — create a per-paper reading profile

## Example

```bash
curl -X POST http://127.0.0.1:8899/v1/capabilities/mcp.paper_radar \
  -H 'Content-Type: application/json' \
  -d '{"tool":"paper_search","input":{"query":"diffusion models","since":"2026-01-01"}}'
```
