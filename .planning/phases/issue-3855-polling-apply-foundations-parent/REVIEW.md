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
| Warning | none | n/a |
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
