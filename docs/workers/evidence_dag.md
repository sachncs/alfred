# evidence_dag

`mcp.evidence_dag` — build an evidence DAG from extracted claims.

## Purpose

Maintain a directed acyclic graph of claims and supporting evidence so the
agent can argue about provenance.

## Tools

- `evidence_update` — add or update an evidence edge
- `evidence_snapshot` — read the current evidence DAG
- `evidence_audit` — verify the DAG's consistency
