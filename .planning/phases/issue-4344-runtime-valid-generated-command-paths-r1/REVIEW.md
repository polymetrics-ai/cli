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
| follow-up | The batch-1 Bitbucket source descriptor carrying the 50 implemented commands is not checked into this base, so the required 28/50 binary sweep cannot be reproduced without importing unrelated connector work. | Recorded explicitly in `RUN-STATE.md`, `TDD-LEDGER.md`, and `VERIFICATION.md`; rerun against the reviewed source-import artifact during integration. |

## Verdict

No code changes requested. The foundation change is safe to submit with the
explicit integration-proof dependency above; it must not be represented as
proof of the absent 50-command artifact.
