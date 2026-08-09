---
phase: github-parity-extract-r1
plan: "08"
type: tdd
wave: 8
depends_on: ["07"]
autonomous: true
files_modified:
  - scripts/github-combined-operation-ledger.mjs
  - scripts/tests/github-combined-operation-ledger.test.mjs
  - scripts/testdata/github-combined-operation-ledger/**
  - internal/connectors/defs/github/sources/github-operation-source-lock.json
  - .planning/phases/github-parity-extract-r1/GITHUB-COMBINED-OPERATION-LEDGER.json
  - cmd/connectorgen/github_documented_surface_test.go
  - cmd/connectorgen/github_api_surface_test.go
  - scripts/github-parity-proof.mjs
  - scripts/tests/github-parity-proof.test.mjs
  - .planning/phases/github-parity-extract-r1/TDD-LEDGER.md
---

# Plan — combined pinned GitHub REST + GraphQL operation inventory

## Goal

Replace the fixed `1220 REST + 4 GraphQL` denominator with one generated, versioned source lock
and operation ledger. The lock is derived from GitHub's official REST OpenAPI artifact at an
immutable source commit and GitHub Docs' public GraphQL schema artifact. It records only
non-secret schema metadata, source URL/commit/hash/capture date, and normalized root operations;
CI checks remain hermetic by reading the checked-in lock rather than fetching a provider source.

This slice is inventory and classification infrastructure, not a claim that every GraphQL root is
implemented or live-proven. The PM-only laboratory remains idle throughout it.

## Source contract

- REST source: `github/rest-api-description` commit
  `b26c240ded1c8b79cb0fb09dee4a21239061fa23`, bundled
  `descriptions/api.github.com/api.github.com.json`; the generated lock retains the exact SHA-256,
  byte count, OpenAPI/info version, and one stable row per documented HTTP method/path.
- GraphQL source: GitHub Docs public schema
  `https://docs.github.com/public/ghec/schema.docs.graphql`; the generated lock retains its capture
  date, SHA-256, byte count, and every top-level `Query` and `Mutation` field. It preserves the
  declaration line, normalized argument/return signature, deprecation and preview directives, but
  never a credential or introspection response.
- `createEnterpriseOrganization` is a mandatory mutation canary. Lock generation and validation
  fail if it is absent or lacks an explicit `not_implemented`/mapped classification in the combined
  ledger.

## TDD slices

### Red 28a — source parser and canary

Add deterministic miniature OpenAPI and SDL fixtures plus a Node test that requires:

1. method/path enumeration to exclude non-operation path keys and derive a stable REST ID;
2. SDL parsing to enumerate every root field despite multiline arguments/descriptions/directives;
3. stable GraphQL IDs `github.graphql.query.<field>` and
   `github.graphql.mutation.<field>`;
4. source hash/provenance fields and a hard failure when `createEnterpriseOrganization` is absent;
5. a generated row for every source operation with protocol, source location, PM mapping,
   implementation state, auth/entitlement/fixture/approval/read-back/cleanup fields, and an exact
   non-secret unblocker for `not_implemented` rows.

The test must fail before the importer exists. It uses only fixture files and no network/provider
request.

### Green 28a — hermetic source lock and combined ledger

Implement `scripts/github-combined-operation-ledger.mjs` with explicit input paths. Its write mode
generates the checked-in source lock and combined ledger from already-downloaded official artifacts;
its check mode rebuilds from the lock and rejects source/ledger drift without a network request.
It compares the lock's REST method/path set to GitHub's declared REST surface so no provider REST
operation silently disappears. Existing fixed GraphQL documents are mapped by their actual root
selection only; all other root fields receive factual `not_implemented` rows rather than a synthetic
`UNTESTABLE` label.

### Red 28b — remove the fixed GraphQL denominator

Make the existing GitHub source-surface/proof tests derive REST totals, method split, and GraphQL
root totals from the lock. They must fail if a test still relies on the literal four-operation
GraphQL denominator or if a fixed document binding names a root field absent from the schema lock.

### Green 28b — source-derived verification

Update the source-surface and proof validators to consume the lock/combined ledger. The legacy four
fixed GraphQL bundle bindings remain explicit mappings, not a completeness denominator. Generated
output must keep REST, GraphQL query, and GraphQL mutation totals distinct and must fail closed for
an absent canary, duplicate stable ID, source-hash mismatch, or source operation without a row.

## Safety and verification

- No provider write, PM provider call, browser action, `gh`, raw API fixture call, or no-mistakes
  run is permitted in this phase.
- Required skills recorded for this plan: `golang-how-to`, `golang-graphql`, `golang-testing`,
  `golang-error-handling`, `golang-security`, and `golang-safety`.
- Inline/manual GSD execution is the documented fallback: the parent delivery contract prohibits
  role spawning. `scripts/gsd doctor`, command-source resolution, and `agentcontractgen check`
  were completed before this plan.
- Local gates before moving to GraphQL runtime work: focused Node importer tests, source-lock check,
  `go test -timeout 20m ./cmd/connectorgen`, `go test -timeout 20m ./internal/connectors/engine`,
  `go run ./cmd/connectorgen validate internal/connectors/defs`, `surface-sync --check`, and
  `git diff --check`.
