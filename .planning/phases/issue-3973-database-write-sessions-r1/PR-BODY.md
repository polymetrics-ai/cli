Closes #3973
Refs #3972

Parent branch: `integration/4015-mvp-flat-r1`
Parent PR: #4100 (`integration/4015-mvp-flat-r1` → `main`)

## Intent

Deliver the driver-neutral transactional database apply/session foundation that
blocks the PostgreSQL native write driver (#3982). No PostgreSQL SQL, DDL,
capability change, connector registration, CLI surface, credential, or live
target execution is included.

## What changed

- Added sealed database write plans, preview-bound one-use approval, canonical
  mode/strategy admission, and exact definition/driver compatibility.
- Added one pinned `WriteSession` with bounded batches, whole-session rollback,
  atomic full-overwrite publish, and explicit committed/rolled-back/unknown
  commit outcomes.
- Added durable target receipt recording through the managed-target delivery
  ledger; only a ledger-recorded confirmed receipt can produce checkpoint
  acknowledgement authority.
- Added fake-driver tests that observe session counts, bounded batches,
  approval ordering, zero mutations on refusal, rollback, unknown-commit
  terminal handling, atomic overwrite sequencing, canonical strategies, and
  the legacy-write trap.

## Red / Green

- Red: `go test -timeout 20m ./internal/connectors/database -run 'TestDatabaseWriteExecutor' -count=1` failed because the write-session contract did not exist; retained at `traces/write-session-red.txt`.
- Green: focused normal and race tests pass; retained at `traces/write-session-green.txt`.

## GSD and skills

- Inline manual GSD fallback: `discuss-phase` → `plan-phase --tdd` →
  `execute-phase` → `verify-work` → `code-review`. The canonical single-worker
  contract forbids role spawning for this foundation.
- Skills: `golang-how-to`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
  `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`,
  and `golang-database`.

## Testing

- `go test -timeout 20m ./internal/connectors/database/...`
- `go test -count=1 -timeout 20m ./internal/app/...`
- `go test -race -timeout 20m ./internal/connectors/database -run 'TestDatabaseWriteExecutor' -count=1`
- `go test -timeout 20m ./internal/connectors/native/postgres -run '^TestWriteUnsupported$' -count=1`
- `go vet ./...`; `go build ./cmd/pm`
- `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`,
  `make agent-contract-check`, `make connectorgen-validate`,
  `make connectorgen-surface-sync`, `make connector-boundary`, and
  `make release-workflow-check`

## Safety and follow-up

- PostgreSQL remains descriptor-only with `write=false`; no generic write tool
  or raw SQL path is introduced.
- Real native PostgreSQL proof is deferred by design to #3982; final
  real-binary capability certification is #3978.

## Review coverage

- Primary: `claude_auto`, pending after PR open for the full range from
  `integration/4015-mvp-flat-r1` to this PR head.
- Fallback: Copilot only if Claude fails, is skipped, or is unavailable; no
  review comments are pending at PR creation.
