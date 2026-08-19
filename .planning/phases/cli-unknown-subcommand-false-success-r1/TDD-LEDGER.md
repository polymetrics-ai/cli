# TDD ledger — cli-unknown-subcommand-false-success-r1

Manual-GSD fallback: this worker generated the required `scripts/gsd prompt` commands and records
the equivalent single-worker red/green evidence here because Pi is unavailable and role spawning is
forbidden.

| Slice | Red test/proof | Green assertion | Status |
| --- | --- | --- | --- |
| Invalid connector help path | New regression invokes a known connector with an unrecognised deep path plus `--help`; current behavior returns 0 and root manual. | Same invocation exits 2 and names the unresolved path. | Green |
| Valid connector help | Existing and expanded tests cover root, group, and real deep-command help. | All retain exit 0 and specific manual content. | Green |
| JSON error envelope | Invalid path with `--json --help` currently returns `CommandManual`. | It returns the existing `usage_error` envelope. | Green |

## Red evidence log

The regression test was added before production changes and run against the unchanged resolver:

```text
$ go test -timeout 20m ./internal/cli -run TestDynamicConnectorDeepHelpPathsResolveOrReportUsage -count=1
--- FAIL: TestDynamicConnectorDeepHelpPathsResolveOrReportUsage/unknown_deep_command
    Run(gong calls definitely-not-real --help) code = 0, want usage error
--- FAIL: TestDynamicConnectorDeepHelpPathsResolveOrReportUsage/unknown_deep_command_JSON
    Run(gong calls definitely-not-real --help --json) code = 0, want usage error
FAIL
```

This proved the live defect before the resolver change: the invalid deep path was discarded by the
help renderer and returned the connector root manual with an exit code of zero. The valid companion
case, `pm gong calls transcript --help`, passed in the same run.

## Green evidence log

After the resolver validates a help path as a declared command or a declared one-segment group,
the same regression test passes:

```text
$ go test -timeout 20m ./internal/cli -run TestDynamicConnectorDeepHelpPathsResolveOrReportUsage -count=1
ok      polymetrics.ai/internal/cli
```

`pm gong calls definitely-not-real --help --json` now returns exit 2 and the existing
`usage_error` envelope. `pm gong calls transcript --help` remains a successful deep-command
manual request.
