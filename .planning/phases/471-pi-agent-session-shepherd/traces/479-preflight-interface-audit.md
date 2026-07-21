# Issue #479 preflight interface audit

Date: 2026-07-21
Parent head audited: `2a89142e1514de0ce1481164d99f21c7f89acadb`
Mode: read-only planning/review (`openai-codex/gpt-5.6-sol`, `xhigh`)

## Outcome

Issue #479 must be the deliberate shared integration slice. The original allowed scope could not
represent autonomous start/resume, durable child/review/human-gate truth, or safe sibling refresh
after the parent branch advances. Production wiring must not begin until #475 and #478 are clean
and integrated and the RED matrix below executes at behavior level.

## Owner and seam decisions

| Seam | Owner and decision |
|---|---|
| Runtime admits only one mutator globally | #475 correction replaces the singleton with bounded canonical workspace/scope leases. Disjoint issue workspaces may coexist; collisions remain denied. |
| Parser requires `--read-only` and caps the old canary surface | #479 owns `arguments.ts`; autonomous start/resume are writable controller actions while canary remains explicit read-only. |
| v1 state projects away scheduler/review/effect truth | #479 owns a validated v2 autonomous DTO/store path while retaining deterministic v1 canary handling. |
| Parent branch cannot be refreshed and stale children cannot rebase/reclaim | #479 narrowly extends the typed Git/workspace adapters. No generic model shell or generic Git command is exposed. |
| Claimed workspace is not a worker file capability | #479 builds an opaque scoped facade that asserts lease ownership before bounded no-follow reads, writes, and edits. |
| #478 defines ports but not production intake/GitHub/review implementations | #479 constructs and injects concrete typed adapters; orchestration policy stays in #478. |
| Human consume and resulting external effect are not atomic | #479 adds prepared/observed/consumed/applied effect journaling and restart reconciliation by exact idempotency key. |
| Review docs still named Claude/Copilot | Program-specific docs now record controller-owned independent Codex 5.6-sol/xhigh coverage only; exact-head human parent merge remains separate. |
| Top-level `read_only` child versus mandatory integration roster | Exact-head review reproduced the impossible lifecycle. #478 must reject top-level read-only child records. #479 keeps read-only research/review as internal roles and never invents integration evidence. |

Child Git mutation leases are already per issue, not globally serialized. #479 holds one outer
parent coordinator lease, uses existing per-child issue leases, and serializes only shared parent
topology/integration. After child A advances the parent, every stale sibling must refresh/rebase or
reclaim and repeat verification plus independent Codex review at its new exact head.

## Required controller composition

```text
extension parser
  -> AutonomousShepherdController
     -> v2 FileStateStore + parent coordinator lease
     -> ParentIntakePort
     -> #474 dependency graph / autonomy policy / reconciler
     -> #475 ShepherdAgentSessionRuntime
     -> #476 WorkspaceAdapter + GitAdapter -> ScopedWorkspaceFactory
     -> typed VerificationPort
     -> concrete #478 GitHubOrchestrationTransport -> GitHubParentOrchestrator
     -> GitHubDecisionBroker
     -> controller-owned CodexReviewPort + durable review repository
```

`start` and `resume` route to the autonomous controller. The existing bounded `ShepherdController`
remains only for the explicit read-only canary until #480 cutover.

## Durable v2 minimum

- Parent repository/issue/worktree identity, run ID, generation, revision, stage/status, timestamps,
  branch/PR/base/head, plan digest, final verification/review, and terminal blocker.
- Per child: dependencies, access, canonical scopes, stable ownership ID, branch/base/head, stage,
  attempt/correction counts, verification, PR, exact review record, dispositions, receipt, and wait.
- Prepared/applied external-effect ledger, bounded retry budgets, pending human request and exact
  binding/head/effect key.
- Structured bounded facts only. Never persist prompts, reasoning, raw model output, secrets,
  credentials, or unrestricted logs.

Resume acquires the parent lease, increments generation, reconciles disk/Git/GitHub/review/decision
truth, and only then schedules. Late results from older generations are discarded.

## Behavior-level RED matrix

1. Full intake -> plan -> parallel children -> PR -> Codex review -> correction -> integration ->
   exact-head human wait; no external shell driver.
2. Dependency ordering, canonical scope collisions, concurrency cap, and deterministic idle reasons.
3. Two disjoint mutating issue workspaces coexist; same workspace/scope collision is denied and each
   lease releases only itself.
4. Child A advances parent; child B refreshes/rebases or reclaims, then repeats exact verification
   and review before integration.
5. Per-stage retry/correction budgets, exhaustion, durable wait, and no score/prose success path.
6. Crash at prepare, publish, observe, consume, apply, and persist checkpoints; restart applies one
   exact effect without duplicate comments, commits, pushes, PRs, integration, or decisions.
7. Stale-head, unauthorized, bot, edited, duplicate, ambiguous, and consumed human replies.
8. Stop during intake, workspace setup, AgentSession creation/run, verification, GitHub mutation,
   decision polling, and backoff; join all accepted work before releasing leases or persisting stop.
9. Stop/shutdown race, late-generation completion fencing, and persistence-failure sibling abort.
10. Commit/push/PR/integration timeout after publication; authoritative reconciliation prevents a
    second external mutation.
11. Stable ownership IDs across resume; changed owner/base/scope fails before mutation unless the
    typed refresh/reclaim transition was durably prepared.
12. Codex findings/dispositions, clean exact-head record, head movement during/after review, and
    mandatory fresh stable-head rerun.
13. Dirty or scope-escaped handoff, wrong branch/base/head, draft PR, incomplete/untrusted CI, and
    review prose all fail closed.
14. Final parent verification/review/human request invalidates on head movement; controller exposes
    no parent-to-main merge action and completes only after observing a human merge.
15. Path traversal, symlink escape, terminal controls, proxy/accessor/cycle/oversize payloads,
    bounded verification argv/output/timeout, and cancellation propagation.
16. Bare/help/invalid/status behavior and stop during unresolved extension initialization.
17. Top-level read-only child records fail during plan validation; internal read-only roles never
    enter the integration roster or fabricate receipts.

## Checkpoint sequence

1. Plan/TDD ledger and compiling fake-port scaffold.
2. One behavior-level RED checkpoint covering the full matrix above.
3. v2 state/effect ledger and autonomous parser/help/status.
4. Scoped workspace, parent refresh/rebase, verification, intake/GitHub, and Codex review adapters.
5. Scheduler plus concurrent child lifecycle.
6. PR/review/correction/integration plus stale-parent recovery.
7. Cancellation, resume reconciliation, and exactly-once human gate.
8. Exact-head readiness, observed human merge, docs, registration, and proportional Shepherd gates.

Broad Go/connector/certification gates remain parent-only after integration. #479 runs focused and
full Shepherd tests, strict TypeScript against Pi 0.80.6, offline Pi RPC, and scope/diff checks.
