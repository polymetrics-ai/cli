# TEST-PLAN — Phase 3: Scheduling

Date: 2026-06-27

---

## Principle

Every behavior task in PLAN.md has a paired red-first test task. Tests must be written and observed failing (red) before the implementation is written. Evidence recorded in `TDD-LEDGER.md`.

---

## Test groups

### Group A — Cron parsing (`internal/schedule/cron_test.go`)

| ID | Description | Input | Expected |
|----|-------------|-------|----------|
| A-1 | Valid expr accepted | `"0 2 * * *"` | no error |
| A-2 | Valid step | `"*/15 * * * *"` | no error |
| A-3 | Valid range | `"0 9-17 * * 1-5"` | no error |
| A-4 | Too few fields | `"0 2 * *"` | error |
| A-5 | Too many fields | `"0 2 * * * *"` | error |
| A-6 | Minute out of range | `"60 * * * *"` | error |
| A-7 | Hour out of range | `"0 25 * * *"` | error |
| A-8 | Month out of range | `"0 0 1 13 *"` | error |
| A-9 | DOW out of range | `"0 0 * * 8"` | error |
| A-10 | `Next()` simple | `"0 2 * * *"`, clock at 2026-06-27 01:00 UTC | 2026-06-27 02:00 UTC |
| A-11 | `Next()` crosses midnight | `"0 2 * * *"`, clock at 2026-06-27 03:00 UTC | 2026-06-28 02:00 UTC |
| A-12 | `Next()` DOW filter | `"0 9 * * 1"`, clock at Monday 2026-06-29 10:00 UTC | 2026-07-06 09:00 UTC |

### Group B — Manifest CRUD (`internal/schedule/schedule_test.go`)

| ID | Description | Expected |
|----|-------------|----------|
| B-1 | Save + Load round-trip | all fields equal |
| B-2 | List returns all saved manifests | count matches |
| B-3 | Duplicate name rejected | error contains "already exists" |
| B-4 | Invalid name slug rejected | error on create |
| B-5 | Delete removes file | Load after Delete → error |
| B-6 | Load non-existent → error | error is non-nil |

### Group C — Unit file rendering (golden tests)

#### C-1 Launchd plist (`internal/schedule/launchd_test.go`)
- `renderPlist(manifest, "/usr/local/bin/pm")` output matches `testdata/launchd_nightly.golden`.
- DOM/month/DOW wildcard maps correctly.
- Non-wildcard cron hour/minute appear as `<integer>` in plist.

#### C-2 Systemd timer (`internal/schedule/systemd_test.go`)
- `renderService` output matches `testdata/systemd_nightly.service.golden`.
- `renderTimer` output matches `testdata/systemd_nightly.timer.golden`.
- `cronToOnCalendar` table-driven: `"0 2 * * *"` → `"*-*-* 02:00:00"`; `"*/15 * * * *"` → `"*-*-* *:0/15:00"`.

#### C-3 Crontab line (`internal/schedule/crontab_test.go`)
- `renderCrontabLine` golden matches expected string including sentinel comment.
- `removeCrontabLine` on content containing the sentinel removes exactly one line.
- `removeCrontabLine` on content without sentinel returns content unchanged (no error).
- Idempotent: remove twice leaves same result as remove once.

### Group D — Backend selection (`internal/schedule/backend_test.go`)

| ID | Description | Env / inputs | Expected backend kind |
|----|-------------|--------------|-----------------------|
| D-1 | No Temporal, darwin | GOOS=darwin, no TEMPORAL_ADDR | launchd |
| D-2 | No Temporal, linux | GOOS=linux | systemd OR crontab (based on systemd detection) |
| D-3 | Temporal addr set + reachable (mocked) | TEMPORAL_ADDR set, probe returns ok | temporal |
| D-4 | --crontab flag forces crontab | any OS | crontab |
| D-5 | Temporal addr set but unreachable | TEMPORAL_ADDR set, probe returns error | falls back to OS backend |

Temporal reachability injected via `ProbeFunc` parameter on `SelectBackend` — avoids real network calls.

### Group E — CLI integration (`internal/cli/schedule_test.go`)

| ID | Invocation | Expected behavior |
|----|------------|-------------------|
| E-1 | `pm schedule create --name x --cron "0 2 * * *" --flow y` | manifest file created in temp root; exit 0 |
| E-2 | `pm schedule list --json` | JSON array contains the created schedule |
| E-3 | `pm schedule install x --crontab` | crontab line written (uses temp crontab file via env override) |
| E-4 | `pm schedule remove x` | manifest deleted; exit 0 |
| E-5 | `pm schedule create` (missing flags) | exit 1, error in stderr |
| E-6 | `pm schedule install unknown` | exit 1, "not found" error |
| E-7 | `pm schedule create --name INVALID` | exit 1, validation error |

All CLI tests use `cli.Run(args, &stdout, &stderr)` against a temp `--root` dir, consistent with existing test patterns (see `internal/cli/runtime_record_test.go`).

---

## TDD Ledger requirement

Before writing any implementation file, run the corresponding test and record:
- file path
- `go test` output showing FAIL
- timestamp

Record in `.planning/phases/schedule/TDD-LEDGER.md`.

---

## Coverage target

All behavior tasks in PLAN.md must be exercised. No skipped tests in CI except those requiring absent binaries (`launchctl`, `systemctl`), guarded by `exec.LookPath` + `t.Skip`.
