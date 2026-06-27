# PRD — Flow Engine (Phase 0)

## Problem

`pm` can run individual ETL syncs and reverse-ETL actions, but has no way to compose them into a
reproducible, auditable multi-step pipeline. Users must chain commands manually, losing
checkpointing, overlap prevention, and visibility across steps.

## Goal

Deliver a declarative flow engine that lets users define YAML flow manifests, validate them, execute
them step-by-step with topological ordering, checkpoint progress to the ledger, and prevent
concurrent execution via leases — all without any new third-party dependencies.

## Non-goals (this phase)

- `action` step kind (reverse-ETL sends) — Phase 1
- `rlm` step kind — Phase 2
- Scheduling — Phase 3
- Model/LLM backends — Phase 4
- Agent Mode — Phase 5

## Users

- Data engineers running local ETL pipelines from the CLI
- Agents (LLM-driven) issuing `pm flow plan|preview|run|status|list --json`

## Requirements

### Functional

FR-1  Parse a YAML flow manifest into a normalized, validated in-memory structure.
FR-2  Build a DAG from step `in`/`out` table declarations and detect cycles.
FR-3  Execute steps in topological order; steps with no dependency may run sequentially
      (parallelism is opt-in, not required in Phase 0).
FR-4  Supported step kinds in this phase: `sync` (maps to `app.ETLRun`) and `query`
      (maps to `app.QuerySQL` materialization).
FR-5  Write a ledger entry (start + finish) per step and per flow run using `internal/ledger`.
FR-6  Acquire a file-backed lease via `internal/state.FileLock` before executing; refuse to
      run if the lock is already held (lease contention).
FR-7  `plan` subcommand: validate manifest + build DAG + run all read-only steps (sync, query)
      for real; return structured JSON output listing executed steps and their results.
FR-8  `preview` subcommand: validate + build DAG + dry-run every step (no actual connector
      calls); return the execution plan with record-count estimates where available.
FR-9  `run` subcommand: full execution including action steps (in Phase 0 there are none, so
      equivalent to `plan`).
FR-10 `status` subcommand: show last-run state for a named flow (reads ledger).
FR-11 `list` subcommand: enumerate flow manifests found under `.polymetrics/flows/`.
FR-12 Checkpoint/resume: on re-run, skip steps already marked `success` in the checkpoint
      store; `--force` flag clears checkpoints.
FR-13 All output honours `--json` flag; human output goes to stdout, errors to stderr.

### Non-functional

NF-1  Zero new go.mod dependencies. Go stdlib only.
NF-2  Every code path exercised by a table-driven test before implementation (strict TDD).
NF-3  `make verify` (fmt + vet + test + build + docs-check) passes at phase boundary.
NF-4  Secrets are never printed or logged.
NF-5  All long operations accept a `context.Context` and honour cancellation.

## Success criteria

- `pm flow plan` against a two-step sync→query manifest executes both steps, writes ledger
  entries, and returns `{"status":"ok"}` with zero non-zero exit codes.
- Cycle detection returns a non-zero exit code and a structured error.
- Re-run with checkpoint skips already-successful steps.
- `make verify` green.
