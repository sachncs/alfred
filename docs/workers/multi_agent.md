# multi_agent

`mcp.multi_agent` — delegate tasks to sub-agents in parallel.

## Purpose

Run a list of tasks as concurrent child agents, returning a status snapshot.

## Tools

- `delegate_task` — kick off a task and return its task ID
- `task_status` — query the status / result of a previously delegated task

## Input

```json
{
  "task": "Summarize arxiv:2401.00001",
  "agentType": "general"
}
```

## Output

```json
{
  "taskId": "delegate-001",
  "status": "queued"
}
```

## Example

```bash
curl -X POST http://127.0.0.1:8899/v1/capabilities/mcp.multi_agent \
  -H 'Content-Type: application/json' \
  -d '{"tool":"delegate_task","input":{"task":"Summarize arxiv:2401.00001"}}'
```
