# PLAN — Phase 3: Scheduling

Date: 2026-06-27
Wave execution: red (tests) → green (impl) → refactor, wave by wave.

---

## Wave 0 — Cron parser (pure, no OS deps)

### T-0.1 [TEST] Cron validation table-driven tests
File: `internal/schedule/cron_test.go`
Type: behavior
Tests: valid exprs pass; out-of-range fields, wrong field count, bad chars fail; `Next()` returns correct next instant for fixed clock.
Must be RED before T-0.2.

### T-0.2 [IMPL] `ParseCron` + `CronExpr.Next()`
File: `internal/schedule/cron.go`
Type: behavior
Implements: 5-field stdlib-only cron parser; `Next(time.Time) time.Time`.
Gate: T-0.1 turns green.

---

## Wave 1 — Manifest CRUD

### T-1.1 [TEST] Manifest load/save/list round-trip
File: `internal/schedule/schedule_test.go`
Type: behavior
Tests: save manifest to temp dir, load it back, field equality; list returns all; name validation rejects invalid slugs; duplicate name errors.
Must be RED before T-1.2.

### T-1.2 [IMPL] `Manifest`, `Save`, `Load`, `List`, `Delete`
File: `internal/schedule/schedule.go`
Type: behavior
Uses: `os.MkdirAll`, `os.ReadFile`, `encoding/json`, `path/filepath`.
Gate: T-1.1 turns green.

---

## Wave 2 — Unit-file rendering (pure functions, golden tests)

### T-2.1 [TEST] Launchd plist golden test
File: `internal/schedule/launchd_test.go`
Type: behavior
Tests: `renderPlist(m Manifest, pmBin string) (string, error)` against fixture in `testdata/launchd_nightly.golden`.
Must be RED before T-2.2.

### T-2.2 [IMPL] `LaunchdBackend` — plist rendering + `startCalendarInterval` conversion
File: `internal/schedule/launchd.go`
Type: behavior
Gate: T-2.1 turns green.

### T-2.3 [TEST] Systemd unit golden tests
File: `internal/schedule/systemd_test.go`
Type: behavior
Tests: `renderService` and `renderTimer` against fixtures; `cronToOnCalendar` table-driven conversion.
Must be RED before T-2.4.

### T-2.4 [IMPL] `SystemdBackend` — service/timer rendering + `cronToOnCalendar`
File: `internal/schedule/systemd.go`
Type: behavior
Gate: T-2.3 turns green.

### T-2.5 [TEST] Crontab line golden test
File: `internal/schedule/crontab_test.go`
Type: behavior
Tests: `renderCrontabLine` golden; `removeCrontabLine` strips exactly the sentinel line; idempotent remove of absent line is no-op.
Must be RED before T-2.6.

### T-2.6 [IMPL] `CrontabBackend` — line render/append/remove
File: `internal/schedule/crontab.go`
Type: behavior
Gate: T-2.5 turns green.

---

## Wave 3 — Backend selection

### T-3.1 [TEST] `SelectBackend` env-driven selection
File: `internal/schedule/backend_test.go`
Type: behavior
Tests: `POLYMETRICS_TEMPORAL_ADDR` set + Temporal reachable → TemporalBackend; `--crontab` flag forces CrontabBackend; GOOS mock for launchd/systemd; no-Temporal + darwin → LaunchdBackend; no-Temporal + linux → SystemdBackend.
Note: Temporal reachability can be faked via an interface seam (avoids real network call in tests).
Must be RED before T-3.2.

### T-3.2 [IMPL] `SelectBackend`, `Backend` interface, stubs for `TemporalBackend`
File: `internal/schedule/backend.go`, `internal/schedule/temporal.go`
Type: behavior
Reuses: `internal/runtimecheck.FromEnv()` to check Temporal addr.

HUMAN GATE: `TemporalBackend.Install` touches network (Temporal server). This path is opt-in and only exercised when `POLYMETRICS_TEMPORAL_ADDR` is set. No actual Temporal network calls in CI tests — the reachability check is mocked via an interface.
Gate: T-3.1 turns green.

---

## Wave 4 — CLI wiring

### T-4.1 [TEST] CLI integration tests for `pm schedule`
File: `internal/cli/runtime_record_test.go` is an existing file; new tests go in `internal/cli/schedule_test.go`
Type: behavior
Tests: `create` writes manifest file; `list` returns JSON array; `install` on dry-run (CrontabBackend in test env) appends line; `remove` cleans up. Uses `cli.Run(args, stdout, stderr)` directly (existing pattern).
Must be RED before T-4.2.

### T-4.2 [IMPL] `internal/cli/schedule.go` — `runSchedule` dispatcher
File: `internal/cli/schedule.go`
Type: behavior
Wire: add `case "schedule":` to `internal/cli/cli.go` switch.
Gate: T-4.1 turns green.

---

## Wave 5 — Docs-only

### D-5.1 [DOCS] Manual entry for `pm schedule`
File: `internal/cli/manual_schedule.go` (or wherever manual strings live)
Type: docs-only
No test needed — follows existing `writeManual` pattern.

---

## Human gates

| Gate | Condition | Action required |
|------|-----------|-----------------|
| Temporal backend activation | `POLYMETRICS_TEMPORAL_ADDR` set in user env | User must explicitly set this env var; install path will diverge to Temporal. Warn in `pm schedule install` output when Temporal is selected. |
| No new go.mod entries | All code uses only stdlib + already-present `go.mod` deps | Verify with `go mod tidy` after wave 4. If any new import appears, stop and get approval. |

---

## Verification gate (after all waves)

```
export GOTOOLCHAIN=auto
gofmt -w internal/schedule internal/cli
go vet ./...
go test ./internal/schedule/... ./internal/cli/...
go build ./cmd/pm
make verify
```
All must be green before the phase is considered done.
