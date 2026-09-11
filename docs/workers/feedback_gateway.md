# feedback_gateway

`mcp.feedback_gateway` — idempotent feedback submission.

## Purpose

Submit user feedback as a structured GitHub issue with deterministic
idempotency keys based on `(title, body)`.

## Tools

- `feedback_submit` — submit feedback (returns issue ID, deduped by key)
- `feedback_status` — query the status of a previously submitted feedback

## Input

```json
{
  "title": "Search providers truncate HTTP body",
  "body": "...",
  "labels": ["bug"]
}
```

## Output

```json
{
  "issueId": "fb-001",
  "status": "created",
  "payload": { "title": "...", "body": "...", "labels": ["bug"] }
}
```
