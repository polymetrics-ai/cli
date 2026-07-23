# Issue #397 PM Orchestrator Extension TDD Ledger

Status: correction round 2 RED planned; GREEN/VERIFY/re-review pending
Manual route: PM-owned PLAN → RED → GREEN → REFACTOR → VERIFY → REVIEW → INTEGRATE because the 69-command GSD registry has no `programming-loop`.

| Risk | RED contract | GREEN target | Status |
|---|---|---|---|
| unavailable command remains normative | focused script rejects PM prompts/contracts that require `gsd-programming-loop` or `scripts/gsd prompt programming-loop` | `/pm-orchestrate` owns the equivalent lifecycle after registry discovery | green |
| competing review owner | focused script rejects canonical PM files loading Claude/Copilot review routes | one exact-head local Codex workflow is required | green |
| self-certified review | focused script requires Shepherd after local Codex and before integration | independent trajectory verdict gates integration | green |
| role cannot bind exact SHA | focused script rejects `pm-reviewer` without read-only git/GitHub access and exact base/head output | reviewer has constrained `bash` for read-only identity/diff commands and emits dispositions | green |
| stale state schema | focused script rejects review enums without `local_codex` and `shepherd` fields | forward records capture both gates; legacy aliases remain parseable | green |
| PR #493 collision | changed-path comparison rejects any PR #493-owned path | extension remains path-disjoint from PR #493 | green |
| historical falsification | scan/diff detects bulk edits under historical phase directories | only new extension and narrow current Wave 1 evidence change | green so far; final diff check pending |
| authoritative PR #493 migration gate | focused script rejects #408 readiness unless durable trace, machine state, and summary require PR #493 integration after Wave 1 | #408 remains `not_spawned_dependency_blocked` until both Wave 1 and PR #493 routing migration integrate | red planned |
| transitive legacy review route | focused script rejects current PM references to generic bot-era handoff/disposition templates | PM-specific templates represent exact-head local Codex plus independent Shepherd only | red planned |
| correction counter split/reset | focused script rejects active `guards.correction_rounds` shape and requires replacement-head/cap fixture | canonical autonomous state persists stable exact-base/candidate-lineage counters and human block on exceed | red planned |

## Evidence log

- 2026-07-23: fetched PR #493 exact head `e21e56339390c5e1946eb4cfaf276eb80a889f29`; recorded its 18-file ownership set before production edits.
- 2026-07-23: `scripts/gsd doctor` passed, registry contains 69 commands, `plan-phase`/`code-review` sources resolve, and `programming-loop` is absent.
- 2026-07-23: one read-only collision scout confirmed the safe canonical PM path; a separate inventory scout could not start because its isolated Pi provider lacked Codex authentication. No implementation evidence is claimed from the failed scout.
- RED: `scripts/tests/pm-orchestrator-contract.sh` exited 1 before canonical guidance edits. It reported missing local Codex workflow/prompt, PM files still requiring Claude/Copilot and the unavailable command, no Shepherd route markers, stale review schema, and an exact-head reviewer without read-only git access.
- GREEN: the focused contract and existing Pi model-routing check pass. Canonical contracts/workflows/prompts now route PM work through one owner, fresh-context exact-head local Codex review, and independent Shepherd; the missing GSD command is registry-discovered rather than invented.
- REFACTOR: the existing durable stage machine remains unchanged; Shepherd validates the REVIEW transition instead of adding a second persistent stage. Legacy bot workflows/agent retain explicit migration pointers and historical records remain untouched.
- Focused safety: `make pi-model-routing-check pi-shepherd-guards-check`, JSON/YAML parse, no dependency delta, and PR #493 path-disjointness all pass.
- REVIEW at `3c88fc78062ba0a3437f79bc88c395286c228c65`: fresh-context local Codex returned five findings. Dispositions are recorded in `REVIEW-DISPOSITION.md`: adapter/migration-gate, schema compatibility, bounded correction, and PR body findings accepted in scope; the Gong operation metadata defect is deferred because the cited files are byte-identical to current main and product changes are forbidden here.
- Correction round 1: removed the unconditional unavailable command example, made PR #493 routing reconciliation a hard post-Wave1 dependency, added canonical pending/clean/blocked plus historical review fixtures, made Shepherd verdict conditional, preserved historical local Codex shapes, and persisted a default four-round correction budget with per-range counters.
- Captain-authorized F2 follow-up: created focused bug https://github.com/polymetrics-ai/cli/issues/497 from the pre-existing Gong evidence without changing product code in PR #495.
- REVIEW at `0665ad7aad1ec083f4bb0572a88ac1a38f417a35`: fresh-context exact-head re-review confirmed F2/F3/F5, then found F6 authoritative queue drift, F7 transitive bot-era templates, and F8 active autonomous counter drift. All three are accepted for correction round 2; no product code or PR #493-owned path is authorized.
- Correction round 2 RED: the extended `scripts/tests/pm-orchestrator-contract.sh` exited 1. It reported absent authoritative trace/run-state/summary PR #493 blockers, current PM dependency edges to both bot-era templates, absent PM-specific templates, and the active autonomous driver/prompt lacking canonical lineage-budget markers. The cap-lineage fixture already encoded replacement-head history and an exceeded human gate.
