# CONTEXT — issue #3718 canonical delivery contract

## Phase mapping

GitHub sub-issue #3718 (Wave 1 of parent #3714) maps to this GSD phase. The phase is a foundation
slice in roadmap workstream 0, "GSD Runtime and Agent Enablement." Waves 2–7 depend on it.

## Locked decisions

- One canonical checked-in source defines `pm-delivery-worker` and the inheriting
  `pm-connector-worker` overlay. Harness-native Claude, Codex, and Pi projections belong to Waves
  2–4 and are not created here.
- One worker owns one job and the parent issue/branch/PR state inline. It does not spawn
  orchestrator, shepherd, planner, reviewer, verifier, or GSD roles. GitHub and GSD artifacts are
  the durable handoff.
- The installed GSD lifecycle is `discuss-phase` → `plan-phase --tdd` → `execute-phase` →
  `verify-work` (including `plan-phase --gaps` and `execute-phase --gaps-only`) → `code-review`.
  `programming-loop` is absent and must not be mandatory.
- The repo stacked-PR contract owns GitHub topology. A deliberate parent seed and draft parent PR
  must exist before production changes. GSD ship is not used for PR creation.
- no-mistakes v1.41.2 has no non-default-base option. Child branches run
  `--skip=push,pr,ci`, then use `gh-axi` for the sub-PR; the integrated parent runs the full
  pipeline. `--yes` is forbidden.
- Away mode never expands authority. Routine reversible decisions fixed by the issue, repo
  contract, or standing authority may be answered against the exact finding with rationale;
  bounded in-scope findings may be auto-fixed and rechecked. Product ambiguity, destructive or
  irreversible work, secrets/auth/security boundaries, dependencies/production impact, generic
  writes, reverse-ETL execution, gate weakening, and final merge pause for the captain.
- Wayfinder remains rejected as a dependency. Only its navigable parent index, outcome-sized child
  issues, and explicit dependency/decision modeling are adopted.

## Implementation boundary

- Add a deterministic repo-native Go generator/checker around a versioned JSON contract.
- Projections use a generated block inside harness-owned wrapper files. Wave 1 records the six
  future paths as optional; later waves make their two files required after adding native wrappers.
- The check validates source invariants, exact generated blocks, overlay inheritance, single-worker
  delegation policy, and installed GSD command resolution.
- Wire the check into `make verify` without touching harness-native agent files.
- Reframe the existing parent-orchestrator contract around single-worker parent ownership while
  preserving child check/review gates and explicit captain approval for final merge.

## Scope fences

- Do not create `.claude/agents/`, `.codex/agents/`, or `.pi/agents/` projections.
- Do not delete legacy project roles or any project-local GSD adapter resources.
- Do not touch global home-directory configuration.
- Do not touch connector engine write/schema or connector validator/conventions lane files named in
  the issue.

## GSD execution note

Commands used: `scripts/gsd prompt discuss-phase issue-3718-canonical-delivery-contract --auto`
and `scripts/gsd prompt plan-phase issue-3718-canonical-delivery-contract --tdd --skip-research`.
The repo-local adapter was healthy. The work is executed inline because the issue explicitly
forbids spawned roles; this is the adapter's documented inline/manual fallback and does not weaken
TDD, verification, or review gates.
