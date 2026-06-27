# PRD — Phase 2: RLM Deterministic Backend

Date: 2026-06-27
Phase: rlm-deterministic
Status: planning

---

## Problem

The `pm` ETL pipeline can collect and materialize contact/company data into warehouse tables (NDJSON files on disk, or Postgres). There is currently no way to score or rank those records — e.g., to identify "likely customers" — without exporting data to an external scoring service or writing ad-hoc SQL by hand.

Users need a reproducible, offline, credential-free scoring stage that:
1. Reads a warehouse table produced by a prior `sync` or `query` step.
2. Applies a user-authored feature/weight spec to produce a score per record.
3. Materializes results into a new warehouse table (e.g., `lead_scores`).
4. Can be wired as a `rlm` step kind inside a flow manifest.

## Goals

1. Ship `internal/rlm` package with an `Analyzer` strategy interface.
2. Implement `deterministic` backend: SQL/weighted-feature scoring, fully offline, reproducible — same input always produces same output.
3. Implement `fixture` backend: canned in-memory table, credential-free, for CI/tests/demos.
4. Stub `model` backend interface (method signatures only, returns `ErrNotImplemented`) — implementation is gated on human approval (Phase 4).
5. Expose `pm rlm run --spec <file> --in <table> --out <table> --mode deterministic|fixture` CLI verb.
6. Wire `rlm` as a flow step kind in `internal/flow` (coordinate via shared types; keep packages independent).
7. End-to-end `likely-customers` example flow runs fully offline using fixture/deterministic mode.

## Non-Goals

- Model/Claude API calls (Phase 4).
- Email drafting (Phase 4).
- Scheduling (Phase 3).
- Any network calls in this phase.
- New third-party Go module dependencies.

## Success Criteria

- `go test ./internal/rlm/...` passes with table-driven tests covering: determinism, feature/weight spec parsing, materialization schema, fixture parity.
- `pm rlm run --spec testdata/likely_customers.yaml --in contacts --out lead_scores --mode deterministic` exits 0 on a local warehouse dir.
- `make verify` (gofmt + vet + test + build) is green.
- Model stub compiles but returns `ErrNotImplemented` and is never invoked in this phase.

## Human Gate

The `model` backend real implementation (Phase 4) requires explicit human approval before any network/credential-touching code is written. This gate is documented in PLAN.md and must not be bypassed.
