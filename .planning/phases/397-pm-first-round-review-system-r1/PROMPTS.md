# Issue #397 PM First-Round Review System Prompt Trace

## Kickoff snapshot

- User objective: implement the captain-authorized audit-backed PM first-round review system after PR #495 integration.
- Exact parent base: `0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20`.
- Branch: `chore/pm-first-round-review-system-r1`.
- Delivery: separate stacked PR to `feat/cli-architecture-v2`, `Refs #397`, no merge.
- GSD command: `scripts/gsd prompt plan-phase 397-pm-first-round-review-system-r1 --skip-research`.
- Programming-loop result: unavailable in registry; active `/pm-orchestrate` lifecycle owner recorded.
- Required RED: two accepted PR #495 findings plus three original preventable misses.
- Review route: one synthesized fresh-context local-Codex verdict, then independent Shepherd, then human authority.
- Safety: no secrets, dependencies, credentialed checks, raw generic write tools, reverse ETL execution, Claude, Copilot, or parent/main merge.
- Downstream artifacts: `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `RUN-STATE.json`, `SETUP-EVIDENCE.md`.
- Plan-check result: initial BLOCKED; gaps corrected before RED.
- RED result: focused semantic test exited 1 for intended classifier, disposition, mutation, threshold, one-way migration, and append-only-history failures.
- GREEN result: focused semantic/compiler/synthesis/measurement tests, canonical PM contract, model routing, shell syntax, Shellcheck, JSON, YAML syntax, and diff checks passed.
- Historical verification result: full local gates and `make verify` passed at `7f1b2d8fe12157b7bea7d6d57553c2aa2b4fe839`; captain's later graph/lab requirements invalidate this as delivery evidence.

## Captain correction and research snapshot

- Requirements read in full: `impact-graph-correction.md`, `counterfactual-review-lab-requirement.md`, and `impact-graph-algorithm-research-requirement.md`.
- GSD command: `scripts/gsd prompt plan-phase 397-pm-first-round-review-system-r1 --research`; read-only scout/security attempts produced no usable result due WebSocket/provider authentication, so inline fallback is recorded.
- Research stop respected: no production graph/lab edits before `.planning/phases/397-pm-first-round-review-system-r1/ALGORITHM-RESEARCH.md`.
- Evidence-selected graph: typed directed multigraph, materialized forward/reverse adjacency, deterministic multi-source relation-policy BFS, three-valued certainty, authoritative `go list`, fail-closed graph/packet bounds, exact-head cold rebuild, no new dependency.
- Required next RED: separate frozen graph/lab correction corpus plus real integration failures for reverse leaf, both-direction script/package, authority, generator, Go test/importer/variant, cycle/bound, unknown, unrelated control, exact coverage, lab denial/limits/cleanup/identity/concurrency/hypothesis/migration.
- Same stable lineage and `0/5` budget continue; captain requirements do not reset the lineage.
- 2026-07-24 decision: Firstmate may conditionally merge the all-green stacked PR into `feat/cli-architecture-v2`; this agent remains no-merge and must report the green open PR. PR #438 into `main` remains separately human-gated.
- Post-correction verification: focused graph/lab/PM gates and all full repository gates including `make verify` passed at exact head `e4ca19ce864b6a3362a2d490aec2d0b6a3717b1f`; final evidence-commit exact-head rerun remains before packet review.
