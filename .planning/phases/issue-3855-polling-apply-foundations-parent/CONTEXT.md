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

All five existing child issues remain open. #3860 is a follow-on documentation child: it depends
on #3856–#3859 and is not a parallel implementation lane in the core DAG above. This scaffold
neither implements any child nor treats a historical parent-scoped change as their completion.

## Reuse ruling

Merged PR #3880 / commit `dc1a6a7171ca901c8dbaf8cd528f67b18e57d9bb` is partial reusable
polling implementation only. It is not closure for #3856, #3857, #3858, or #3859 and must not be
cherry-picked or reimplemented. The topology report remains the source for its remaining gaps.

## Scope fence

This phase may add only the bounded parent planning/acceptance ledger and the associated draft PR.
It must not implement #3856–#3859, #3860 documentation, #3864, PostgreSQL, connector runtime
behavior, or a product capability claim. No `cmd/` or `internal/` path is in scope.

## Discussion record

The supplied brief fixes the branch, temporary base, permanent target, draft state, child DAG,
reuse classification, and non-goals. There are no remaining choices needed to create this seed, so
the generated `discuss-phase` workflow is executed inline without reopening those decisions.

The polling capability-taxonomy conflict identified in the topology report is deliberately deferred
to #3857 planning; this parent scaffold does not resolve it or claim executable behavior.

## Scope recovery and live GitHub note

The completed local no-mistakes run `01KZPY9EBYX84WZM11EN1F6C83` accepted the planning scaffold at
`b2d0f63029a34f8d647e57f34a747ad3983e1578`, then produced terminal commit
`81f37c2b2cf638bfd6b06b35393256232ea6e23d` with an out-of-scope inherited-document edit. Guarded
custody recovery retained that commit in ancestry and whole-revert
`4099335632a79e3c70ce20c004a5f63933171280` restored the accepted tree exactly.

`docs/architecture/github-postgres-warehouse-certification.md` is protected #4015 content. Its
required temporary-base blob is `889e4eddffabb76aa8be46a934c3e9abe0610f4c`; this phase must not
edit, remove, reformat, or correct it.

`gh-axi` REST initially verified five open children (#3856–#3860) and draft PR #4041, open from
`feat/3855-polling-apply-foundations` to temporary base
`feat/3862-any-to-any-transport` at recovered head
`4099335632a79e3c70ce20c004a5f63933171280`. No raw GitHub CLI, alternate identity, or auth-scope
change is used. Fresh no-mistakes validation will re-inspect the final evidence head.

## Accepted transport-seam refresh

On 2026-08-12, the live temporary base advanced from
`30b2fb4aeb121641b6158903fe1d3b54668599a6` to
`c67f40a5ff67a131950f3123e70527027dca8493` when transport child PR #4059 merged into PR #4019.
`git merge-base --is-ancestor c67f40a5... b61d0fa7...` returned false, so leaving this parent on
the old base would prevent future #3856 work from inheriting the accepted transport seam.

Before the refresh, guarded `no-mistakes axi sync --recover` returned custody of its completed
unpublished planning-only head `b61d0fa7eefc719c39593e44afbcb1b7a3f76613`. The safe replay
`git rebase --onto origin/feat/3862-any-to-any-transport 30b2fb4a...` completed without conflicts.
`git range-diff 30b2fb4a...b61d0fa7... c67f40a5...HEAD` maps all eight parent commits as
patch-equivalent, and the post-rebase PR-relative diff remains only this phase's nine files. This
is a topology-only recovery: it neither implements a child nor consumes a substantive correction
round. The PR remains draft, dependency-only, and unmerged.
