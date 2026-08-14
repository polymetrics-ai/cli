# Code review: issue #4087 legacy sync-mode bypass

**Mode:** inline/manual GSD fallback
**Verdict:** pass — no actionable findings

## Reviewed concerns

- The public aliases remain accepted, with their canonical contract and generic capability decision held only in `internal/synccontract/public_modes.go`.
- Normal and persisted-legacy parsing use that same table; both aliases have a non-empty contract and bypass the legacy branch.
- In an unadmitted source/destination pairing the application returns the typed `ModeNotExecutableError` before a source read. The CLI certification report now treats that deliberate `Error`/exit-1 shape as passing evidence, rather than as an ETL failure.
- The existing public compatibility adapters and the closed canonical modes retain their tested parse/runtime behavior. No connector-specific branch or literal was added to shared production code.
- Help, generated CLI docs, website docs, and generated website documentation data describe the same typed-admission behavior.

## Evidence reviewed

- `internal/synccontract/public_modes.go`, `internal/app/sync_modes.go`, and their focused regression/control tests.
- `internal/connectors/certify` stage and scripted-CLI coverage, plus the real CLI certification route test.
- Generated documentation diffs and the completed command matrix in `VERIFICATION.md`.

External pull-request review is intentionally pending: this worker was instructed to stop after the commit, before Firstmate's no-mistakes/PR workflow.
