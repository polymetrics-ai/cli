---
name: cli-architecture-v2-delivery
description: >-
  Delivery and orchestration rules for Polymetrics CLI Architecture v2. Use for issue #397,
  parent PR #438, branch feat/cli-architecture-v2, its S0, P01-P22, P18B, and D-TUI phases,
  direct or nested issues #398-#437, #462, #463, or #469, safety follow-up #453, and any audit,
  planning, implementation, verification, review, integration, continuation, completion,
  scheduling, or human-readiness work on that program. Covers Cobra/Viper and typed configuration,
  the event bus and TTY/NDJSON/Bubble Tea/accessibility phases, slog/OpenTelemetry, plain/JSON
  parity, stacked PRs, GSD/TDD evidence, exact-head review, Shepherd validation, and parent readiness
  while distinguishing parent-branch implementation truth from open delivery state.
compatibility: codex,opencode,claude
metadata:
  audience: agents
  purpose: program-delivery
---

# CLI Architecture v2 Delivery

This skill governs **how to deliver** CLI Architecture v2. It does not teach Cobra, Viper, Bubble
Tea, telemetry, or Go testing mechanics. It MUST be loaded with the specialists routed below.
Connector Architecture v2 is a separate sibling program; its migration handoff is not the roadmap
for issue #397.

## Mandatory Intake

Before classification, planning, review, or edits, the worker MUST:

1. Read [AGENTS.md](../../../AGENTS.md), [required skill routing](../../agentic-delivery/references/required-skills-routing.md), the [parent-orchestrator contract](../../agentic-delivery/contracts/parent-orchestrator-contract.md), the [issue-agent contract](../../agentic-delivery/contracts/issue-agent-contract.md), the [universal runtime loop](../../agentic-delivery/workflows/gsd-universal-runtime-loop.md), and the [Shepherd validator](../../agentic-delivery/workflows/shepherd-validator.md).
2. Read the [source plan](../../../docs/plans/cli-architecture-v2-improvement-plan.md), applicable ADRs, current parent orchestration artifacts, issue/PR state, and code. Parent artifacts, not this skill, own mutable edges and status.
3. Fetch remote truth and pin `main`, `feat/cli-architecture-v2`, the worker branch, relevant PR head,
   and merge-base. Re-fetch before handoff or integration.
4. Treat parent-branch implementation truth separately from GitHub OPEN/CLOSED delivery state. An
   issue remaining open does not prove its implementation is absent.
5. Load the [GSD Core skill](../../../.pi/skills/gsd-core/SKILL.md), run `scripts/gsd doctor`,
   `scripts/gsd list`, and `scripts/gsd sources <command>`, and use only commands that exist. If
   `programming-loop` is absent, record that fact and follow the manual universal loop; never invent
   the command.

## State And Ownership

Workers MUST use these distinct states: `parent_branch_satisfied_at`, `active_ready`,
`dependency_blocked`, `human_decision_blocked`, `integrated_review_debt`, `deferred_by_human`, and
`default_branch_complete`. Follow the [state and dependency model](references/state-and-dependency-model.md).

A commit message, issue checkbox, branch, or PR alone MUST NOT establish completion. Require code,
tests, GSD artifacts, exact integration evidence, and delivery state. Hierarchy, native dependency
edges, planning-only edges, and write-scope collisions are different facts.

One mutating worker owns one issue, branch, worktree, declared write scope, and current head. The
parent orchestrator alone owns shared program state, queue arbitration, promotion to the parent,
PR #438, and parent readiness. Workers MUST NOT launch autonomous drivers merely because this skill
loaded.

Integrated phases are frozen. Review or evidence debt SHOULD be repaired by exact-range review,
disposition, and an explicit waiver when authorized—not duplicate implementation or fabricated PR
history.

## Plan And Track Routing

Before production edits, the worker MUST update PLAN, TDD ledger, VERIFY checklist, prompt/run
state where used, owner, and write scope. Execute RED → GREEN → REFACTOR and retain exact evidence.
Use the [phase delivery checklist](references/phase-delivery-checklist.md).

- **CLI/config/help (ADR-0002):** load `golang-how-to`, `golang-cli`, `golang-spf13-cobra`,
  `golang-spf13-viper`, testing, security, documentation, and CLI help/docs/website parity
  guidance as applicable.
- **Events/TUI (ADR-0003):** additionally load [bubble-tea-tui-design](../bubble-tea-tui-design/SKILL.md),
  context, concurrency, safety, security, testing, and documentation. That skill remains the sole
  detailed authority for models, key maps, layout, charts, accessibility, and TUI tests.
- **slog/OpenTelemetry (ADR-0004):** load `golang-observability`, context, security,
  error-handling, and testing. Load `golang-benchmark` and `golang-performance` when a phase
  measures or optimizes telemetry overhead. A new or beta library always requires dependency approval.
- **Compact handoffs:** [caveman](../caveman/SKILL.md) MAY reduce prose but MUST NOT compress gates,
  evidence, commands, or warnings.

## Non-Negotiable Gates

Every phase MUST preserve exit codes, stdout/stderr ownership, one-envelope JSON, NDJSON ordering,
plain/non-TTY behavior, TTY bypasses, cancellation, sanitation, and secret redaction. CLI-visible
work MUST satisfy runtime help, bare namespace, command help, generated help/manual, `docs/cli/**`,
website, completion metadata, and tests, or record why an item is not applicable.

No worker may introduce generic shell, HTTP, or SQL write capability. Command arguments remain
untrusted. Reverse ETL remains plan → preview → approval → execute. Credentials, paths, telemetry,
and dependencies retain their existing approvals and safety rules; no issue body implies dependency
approval.

Review MUST bind findings and dispositions to an exact-head identity, then run independent VERIFY,
Shepherd trajectory validation, and parent reruns before promotion. Follow
[parent integration and review](references/parent-integration-and-review.md). Workers MUST obey
active review constraints and stop when they conflict with a configured route; this skill does not
hard-code a review bot.

The parent PR stays draft unless a human authorizes readiness. Parent merge to `main` is human-only.
Agents MUST NOT merge it.

## Definition of Done

A program slice is done only when:

- current remote, parent, worker, PR, and merge-base identities are recorded without drift;
- PLAN/TDD/VERIFY evidence proves RED → GREEN → REFACTOR and focused/broad checks;
- machine, TTY, security, dependency, reverse ETL, and help/docs/website gates pass;
- exact-head review findings are dispositioned and independent VERIFY plus Shepherd pass;
- the parent orchestrator accepts the handoff and reruns the parent integration gates;
- stacked metadata uses the parent base and `Refs`, while the human-only parent merge remains open.

A worker SHOULD hand off blockers with the precise state and evidence. It MAY recommend the next
safe action, but it MUST NOT mutate parent ownership or claim default-branch completion early.
