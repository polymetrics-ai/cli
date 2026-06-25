# OBSERVABILITY: Agentic ETL Platform

## Required Signals

- ETL run status and counts.
- Batch count and checkpoint metadata.
- Structured error category/code.
- Runtime doctor status when dependency-backed mode is requested.

## Current Output Targets

- Human summaries on stdout for non-JSON commands.
- Warnings and errors on stderr.
- Machine JSON on stdout for `--json`.

## Deferred

- Metrics and tracing exporters are deferred until a runtime-backed service boundary is introduced.
