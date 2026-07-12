# Phase Worker Contract

Phase: 325-agentloop-characterization

- Parent-spawned worker: issue #325 in `wt-325-agentloop-characterization`.
- Mutation scope: exactly the paths listed in issue #325.
- Parent branch and PR #324 artifacts are read-only.
- Inner planner, backend, tester, security, reliability/observability, and reviewer passes are
  `local_critical_path` because they inspect one coupled Phase 0 slice in this isolated worktree.
- Read-only automated review occurs after the stacked PR is open; it is not pre-claimed here.
- Human gates: dependency/auth/ruleset changes, secrets, destructive/live provider actions,
  quality-gate reduction, any merge, and parent-to-main delivery.
