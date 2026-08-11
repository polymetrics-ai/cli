# #3897 Code Review Record

**Mode:** Manual deep review via the resolved `code-review` GSD prompt under
the documented inline/manual-GSD fallback.

**Result:** PASS — correction 3 / 5 verified locally.

## Reviewed risks

- Selected query views cannot include a different connection's Parquet file.
- Selected action rows use the existing safe table lookup rather than raw SQL
  identifier concatenation.
- Generic SQL remains typed ambiguous when duplicate bare names are queried;
  only the flow boundary supplies an actionable manifest remedy.
- `_unattributed` bypasses the connection registry deliberately and filters to
  root ownership only.
- Error wrapping preserves `errors.As` for `*warehouse.AmbiguousTableError`.
- Fault-aware bare binding preserves `errors.As` for
  `*warehouse.FaultError` when damaged ownership records hide a duplicate.
- Parsed DuckDB replacement scans retain typed errors for quoted legal table
  names without SQL-name regex parsing.
- A query reuses one immutable warehouse inventory across all bare views and
  replacement scans, rather than rescanning per table name.
- Action source identity reaches the current runner boundary without adding
  dispatch, provider writes, or a new approval lifecycle.

## Correction 1 disposition — #4032

R1 was legitimate: the flow action adapter delegated an uncapped source
semantic to public `QueryTable(..., Limit: 0)`, whose intentional user-facing
default is 100. A selected source larger than that could be marked successful
after only its prefix reached the local action runner. The correction adds an
action-specific request with no limit and routes only the action adapter to
the internal uncapped warehouse read. It keeps connection selection,
`_unattributed`, typed ambiguity, and the public cap in their existing paths.
The new real-Parquet regression uses 101 selected rows and a local failed
runner to prove no success checkpoint is written on a failed complete dispatch.
Focused app/flow/CLI verification passed with all 101 selected rows at both
local runner boundaries.

## Correction 2 disposition — #4037

R3 and R4 were legitimate one-layer binding defects. `registerViews` used the
healthy `warehouse.Tables` subset to decide bare-view uniqueness, so a damaged
owner record could hide a competing table and permit an unscoped read. Its
missing-table regex also excluded legal warehouse names that require quoted
DuckDB identifiers. The correction asks `warehouse.FindTable` before every
bare view, quotes identifiers by doubling embedded quotes, and uses the pinned
replacement-scan table-name callback to retain the first original lookup error
with `%w`. Focused app verification proves hidden-owner `FaultError`, selected
healthy and unrelated reads, `_unattributed`, legal quoted names, omitted
typed ambiguity, aggregate reads, and cancellation. Fixtures only write local
warehouse data; no provider mutation or reverse-ETL behavior changed.

## Correction 3 disposition — #4040

R5 was legitimate: every bare view called `warehouse.FindTable`, which rescans
the full warehouse inventory. The correction extracts the same fail-closed
rules into `warehouse.TableResolver`, retains `FindTable` as its compatibility
wrapper, and builds one immutable resolver at `QuerySQL` entry. The resolver is
passed unchanged into view registration and the DuckDB replacement callback,
so typed `FaultError` and `AmbiguousTableError` behavior remains attached to
the same inventory. The RED two-table query built three snapshots; the focused
race matrix proves one snapshot for multi-table and replacement-scan paths,
alongside the inherited flow and warehouse behavior matrix.

## Baseline observations

The repository-wide direct Go lint command reports pre-existing diagnostics
outside this diff; changed-diff lint against the parent is clean and `make
lint` passes. Website lint reports existing unrelated warnings; website docs
generation and type-check pass. Neither observation is an #3897 defect or a
correction round.
