# Issue #4292 — captain pre-merge operation-evidence gate

Status: **blocked-common-foundations**.

No batch 8–10 provider-defined operation can be presented as merge-ready until
one operation-level machine projection proves all of these cells, separately
for ETL, reverse ETL, direct read, direct write, binary download, and binary
upload:

1. Provider source URL, version, byte/hash pin, and source location.
2. Canonical connector mapping.
3. Enabled runtime reachability.
4. Generated connector CLI command.
5. Generated website row.
6. Executable fixture or conformance evidence.

`N/A` is admissible only where the pinned provider evidence proves the
capability absent. Connector scope/tier, destructive classification, elevated
scope, and safety rules remain typed runtime metadata, confirmation, and
authorization requirements; they are never an operation-disablement reason.

The current seven-surface ledger records source/canonical disposition and every
typed action's CLI status, but it does not yet have the common operation-level
runtime/website/conformance projection. That omission is explicit in
`SEVEN-SURFACE-LEDGER.json.pre_merge_gate` and is a hard pre-merge blocker.

`traces/generate-foundation-gap-ledger.mjs` materializes the shared gap fan-out
in `FOUNDATION-GAP-LEDGER.json`: every provider operation receives a stable row
with exact source trace, canonical class, affected surfaces, owner, status, and
closure verification. The gap catalog deduplicates the provider-neutral
capability; batch and portfolio rollups make the non-merge-ready fan-out
auditable. Connectors whose provider operation inventory is unavailable or
dynamic-instance-dependent receive a separate shared inventory gap row rather
than silently disappearing from portfolio coverage.
