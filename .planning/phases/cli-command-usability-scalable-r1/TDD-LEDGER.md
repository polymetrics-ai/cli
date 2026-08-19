# Issue #4193 TDD ledger

| Slice | Red | Green | Refactor | Status |
| --- | --- | --- | --- | --- |
| Exhaustive legacy leaf help | `go test -run TestEveryLegacyLeafHelpIsExecutable` failed on base because `init`, `help`, `man`, `extract`, and `worker` had no manual. | Added a shared wrapper-level `containsHelpFlag` path and complete wrapper manuals; the exhaustive test passes. | Removed `legacyLeafManualTopic`; no per-command help switch remains. | green |
| Error boundary | `TestLeafHelpDoesNotMaskInvalidCommandsOrApprovalCarrierSyntax` failed because unknown `--help` returned exit 1 and `schedule create --help` required flags. | Unknown command help now returns usage exit 2; malformed approval syntax remains a usage error; required-flag help renders its manual. | `writeError` remains the sole exit-code reporter. | green |
| Built-binary proof | Base binary enumeration records failures and accidental execution. | Rebuilt and verified 63 legacy paths × `--help`/`-h` × empty/initialized project = 252 successful manual renders; all 36 declaration-backed connector roots passed the equivalent 144-request binary sweep. | Regenerated five newly tracked CLI manuals and updated the transcript fixture. | green |
| Declaration-backed leaves | The original two-row test did not enumerate the provider command surface. | A registry-driven test checks both help spellings for all 8,900 declared connector leaves before dispatch (17,800 variants). | The test consumes `CommandSurface.Commands`, so new generated commands enter the sweep automatically. | green |

## Baseline red evidence

Built `./cmd/pm` at `ff6a87101` and invoked each leaf in an isolated directory
without `.polymetrics`. The sample confirmed failures for credentials, catalog,
query, flow, reverse, schedule, and many ETL paths. It also found a more serious
variant: `agent plan --help`, `runtime doctor --help`, and performance paths can
enter handlers instead of rendering help. The complete input inventory is in
`VERIFICATION.md`; focused red-test failure and green-test commands are recorded
there before broader verification.
