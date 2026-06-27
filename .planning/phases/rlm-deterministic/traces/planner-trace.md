# Planner Trace — Phase: rlm-deterministic

Date: 2026-06-27
Planner model: claude-sonnet-4-6

---

## Inputs read

- `docs/prompts/gsd-flow-rlm-agent-mode-tdd-prompt.md` — canonical prompt, Phase 2 section
- `internal/app/types.go` — existing `Run`, `ReverseRun`, `StreamState` types; established JSON field naming conventions
- `internal/app/local_warehouse.go` — `localRawRecord` struct, `localWarehouseTablePath`, atomic write pattern (temp+rename)
- `internal/ledger/ledger.go` — `RunRecord`, `JSONLedger`, `PostgresLedger`; existing ledger interface pattern
- `internal/state/store.go` — `JSONStore[T]`, `Locker` interface pattern
- `internal/cli/cli.go` — CLI dispatcher switch, `withApp` pattern, `envelope` type
- `.planning/phases/flow-engine/` — phase artifact naming conventions

---

## Key decisions made

1. **Spec format: JSON not YAML** — stdlib only constraint eliminates YAML. Documented in ADR-001.
2. **Scoring in Go not SQL** — DuckDB is optional/build-tagged; pure Go is simpler and fully testable. ADR-003.
3. **Flat OutTable NDJSON** — not wrapped in `localRawRecord` envelope. Makes output directly queryable. ADR-002.
4. **LedgerAppender interface** — thin interface in `internal/rlm` avoids direct coupling to `internal/ledger`. ADR-004.
5. **ModelAnalyzer: unconditional stub** — no feature flag, no env var. Only replaced in Phase 4 after human gate. ADR-005.
6. **Map iteration non-determinism** — identified as a key risk. POSTMORTEM-TEMPLATE documents it. Test T2.1 `TestDeterminismSameInputSameOutput` catches it.
7. **Path traversal mitigation** — `validateTableName` function must be called before any `filepath.Join`. THREAT-MODEL T1 and POSTMORTEM-TEMPLATE item 2.

---

## Wave ordering rationale

- Wave 0 (scaffolding) is the only non-TDD wave — types and interfaces only, no behavior.
- Wave 1 (spec parsing) before Wave 2 (scoring) because the scorer depends on `*Spec`.
- Wave 2 (scoring engine) before Wave 3 (materialization) because materialization wraps scoring.
- Wave 4 (fixture) after Wave 3 because fixture reuses `scoreRecords` from Wave 2/3 — can write its tests earlier (after T2.2).
- Wave 5 (CLI) last behavior wave — integration layer over all prior waves.
- Wave 6 (flow step) is docs-only in this phase (flow-engine package may not exist yet).
- Wave 7 (e2e) confirms the fixture path works end-to-end; test can be written after Wave 4.

---

## Human gates identified

1. **Model backend (Phase 4)** — documented in PLAN.md, THREAT-MODEL.md, ADR-005, RELEASE-NOTES.md. Must not be bypassed.

---

## Files to be created by the implementation agent

```
internal/rlm/rlm.go
internal/rlm/spec.go
internal/rlm/deterministic.go
internal/rlm/fixture.go
internal/rlm/model.go
internal/rlm/rlm_test.go
internal/rlm/spec_test.go
internal/rlm/deterministic_test.go
internal/rlm/fixture_test.go
internal/rlm/model_stub_test.go
internal/rlm/e2e_test.go
internal/rlm/testdata/likely_customers.json
internal/cli/rlm_cli.go
internal/cli/rlm_cli_test.go
```

## Files to be modified by the implementation agent

```
internal/cli/cli.go   — add `case "rlm":` to switch cmd
```

---

## Stop conditions not triggered

- No missing required context.
- Verification can run (`go test ./internal/rlm/...`).
- No human gate reached yet (model stub is compile-only).
- No repeated failures.
