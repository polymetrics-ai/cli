# Planner Trace — Phase: schedule

Date: 2026-06-27
Planner: claude-sonnet-4-6

---

## Context gathered

1. Read `docs/prompts/gsd-flow-rlm-agent-mode-tdd-prompt.md` — Phase 3 section is the authoritative spec.
2. Read `internal/runtimecheck/runtimecheck.go` — confirmed `FromEnv()` reads `POLYMETRICS_TEMPORAL_ADDR`; `Doctor()` probes all three services; `CheckResult` has `Status` field for ok/degraded detection.
3. Read `internal/cli/cli.go` — confirmed stdlib dispatcher pattern; `case "schedule":` slot is absent (to be added); `withApp` pattern for commands needing the app layer; `parseGlobal` for `--root`/`--json`.
4. Confirmed module path: `polymetrics.ai`.
5. Confirmed no `internal/flow` or `internal/schedule` packages exist yet — greenfield.
6. Confirmed no `.planning/phases/schedule/` directory existed.

---

## Design decisions made

- Cron parser: stdlib-only, 5-field, supports `*`, numbers, ranges, lists, steps per field.
- `CronExpr.Next()`: needed for `pm schedule list` next-run display; pure function, no daemon.
- Backend selection: `SelectBackend` with injected `ProbeFunc` for testability.
- `TemporalBackend` uses already-present Temporal SDK from `go.mod` — no new deps.
- Unit files: `0600` permissions per threat model.
- CLI wiring: new `internal/cli/schedule.go` file + `case "schedule":` in existing switch.
- Golden tests: `testdata/` subdirectory under `internal/schedule/`, following existing repo patterns.

---

## Human gates identified

1. Temporal backend activation — user must set `POLYMETRICS_TEMPORAL_ADDR`. Flagged in PLAN.md Wave 3, ADR-003, API-CONTRACT.md.
2. No new `go.mod` entries — verified and required. Flagged in PLAN.md Wave 4.

---

## Artifacts created

- PRD.md
- SPEC.md
- PLAN.md (5 waves, 12 tasks: 6 TEST, 5 IMPL, 1 DOCS)
- TEST-PLAN.md (groups A–E, 28 test cases)
- THREAT-MODEL.md (6 threats, all mitigated)
- RUNBOOK.md
- API-CONTRACT.md
- DATA-MODEL.md
- OBSERVABILITY.md
- EVAL-PLAN.md
- RELEASE-NOTES.md
- POSTMORTEM-TEMPLATE.md
- ADR.md (5 decisions)

---

## Stop conditions not triggered

- No missing required context.
- Verification can run: `go test ./internal/schedule/... ./internal/cli/...` is executable once code is written.
- No human gate reached during planning itself (gates are flagged for execution time).
- No repeated failures.
