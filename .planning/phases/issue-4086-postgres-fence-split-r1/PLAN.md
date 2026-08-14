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
   fail-closed CDC fence into the already CDC-owned `cdc_capability_fence_test.go`.
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
6. Prove registration-order changes inert with Go's shuffled test mode at the
   exact comparison base and original move head, using both affected packages
   and deterministic seeds `408601` and `408602`.

## Registration-order planning trace

The proof uses Go's `-shuffle=on` facility in deterministic numeric form:
`go test -count=1 -timeout 20m -shuffle=<seed> <package>`.

| Revision | Seed | `./internal/connectors/database` | `./internal/connectors/native/postgres` |
| --- | --- | --- | --- |
| Base `5a457970b3bc15343e5ba6b7b4acf48994b63add` | `408601` | `ok 6.452s` | `ok 0.881s` |
| Base `5a457970b3bc15343e5ba6b7b4acf48994b63add` | `408602` | `ok 6.229s` | `ok 0.878s` |
| Original move head `6e31ac1abfc4a46fd1dbbef3ec54086da85b682e` | `408601` | `ok 6.416s` | `ok 0.875s` |
| Original move head `6e31ac1abfc4a46fd1dbbef3ec54086da85b682e` | `408602` | `ok 6.652s` | `ok 0.870s` |

## Explicit exclusions

No production declaration, connector definition JSON, generated file, command,
help text, documentation, runtime behavior, test assertion, or test semantics
is changed. Any discovered defect is recorded only as a follow-up candidate.
