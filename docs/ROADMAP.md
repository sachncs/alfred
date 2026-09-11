# Alfred Roadmap

This document tracks Alfred's planned phases and current state.

## Phases shipped

- **Phase 1**: OOP hierarchy (`Tool`, `Agent`, `MCPServer`, `Capability`), `ChatAgent`, `EchoServer`, binary boot
- **Phase 2**: HTTP/SSE runtime, turn loop with tool dispatch, compaction, token economy, history hygiene, steering queue, sub-agent delegation, usage tracking, tool budgets, prompt cache, memory store, skills loader, all built-in tools (`Read`/`Write`/`Edit`/`ApplyPatch`/`Bash`/`Grep`/`Find`/`Ls`), code review, fork/resume, goals/todos, all HTTP routes
- **Phase 3**: SQLite persistence (pure Go via `modernc.org/sqlite`), schema migrations, hybrid thread store (SQLite index + JSONL body), retention pruning, JSONL migration, legacy Kun config migration
- **Phase 4**: Web UI (htmx + Go html/template) under `/` and `/design`
- **Phase 5**: Capability broker + initial workers (`search`, `model_router`, `plan_gateway`, `write_assist`, `inspector`)
- **Phase 6**: Settings persistence (`AppSettingsV1`) via SQLite settings table
- **Phase 7**: Research / content / workspace workers (paper radar, multi-agent, evidence DAG, image generation, scientific plotting, PPT, schedule, workflow, remote executor, workspace previews)

## Worker catalog

The 25+ MCP workers registered in `cmd/alfred/main.go` cover:

- **Research** (Phase 7a): `paper_radar`, `multi_agent`, `sci_modality`, `evidence_dag`, `project_dag`, `feedback_gateway`, `bgc_discovery`
- **Content / media** (Phase 7b): `image_generation`, `scientific_plotting`, `visual_document`, `ppt_master`
- **Workspace previews** (Phase 7c): `workspace_intel`, `workspace_bioimaging`, `workspace_deck`, `workspace_molecular`, `workspace_omics`, `workspace_sequence`, `workspace_spectra`, `workspace_tabular`
- **System** (Phase 7d): `workflow`, `schedule`, `remote_executor`, `gui_owl`
- **Phase 5**: `search`, `model_router`, `plan_gateway`, `write_assist`, `inspector`

## Future phases

- **Phase 8**: UX features — plan mode, write mode, anchored comments, multi-thread views
- **Phase 9**: Remote channel runtime, cutover from `SciForge`

## See also

- `README.md` for the binary overview, run instructions, and recipes
- `CHANGELOG.md` for per-phase release notes
- `cmd/alfred/main.go` for the worker registration list
