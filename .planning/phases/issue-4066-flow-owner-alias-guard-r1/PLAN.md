---
issue: 4066
parent_issue: 3897
phase: issue-4066-flow-owner-alias-guard-r1
type: tdd
wave: 1
depends_on: []
files_modified:
  - .planning/phases/issue-4066-flow-owner-alias-guard-r1/
  - internal/app/types.go
  - internal/app/query_engine_duckdb.go
  - internal/cli/flow_cli.go
  - internal/cli/flow_cli_test.go
autonomous: true
requirements: []
---

# #4066 TDD Plan: fail closed for omitted flow aliases

## Manual-GSD execution record

The active phase is a no-mistakes review correction. The parent #3897 phase
already records the repository's manual-GSD fallback because its roadmap is an
archive rather than an active numbered phase. No role is spawned. This review
phase updates issue-specific context, plan, and TDD evidence before production
edits; the outer executor owns subsequent lifecycle gates.

Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
`golang-error-handling`, `golang-security`, `golang-safety`, `golang-lint`,
`golang-database`, `golang-design-patterns`, and `golang-structs-interfaces`.

## Slice 1 — RED: omitted flow selector bypasses ambiguity

Create two structurally owned real Parquet `records` tables for `acme` and
`globex`. Drive a flow query with no `connection` through the production flow
adapter using both unquoted and quoted generated
`records__<globex-connection-id>` aliases. Before the fix, each query succeeds
instead of returning `*warehouse.AmbiguousTableError` for `records` with the
existing flow `connection` remedy.

The same regression also invokes `pm query run --sql` using that exact alias;
it must keep returning the `globex` row. It records a failed flow result, no
successful step checkpoint, and no returned flow rows on the rejected paths.

## Slice 2 — GREEN: snapshot-backed flow alias policy

Add an internal query-origin policy to `app.QuerySQLRequest`; only the flow
adapter selects the flow origin. Build the policy from the one
`warehouse.TableResolver` snapshot already captured by the DuckDB query. For
an unscoped flow, suppress generated aliases during view registration and map
those exact generated identifiers back to the ambiguous base table in the
replacement-scan callback. Preserve the resolver's typed error by resolving
the base name through that same snapshot.

Do not interpret SQL text. Do not change generic query, manifest syntax,
action-source reads, connector metadata, or #4063 artifacts.

## Slice 3 — focused verification

Run one focused CLI race test command covering the new alias regression plus
the existing explicit selector, bare ambiguity, `_unattributed`, `SELECT 1`,
and action-source cases. Record the result in the ledger after all source
changes are complete.
