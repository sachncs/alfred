# runtime_inspector

`mcp.inspector` — runtime introspection.

## Purpose

Expose read-only diagnostics about the running alfred process (env, git status,
event counts, replay buffer sizes).

## Tools

- `runtime_env` — emit selected `ALFRED_*` env vars
- `git_status` — return the workspace git status summary

## Examples

```bash
curl -X POST http://127.0.0.1:8899/v1/capabilities/mcp.inspector \
  -H 'Content-Type: application/json' \
  -d '{"tool":"runtime_env","input":{}}'
```
