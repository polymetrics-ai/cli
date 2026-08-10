# #3897 Code Review Record

**Mode:** Manual deep review via the resolved `code-review` GSD prompt under
the documented inline/manual-GSD fallback.

**Result:** PASS — correction 1 / 5 verified locally.

## Reviewed risks

- Selected query views cannot include a different connection's Parquet file.
- Selected action rows use the existing safe table lookup rather than raw SQL
  identifier concatenation.
- Generic SQL remains typed ambiguous when duplicate bare names are queried;
  only the flow boundary supplies an actionable manifest remedy.
- `_unattributed` bypasses the connection registry deliberately and filters to
  root ownership only.
- Error wrapping preserves `errors.As` for `*warehouse.AmbiguousTableError`.
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

## Baseline observations

The repository-wide direct Go lint command reports pre-existing diagnostics
outside this diff; changed-diff lint against the parent is clean and `make
lint` passes. Website lint reports existing unrelated warnings; website docs
generation and type-check pass. Neither observation is an #3897 defect or a
correction round.
