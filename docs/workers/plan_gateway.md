# plan_gateway

`mcp.plan_gateway` — generate a structured plan for a coding task.

## Purpose

Wrap OpenCode's plan generation so the chat agent can produce a multi-step,
dependency-aware plan before executing.

## Tools

- `generate_plan` — build a plan with steps, dependencies, and acceptance criteria

## Input

```json
{
  "task": "Refactor the SQLite event insert path to use a write-ahead marker",
  "context": "Current code lives in internal/store/hybrid_thread.go"
}
```

## Output

A structured plan object with `steps`, `dependencies`, and `acceptanceCriteria`.

## Example

```bash
curl -X POST http://127.0.0.1:8899/v1/capabilities/mcp.plan_gateway \
  -H 'Content-Type: application/json' \
  -d '{"tool":"generate_plan","input":{"task":"Add docs site"}}'
```
