# PLAN — Issue #3974 typed database foundation recovery

## GSD execution record

The following commands were resolved through `scripts/gsd sources` and generated with
`scripts/gsd prompt`: `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and
`code-review`. `scripts/gsd doctor` and `go run ./cmd/agentcontractgen check` passed.

The official workflow expects a numbered roadmap phase and isolated roles. #3974 is an
issue-local foundation and `.agents/agentic-delivery/canonical/delivery-contract.json` requires
one inline worker with delegation disabled. This is the explicit manual-GSD fallback; it preserves
the required lifecycle and artifacts instead of silently skipping it.

Required skills loaded: `golang-how-to`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-naming`, `golang-error-handling`, `golang-security`,
`golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`, `golang-database`,
`golang-documentation`, `golang-lint`, `github-issue-first-delivery`, `no-mistakes`, and the five
GSD lifecycle skills.

## Scope

Recover the exact 12-commit range from `0be21e1293a9e5c913d9b4e23a846237594630c7` through
`9e763fe07a69b48f8ed88484ccb18580c63ae160` on
`feat/3974-typed-database-foundation-recovery`, based at
`8242bedee0cfa75fc61d1fb0e47d73251605162d`.

## TDD/replay sequence

1. **Red — typed admission:** before replay, run the focused absent-package contract test and
   preserve its failure. This proves the parent does not already contain the F1 contract.
2. **Red — semantic replay:** cherry-pick the ordered range. Preserve conflict state/output for
   the three known semantic overlaps before resolving them.
3. **Green — replay:** retain all twelve commits in order and resolve only the known newer-canon
   overlaps.
4. **Green — admission contract:** run the topology-report database, warehouse, synccontract,
   engine, PostgreSQL, connector-registry, and application suites. Confirm every listed F1 test
   passes and that the `database.json` admits no modes.
5. **Refactor/verify:** format, run static/build and individual repository gates, then complete
   inline `verify-work`, any necessary gaps-only loop, and deep code review.

## Conflict dispositions

| Path | Disposition |
| --- | --- |
| `docs/architecture/connector-architecture-v2-design.md` | Preserve #4003's current-canon framing and only integrate the F1 database declaration/warehouse-mediation rules that do not weaken it. |
| `docs/migration/conventions.md` | Preserve current bundle-authoring canon and add only the strict optional native `database.json` policy recipe. |
| `internal/connectors/native/postgres/connector.go` | Retain current TLS and CDC fail-closed behavior; add the reference seam documentation/assertion only. |

## Acceptance checks

- `database.json` is strict, secret-safe, bounded, and exposes defensive projections.
- Logical compatibility is exact/lossless only; structured identities include workspace, connector,
  and connection IDs without credentials.
- Driver declarations require registered, compatible, leg-specific native admission.
- Every database leg is warehouse-mediated and cannot represent a direct source-to-target pair.
- PostgreSQL retains `write:false`, `query:false`, and `cdc:false`, with no executable target,
  direct query, canonical mode, or CDC v2 behavior.

## Commit/push checkpoints

1. GSD planning checkpoint (this artifact set).
2. Ordered recovery replay with semantic dispositions and focused Green evidence.
3. Verification/review/no-mistakes correction checkpoint(s), capped at five correction loops.

## Verification commands

Run the focused tests named in the topology report, then `./internal/app`, `./internal/cli`,
formatting, `go vet ./...`, `go build ./cmd/pm`, and each required make gate individually.
Runtime-backed PostgreSQL checks are not applicable: this F1 slice neither enables an executor nor
uses credentials or performs a database mutation.
