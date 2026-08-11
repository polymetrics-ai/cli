# #4066 — Guard generated owner aliases in unscoped flows

**Parent issue:** #3897  
**Correction:** 5 / 5  
**Starting head:** `c5b91917e3f5c07a010db2bdf58348cbc73cb9d5`

## Problem

An unscoped flow query is supposed to leave a same-named warehouse table
ambiguous. The generic DuckDB query path correctly installs generated
`<table>__<connection-id>` views for interactive SQL, but a flow query shares
that path and can name one of those aliases directly. That bypasses the
resolver's typed ambiguity for the base table.

## Locked decisions

- Flow queries with an omitted connection must reject generated owner aliases
  through `*warehouse.AmbiguousTableError` for the base table and retain the
  existing flow `connection` remedy.
- The policy is derived once from the immutable `warehouse.TableResolver`
  snapshot and is carried at the app/DuckDB request boundary.
- The policy must not inspect, parse, regex-match, or rewrite SQL text.
- Generic `pm query run --sql` retains generated owner aliases and returns the
  selected owner's rows.
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
`golang-database`, `golang-design-patterns`, and `golang-structs-interfaces`.
