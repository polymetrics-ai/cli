# Agent Trace: tester

## Rendered Prompt Or Prompt Reference

GSD TDD tester for Phase: schedule — write failing (RED) tests before implementation.

## Actions taken

1. Read PLAN.md, TEST-PLAN.md, SPEC.md to understand 5-wave task breakdown and 28 test cases (Groups A-E).
2. Read `internal/cli/runtime_record_test.go` and `internal/cli/cli.go` for existing `cli.Run` pattern.
3. Checked go.mod: confirmed `go.temporal.io/sdk` and `github.com/stretchr/testify` already present.
4. Created `internal/schedule/` directory with stub source files (schedule.go, cron.go, launchd.go, systemd.go, crontab.go, temporal.go, select.go). All behavior functions return `errors.New("X: not implemented")`.
5. Created test files: cron_test.go (Group A), manifest_test.go (Group B), render_test.go (Group C), select_test.go (Group D).
6. Created CLI test file: `internal/cli/schedule_test.go` (Group E).
7. Created golden fixtures in `internal/schedule/testdata/`.
8. Verified `go build ./internal/schedule/...` and `go vet ./internal/schedule/...` both clean.
9. Ran `go test ./internal/schedule/...` — 19 tests FAIL for correct reason ("not implemented").
10. Ran `go test ./internal/cli/...` — build failed due to pre-existing broken `runtime_record_test.go` (not our change).
11. Recorded evidence in TDD-LEDGER.md.

## Pre-existing issue

`internal/cli/runtime_record_test.go` (untracked in git, `??` status) references undefined functions `runtimeETLLeaseRequest` and `runtimeETLRunRecord`. This blocks the cli test binary build. Our schedule_test.go is correct and will be red on "unknown command schedule" once that pre-existing issue is resolved.

## Red run evidence (schedule package)

```
FAIL    polymetrics.ai/internal/schedule    0.170s

Failing: TestParseCron_Valid (A-1,2,3), TestCronExprNext (A-10,11,12),
         TestManifestSaveLoad (B-1), TestManifestList (B-2),
         TestManifestDuplicateRejected (B-3), TestManifestDelete (B-5),
         TestRenderPlist_Golden (C-1), TestRenderService_Golden (C-2),
         TestRenderTimer_Golden (C-2), TestCronToOnCalendar (C-2 x2),
         TestRenderCrontabLine_Golden (C-3), TestRemoveCrontabLine_* (C-3 x3),
         TestSelectBackend_Darwin (D-1)
```
