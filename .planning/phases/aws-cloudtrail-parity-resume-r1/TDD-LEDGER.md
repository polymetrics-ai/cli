# TDD ledger: AWS CloudTrail parity resume r1

| Slice | Red evidence | Green evidence | Notes |
| --- | --- | --- | --- |
| Native command surface | `New()` does not satisfy `connectors.CommandSurfaceProvider`; a native contract test must fail before the delegate exists. | The native connector exposes all 60 bundle command rows: 57 implemented, 3 policy-disallowed. | Delegate only; no shared engine change. |
| Runtime registration | Before the fix, `pm aws-cloudtrail --help` returns `unknown command`. | Runtime help lists the generated groups/commands; representative direct-read and reverse-ETL routes reach normal typed argument/plan validation rather than connector lookup failure. | No credentials and no provider requests. |
| Citation research | The historical field ledger has field names/types/requiredness but no per-field source metadata. | A connector-local research matrix contains an AWS URL/Request Parameters section, provider-reference evidence type, high confidence, and AWS Required Yes/No rationale for every documented request field. | This trace is not a competing bundle schema; shared citation-convention landing will be rebased before final validation. |
| Final verification | N/A | Focused current-main parity gates, generated website data, and diff checks pass. | Full `go test ./...` and `make verify` are intentionally not run in this bounded lane. |

## Manual-GSD decision

The project adapter recognizes no `programming-loop` command, so this ledger is the manual-GSD fallback record. The orchestration decision for the plan/TDD cycle is `local_critical_path`: the user assigned one connector and the runtime instructions prohibit unsolicited subagent delegation.
