# PRD: Agentic ETL Platform

## Goal

Build the next production slice of the Polymetrics Go CLI monolith so agents can safely inspect connectors, run large ETL jobs, generate docs/skills, and prepare reverse ETL actions without receiving secrets or unsafe generic tools.

## Scope

- Structured CLI error contracts with stable exit codes and JSON errors.
- Shared validation and terminal sanitization for agent-supplied values.
- Connector manifests that drive inspection, documentation, and skill generation.
- Generated agent-facing skills and context docs.
- Bounded streaming ETL execution for large paginated reads.
- Checkpoint/progress metadata for ETL runs.
- Preservation of dependency-free execution as the default.

## Non-Goals

- No web UI.
- No untrusted plugin loading.
- No generic shell, generic HTTP write, or generic SQL write tools.
- No live external write tests without explicit approval.
- No new Go dependencies in this phase unless separately approved.

## Acceptance

- `make verify` passes.
- `poly help`, `poly docs generate`, and `poly skills generate` work.
- Connector inspection includes manifest data without reading secrets.
- Large source reads are loaded in bounded batches rather than one full in-memory buffer.
- CLI errors have stable categories and JSON shape when `--json` is supplied.
- Generated docs and skills do not contain credential values.
