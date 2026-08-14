# Code review: issue #4087 legacy sync-mode bypass

**Mode:** inline/manual GSD fallback
**Verdict:** pass — no actionable findings

## Reviewed concerns

- The public aliases remain accepted, with their canonical contract and generic capability decision held only in `internal/synccontract/public_modes.go`.
- Normal and persisted-legacy parsing use that same table; both aliases have a non-empty contract and bypass the legacy branch.
- In an unadmitted source/destination pairing the application returns the typed `ModeNotExecutableError` before a source read. The CLI certification report now treats that deliberate `Error`/exit-1 shape as passing evidence, rather than as an ETL failure.
- The existing public compatibility adapters and the closed canonical modes retain their tested parse/runtime behavior. No connector-specific branch or literal was added to shared production code.
- Help, generated CLI docs, website docs, and generated website documentation data describe the same typed-admission behavior.
- The shared engine regression remains connector-neutral, while exact generated-content validation now covers connector manuals, skills, index, and catalog output; regenerated capability-flow artifacts passed their drift check without changing the certification report contract or version.
- The shared effective cursor projects a schema `x-cursor-field` or a validated incremental cursor, while an explicitly conflicting pair remains separately rejected before incremental modes are advertised.

## Evidence reviewed

- `internal/synccontract/public_modes.go`, `internal/app/sync_modes.go`, and their focused regression/control tests.
- `internal/connectors/certify` stage and scripted-CLI coverage, plus the real CLI certification route test.
- Generated documentation diffs and the completed command matrix in `VERIFICATION.md`.
- The focused engine and connectorgen regression tests, including schema-only and incremental-only cursor coverage.
- Required skills used: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, and `golang-lint`.

External pull-request review is intentionally pending: this worker was instructed to stop after the commit, before Firstmate's no-mistakes/PR workflow.
