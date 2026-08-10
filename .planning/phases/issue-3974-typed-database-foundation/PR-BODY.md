## Intent

Establish the typed, non-executing database connector foundation required by
the PostgreSQL parity programme.

Refs #3974

## What changed

- Added strict `database.json` loading, closed logical types, bounded
  resources, structured identity/catalog/read-plan values, and defensive
  projections.
- Added exact driver registration and native-admission validation; a declaration
  alone cannot become executable.
- Added the captain-mandated warehouse mediation seam: a connector can form
  only one leg at a time — database source → shared warehouse artifact, or
  shared warehouse artifact → database target. No direct source/destination
  pair, executor, write session, receipt, or zero-copy path was added.
- Kept layer one connector-agnostic: the existing warehouse materializer writes
  WAL then Parquet, the existing reverse path reads published Parquet, and
  `warehouse.ArtifactIdentity`/`ArtifactRef` provide their shared typed
  address. PostgreSQL owns neither those values nor a shared operation.
- Made native admission per-leg. An inbound descriptor cannot be reused as an
  outbound apply declaration.
- Added PostgreSQL's compile-time reference seam only; public `write` and
  `cdc` capability flags remain false.

## MySQL extension proof

A future MySQL author changes only layer two: its strict `database.json`, two
per-leg descriptor/evidence admissions, and native extraction/apply code and
tests. `TestMySQLLayerTwoReferenceCompilesAgainstSharedWarehouseArtifact`
constructs both legs with no PostgreSQL import and no change to
`internal/warehouse`, the generic app mediator, or the shared `database`
package. It is a type-level seam demonstration, not a MySQL capability or
executor declaration.

## TDD and GSD

- Red evidence is retained in `traces/red-run.txt`,
  `traces/warehouse-layer-red-run.txt`,
  `traces/warehouse-owner-identity-red-run.txt`, and
  `traces/warehouse-multi-admission-red-run.txt`.
- Green evidence is retained in `traces/green-run.txt` and
  `traces/warehouse-layer-green-run.txt`.
- Used the issue-first GSD sequence: `discuss-phase` → `plan-phase --tdd` →
  `execute-phase` → `verify-work` → `code-review`. The issue is not a numbered
  roadmap phase, so the repository-approved manual inline fallback is recorded
  in the phase artifacts.
- Required Go skills: `golang-how-to`, design patterns, structs/interfaces,
  error handling, security, safety, testing, context, concurrency, and
  database, naming, documentation, and lint; also `no-mistakes` and the
  relevant GSD skills.

## Testing

- Focused `warehouse`, `database`, `synccontract`, engine, and native Postgres
  tests; race tests for the shared packages; PostgreSQL capability test.
- `go vet ./...`, `go build ./cmd/pm`, direct shared-package lint, and
  `go test -timeout 20m ./internal/cli -count=1`.
- `make tidy-check`, `lint`, `docs-check`, `smoke-no-build`,
  `agent-contract-check`, `connectorgen-validate`,
  `connectorgen-surface-sync`, `connector-boundary`, and
  `release-workflow-check`.

## Scope and safety

No generic SQL executor, database-shaped REST operation, target DDL,
provisioning, write session, transaction/receipt/checkpoint, CDC/changefeed,
or second sync-mode enum is included. No credentials, DSNs, or raw connection
material enter the new typed references.

## Review and pipeline

Manual GSD review found no unresolved scope, security, or type-boundary issues;
the disposition is in `REVIEW.md`. The branch is pushed in two committed
checkpoints. `no-mistakes`, PR creation, remote checks, and automated PR review
remain pending the firstmate-directed delivery phase.
