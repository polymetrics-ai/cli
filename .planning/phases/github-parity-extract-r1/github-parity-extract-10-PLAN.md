---
phase: github-parity-extract-r1
plan: "10"
type: tdd
wave: 10
depends_on: ["07", "08", "09"]
autonomous: true
files_modified:
  - scripts/github-combined-operation-ledger.mjs
  - scripts/tests/github-combined-operation-ledger.test.mjs
  - scripts/testdata/github-combined-operation-ledger/mini-schema.graphql
  - internal/connectors/defs/github/sources/github-operation-source-lock.json
  - .planning/phases/github-parity-extract-r1/GITHUB-COMBINED-OPERATION-LEDGER.json
  - .planning/phases/github-parity-extract-r1/TDD-LEDGER.md
---

# Plan — typed GraphQL source import and combined inventory contract

## Goal

Upgrade the GitHub source lock from root-name/signature inventory to a compact, versioned,
type-bearing GraphQL source model.  This is the prerequisite for generated fixed GraphQL operation
contracts: each root must retain its typed argument declarations, return type, and the type-system
facts required to derive bounded selections and the `node`/`nodes` projection matrix.  The
generated combined ledger remains a factual inventory in this slice; it does not claim that the
305 GraphQL roots are implemented or live-proven.

The scope is GitHub source tooling and generated GitHub/planning artifacts only.  The prior
provider-neutral GraphQL runtime remains unchanged.  No provider fixture, `pm`, credential, or
browser operation is permitted in this slice.

## Pinned input and provenance

- The official, public, unauthenticated GitHub Docs SDL is
  `https://docs.github.com/public/ghec/schema.docs.graphql`.
- The current source lock pins its exact 2026-08-09 bytes: `1,546,421` bytes and SHA-256
  `c09aba9911b08d2aa8a022578edaf256aa040f38d7fb7196656356ea236c249d`.
- A local in-memory fetch reproduced those exact facts and counts (31 `Query`, 274 `Mutation`,
  including `createEnterpriseOrganization`) before this plan was written.  It was a public source
  artifact read, not a provider API/fixture request and performed no provider write.
- Normal hermetic CI validates the checked-in typed lock and generated ledger.  A scheduled
  source-drift job can fetch the public artifact and compare its bytes/hash; no ordinary test may
  require network access.

## Typed-lock contract

- `schema_version: 2` retains all v1 REST and GraphQL provenance/count facts and adds a compact
  GraphQL type graph.  Each root field has a structured ordered `arguments` array and structured
  `return_type`, in addition to the human-readable signature.
- The graph records named input objects/fields, enums, objects, interfaces, unions, and scalar
  fields/types reachable from a root.  Type references preserve list and non-null structure.
  The importer validates references against built-in scalars or a declared type and rejects
  duplicate root/type/field identities rather than silently accepting drift.
- The graph derives interface/union possible object types, so generic `node` and `nodes` ledger
  rows can be source-backed projection matrices.  It must not fabricate selectable fields or
  caller-configurable selections.
- `createEnterpriseOrganization` is both an inventory and type canary: import/check fails if the
  mutation, its `input: CreateEnterpriseOrganizationInput!` argument, or the referenced input
  object disappear or become unclassified.
- The combined ledger adds source type facts and projection availability without changing the
  terminal implementation state of an unimplemented root.  Its validator remains fail-closed for
  missing source rows, a source-hash mismatch, forbidden `UNTESTABLE`, or a missing canary.

## TDD slices

### Red 30a — type-bearing source lock is mandatory

Extend the hermetic mini SDL with an input object, enum, object/interface relationship, union, and
`node`/`nodes`-compatible return definitions.  Add tests that require the source lock to retain
root argument/return type references, nested input fields, declared enum values, and source-derived
possible object types.  The current v1 lock/parser must fail because it only carries formatted root
signatures.

### Red 30b — malformed type graph and enterprise canary fail closed

Add tests that remove a referenced input type, duplicate a type/field, or remove the typed
`createEnterpriseOrganization` argument.  Each must reject construction/checking before a ledger
can be written.  Add a ledger test proving `node`/`nodes` projections use the lock's type graph,
not a regex over arbitrary command documents.

### Green 30 — compact SDL type import and regenerated GitHub facts

Extend the deliberately small SDL lexer/parser only as far as the public schema requires.  Produce
the v2 source lock from the verified exact SDL and regenerate only the GitHub combined ledger.  The
parser must be deterministic, dependency-free, bounded to its input, and never execute GraphQL.
Keep the raw SDL out of runtime bundles and do not add an unrestricted GraphQL transport.  Update
the lock checker and source-inventory test so a future v1/root-only lock cannot pass.

### Refactor/checkpoint 30 — reusable typed source boundary

Keep the source importer independent of command generation.  The follow-on command-generator
slice consumes this typed lock to emit fixed documents, closed variable schemas, bounded selection
sets, declared pagination, source-derived command/docs artifacts, and per-root factual
classification.  It must not be smuggled into this importer as a broad raw GraphQL escape hatch.

## Verification

All verification in this phase is hermetic/local after the one pre-plan public artifact hash
verification described above:

```bash
node --test scripts/tests/github-combined-operation-ledger.test.mjs
node scripts/github-combined-operation-ledger.mjs --check
node --test scripts/tests/github-live-lab.test.mjs
node scripts/github-live-lab.mjs --check-boundary \
  --boundary .planning/phases/github-parity-extract-r1/GITHUB-LIVE-LAB-BOUNDARY.json
go test -timeout 20m ./internal/connectors/engine -count=1
go test -timeout 20m ./internal/connectors/commandrunner -count=1
go run ./cmd/connectorgen validate internal/connectors/defs
go run ./cmd/connectorgen surface-sync --check
git diff --check
```

CLI help/manual/website parity is intentionally not applicable in this importer-only slice:
no user-visible command, flag, output, or generated command catalog changes.  The follow-on
generated-command slice must run the full CLI parity checklist before claiming new commands.

## GSD and skill record

`scripts/gsd doctor`, all five `scripts/gsd sources` resolutions, `agentcontractgen check`, and
the generated `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and
`code-review` prompts ran in this worktree.  The inline/manual fallback is required because this
is the canonical single-worker parent lane and compatible GSD role spawning is forbidden.  Skills
loaded: `golang-how-to`, `golang-cli`, `golang-graphql`, `golang-testing`,
`golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, and
`golang-structs-interfaces`.
