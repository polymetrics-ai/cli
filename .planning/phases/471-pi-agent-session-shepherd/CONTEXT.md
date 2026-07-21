# Context: Autonomous In-Process Shepherd

Issue: https://github.com/polymetrics-ai/cli/issues/471
Parent PR: https://github.com/polymetrics-ai/cli/pull/472
Parent branch: `feat/471-pi-agent-session-shepherd`
Base: `main`
First consumer/canary: issue #397 and draft PR #438 (CLI Architecture v2)

## Authoritative decision

The standalone Go/tmux Shepherd is abandoned. Issues #372, #389, and #470 are closed as
`not_planned`; draft PRs #390 and #456 are closed unmerged. Their branches, worktrees, commits, and
discussion remain historical evidence and must not be deleted or transplanted wholesale.

Issue #471 replaces that program with a complete autonomous Shepherd implemented as a first-class
Pi extension. It must own the full parent-issue delivery loop, not merely observe or score another
driver. The existing read-only `AgentSession` code on the parent branch is a hardened control-plane
seed; it is neither the product boundary nor a supported fallback.

The legacy shell loops remain rollback-only until #481 proves the replacement. They do not define
the new architecture. #480 prepares and verifies the reversible cutover, but deprecation is
activated only by the parent-owned post-canary finalization after #481 passes; a failed canary must
leave the rollback path unchanged.

## Required outcome

Given one objective or parent issue, Shepherd must:

1. reconcile durable state with Git and GitHub truth;
2. research unknowns and create a GSD parent plan;
3. create dependency-linked sub-issues, the parent branch, and a draft parent PR;
4. select all ready non-colliding issues and dispatch isolated in-process Pi `AgentSession` workers;
5. enforce plan-first red-green-refactor, exact verification, review, disposition, and correction;
6. integrate eligible sub-PRs into the parent branch and maintain review-coverage evidence;
7. survive interruption by reconciling intent, worktrees, refs, issues, PRs, checks, reviews, and
   decisions before any new mutation;
8. request human decisions on the ideal issue/PR, wait durably, consume one authenticated answer,
   revalidate, and resume; and
9. merge the exact verified parent head only after a fresh explicit human `approve-merge` decision.

## Pi SDK boundary

- Installed Pi version: `0.80.6`.
- Use the public `createAgentSession` API with explicit resource loading and in-memory child session
  and settings managers.
- Do not spawn another `pi` process, use tmux as transport, or give child sessions recursive
  orchestration authority.
- Implementation/correction roles use `openai-codex/gpt-5.6-sol`/`high`; all planning, research,
  issue proposal, verification, review, disposition, and orchestration roles use
  `openai-codex/gpt-5.6-sol`/`xhigh`.
- Child sessions share the parent process and crash domain. Durability is persisted intent plus
  authoritative reconciliation, not process isolation.
- Mutating sessions require one issue, branch, isolated worktree, declared canonical write scope,
  PR base, and least-authority tools. Typed host adapters—not model shell commands—own external
  mutations.

## Security and durability boundary

- Never place secrets, auth values, prompts, reasoning, or unrestricted output in state, logs,
  comments, or child context. GitHub auth remains in the host/keychain boundary.
- The macOS state root is a private trusted same-user local directory. Without native `openat`/
  descriptor-relative operations, do not claim resistance to a hostile same-UID swap-and-restore
  attacker.
- Agent output is untrusted evidence. Deterministic host validation and exact Git/GitHub state
  govern every transition.
- Parent merge approval is bound to a durable request ID, allowlisted actor, PR, generation, and
  exact head SHA. Silence, emoji, CI, review prose, or a score is not approval.

## Workflow adapter status

The repository-local GSD adapter on this base does not expose `programming-loop`; the previously
recorded manual-GSD fallback remains active for #473 until the adapter is available. Every child
still records plan, RED, GREEN, refactor, verification, review, and handoff evidence. The parent
orchestrator contract and issue bodies #473-#481 are authoritative for scheduling.
