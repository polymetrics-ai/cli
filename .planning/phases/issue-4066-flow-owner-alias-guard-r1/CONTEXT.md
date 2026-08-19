# #4066 — Guard generated owner aliases in unscoped flows

**Parent issue:** #3897
**Correction:** 5 / 5
**Starting head:** `c5b91917e3f5c07a010db2bdf58348cbc73cb9d5`
**Collision correction head:** `bda85b778f89f4320760b8d83826ac9d393b0220`
**Case-variant correction head:** `08acb08a8521ae7485152092810d4318ced29086`
**Bare-name correction head:** `4923b17648575d8947887139bb8058d2a5805a78`

## Problem

An unscoped flow query is supposed to leave a same-named warehouse table
ambiguous. The generic DuckDB query path correctly installs generated
`<table>__<connection-id>` views for interactive SQL, but a flow query shares
that path and can name one of those aliases directly. That bypasses the
resolver's typed ambiguity for the base table.

A legitimate catalog table can itself be named
`records__<connection-id>`. The original guard only intercepts generated
aliases after `registerViews` has already installed a unique real table as a
bare view, so an unscoped flow can use that colliding identifier instead of
getting the required `records` ambiguity.

DuckDB compares quoted and unquoted identifiers with ASCII case equivalence.
An uppercase real collision therefore evades the exact Go-string collision map,
allowing the same bypass and generic duplicate-view failure through a different
spelling of the generated alias.

The resolver still groups raw catalog names separately. With ambiguous
lowercase `records` and a unique uppercase `RECORDS`, an unscoped flow installs
the uppercase bare view and DuckDB resolves every case spelling through it,
bypassing the lowercase resolver ambiguity.

## Locked decisions

- Flow queries with an omitted connection must reject generated owner aliases
  through `*warehouse.AmbiguousTableError` for the base table and retain the
  existing flow `connection` remedy.
- The policy is derived once from the immutable `warehouse.TableResolver`
  snapshot and is carried at the app/DuckDB request boundary.
- The policy must not inspect, parse, regex-match, or rewrite SQL text.
- Generic `pm query run --sql` retains generated owner aliases and returns the
  selected owner's rows.
- Generated aliases are additive generic-query views, not a reserved table
  namespace. When one equals a real catalog table from the snapshot, generic
  query exposes the real table and does not register the generated alias.
- An unscoped flow suppresses a colliding real bare view before registration and
  routes that identifier to the ambiguous base table through the same snapshot.
- All snapshot-derived DuckDB policy keys use ASCII-only case canonicalization;
  real catalog names and typed resolver error identities retain their original
  spelling.
- An unscoped flow also suppresses every canonical bare-name variant of an
  ambiguous resolver table and maps replacement scanning to that original
  resolver name.
- Explicit flow connections, bare-name ambiguity, `_unattributed`, `SELECT 1`,
  and action-source reads retain their existing behavior.
- #4063's discovery metadata correction remains untouched.

## Delivery record

This review correction uses the repository's documented inline/manual GSD
fallback. The outer no-mistakes executor owns later validation, push, PR, and
CI phases; this phase records the RED/GREEN evidence and changes only the
affected source and focused regression.

Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
`golang-error-handling`, `golang-security`, `golang-safety`, `golang-lint`,
`golang-database`, `golang-stretchr-testify`, `golang-design-patterns`, and
`golang-structs-interfaces`.
