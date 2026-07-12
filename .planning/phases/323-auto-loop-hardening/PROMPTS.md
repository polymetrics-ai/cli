# Prompt Snapshots: Autonomous Delivery Control-Plane Hardening

## Parent kickoff

- Role: active parent orchestrator
- Parent issue: #323
- Objective: implement phases 0-9 as one-primary-issue stacked PRs on a dedicated main-based parent
  branch while preserving the final human merge gate.
- Inputs: issue #323, `ANALYSIS-CODEX.md`, issue/parent contracts, universal runtime workflow,
  Shepherd/driver contracts, required skill routing.
- Execution decision: `local_critical_path` for parent scaffold and native child issue creation;
  three read-only architecture/test/control-design subagents were then spawned.
- Downstream artifacts: `.planning/phases/323-auto-loop-hardening/PLAN.md`, draft PR #324, fourteen
  native sub-issues #325-#338.
- Verification result: parent-scaffold JSON/diff checks passed; production verification pending.

Exact worker prompts and compact handoffs will be recorded per child issue. Safety gates, commands,
test output, and security warnings remain uncompressed.

## Phase 0 worker dispatch

- Issue: #325
- Worker: Codex `phase0_worker`
- Worktree: `/Users/karthiksivadas/Development/polymetrics-cli-agents/wt-325-agentloop-characterization`
- Branch/base: `fix/325-agentloop-characterization` -> `fix/323-auto-loop-hardening`
- Scope: sanitized fixture/replay/policy/redaction/safety core, replay/safety CLI, hard driver fuse,
  shell boundary tests, Makefile gate, and issue artifacts only.
- Explicit boundary: no historical commit cherry-pick, unmerged Pi runtime, connector change, live
  action, dependency, merge, or parent artifact mutation.
- Verification result: pending worker handoff.
