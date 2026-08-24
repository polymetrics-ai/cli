# Inline code review — #4344

## Scope reviewed

- `cmd/connectorgen/sourceprojection.go`
- `cmd/connectorgen/sourceprojection_test.go`
- GSD/TDD evidence in this phase directory

## Findings and dispositions

| Severity | Finding | Disposition |
| --- | --- | --- |
| none | New command paths are a two-segment parser-valid identity: `api op-<hex(METHOD path)>`. Hex is injective over the endpoint key, deterministic, and does not expose `{`/`}`. | Accepted. |
| none | The migration predicate changes only the generator's old raw-ID path when the runtime's own identifier validator rejects it. Valid legacy paths, including existing GitHub names, remain untouched. | Accepted; behavioral test covers both sides. |
| none | Path-parameter command flags remain generated from the action contract. The test exercises `commandrunner.BuildWriteCommand` and asserts the bound `workspace`, `repo_slug`, and `id` record fields. | Accepted. |
| none | The batch-1 Bitbucket source descriptor is not checked into this base, but its measured fixture was regenerated in an isolated temporary directory with this PR's generator and tested through an overlay binary. | Accepted: 50/50 reached `missing --credential`; 0 unknown or invalid paths; no connector-local artifact was absorbed. |

## Verdict

No code changes requested. The foundation change is safe to submit. The
isolated exact-batch fixture provides the required 50-command binary proof
without expanding this PR to include connector-local reconciliation.
