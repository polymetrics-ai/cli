# #4069 — Case-equivalent unique warehouse tables

**Parent issue:** #3897
**Specification owner:** #4066 (exhausted at correction 5 / 5)
**Fresh delivery lineage:** correction 0 / 5
**Branch:** `fix/4069-flow-case-equivalent-unique-tables-r1`
**Immutable starting head:** `659efd8a0d69f26b55fcbd3c02150e995c159519`

## Objective

Keep valid, connection-owned Parquet tables usable when their exact names are
different but DuckDB treats their identifiers as equivalent. The required
inventory is `acme/records` and `globex/RECORDS`, with distinct rows and one
exact owner for each spelling.

## Background

The immutable Sol audit of draft PR #4060 found that the existing #4066
per-query policy groups the resolver snapshot by exact Go strings. It therefore
tries to create both `records` and `RECORDS` views. DuckDB rejects the second
view before the SQL is evaluated, making even generic `SELECT 1` unavailable
and preventing an omitted flow from returning the typed warehouse ambiguity.

#4066 owns the connection-scoping contract but has exhausted its 5 / 5
correction budget. Captain disposition authorizes this issue as the one fresh
#3897 child and prohibits treating it as correction loop 6.

## Product requirements

1. Explicit `acme` and `globex` connection-scoped reads return only their own
   `records` and `RECORDS` rows respectively.
2. DuckDB case equivalence cannot make unrelated generic SQL such as
   `SELECT 1` fail during view registration.
3. An omitted-connection flow against either canonical-equivalent bare
   identifier fails through `*warehouse.AmbiguousTableError` and receives the
   existing manifest `connection` remedy, not a raw DuckDB catalog error.
4. The established #4066 regressions retain generated-owner-alias refusal,
   three-table case variants, generic availability, action selectors,
   reverse/read isolation, and schedule re-entry.

## Scope and exclusions

The implementation is limited to the immutable resolver-snapshot policy at the
app/DuckDB boundary and its real Parquet/DuckDB flow/CLI regression. It does
not add credentials, provider calls, connector certification claims, transport
registration, production warehouse wiring, new CLI syntax, or unrelated
planning cleanup.

## Sources

- #3897 and #4066
- Draft PR #4060 at the immutable starting head above
- `data/cli-github-4060-final-sol-audit-r1/report.md`
- `data/cli-github-4060-final-sol-audit-r1/case-equivalent-unique-table-disposition.md`
- `data/cli-connector-release-certification-r1/FINISH-AND-PARALLELIZATION-PLAN.md`

## Correction 1 / 5 — same-owner case-equivalent inventory

Independent Sol audit of exact head
`d9022359e7b7bc2f7eb262c16177b52010681192` found an omitted partition of
this issue's warehouse-identifier contract: one local-warehouse connection can
declare distinct stream destinations `records` and `RECORDS`. DuckDB treats
them as one identifier, and a case-insensitive filesystem can address one
Parquet path for both spellings. The previous cross-owner policy correctly does
not cover that configuration.

The accepted policy is deliberately narrow: reject a newly created
local-warehouse connection with distinct ASCII-case-equivalent effective
destination tables after defaults and before save; keep legacy state open and
unchanged; refuse its next sync before `beginRun`, WAL, checkpoint, owner,
directory, temporary, or Parquet mutation; and return a new typed same-owner
collision error only when SQL references the irreducible identifier. Unrelated
SQL such as `SELECT 1` remains available. Exact direct/action/reverse reads
remain limited to the physical spelling the resolver can prove exists.

This is correction **1 / 5** in the fresh #4069 lineage, not a new issue and
not correction 6 on #4066 or #4060.
