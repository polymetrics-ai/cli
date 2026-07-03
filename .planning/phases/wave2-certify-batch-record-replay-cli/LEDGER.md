# TDD Ledger

Phase: wave2-certify-batch-record-replay-cli

Scope: certification design §B (batch mode), §E (Tier-1 record/replay), and CLI wiring
(`pm connectors certify <name> | --all --credentials-file | --sweep`) per
docs/architecture/connector-certification-design.md. Ran concurrently in the same working
tree as another agent's write-protocol phase (`wave2-certify-write-stages`:
ledger.go/pairing.go/stages_write.go/sweeper.go/stages_glue.go) — see "Concurrent work"
note below.

| Task | Status | Evidence | Tests green |
| --- | --- | --- | --- |
| credsfile.go — creds.yaml parsing (version/defaults, per-connector credential from_env/exec, sandbox/write/rate_limit/skip/reason), ResolveSecrets, EffectiveOptions, ConnectorNames | red-confirmed → green | credsfile_test.go | 15 |
| record.go — RecordingTransport (RoundTripper wrapper, per-stage sequential cassette files, Authorization/Cookie/X-Api-Key header scrub, secret-value scrub exact/base64/urlencoded, caller-visible body untouched) | red-confirmed → green | record_test.go | 5 |
| replay.go — ReplayTransport (per-stage sequence cursor, method+path match, unmatched request/exhausted sequence/unknown stage/method-mismatch → hard failure, shares cassette format with record.go) | red-confirmed → green | replay_test.go | 8 |
| batch.go — RunBatch (worker pool bounded by Defaults.Parallel, min 1; skip:true entries never invoke RunnerFactory; --resume skips connectors with a report newer than batch start; per-connector Runnable error captured not fatal; progress.json written; SummaryMatrix with leak rows sorted first) | red-confirmed → green | batch_test.go | 15 |
| report.go — additive Leak/ExitCodeFor/ScheduleResult/WriteActionResult/Capabilities.Flow·Schedule·WriteActions/Report.Leaks (needed by batch.go; landed concurrently with the write-protocol phase converging on the identical shape) | n/a (infra, covered transitively) | report_test.go (pre-existing, still green after additive fields) | 7 |
| cliharness.go — SetCLIRunFuncseam: certify can no longer import internal/cli directly once internal/cli/certify_cli.go imports certify (import cycle); production wiring moved to cmd/pm/main.go | red-confirmed → green | certify_testmain_test.go (TestMain wiring), full certify suite | n/a (refactor, zero behavior change verified via full suite rerun) |
| internal/cli/certify_cli.go — single/batch/sweep dispatch, Options-from-flags, JSON+text rendering, exit-code mapping (0/1/2/3) via new certifyExitErrorf + cliError.exitOverride/alreadyReported (additive to errors.go, zero effect on existing categories) | red-confirmed → green | certify_cli_test.go, certify_exit_test.go | 19 |
| cli.go wiring — runConnectors gains `case "certify"` calling runCertify; signature widened to (ctx, root, args, stdout, jsonOut) at its one call site in Run() | red-confirmed → green | certify_cli_test.go TestCertifyCLIDoesNotBreakExistingConnectorsSubcommands (regression guard) | 2 |

## Self-verify (final, clean run)

```
go build ./...                                  # exit 0
go test ./internal/connectors/certify -count=1  # ok, 25.5s, 0 failures (full package incl. concurrent write-protocol phase)
go test ./internal/cli/... -skip 'TestScheduleCLI_Remove$' -count=1  # ok, 10.2s
make lint                                       # 0 issues
go vet ./...                                    # clean
gofmt -l <every file this phase touched>        # clean
go mod tidy && git diff go.mod go.sum           # yaml.v3 promoted indirect->direct only, go.sum unchanged
```

Manual end-to-end smoke (built binary, real `pm init`-style ephemeral root, sample connector):
single-connector JSON + text (exit 0 pass, exit 2 on forced `--stream doesnotexist` failure),
batch JSON + text matrix (exit 0), `--sweep` with and without `--credentials-file` (exit 2 usage
error / exit 0 no-op respectively). Verified stdout carries exactly one JSON envelope per
invocation even on the certify-failure exit path (`alreadyReported` suppresses writeError's
would-be duplicate `Error` envelope).

## Deviations from the literal file list

Task said "in internal/connectors/certify/{credsfile.go,record.go,replay.go}". Batch-mode
orchestration (RunBatch/BatchReport/SummaryMatrix) didn't fit inside those three files, so it
landed in a new sibling file `batch.go`, per the design doc's own package layout (§E) which
lists batch orchestration as part of `certify.go`/report.go's territory, not those three. Also
added: `report.go` additive fields (Leak/ExitCodeFor/etc., required by batch.go, see below),
`cliharness.go` SetCLIRunFunc seam (required to resolve an import cycle — see below),
`certify_testmain_test.go` (TestMain wiring for the package's test binary), `cmd/pm/main.go`
(one-line production wiring of the same seam).

## Import cycle: internal/cli/certify_cli.go <-> internal/connectors/certify

`certify`'s `cliharness.go` drives the real CLI in-process via `cli.Run` (by design — see design
doc §A execution model). Wiring `pm connectors certify` into `internal/cli/cli.go` means
`internal/cli` must import `certify` too, i.e. `cli -> certify -> cli`, which Go forbids.
Resolved by removing certify's compile-time dependency on `internal/cli`: `cliharness.go` now
calls a package-level `cliRun` function variable, set once via the new exported
`certify.SetCLIRunFunc(run func([]string, io.Writer, io.Writer) int)`. `cmd/pm/main.go` (a leaf
package that already imports both) calls `certify.SetCLIRunFunc(cli.Run)` before entering
`cli.Run` itself. Every `certify` package test now wires this once via a `TestMain` in
`certify_testmain_test.go`; `internal/cli/certify_cli_test.go` does the same for its own test
binary (needed since its tests drive `pm connectors certify` through the real `cli.Run`, which
recurses into `certify.Runner`/`certify.RunBatch`, which need `cliRun` set). This is the only
change to a wave0 "already done" file's *contract* (a one-line call-site swap
`cli.Run(...)` -> `cliRun(...)` in `Harness.Run`); no stage logic changed.

## Exit-code wiring: internal/cli/errors.go

Certification design §A's exit codes (0 pass / 1 usage-internal / 2 certification failures /
3 leaked resources) are on a different scale than every other command's category-based mapping
in errors.go (usage=2, validation=3, auth=4, ...). Added two purely-additive fields to the
unexported `cliError` struct — `exitOverride *int` (checked first in `exitCodeFor`, nil
everywhere except certify's path) and `alreadyReported bool` (tells `writeError` to skip its own
stdout/stderr writes, since certify already wrote its report to stdout before returning the
sentinel error — cli.Run's one-JSON-envelope-per-invocation contract must hold even on a
certify-failure exit). New constructor `certifyExitErrorf(code, format, args...)` is the only
caller of either field; every pre-existing error constructor and category is untouched.

## Concurrent work in the same working tree

A second agent was independently implementing the write-protocol phase (ledger.go, pairing.go,
stages_write.go, sweeper.go, stages_glue.go, plus stages_source.go/certify.go/report.go edits)
in this exact directory throughout this session, with visible in-flight compile breaks that
resolved themselves over several minutes each time (confirmed by polling `go build`/`go vet`
every ~15s until stable). None of those files were edited by this phase except report.go, whose
additive fields (Leak/ExitCodeFor/ScheduleResult/WriteActionResult/Capabilities.Flow·Schedule·
WriteActions) were independently converged upon by both agents to the identical shape (verified
by diffing an earlier observed version of report.go from the other agent's WIP against what was
implemented here). By the end of this session the full `internal/connectors/certify` package
(89 tests total across both phases) and `internal/cli` package are green together.

## Known pre-existing issue (not touched, not in scope)

`internal/cli.TestScheduleCLI_Remove` (schedule_test.go) hangs indefinitely in this sandbox: it
calls `pm schedule remove` without `PM_CRONTAB_FILE` set, so `runScheduleRemove` shells out to the
real `crontab -l`/`crontab -` binary, which blocks with no interactive terminal/crontab daemon in
this environment. Reproduced identically with this phase's changes fully stashed out (git stash
of internal/cli/errors.go, the only cli.go-family file this phase's diff touches that predates
the certify dispatch case), confirming it is environmental, not a regression from this phase.
Not fixed — unrelated to certify, and destructive/global changes to schedule.go are out of scope.
