# TDD ledger — #3855 parent topology scaffold

This is a planning-only phase. “RED” and “GREEN” describe observable delivery-topology conditions,
not product behavior; no production or test code is permitted in this seed.

| ID | Required condition | Red evidence | Green evidence | Status at seed creation |
| --- | --- | --- | --- | --- |
| P1 | Requested parent branch has one precise remote-derived start point | `refs/heads/feat/3855-polling-apply-foundations` and its remote counterpart were absent; the supplied worktree was detached | Branch is created from `origin/feat/3862-any-to-any-transport` at `30b2fb4aeb121641b6158903fe1d3b54668599a6` | green locally |
| P2 | The parent inventories all children and retains the core DAG | No current parent acceptance ledger bound #3856 → #3857 → (#3858 \|\| #3859) plus the #3860 follow-on | `PLAN.md` names each existing child, prerequisite, parallel boundary, and #3860 follow-on dependency | green locally |
| P3 | Historical polling work cannot be mistaken for child completion | #3880 could be confused with child implementation because it is parent-scoped | Context, plan, and PR body name `dc1a6a7171ca901c8dbaf8cd528f67b18e57d9bb` as partial reusable work only for every child | green locally |
| P4 | A draft parent records dependency without unauthorized integration | No current draft parent PR/seed is available to represent the #3862 dependency | Draft PR #4041 is open/draft with the exact #3862 base and #3855 head; the body prohibits merge to #3862 or `main` and requires later retarget | green — fresh validation pending |
| P5 | The scaffold cannot smuggle product work | A parent documentation change could accidentally widen into connector/transport/PostgreSQL code | Whole-revert restored the accepted tree; changed-path and protected-blob checks permit only this phase directory | green — fresh validation pending |

## Execution evidence

- **Red:** read-only worktree and ref checks established P1’s missing branch state before creation.
- **Green:** the branch is now checked out at the verified #3862 remote head; this phase supplies
  the bounded acceptance ledger before any product edit.
- **Refactor:** not applicable. Any later source/test change belongs to an existing child or a new,
  narrowly scoped #3855 child, never to this scaffold.

The draft-PR green transition is complete: `gh-axi` REST inspected #4041 as the one open draft with
the exact temporary base/head. The accepted local tree was restored by whole-reverting terminal
commit `81f37c2b2cf638bfd6b06b35393256232ea6e23d`; this custody recovery does not consume a
correction round. Fresh validation remains pending for the evidence commit and must preserve the
protected inherited #4015 architecture blob.
