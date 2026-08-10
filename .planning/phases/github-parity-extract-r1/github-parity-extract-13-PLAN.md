---
phase: github-parity-extract-r1
plan: "13"
type: tdd
wave: 13
depends_on: ["11", "12"]
autonomous: true
files_modified:
  - internal/connectors/engine/schema/api_surface.schema.json
  - internal/connectors/engine/*.go
  - internal/connectors/commandrunner/*.go
  - cmd/connectorgen/*.go
  - scripts/gen-github-parity.py
  - scripts/gen-github-graphql-parity.mjs
  - scripts/tests/*.mjs
  - internal/connectors/defs/github/{api_surface,operations,writes,cli_surface}.json
  - internal/connectors/defs/operation_endpoint_ledger.json
  - .planning/phases/github-parity-extract-r1/*
---

# Plan — complete GitHub REST + GraphQL parity closure

## Captain outcome

Make every declared GitHub REST and generated GraphQL operation executable through a fixed,
typed PM command, with no `unsafe_or_disallowed` rows for documented operations.  `auth token`
and `api` are instead explicit `unsupported_local` safety boundaries: one would disclose a
credential and the other would create an unrestricted authenticated-request escape hatch.  They
have no operation, write, or API-surface binding and therefore remain non-executable until
separately authorized.  No raw API, user-supplied GraphQL document, or weakened
approval/confirmation flow is permitted.

This continuation also completes the target-bound PM-only live cohort for immutable repository
`1327549621`, preserving the authenticated-repository discovery Error-envelope defect as a
first-class finding.  Provider writes remain paused until local safety, deployment-key, and
read-back gates are green for each affected slice.

## Reconciled source facts at plan start

The captain's historical REST proof ledger remains a valid preserved artifact but is not the
plan-start all-source denominator:

| artifact | mechanically derived fact |
| --- | --- |
| `OPERATION-PROOF-LEDGER.json` / `COMMAND-PROOF-LEDGER.json` | legacy REST binding: 1,224 endpoints; 1,179 commands; 1,081 implemented; 37 partial; 77 endpoint rows blocked |
| `api_surface.json` | 1,220 pinned REST rows plus 4 legacy GraphQL bindings; 1,147 covered REST rows and 77 blocked rows |
| `cli_surface.json` | 1,484 current commands: 1,381 implemented, 37 partial, 8 planned, 10 unsafe/disallowed, 27 unsupported-api, 21 unsupported-local |
| `GITHUB-COMBINED-OPERATION-LEDGER.json` | 1,525 authoritative source operations: 1,220 REST + 31 GraphQL query + 274 GraphQL mutation; 1,345 classified implemented |

The stated historical `143` is `1224 - 1081`; it is not a mutually exclusive current command
partition (`45 + 37 + 8 = 90`).  This plan therefore makes every closure claim from generated,
source-derived ledgers rather than preserving a contradictory hand count.  The first executable
test locks this reconciliation and fails on any unresolved current classification.

## Current-ref — merged GitHub surface

The authoritative source-derived count and its provenance are in
[VERIFICATION.md](VERIFICATION.md). This plan does not duplicate that generated-surface
measurement.

The plan-start table remains historical evidence preserved at base ref
`4df0b0416e46958d9acb1b02708464570c070e0f` on 2026-08-10.

## Scope and safety boundary

- The target connector is exactly `github`.  Shared changes must be declaration-bound and pass
  a non-GitHub regression; no shared production branch may name GitHub.
- REST blocked rows are converted only to fixed operations with closed parameter/body contracts
  derived from the pinned `github/rest-api-description` commit.  GETs receive bounded direct-read
  contracts; writes retain plan → preview → approval → execute and typed destructive confirmation.
- Root union bodies remain explicit named actions.  The two `anyOf` custom-pattern updates use the
  existing documented at-least-one object model; no raw request-body escape hatch is added.
- A source operation that cannot become executable must have a schema-validated named dependency
  that names the exact missing capability.  The captain's build-inline rule means ordinary
  foundations are implemented here, not silently deferred.  Credential-minting/token-printing and
  raw-API safety boundaries remain `unsupported_local`, never `unsafe_or_disallowed` or
  mislabelled `implemented`.
- All generic ETL/reverse-ETL routes remain valid proof routes.  A direct connector command is
  added only for an actual documented or existing PM alias contract.

## TDD slices

### Red 33a / Green 33a — terminal parity inventory and named dependencies

Add a source-derived test that fails while any documented REST/GraphQL operation lacks an exact
command or a schema-valid named dependency, while any documented command remains `partial`,
`planned`, or `unsafe_or_disallowed`.  Add the generic
schema/validator representation for a named dependency and test that an unnamed or unresolved
dependency cannot pass the inventory.  Regenerate operation/command proof ledgers from the same
source facts.

### Red 33b / Green 33b — all remaining REST endpoint contracts

Generate exact contracts from the pinned REST artifact for the 77 historical blocked rows.  Keep
the installation access-token mint endpoint named as the held credential-output dependency; make
the remaining typed GET, status, text, write, `anyOf`, and compatibility routes executable with
bounded response policies, redaction, declared fields, and existing write gates.  Add provider
double tests that assert method, path, query, body, output policy, and zero dispatch for rejected
input.

### Red 33c / Green 33c — close partial/planned/unsafe command aliases

Promote structured object/array write fields through the existing declaration-bound JSON-field
preflight, not a generic request body.  Bind legacy aliases (`issue/pr/release/workflow/run/ruleset
view`, project/discussion creation, issue transfer/delete, PR revert, and supported status/search
operations) to exact REST or fixed GraphQL contracts.  Implement an environment-only typed GraphQL
secret-input channel if required by generated sensitive mutations; it must never put secret JSON
in argv, plans, previews, transcripts, or evidence.  Retain `auth token` and `api` only as
non-executable `unsupported_local` safety boundaries.

### Green 33d — generation, reachability, and live evidence

Regenerate source-owned GitHub bundle, operation ledger, manuals/catalogs, and golden transcripts.
Run the newly built binary over every command in its own initialized project; no declared command
may return `unknown command`.  Then execute the PM-only current-head lab cohorts against only
repository ID `1327549621`, retain the safe PM Error envelope as a defect finding, and record one
terminal result per implemented command.  Do not treat help, fixtures, preflight, or history as
live proof.

## Required verification

- Focused RED and GREEN tests for each slice, including non-GitHub coverage for shared structured
  JSON/named-dependency behavior.
- `go run ./cmd/connectorgen validate internal/connectors/defs`
- `go run ./cmd/connectorgen surface-sync --check`
- source-owned GitHub generator/ledger tests and `make github-parity-artifacts-check`
- `go test -timeout 20m` for changed engine, commandrunner, connectorgen, GitHub, app, and CLI
  packages; `go vet ./...`; `go build ./cmd/pm`.
- CLI parity: `pm github`, representative help for direct read, safe write, destructive write,
  GraphQL mutation, named dependency, and the two held aliases; regenerated docs/website checks.
- PM-only boundary/lab tests and an isolated binary reachability sweep.

## GSD and skill record

The installed GSD adapter was checked with `scripts/gsd doctor`; the `discuss-phase`,
`plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` prompts were resolved, and
`go run ./cmd/agentcontractgen check` passed.  The canonical single-worker contract prohibits
role spawning, so this is the recorded inline/manual fallback.  Loaded skills: `golang-how-to`,
`golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`,
`golang-design-patterns`, `golang-structs-interfaces`, `golang-graphql`,
`golang-documentation`, and `javascript-testing-patterns`.
