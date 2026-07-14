# Requirements

This file is the explicit capability and coverage contract for the project.

## Active

### R001 — After validated intake, one `shepherd supervise --config <path> --issue <N>` invocation continuously reconciles canonical GSD state, selects exactly one fenced unit, and advances without routine operator command selection until a human decision or final merge gate is required.
- Class: primary-user-loop
- Status: active
- Description: After validated intake, one `shepherd supervise --config <path> --issue <N>` invocation continuously reconciles canonical GSD state, selects exactly one fenced unit, and advances without routine operator command selection until a human decision or final merge gate is required.
- Why it matters: This is the core autonomous replacement for the current operator-driven unit command loop.
- Source: spec
- Primary owning slice: M001/S01
- Supporting slices: M001/S02, M001/S03, M001/S04, M001/S05, M001/S06, M001/S07
- Validation: mapped
- Notes: The loop must remain single-unit-at-a-time and fail closed on unknown canonical state.

### R002 — Shepherd deterministically classifies dead worker, silent tool, false green, stale head, dirty tree, scope breach, and repeated failure, then selects only an allowed bounded recovery action for the class.
- Class: failure-visibility
- Status: active
- Description: Shepherd deterministically classifies dead worker, silent tool, false green, stale head, dirty tree, scope breach, and repeated failure, then selects only an allowed bounded recovery action for the class.
- Why it matters: Autonomy is unsafe if failures are retried generically, hidden, or allowed to consume unbounded attempts.
- Source: spec
- Primary owning slice: M001/S02
- Supporting slices: M001/S01, M001/S07
- Validation: mapped
- Notes: Unknown or exhausted failures stop or enter awaiting_decision; recovery never grants new authority.

### R003 — Every unit attempt runs in a transactional worktree; failed or cancelled attempts leave the canonical issue worktree unchanged, while only independently verified in-scope commits may be promoted after a fresh head check.
- Class: continuity
- Status: active
- Description: Every unit attempt runs in a transactional worktree; failed or cancelled attempts leave the canonical issue worktree unchanged, while only independently verified in-scope commits may be promoted after a fresh head check.
- Why it matters: Attempt isolation prevents partial worker output and cleanup failures from corrupting the canonical delivery checkout.
- Source: spec
- Primary owning slice: M001/S03
- Supporting slices: M001/S01, M001/S07
- Validation: mapped
- Notes: Cleanup ownership and crash recovery must be deterministic and must not delete worktrees owned by a live generation.

### R004 — Planning, recovery planning, coordination, independent validation, and UAT prove `openai-codex/gpt-5.6-sol` with `high` thinking; implementation and delegated execution prove `openai-codex/gpt-5.5` with `high` thinking; any model or thinking drift fails closed.
- Class: constraint
- Status: active
- Description: Planning, recovery planning, coordination, independent validation, and UAT prove `openai-codex/gpt-5.6-sol` with `high` thinking; implementation and delegated execution prove `openai-codex/gpt-5.5` with `high` thinking; any model or thinking drift fails closed.
- Why it matters: The specification assigns distinct authority and execution roles and requires independently observable routing rather than configuration intent alone.
- Source: spec
- Primary owning slice: M001/S04
- Supporting slices: M001/S01, M001/S02, M001/S07
- Validation: mapped
- Notes: Both admission settings and post-unit native observations must agree.

### R005 — `awaiting_decision` is a durable resumable run state backed by an atomic generation-bound decision request and exactly-once consumption of one accepted answer across process or host restart.
- Class: continuity
- Status: active
- Description: `awaiting_decision` is a durable resumable run state backed by an atomic generation-bound decision request and exactly-once consumption of one accepted answer across process or host restart.
- Why it matters: Human latency must not lose supervisor progress, duplicate authority, or replay an answer.
- Source: spec
- Primary owning slice: M001/S05
- Supporting slices: M001/S01, M001/S06, M001/S07
- Validation: mapped
- Notes: Requests bind issue, PR when present, canonical unit, generation, and exact head; expiry defaults to a safe stop.

### R006 — Shepherd publishes exactly one marker-owned bounded GitHub question on the bound issue or PR, mentions `@karthik-sivadas`, polls replies, and advances only for the configured allowlisted human's unedited exact-syntax answer matching the live request, generation, and head.
- Class: integration
- Status: active
- Description: Shepherd publishes exactly one marker-owned bounded GitHub question on the bound issue or PR, mentions `@karthik-sivadas`, polls replies, and advances only for the configured allowlisted human's unedited exact-syntax answer matching the live request, generation, and head.
- Why it matters: GitHub is V1's decision source of truth and GitHub Mobile is the notification path; authorization and replay resistance are product-critical.
- Source: spec
- Primary owning slice: M001/S06
- Supporting slices: M001/S05, M001/S07
- Validation: mapped
- Notes: Bots, unauthorized users, stale generations or heads, edited comments, malformed syntax, and duplicate deliveries are rejected and recorded.

### R007 — The supervisor emits a bounded event or heartbeat at least every 15 seconds, monitors process liveness, progress, write scope, deadlines, and budgets, and resumes after restart without duplicating a unit, checkpoint, comment, summary, external effect, or accepted answer.
- Class: quality-attribute
- Status: active
- Description: The supervisor emits a bounded event or heartbeat at least every 15 seconds, monitors process liveness, progress, write scope, deadlines, and budgets, and resumes after restart without duplicating a unit, checkpoint, comment, summary, external effect, or accepted answer.
- Why it matters: An unattended supervisor must remain diagnosable and crash-safe under ordinary process and host failures.
- Source: spec
- Primary owning slice: M001/S01
- Supporting slices: M001/S02, M001/S03, M001/S05, M001/S06, M001/S07
- Validation: mapped
- Notes: Durable idempotency and fencing, not timing assumptions, provide exactly-once effects.

### R008 — A deterministic incident replay suite plus a merge-disabled GitHub sandbox canary proves one supervise command progresses known Twenty and Asana failure scenarios from issue state to the final human-gated parent PR without manual per-unit commands.
- Class: launchability
- Status: active
- Description: A deterministic incident replay suite plus a merge-disabled GitHub sandbox canary proves one supervise command progresses known Twenty and Asana failure scenarios from issue state to the final human-gated parent PR without manual per-unit commands.
- Why it matters: Unit contracts alone cannot prove the assembled supervisor, GSD Pi, Git, GitHub, persistence, restart, and validation boundaries work together.
- Source: spec
- Primary owning slice: M001/S07
- Supporting slices: M001/S01, M001/S02, M001/S03, M001/S04, M001/S05, M001/S06
- Validation: mapped
- Notes: The canary must be merge-disabled and must exercise restart and exactly-once decision behavior.

### R009 — Logs, comments, decision summaries, and durable records contain only bounded redacted evidence and never secrets, credentials, raw prompts, chain-of-thought, command arguments, or unrestricted tool output.
- Class: compliance/security
- Status: active
- Description: Logs, comments, decision summaries, and durable records contain only bounded redacted evidence and never secrets, credentials, raw prompts, chain-of-thought, command arguments, or unrestricted tool output.
- Why it matters: The supervisor processes agent and tool activity unattended and publishes to GitHub; unsafe telemetry would leak sensitive or untrusted content.
- Source: spec
- Primary owning slice: M001/S06
- Supporting slices: M001/S01, M001/S02, M001/S04, M001/S05, M001/S07
- Validation: mapped
- Notes: PR summaries retain decision, actor, concise basis, canonical unit, and exact head only.

### R010 — GitHub question publishing and answer reading are exposed through replaceable narrow ports so bootstrap `gh` authentication can later move to a least-privilege Shepherd App without changing supervisor policy.
- Class: integration
- Status: active
- Description: GitHub question publishing and answer reading are exposed through replaceable narrow ports so bootstrap `gh` authentication can later move to a least-privilege Shepherd App without changing supervisor policy.
- Why it matters: The current authenticated human identity is explicitly temporary and transport concerns must not become policy dependencies.
- Source: inferred
- Primary owning slice: M001/S06
- Supporting slices: M001/S05
- Validation: mapped
- Notes: V1 implements only the currently authenticated GitHub transport.

## Validated

## Deferred

### R011 — Support Telegram, WhatsApp, or Signal notification adapters as alternate human-decision notification paths.
- Class: admin/support
- Status: deferred
- Description: Support Telegram, WhatsApp, or Signal notification adapters as alternate human-decision notification paths.
- Why it matters: Additional channels may improve response latency after V1 establishes a stable decision broker.
- Source: spec
- Primary owning slice: none
- Supporting slices: none
- Validation: unmapped
- Notes: Explicitly deferred; GitHub and GitHub Mobile are V1's sole source-of-truth and notification path.

### R012 — Authenticate GitHub operations with a dedicated least-privilege Shepherd GitHub App or service account.
- Class: admin/support
- Status: deferred
- Description: Authenticate GitHub operations with a dedicated least-privilege Shepherd GitHub App or service account.
- Why it matters: A dedicated identity improves operational separation after the autonomous canary stabilizes.
- Source: spec
- Primary owning slice: none
- Supporting slices: none
- Validation: unmapped
- Notes: V1 bootstraps with the currently authenticated human GitHub identity behind replaceable ports.

## Out of Scope

### R013 — Shepherd must never merge the parent pull request to `main`, refresh authentication scope, provision secrets, perform destructive production effects, or remove legacy controllers before the replacement canary passes.
- Class: anti-feature
- Status: out-of-scope
- Description: Shepherd must never merge the parent pull request to `main`, refresh authentication scope, provision secrets, perform destructive production effects, or remove legacy controllers before the replacement canary passes.
- Why it matters: These actions require new authority or would bypass the explicit final human gate and safe migration sequence.
- Source: spec
- Primary owning slice: none
- Supporting slices: none
- Validation: n/a
- Notes: Final readiness is surfaced as a human decision and merge remains external to Shepherd.

## Traceability

| ID | Class | Status | Primary owner | Supporting | Proof |
|---|---|---|---|---|---|
| R001 | primary-user-loop | active | M001/S01 | M001/S02, M001/S03, M001/S04, M001/S05, M001/S06, M001/S07 | mapped |
| R002 | failure-visibility | active | M001/S02 | M001/S01, M001/S07 | mapped |
| R003 | continuity | active | M001/S03 | M001/S01, M001/S07 | mapped |
| R004 | constraint | active | M001/S04 | M001/S01, M001/S02, M001/S07 | mapped |
| R005 | continuity | active | M001/S05 | M001/S01, M001/S06, M001/S07 | mapped |
| R006 | integration | active | M001/S06 | M001/S05, M001/S07 | mapped |
| R007 | quality-attribute | active | M001/S01 | M001/S02, M001/S03, M001/S05, M001/S06, M001/S07 | mapped |
| R008 | launchability | active | M001/S07 | M001/S01, M001/S02, M001/S03, M001/S04, M001/S05, M001/S06 | mapped |
| R009 | compliance/security | active | M001/S06 | M001/S01, M001/S02, M001/S04, M001/S05, M001/S07 | mapped |
| R010 | integration | active | M001/S06 | M001/S05 | mapped |
| R011 | admin/support | deferred | none | none | unmapped |
| R012 | admin/support | deferred | none | none | unmapped |
| R013 | anti-feature | out-of-scope | none | none | n/a |

## Coverage Summary

- Active requirements: 10
- Mapped to slices: 10
- Validated: 0
- Unmapped active requirements: 0
