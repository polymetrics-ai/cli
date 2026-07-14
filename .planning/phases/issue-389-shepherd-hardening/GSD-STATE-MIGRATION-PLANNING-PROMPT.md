# GSD and Shepherd State Migration — Planning Prompt

Act as the official GSD Pi architecture planner for Polymetrics issue #389, stacked under parent
issue #372 and parent PR #390. This is a read-only planning turn. Do not edit files, create commits,
push, mutate GitHub, or expose secrets.

## Objective

Design the complete, safe sequence that lets us:

1. finish the official GSD Pi project/workflow implementation;
2. finish the standalone Go Shepherd supervisor implementation;
3. make Shepherd the authoritative controller for issue execution and recovery;
4. keep official GSD Pi as the owner of workflow state;
5. make per-run `.gsd/` and `.planning/` state local and untracked without breaking the repository's
   GSD command adapter, agent contracts, TDD evidence, or issue-to-PR workflow.

## Current facts to verify

- Official GSD Pi v1.11.0 runs through `gsd` and persists project workflow state in `.gsd/gsd.db`.
- Shepherd invokes official GSD Pi through `gsd headless`, with an issue worktree as
  `GSD_PROJECT_ROOT`, a controlled `GSD_HOME`, and protected controller state outside the worktree.
- `scripts/gsd` is a separate repository prompt adapter. It currently reads tracked resources from
  `.gsd/commands.json`, `.gsd/local-commands.json`, `.gsd/prompts/`, `.gsd/official-docs/`, and
  `.gsd/upstream.lock.json`.
- `.planning/` currently mixes repository-wide legacy planning, issue plans, TDD ledgers, prompts,
  and verification evidence. Repository contracts still name these paths.
- The issue #389 worktree contains intentional unfinished RED/GREEN code. Do not recommend deleting,
  resetting, or overwriting it.
- The target runtime is local host GSD Pi; do not add or restore a Podman dependency.
- No new dependencies are authorized.

## Decisions the plan must make

1. Define the exact source-of-truth hierarchy for Git/GitHub, official GSD workflow state, Shepherd
   controller authority, issue contract, tests, and audit evidence.
2. Decide which tracked `.gsd/` adapter assets move to a stable repository-owned path and provide an
   exact old-path to new-path mapping.
3. Decide what happens to existing tracked `.planning/` history: tracked archive, generated local
   projection, or another explicit category. Do not silently discard durable evidence.
4. Define the final `.gitignore` rules and explain how already tracked files are migrated safely.
5. Decide whether this directory migration belongs in issue #389 or requires a separate sub-issue
   under #372. Preserve one-primary-issue-per-PR scope.
6. Define bootstrap, adoption, restart, collision, rollback, and upgrade behavior for one fresh GSD
   project per issue worktree.
7. Define how `scripts/gsd`, `.pi/extensions/gsd`, Shepherd, tests, docs, and AGENTS contracts locate
   resources after the migration without relying on ignored runtime files.
8. Define tests that fail before implementation and prove no cross-issue state reuse, no adapter
   resource loss, no untracked-state merge, correct model routing, and restart-safe controller
   ownership.
9. Define an incremental migration order that keeps at least one working GSD command path at every
   checkpoint.

## Required output

Return a concise but complete architecture plan with these sections:

1. `VERDICT`: feasible as stated, feasible with corrections, or not feasible.
2. `SOURCE OF TRUTH`: an ownership table.
3. `TARGET TREE`: exact target directory structure.
4. `ISSUE SPLIT`: issue/PR boundaries and dependency order.
5. `MIGRATION`: ordered checkpoints, exact file/path changes, compatibility bridge, and rollback.
6. `TDD`: RED/GREEN/refactor tests for every behavior-changing checkpoint.
7. `VERIFICATION`: exact local commands and acceptance evidence.
8. `RISKS AND HUMAN GATES`: especially deletion, broad rewrites, auth, dependencies, and main merge.
9. `IMPLEMENTATION PROMPTS`: one bounded GPT-5.5/high prompt per executable issue, followed by one
   GPT-5.6 Sol/high independent validation prompt.

The plan must distinguish controller source of truth from workflow source of truth. Shepherd may be
the sole controller, but must not claim ownership of GSD's private workflow database or Git/GitHub
delivery facts.
