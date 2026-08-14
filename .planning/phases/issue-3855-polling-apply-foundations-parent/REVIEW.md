# Code review — #3855 parent topology scaffold

**Mode:** standard-depth inline/manual GSD fallback
**Reason:** the canonical single-worker contract forbids spawning `gsd-code-reviewer`.

## Reviewed scope

- all files in this phase directory;
- branch/base/retarget assertions in `RUN-STATE.json` and `PLAN.md`;
- issue links, draft-only wording, child order, and #3880 reuse wording in `PR-BODY.md`;
- changed-path and secret-free scope.

## Findings

| Severity | Finding | Disposition |
| --- | --- | --- |
| Critical | none | n/a |
| Warning | Historical recovery text prescribed an explicit-lease force push and did not distinguish the active serial programme schedule from the logical issue graph. | Corrected in the bounded nine-file GSD gap checkpoint; no product scope widened. |
| Info | none | n/a |

## Result

The seed is bounded to planning artifacts. It preserves `#3856 -> #3857 -> (#3858 || #3859)`,
records #3860 as a follow-on documentation child, uses the exact temporary #3862 base and final
#4015 retarget rule, records #3880 as partial reuse only, and does not claim certification or
executable product behavior. Draft PR #4041 was inspected through `gh-axi` REST as open/draft with
the exact temporary base and #3855 head. The terminal inherited-document edit was whole-reverted,
leaving the #4015 architecture blob protected and the net range phase-only. Fresh validation of
this evidence commit remains pending; the draft is not an authorization to merge.

## 2026-08-12 topology-refresh review

The fresh parent head `c67f40a5ff67a131950f3123e70527027dca8493` contains accepted transport PR
#4059 and was absent from the recovered #3855 ancestry. The selected `--onto` rebase uses the exact
old parent base `30b2fb4aeb121641b6158903fe1d3b54668599a6`; all eight replayed commits are
patch-equivalent by `git range-diff`, with no conflicts and no paths outside this phase directory.
This is the smallest safe refresh that makes the accepted seam available to future child #3856.

No product, test, connector, PostgreSQL, or inherited #4015 document path is introduced. Fresh
no-mistakes and final live PR inspection remain required before the recovery is handed off.

## Audited bridge review

The reproduced pre-bridge rejection is a custody topology condition, not a product defect: the
fresh run could not start because no-mistakes' internal ref remained at `b61d0fa7...` while the
accepted transport refresh is `e541170e...`. The selected one non-force `--no-ff -s ours` bridge
passed at `72a4fc32...` after fixed-SHA rechecks: it retains checkpoint `3b6e3e2e...` byte-for-byte,
has the prescribed parents, carries remote/preserved/transport ancestry, and changes no non-phase
or protected inherited path. The inline GSD verification/review and planning-only local gates also
passed. Fresh no-mistakes review remains pending; the draft is never an authorization to merge.
