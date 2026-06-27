# RELEASE-NOTES — Phase 2: RLM Deterministic Backend

## Summary

Phase 2 adds the RLM (Reasoning/Learning Module) engine to `pm` — specifically the fully offline, reproducible deterministic scoring backend. No network calls, no credentials, no new dependencies.

---

## New CLI command

```
pm rlm run --spec <file.json> --in <table> --out <table> --mode deterministic|fixture
```

Score a warehouse table using a weighted-feature spec. Results are materialized as a new warehouse table, sorted by score descending.

## New package: `internal/rlm`

- `Analyzer` strategy interface — cleanly separable backends.
- `DeterministicAnalyzer` — offline weighted-feature scoring; same input always produces same output.
- `FixtureAnalyzer` — canned rows for CI and offline demos; no InTable required.
- `ModelAnalyzer` stub — returns `ErrNotImplemented`; real implementation gated on Phase 4 approval.
- `ParseSpec([]byte)` — parse and validate JSON scoring spec.

## Flow step kind

`kind: rlm` is now a recognized step type in flow manifests. Wire an RLM step between `sync` and `action` steps to score records before any reverse-ETL push.

## Backwards compatibility

- No existing CLI commands changed.
- No existing package APIs changed.
- `internal/rlm` does not import `internal/flow` (clean dependency direction).
- Warehouse NDJSON format unchanged; OutTable adds `_rlm_*` prefix fields only.

## Human gates

- Model backend (Phase 4): requires explicit approval before any implementation. Current stub returns an error.

## Verification

`make verify` green. All new tests pass with `-count=1` (no caching).

---

## What is NOT in this release

- Claude API / model scoring (Phase 4).
- Email drafting (Phase 4).
- Scheduling (Phase 3).
- Prometheus/OTel metrics export.
