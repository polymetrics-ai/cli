# Context — #3855 polling/apply foundations parent

**Phase:** `issue-3855-polling-apply-foundations-parent`
**Primary issue:** #3855 — shared polling/watermark and native apply foundations
**Phase type:** planning-only parent scaffold; no product implementation

## Locked delivery topology

- The parent branch is `feat/3855-polling-apply-foundations`.
- Its initial draft PR base is `feat/3862-any-to-any-transport` (transport parent #3862 / PR #4019).
- That base is a dependency record only. It does not authorize merging #3855 into #3862.
- Before final parent integration, retarget #3855 to
  `docs/4015-connector-release-certification` (#4015 / PR #4016) after the reviewed transport seam is present there.
- Never target or merge this parent to `main`.

## Child order and acceptance boundary

The only permitted child order is:

```text
#3856 -> #3857 -> (#3858 || #3859)
```

All four existing child issues remain open. This scaffold neither implements them nor treats a
historical parent-scoped change as their completion.

## Reuse ruling

Merged PR #3880 / commit `dc1a6a7171ca901c8dbaf8cd528f67b18e57d9bb` is partial reusable
polling implementation only. It is not closure for #3856, #3857, #3858, or #3859 and must not be
cherry-picked or reimplemented. The topology report remains the source for its remaining gaps.

## Scope fence

This phase may add only the bounded parent planning/acceptance ledger and the associated draft PR.
It must not implement #3856–#3859, #3864, PostgreSQL, connector runtime behavior, or a product
capability claim. No `cmd/` or `internal/` path is in scope.

## Discussion record

The supplied brief fixes the branch, temporary base, permanent target, draft state, child DAG,
reuse classification, and non-goals. There are no remaining choices needed to create this seed, so
the generated `discuss-phase` workflow is executed inline without reopening those decisions.

The polling capability-taxonomy conflict identified in the topology report is deliberately deferred
to #3857 planning; this parent scaffold does not resolve it or claim executable behavior.

## Live GitHub note

`gh-axi` live issue/PR reads were attempted before mutation and returned `RATE_LIMITED`. No raw
GitHub CLI, alternate identity, or auth-scope change was used. The supplied current topology report
and locally verified remote refs are recorded provisionally; final live PR validation is retried
through `gh-axi` after local gates.
