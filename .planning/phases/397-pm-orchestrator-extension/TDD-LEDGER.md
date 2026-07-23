# Issue #397 PM Orchestrator Extension TDD Ledger

Status: PLAN complete; RED pending
Manual route: PM-owned PLAN → RED → GREEN → REFACTOR → VERIFY → REVIEW → INTEGRATE because the 69-command GSD registry has no `programming-loop`.

| Risk | RED contract | GREEN target | Status |
|---|---|---|---|
| unavailable command remains normative | focused script rejects PM prompts/contracts that require `gsd-programming-loop` or `scripts/gsd prompt programming-loop` | `/pm-orchestrate` owns the equivalent lifecycle after registry discovery | pending |
| competing review owner | focused script rejects canonical PM files loading Claude/Copilot review routes | one exact-head local Codex workflow is required | pending |
| self-certified review | focused script requires Shepherd after local Codex and before integration | independent trajectory verdict gates integration | pending |
| role cannot bind exact SHA | focused script rejects `pm-reviewer` without read-only git/GitHub access and exact base/head output | reviewer has constrained `bash` for read-only identity/diff commands and emits dispositions | pending |
| stale state schema | focused script rejects review enums without `local_codex` and `shepherd` fields | forward records capture both gates; legacy aliases remain parseable | pending |
| PR #493 collision | changed-path comparison rejects any PR #493-owned path | extension remains path-disjoint from PR #493 | pending |
| historical falsification | scan/diff detects bulk edits under historical phase directories | only new extension and narrow current Wave 1 evidence change | pending |

## Evidence log

- 2026-07-23: fetched PR #493 exact head `e21e56339390c5e1946eb4cfaf276eb80a889f29`; recorded its 18-file ownership set before production edits.
- 2026-07-23: `scripts/gsd doctor` passed, registry contains 69 commands, `plan-phase`/`code-review` sources resolve, and `programming-loop` is absent.
- 2026-07-23: one read-only collision scout confirmed the safe canonical PM path; a separate inventory scout could not start because its isolated Pi provider lacked Codex authentication. No implementation evidence is claimed from the failed scout.
