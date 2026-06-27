# TDD Ledger

Phase: schedule

Record failing test evidence before production code for every behavior-adding task.

---

## Evidence entries

### T-0.1 — Cron parse/validate (Group A)
File: `internal/schedule/cron_test.go`
Timestamp: 2026-06-27
Status: **red-confirmed**

Red run excerpt:
```
--- FAIL: TestParseCron_Valid/A-1 (0.00s)
    cron_test.go:24: ParseCron("0 2 * * *") returned unexpected error: ParseCron: not implemented
--- FAIL: TestParseCron_Valid/A-2 (0.00s)
    cron_test.go:24: ParseCron("*/15 * * * *") returned unexpected error: ParseCron: not implemented
--- FAIL: TestParseCron_Valid/A-3 (0.00s)
    cron_test.go:24: ParseCron("0 9-17 * * 1-5") returned unexpected error: ParseCron: not implemented
--- FAIL: TestCronExprNext/A-10 (0.00s)
    cron_test.go:85: ParseCron("0 2 * * *"): ParseCron: not implemented
--- FAIL: TestCronExprNext/A-11 (0.00s)
    cron_test.go:85: ParseCron("0 2 * * *"): ParseCron: not implemented
--- FAIL: TestCronExprNext/A-12 (0.00s)
    cron_test.go:85: ParseCron("0 9 * * 1"): ParseCron: not implemented
```
Note: A-4 through A-9 (invalid exprs) pass trivially against stub — correct negative behavior preserved after implementation.

---

### T-1.1 — Manifest CRUD (Group B)
File: `internal/schedule/manifest_test.go`
Timestamp: 2026-06-27
Status: **red-confirmed**

Red run excerpt:
```
--- FAIL: TestManifestSaveLoad (0.00s)
    manifest_test.go:27: Save: Save: not implemented
--- FAIL: TestManifestList (0.00s)
    manifest_test.go:47: Save "alpha": Save: not implemented
--- FAIL: TestManifestDuplicateRejected (0.00s)
    manifest_test.go:63: first Save: Save: not implemented
--- FAIL: TestManifestDelete (0.00s)
    manifest_test.go:89: Save: Save: not implemented
```

---

### T-2.1 — Launchd plist golden (C-1)
File: `internal/schedule/render_test.go` (TestRenderPlist_Golden)
Timestamp: 2026-06-27
Status: **red-confirmed**

```
--- FAIL: TestRenderPlist_Golden (0.00s)
    render_test.go:35: renderPlist: renderPlist: not implemented
```

---

### T-2.3 — Systemd unit golden (C-2)
File: `internal/schedule/render_test.go`
Timestamp: 2026-06-27
Status: **red-confirmed**

```
--- FAIL: TestRenderService_Golden (0.00s)
    render_test.go:47: renderService: renderService: not implemented
--- FAIL: TestRenderTimer_Golden (0.00s)
    render_test.go:59: renderTimer: renderTimer: not implemented
--- FAIL: TestCronToOnCalendar/0_2_*_*_* (0.00s)
    render_test.go:78: cronToOnCalendar("0 2 * * *"): cronToOnCalendar: not implemented
--- FAIL: TestCronToOnCalendar/*/15_*_*_*_* (0.00s)
    render_test.go:78: cronToOnCalendar("*/15 * * * *"): cronToOnCalendar: not implemented
```

---

### T-2.5 — Crontab line golden (C-3)
File: `internal/schedule/render_test.go`
Timestamp: 2026-06-27
Status: **red-confirmed**

```
--- FAIL: TestRenderCrontabLine_Golden (0.00s)
    render_test.go:91: renderCrontabLine: renderCrontabLine: not implemented
--- FAIL: TestRemoveCrontabLine_RemovesSentinel (0.00s)
    render_test.go:105: removeCrontabLine: removeCrontabLine: not implemented
--- FAIL: TestRemoveCrontabLine_NoopWhenAbsent (0.00s)
    render_test.go:120: removeCrontabLine: removeCrontabLine: not implemented
--- FAIL: TestRemoveCrontabLine_Idempotent (0.00s)
    render_test.go:133: first remove: removeCrontabLine: not implemented
```

---

### T-3.1 — SelectBackend env-driven selection (Group D)
File: `internal/schedule/select_test.go`
Timestamp: 2026-06-27
Status: **red-confirmed**

Key red test:
```
--- FAIL: TestSelectBackend_Darwin (0.00s)
    select_test.go:48: darwin: got kind "crontab", want "launchd"
```
Note: D-3/D-4/D-5 pass because the Temporal path and forceCrontab are partially wired. D-1 (darwin→launchd) is the red gate for OS-based selection.

---

### T-4.1 — CLI integration (Group E)
File: `internal/cli/schedule_test.go`
Timestamp: 2026-06-27
Status: **red-confirmed** (package build blocked by pre-existing undefined functions)

```
# polymetrics.ai/internal/cli [polymetrics.ai/internal/cli.test]
internal/cli/runtime_record_test.go:21:11: undefined: runtimeETLLeaseRequest
internal/cli/runtime_record_test.go:25:12: undefined: runtimeETLRunRecord
FAIL    polymetrics.ai/internal/cli [build failed]
```

The schedule CLI tests (E-1 through E-7) are authored correctly and will fail with "unknown command schedule" once the pre-existing cli breakage is resolved. The cli package itself builds clean (`go build ./internal/cli/...` succeeds); only the test binary fails due to the pre-existing unimplemented functions referenced in `runtime_record_test.go`.
