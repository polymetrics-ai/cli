---
phase: github-parity-extract-r1
plan: "11"
type: tdd
wave: 11
depends_on: ["09", "10"]
autonomous: true
files_modified:
  - internal/connectors/connectors.go
  - internal/connectors/commandrunner/runner.go
  - internal/connectors/commandrunner/runner_test.go
  - internal/connectors/engine/graphql_operation.go
  - internal/connectors/engine/graphql_operation_test.go
  - internal/connectors/engine/connector.go
  - internal/connectors/engine/bundle.go
  - internal/connectors/engine/operation_endpoint_ledger.go
  - internal/connectors/engine/direct_write.go
  - internal/connectors/engine/schema/api_surface.schema.json
  - cmd/connectorgen/validate.go
  - cmd/connectorgen/*_test.go
  - internal/connectors/conformance/{static,static_test,github_exhaustive_proof_internal_test,github_exhaustive_proof_test}.go
  - internal/connectors/certify/{stages_surface_inventory,stages_surface_inventory_internal_test,github_source_lock_internal_test}.go
  - scripts/gen-github-graphql-parity.mjs
  - scripts/tests/gen-github-graphql-parity.test.mjs
  - scripts/github-combined-operation-ledger.mjs
  - scripts/tests/github-combined-operation-ledger.test.mjs
  - scripts/github-source-drift.mjs
  - scripts/tests/github-source-drift.test.mjs
  - .github/workflows/github-source-drift.yml
  - Makefile
  - internal/connectors/defs/github/{api_surface,operations,cli_surface}.json
  - internal/connectors/defs/operation_endpoint_ledger.json
  - docs/connectors/{github/{MANUAL,SKILL},catalog/{all-connectors.md,all-connectors.json}}
  - website/{scripts/{gen-github-cli-surface,gen-github-cli-surface.test}.mjs,content/docs/github-cli-surface.mdx,data/connectors.generated.json,lib/{docs.generated.ts,connectors.catalog.data.generated.json}}
  - .planning/phases/github-parity-extract-r1/{GITHUB-COMBINED-OPERATION-LEDGER.json,TDD-LEDGER.md,github-parity-extract-11-{VERIFICATION,UAT,REVIEW,SUMMARY}.md}
---

# Plan — generated, typed GitHub GraphQL root contracts

## Goal

Turn every root retained by the checked-in GitHub GraphQL source lock into one fixed, typed PM
operation contract without introducing a caller-controlled GraphQL document, selection, endpoint,
header, or cursor.  This is the command-generation bridge between the v2 source importer and the
later PM-only live cohorts.  It must leave the historic live-proof evidence and the private lab
directory untouched, and it must not make a provider request.

The authoritative denominator is the typed source lock: 31 `Query` roots and 274 `Mutation` roots.
For each source root the generator will emit exactly one source-derived operation ID and one PM
command under `pm github graphql query …` or `pm github graphql mutation …`, except that the
captain-recorded `deleteIssue` policy remains explicitly non-executable.  Existing legacy commands
remain as compatibility bindings; they do not alter the root denominator.

## Fixed contract boundary

- The generated document has a deterministic operation name, source-declared variable types, the
  exact root field, and a bounded declaration-owned selection.  It never accepts raw GraphQL,
  caller selections, caller headers, or a caller endpoint.
- Composite results use a minimal fixed `__typename` projection; generated connection documents add
  a bounded `pageInfo { hasNextPage endCursor }` contract only when the source type proves it.  The
  `node`/`nodes` rows retain a source-derived possible-object matrix; their fixed `__typename`
  projection is valid for every listed type and is not a claim to expose arbitrary fields.
- Variables are generated from the typed source graph as recursively closed JSON Schema with
  `additionalProperties: false`, array limits, and requiredness matching GraphQL non-null markers.
  The source graph has no input-object cycles and a measured maximum nesting depth of five, so a
  complete finite schema can be generated rather than silently truncating a recursive type.
- A structured `json` flag is admitted only for a top-level `body.<variable>` of a named,
  fixed-GraphQL operation whose own closed variable schema declares that variable as an object or
  array.  Runtime preflight and static generation share the engine validator.  No other direct
  read/write JSON body is widened.
- A single physical `POST /graphql` transport row is covered by an exact generated list of operation
  IDs.  It is deliberately separate from the pinned REST count; the runtime endpoint ledger binds
  each fixed query operation ID to the transport, while GraphQL mutation preflight binds the same
  source-owned transport through that exact coverage list.
- All mutations stay inside the existing plan → preview → approval → execute lifecycle.  Source
  roots not explicitly classified as an approval-only non-destructive verb are fail-closed as
  destructive and receive the existing typed confirmation.  `deleteIssue` remains
  `unsafe_or_disallowed` under the recorded provider/product decision.

## TDD slices

### Red 31a — structured JSON must not be a generic direct-operation escape hatch

Add an engine/commandrunner fixture with one fixed GraphQL root that has a closed object variable.
The real preflight and command shaper must accept exactly that top-level `body.input` JSON value,
reject malformed/scalar values, reject nested/unknown variables, and continue rejecting `json` on
ordinary REST direct reads/writes.  Before the narrow GraphQL-variable admission exists, the valid
fixture fails at the reverse-ETL-only rule.

### Red 31b — a type-bearing source lock must produce every root contract

Add a hermetic mini-schema generator test that asserts one fixed operation, command, variable
schema, selection, and transport coverage entry per root.  It must fail for a missing root, an
unbounded array/object variable, a caller-visible raw selection/document field, a stale
`createEnterpriseOrganization` contract, duplicate generated path/ID, or an unclassified
`deleteIssue` mutation.

### Green 31 — source-only generation and exact physical binding

Implement `scripts/gen-github-graphql-parity.mjs` as a deterministic generator over the checked-in
v2 lock.  Regenerate only GitHub operation/CLI/API artifacts and the shared runtime endpoint
ledger.  Update the common endpoint coverage model only to enumerate exact operation IDs sharing a
physical transport; do not treat the transport as a REST OpenAPI operation.  Extend the combined
ledger so an executable, complete fixed GraphQL binding reports implementation separately from live
proof, and retain an explicit factual blocker for `deleteIssue`.

### Green 31b — runtime and static gates use the same typed-variable authority

Expose a deliberately narrow engine preflight for one closed GraphQL variable, wire it through the
command runner, and have `connectorgen validate` call the same engine rule.  Add no generic HTTP
or GraphQL request interface.  The endpoint/operation ledger must reject a GraphQL root whose
fixed operation ID is absent from the transport coverage.

### Refactor/checkpoint 31 — help, docs, and generated proof readiness

Regenerate help/manual/website artifacts if their generators include connector commands, then run
the CLI help parity checklist for representative query, mutation, destructive mutation, and the
blocked `deleteIssue` command.  The following live-proof phase may use only generated PM commands
and the existing fail-closed lab boundary; it may not use this implementation checkpoint as live
acceptance evidence.

### Review hardening 31c — error and operation-name boundaries stay closed

Review must prove that a GraphQL HTTP error cannot bypass the bounded GraphQL error sanitizer and
that a `graphql_query` cannot hide an appended mutation selected through `operationName`. The
runtime must require exactly one named fixed operation of the declared kind before it resolves
configuration or sends a request. Static capability accounting must count fixed GraphQL queries as
reads even though their physical transport is POST.

## Verification

All commands in this checkpoint are local and hermetic.  No credential, `pm` provider invocation,
browser action, fixture creation, cleanup, or source-network request is authorized.

```bash
go test -timeout 20m ./internal/connectors/engine -run 'GraphQL|OperationEndpoint' -count=1
go test -timeout 20m ./internal/connectors/commandrunner -run 'GraphQL|StructuredJSON' -count=1
go test -timeout 20m ./cmd/connectorgen -run 'GitHub|GraphQL|CLISurface' -count=1
node --test scripts/tests/github-combined-operation-ledger.test.mjs \
  scripts/tests/gen-github-graphql-parity.test.mjs \
  scripts/tests/github-source-drift.test.mjs
node scripts/gen-github-graphql-parity.mjs --check
node scripts/github-combined-operation-ledger.mjs --check
make github-parity-artifacts-check
go run ./cmd/connectorgen validate internal/connectors/defs
go run ./cmd/connectorgen surface-sync --check
node --test scripts/tests/github-live-lab.test.mjs
node scripts/github-live-lab.mjs --check-boundary \
  --boundary .planning/phases/github-parity-extract-r1/GITHUB-LIVE-LAB-BOUNDARY.json
go build ./cmd/pm
./pm github graphql
./pm github graphql query viewer --help
./pm github graphql mutation create-enterprise-organization --help
./pm github graphql mutation delete-issue --help
./pm docs generate --dir docs/cli
(cd website && pnpm run gen:website-data && pnpm run test:scripts)
git diff --check
```

## GSD and skills record

This is the canonical single-worker continuation.  The installed GSD adapter was checked, its
`discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` prompts
were resolved/read inline, and `agentcontractgen check` passed before this plan.  The manual inline
fallback is required because the parent lane prohibits GSD role spawning.  Skills loaded for the
continuation: `golang-how-to`, `golang-cli`, `golang-graphql`, `golang-testing`,
`golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, and
`golang-structs-interfaces`.
