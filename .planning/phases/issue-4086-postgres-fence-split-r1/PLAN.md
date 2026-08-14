# PLAN — Issue #4086 PostgreSQL database/connector fence split

## GSD and skills evidence

- `scripts/gsd doctor`, all five `scripts/gsd sources` resolutions, and
  `go run ./cmd/agentcontractgen check` passed before planning.
- Generated prompts were reviewed for `discuss-phase 4086 --auto`,
  `plan-phase 4086 --tdd --skip-research`, `execute-phase 4086 --interactive`,
  `verify-work 4086 --auto`, and `code-review 4086 --depth=standard`.
- Inline/manual fallback is required by the task's no-role-spawning constraint.
- Required skills loaded: `golang-how-to`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
  `golang-safety`, `golang-testing`, and `golang-database`.

## Exact move plan

1. Split `internal/connectors/native/postgres/postgres_test.go` into a stable
   capability-surface test file and a source test file. Move the existing
   fail-closed CDC fence into the already CDC-owned `cdc_test.go`.
2. Split `internal/connectors/database/database_test.go` into mapping-definition,
   source-read-plan, and target-admission files, moving the existing test-only
   helpers unchanged to a neutral helper file.
3. Verify declaration equivalence from the Git diff: every declaration removed
   from either monolith must appear byte-identically in exactly one destination.
4. Build and compare the listed runtime commands against a binary built at the
   explicit base SHA. Compare stdout and stderr byte-for-byte.
5. Run focused and complete touched-package tests; prove generated capability
   parity via `connectorgen surface-sync --check` plus a SHA-256 comparison of
   the generated capability ledger before and after.

## Explicit exclusions

No production declaration, connector definition JSON, generated file, command,
help text, documentation, runtime behavior, test assertion, or test semantics
is changed. Any discovered defect is recorded only as a follow-up candidate.
