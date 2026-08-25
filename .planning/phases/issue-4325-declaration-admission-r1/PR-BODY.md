## Intent

Refs #4325. Add the shared source-declaration admission foundation for Batch 1
without converting connector-owned definitions or claiming live certification.

## What Changed

- Added `connectorgen declaration-admission [dir] [--json]`, a required
  independently counted source-cohort schema, and a separate declaration
  catalog schema.
- Required exactly one declaration per source row, a canonical lane/endpoint,
  a discoverable `cli_surface.json` command, delete metadata, and either a
  proven runtime binding or named deferred foundation.
- Added `deferred` command metadata and a typed pre-execution
  `system/missing_foundation` refusal, while retaining existing implemented
  runtime preflight behavior.
- Added focused red/green tests and certificate-separation documentation.
- Deferred source rows now name a bounded missing implementation component and
  its evidence plus one exact API-surface target; excluded, policy-only,
  duplicate, mismatched, and stale targets are rejected before the declared
  command can return its named missing foundation.
- Reused the source-import public HTTPS citation policy and the real no-I/O
  command preflight so citation and resolver semantics cannot drift.
- Required semantic destructive metadata from the exact target, including
  non-`DELETE` `destructive_action` operations, and rejected false implemented
  bindings against excluded, policy-only, or duplicate surface rows.
- Added GitHub `label delete` as the implemented destructive green control:
  endpoint-specific missing action contracts are deferred, not delete
  operations as a class.
- Required exact source/declaration/runtime binding identities, including for
  GraphQL operations sharing one transport endpoint; empty provider-native
  operation IDs remain valid when the stable source identity and citation are
  exact.
- Embedded the compact source target ledger used by deferred production
  preflight, while retaining the existing omission of full `api_surface.json`.
  App planning now preflights before credentials, oversized foundation metadata
  remains typed, and the public CLI preserves `missing_foundation`.
- Made source-operation uniqueness provenance-only and binding uniqueness a
  separate invariant, so changing a local binding cannot hide a duplicate
  provider row.
- Kept canonical provider identities separate from physical runtime endpoints,
  with closed named equivalence for declared base paths/placeholders, registered
  hooks, GraphQL `POST /graphql` transport, queries, suffixes, and operation
  annotations. Arbitrary method/path substitution fails closed.
- Required deferred rows to fail the real implemented commandrunner preflight
  before an executor-specific foundation gap is accepted; runnable GitHub
  delete and GraphQL read/write controls cannot be relabelled deferred.
- Classified the compact runtime declaration ledger exactly in the production
  embed inventory and added a real Outreach `/api/v2` ETL/destructive no-I/O
  integration gate without changing the Outreach bundle.

The accompanying investigation records the exact distinction: GitHub's cited
`label delete` endpoint has a typed delete action, CLI binding, fixture, and
runtime-preflight control; Stripe's `/v1/accounts/{account}` is an existing
blocked source mapping with no action or command binding. See
`DELETE-CONTROL-AND-STRIPE-INVESTIGATION.md`.

## Scope and Safety

- No provider I/O, credentials, writes, deletes, retained source artifacts,
  hashes, generated connector evidence, or Batch-1 connector definitions.
- Source-lock, surface-sync, runtime preflight, credential-bound binary proof,
  and live certification remain strict and separate.
- Per captain clarification 007, individual provider declarations and CLI
  projections remain in their mapping lane; this PR commits only generic
  admission semantics and regression infrastructure.

## TDD and GSD Evidence

- Red: the admission suite initially failed to build without the document,
  checker, and state contract; a later red test caught mismatched implemented
  write endpoints.
- Green: focused acceptance tests plus full changed Go packages pass. Details:
  `.planning/phases/issue-4325-declaration-admission-r1/TDD-LEDGER.md`.
- Lifecycle was executed inline using the resolved `scripts/gsd prompt`
  workflow because the repository contract forbids role spawning here.
- Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, and
  `golang-documentation`, plus `golang-lint` for static verification.

## Testing

- `go test -timeout 20m ./cmd/connectorgen` (pass)
- `go test -timeout 20m ./internal/connectors/commandrunner` (pass)
- `go test -timeout 20m ./internal/connectors` (pass)
- `go test -timeout 20m ./internal/connectors/engine` (pass)
- `go vet ./...`; `go build ./cmd/pm`; docs, lint, smoke, generated artifact,
  certification projection, connector boundary, canon, and release-workflow
  gates (all pass)
- Audit repair: focused declaration-admission, deferred-command, and engine
  projection suites pass, including the exact-SHA DA-001/002/003/004/005/006/
  010/011 repairs; repository-wide vet/build, tidy, lint, agent-contract, and
  GSD evidence checks are recorded in verification.
- R2 exact head `f97dede07`: all 477 connectorgen tests/examples pass across
  deterministic shards; engine, commandrunner, connector, definitions, Notion
  hook, app, and CLI packages pass; the fleet census proves 243 non-GraphQL and
  4 GraphQL aliases; real Outreach passes; boundary reports 322 files / 553
  connectors / 0 findings; exact-head vet, lint, build, docs, smoke, admission,
  surface-sync, release, and GSD workflow gates pass.

The aggregate `go test -timeout 20m ./...` and serial `make verify` are left
to CI, per the repository’s per-command-timeout guidance. Full commands and
results are in `.planning/phases/issue-4325-declaration-admission-r1/VERIFICATION.md`.

## CLI/Docs Parity

This is an internal `connectorgen` command, not a new shipped `pm` command.
Its usage and connector certification/canon docs are updated; `docs/cli/**`,
website/manual generation, `pm` bare namespace behavior, and shell completion
are not applicable.

## Review Route

The original route was `claude_auto` on this non-draft, `main`-targeted PR. The
audit repair is a new unreviewed commit and receives one deliberate fresh
review request after push, plus Firstmate's independent-Codex re-audit handoff.
Automated findings, if any, require disposition before merge.
