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

`gh-axi` REST verifies five open children (#3856–#3860) and draft PR #4041, open from
`feat/3855-polling-apply-foundations` at `4099335632a79e3c70ce20c004a5f63933171280` to temporary
base `feat/3862-any-to-any-transport`. No raw GitHub CLI, alternate identity, or auth-scope change
is used. Fresh no-mistakes validation remains pending for this evidence commit.
