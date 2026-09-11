# Workers

Alfred ships with 25+ MCP workers. Each worker is registered as a `Capability`
with the broker in `cmd/alfred/main.go` and exposed via `GET /v1/capabilities`.

## Catalog

| Worker | Purpose |
| --- | --- |
| [`search`](./workers/search.md) | Web / paper search across DuckDuckGo, arxiv, biorxiv, Europe PMC, Semantic Scholar, Tavily |
| [`model_router`](./workers/model_router.md) | Proxy to OpenAI / Anthropic upstream APIs |
| [`plan_gateway`](./workers/plan_gateway.md) | Plan-mode entry point for multi-step tasks |
| [`write_assist`](./workers/write_assist.md) | Write-mode helpers (diffs, summaries, doc skeletons) |
| [`inspector`](./workers/runtime_inspector.md) | Runtime introspection (events, replay buffer, tools) |
| [`paper_radar`](./workers/paper_radar.md) | Track recent papers across arxiv / biorxiv / PubMed |
| [`multi_agent`](./workers/multi_agent.md) | Delegate tasks to sub-agents in parallel |
| [`sci_modality`](./workers/sci_modality.md) | Scientific modality handlers |
| [`evidence_dag`](./workers/evidence_dag.md) | Build an evidence DAG from extracted claims |
| [`project_dag`](./workers/project_dag.md) | Project planning DAG |
| [`feedback_gateway`](./workers/feedback_gateway.md) | Idempotent feedback → GitHub issue submission |
| [`bgc_discovery`](./workers/bgc_discovery.md) | Biosynthetic gene cluster discovery |
| [`image_generation`](./workers/image_generation.md) | Image generation worker |
| [`scientific_plotting`](./workers/scientific_plotting.md) | Plot generation (matplotlib / plotly) |
| [`visual_document`](./workers/visual_document.md) | Visual document composer |
| [`ppt_master`](./workers/ppt_master.md) | PowerPoint deck builder |
| [`workspace_intel`](./workers/workspace_intel.md) | Workspace intel preview |
| [`workspace_bioimaging`](./workers/workspace_bioimaging.md) | Bioimaging workspace preview |
| [`workspace_deck`](./workers/workspace_deck.md) | Slide-deck workspace preview |
| [`workspace_molecular`](./workers/workspace_molecular.md) | Molecular workspace preview |
| [`workspace_omics`](./workers/workspace_omics.md) | Omics workspace preview |
| [`workspace_sequence`](./workers/workspace_sequence.md) | Sequence workspace preview |
| [`workspace_spectra`](./workers/workspace_spectra.md) | Spectra workspace preview |
| [`workspace_tabular`](./workers/workspace_tabular.md) | Tabular workspace preview |
| [`workflow`](./workers/workflow.md) | Workflow engine |
| [`schedule`](./workers/schedule.md) | Scheduled task worker |
| [`remote_executor`](./workers/remote_executor.md) | Remote executor |
| [`gui_owl`](./workers/gui_owl.md) | GUI automation |

## Discovery

Hit `GET /v1/capabilities` against a running alfred to see the live list of
worker IDs and their tool counts.

## Adding a new worker

1. Create a new package under `internal/mcp/worker/<name>/` with a
   `NewXxxServer()` constructor.
2. Register it in `cmd/alfred/main.go` alongside the existing workers.
3. Add a documentation page under `docs/workers/<name>.md` and link it from
   `docs/workers/README.md`.
