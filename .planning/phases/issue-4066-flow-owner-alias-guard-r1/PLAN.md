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
`golang-database`, `golang-stretchr-testify`, `golang-design-patterns`, and
`golang-structs-interfaces`.

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

## Slice 4 — RED/GREEN: real-table collision remains structurally distinct

Create a third, legitimate Parquet table named
`records__<globex-connection-id>` while `records` remains owned by both `acme`
and `globex`. RED drives that identifier through an omitted-connection flow in
quoted and unquoted forms: the current bare-view registration succeeds against
the real table instead of returning the `records` ambiguity and remedy. The
same generic query controls must return the real table in both forms.

GREEN extends the immutable resolver-snapshot policy with generated-alias
collisions before either bare-view registration or replacement-scan routing.
For generic origin, skip only the generated view that collides and preserve the
real table. For an unscoped flow, suppress the colliding real bare view and
route its identifier to the ambiguous base table. Do not reserve names, rename
tables, or inspect SQL text.

## Slice 5 — RED/GREEN: ASCII-equivalent collisions

Create the same third table with an uppercase `RECORDS__<CONNECTION-ID>` name.
RED exercises lowercase and uppercase spellings in both quoted and unquoted
queries. The existing exact-string map lets an omitted flow read the real table
and makes generic query fail during duplicate view registration.

GREEN canonicalizes only DuckDB identifier keys derived from the immutable
resolver snapshot and the replacement-scan identifier with DuckDB's documented
ASCII case equivalence. The stored catalog name remains the exact resolver key,
so generic query exposes the real uppercase table while unscoped flow resolves
every case variant to the typed `records` ambiguity. Do not lowercase SQL text
or catalog table names.

## Slice 6 — RED/GREEN: case-equivalent bare-table ambiguity

Create a unique uppercase `RECORDS` table alongside the two owned lowercase
`records` tables. RED drives quoted and unquoted lowercase and uppercase bare
identifiers through an omitted flow. The raw-name grouping registers the
uppercase view, so all forms succeed instead of returning the `records` typed
ambiguity and remedy. Generic controls must retain access to the uppercase real
table.

GREEN records canonical bare-name keys only for unscoped flow origin, mapping
each key back to the original ambiguous resolver name. Suppress bare views for
that flow policy before registration and resolve replacement scans through the
stored original name. Keep generic origin, user table spelling, aliases, and
SQL text untouched.
