# M001: Autonomous Shepherd Supervisor V1

**Gathered:** 2026-07-14
**Status:** Ready for planning

## Project Description

Autonomous Shepherd Supervisor V1 turns the existing standalone Go Shepherd runtime into an operator-started, issue-to-PR supervisor. One invocation, `shepherd supervise --config <path> --issue <N>`, accepts one already-validated GitHub issue, continuously reconciles canonical GSD state, selects exactly one fenced unit at a time, supervises the pinned official GSD Pi subprocess, and advances without routine operator command selection.

The supervisor owns routine unit selection, process monitoring, classified bounded recovery, durable checkpointing, restart reconciliation, and GitHub decision polling. It remains fail-closed: scope or safety ambiguity, unknown or exhausted failures, and final parent-PR merge authority stay with the configured allowlisted human.

## Why This Milestone

Shepherd already has typed run and authority state, validated dispatch contracts, SQLite/WAL leases and outbox state, allowlisted native event parsing, bounded heartbeats, exact model and thinking validation, live write-scope cancellation, exact-head Git snapshots, local checkpointing, durable decision provenance, and idempotent PR decision summaries. These capabilities do not yet form an autonomous operating loop: an operator still selects units, classifies failures, resumes retries, and polls human decisions.

This milestone closes that gap so one validated issue can safely progress from canonical intake to a human-ready parent PR without manual per-unit commands. It preserves the central safety promise: autonomy removes routine intervention but never expands Shepherd's authority or bypasses the final human merge gate.

## User-Visible Outcome

### When this milestone is complete, the user can:

- Run `shepherd supervise --config <path> --issue <N>` for one validated issue and watch Shepherd autonomously advance its canonical GSD delivery units.
- Interrupt the foreground process, restart it, and observe the same run resume from durable state without duplicating a unit, accepted answer, comment, summary, checkpoint, commit promotion, or other external effect.
- Inspect concise live transitions and machine-readable durable status/history while Shepherd performs allowed bounded retries or waits for an allowlisted human's GitHub decision.
- Receive a verified, reviewed, human-ready parent PR that Shepherd will not merge automatically.

### Entry point / environment

- Entry point: `shepherd supervise --config <path> --issue <N>` in the standalone `agent-runtime/shepherd` CLI
- Environment: foreground local CLI process, with a production-like merge-disabled GitHub sandbox used for canary acceptance
- Live dependencies involved: pinned official GSD Pi subprocess, Git, transactional worktrees, local SQLite/WAL state, GitHub issue/comment/review/PR APIs, and configured model providers

## Completion Class

- Contract complete means: deterministic tests prove state transitions, single-unit fencing, failure classification, bounded recovery, transactional promotion, durable decision generation, exact-once answer consumption, redaction, restart reconciliation, and machine-readable status behavior.
- Integration complete means: the actual Shepherd CLI drives the pinned official GSD Pi subprocess, transactional Git worktrees, SQLite state, and real GitHub issue/branch/comment/review/PR state through their narrow ports.
- Operational complete means: a foreground run emits bounded events or heartbeats at least every 15 seconds, survives SIGINT/process restart, resumes durable waits and retries, rejects stale or unauthorized authority, avoids duplicate external effects, and reaches the final human gate in a merge-disabled sandbox canary.

## Final Integrated Acceptance

To call this milestone complete, we must prove:

- One real validated sandbox issue progresses from canonical issue state to a verified, automated-review-covered, human-ready parent PR through one `supervise` invocation, without manual per-unit commands and without Shepherd merging it.
- During a real run, the process is interrupted while work or a human decision is pending; restart reconciles the same generation and head, resumes polling or execution, and produces no duplicate unit, checkpoint, GitHub question, accepted answer, summary, commit promotion, or other external effect.
- Known Twenty and Asana failure scenarios pass a deterministic incident replay suite, including dead worker, silent tool, false green, stale head, dirty tree, scope breach, and repeated failure outcomes.
- The final canary cannot substitute fixtures for the pinned official GSD Pi subprocess or for real GitHub issue, branch, marker-owned question, allowlisted answer, review, and parent-PR state. Controlled failure injection may be simulated, but the external integration path and restart behavior must be live.

## Architectural Decisions

### One Issue Per Foreground Invocation

**Decision:** V1 preserves the exact `--issue <N>` contract and supervises one validated issue per foreground process. It does not accept a sequential multi-issue queue.

**Rationale:** This aligns with R001, keeps ownership and restart semantics unambiguous, and retires the core autonomy risk before introducing queue continuation policy or cross-issue scheduling.

**Alternatives Considered:**
- Sequential issue list — deferred because per-issue failure continuation and queue durability enlarge V1 without proving the core loop.
- Durable queue config — deferred because it introduces another state machine and operator contract.
- GitHub label discovery — excluded because V1 admission is explicit, not autonomous discovery.

### Reconciliation With One Fenced Canonical Unit

**Decision:** The supervisor continuously reconciles durable canonical state and selects exactly one fenced unit at a time. Every attempt runs in a transactional worktree, and promotion requires independent verification plus a fresh exact-head and write-scope check.

**Rationale:** Fencing and transactional isolation make crash recovery and cancellation safe. Failed or cancelled attempts cannot contaminate the canonical issue worktree or gain authority through retry.

**Alternatives Considered:**
- Direct execution in the canonical issue worktree — rejected because partial or failed attempts could leave ambiguous state.
- Timing-based single-worker assumptions — rejected because restart and stale-worker races require durable generations, leases, and head checks.

### Classified Bounded Recovery

**Decision:** Only deterministic, allowlisted failure classes receive persisted bounded recovery actions with backoff. Unknown, unsafe, scope-ambiguous, stale-authority, or exhausted failures stop or enter `awaiting_decision`; recovery never grants new authority.

**Rationale:** Generic or unbounded retries can conceal false greens, consume budgets, repeat external effects, or continue after the run's authority is stale.

**Alternatives Considered:**
- One generic retry for every failure — rejected because different failures require different safe actions.
- Fully autonomous replan/replace behavior — rejected for V1 because it expands policy authority and complicates proof.
- Human decision after every failure — rejected because it preserves the routine operator loop this milestone exists to remove.

### Durable GitHub Human-Decision Channel

**Decision:** Safety, scope, and final-gate decisions use a generation- and exact-head-bound durable `awaiting_decision` state. Shepherd publishes one marker-owned bounded GitHub question, mentions `@karthik-sivadas`, remains alive with bounded polling and heartbeats, and consumes exactly one unedited exact-syntax answer from the configured allowlisted human.

**Rationale:** GitHub is the canonical collaboration surface and survives process restarts. Binding the request and answer to issue, PR when present, unit, generation, and exact head prevents stale or unauthorized decisions from advancing work.

**Alternatives Considered:**
- Attached terminal input — rejected because it is not durable across process or host restart.
- Exit immediately after persisting a wait — not chosen for V1 because the desired operator experience is a continuously supervising foreground process.
- Accept either CLI or GitHub answers — rejected because dual authority channels create conflict and replay rules.

### Human-Ready Parent PR Is the Terminal Success State

**Decision:** Successful supervision ends when the parent PR contains integrated, independently verified work and decision summaries, has required automated review coverage, and is ready for final human merge approval. Shepherd never merges the parent PR.

**Rationale:** This provides a literal issue-to-PR outcome while preserving the repository's non-bypassable human merge gate.

**Alternatives Considered:**
- Stop after sub-PR integration — rejected because parent-PR readiness would still require routine operator completion work.
- Continue until human approval is recorded — not required because review latency and final approval are human workflow concerns.
- Automatic merge — explicitly excluded because it violates the authority boundary.

### Foreground Lifecycle With Checkpointed Stop

**Decision:** On SIGINT or termination, Shepherd cancels and fences active work, persists canonical state and retry metadata, releases owned claims safely, and stops. Restart reconciles state before taking any action and must not duplicate external effects.

**Rationale:** A foreground CLI must support normal operator interruption without corrupting work or requiring manual reconstruction.

**Alternatives Considered:**
- Always finish the current unit before exit — rejected because unsafe or long-running work may need prompt cancellation.
- Immediate uncheckpointed exit — rejected because cleanup ownership and resume position would become ambiguous.

## Error Handling Strategy

Shepherd uses typed failure classes and fail-closed propagation rather than generic retries. It deterministically distinguishes at least dead worker, silent tool, false green, stale head, dirty tree, scope breach, repeated failure, deadline/budget exhaustion, malformed or unauthorized human answers, and persistence or GitHub integration failures.

Each known recoverable class maps to a bounded allowlisted action such as reconcile, resume, or retry. Attempt counts, next eligible retry time, generation, exact head, and selected action are persisted before execution. Backoff and retry limits are configuration validated; concrete defaults remain a planning decision. Unknown classes, exhausted limits, authority drift, model/thinking drift, and any ambiguous safety or write-scope condition stop or enter durable `awaiting_decision` rather than guessing.

All side effects use durable fencing and idempotency. Transactional attempts leave the canonical worktree unchanged on failure or cancellation. GitHub questions and summaries are marker-owned and idempotent. Answers are accepted only from the allowlisted human, with exact syntax and matching live generation/head; bots, edited comments, stale answers, malformed syntax, and duplicate delivery are rejected and recorded as bounded evidence.

User-facing errors and status records identify the failure class, safe current state, exhausted or remaining recovery budget, and required next authority without exposing secrets. Logs, comments, summaries, and durable records never include credentials, raw prompts, chain-of-thought, unrestricted tool output, or unbounded command arguments.

## Risks and Unknowns

- Exactly-once external effects across crash boundaries — checkpoints, Git commits, GitHub comments, and accepted answers can straddle process failure, so idempotency and reconciliation must be proven rather than assumed.
- Stale worker or generation races — a prior process may remain alive after restart; leases, generation fencing, write-scope cancellation, and exact-head checks must prevent dual authority.
- False-green classification — treating incomplete or unverified work as success could promote bad commits and advance the wrong unit.
- Retry policy quality — limits that are too permissive can hide persistent failures, while limits that are too strict can create unnecessary human waits. Exact counts and backoff defaults remain for planning.
- GitHub decision authenticity and replay — edited, stale, duplicated, bot-authored, or unauthorized replies must never advance a live request.
- Operational liveness — the supervisor must distinguish a quiet but healthy tool from a silent or dead worker while honoring the 15-second bounded event/heartbeat contract.
- Bootstrap GitHub authentication — V1 may use existing `gh` authentication behind narrow ports, but the design must permit later migration to a least-privilege Shepherd App without changing supervisor policy.
- Model availability and identity — required role-specific model and thinking settings must be observed exactly; drift or unavailable models fail closed.

## Existing Codebase / Prior Art

- `agent-runtime/shepherd/` — standalone nested Go module that owns Shepherd without compiling it into `pm`.
- `agent-runtime/shepherd/cmd/` — existing Shepherd CLI entry points where the `supervise` command and status surface belong.
- `agent-runtime/shepherd/internal/` — existing typed run state, dispatch, persistence, Git/GitHub, execution, telemetry, and policy packages to compose into the supervisor loop.
- `agent-runtime/shepherd/shepherd.example.json` — prior configuration shape to extend with validated supervisor recovery and lifecycle settings.
- `.gsd/REQUIREMENTS.md` — authoritative R001-R010 capability, safety, integration, and launchability contract.
- `scripts/gsd` and `.gsd/upstream.lock.json` — repository adapter and source lock for the pinned official GSD Pi runtime.

## Relevant Requirements

- R001 — Defines the one-command, one-issue autonomous reconciliation loop and final human gate.
- R002 — Requires deterministic failure classification and only allowed bounded recovery actions.
- R003 — Requires transactional attempt worktrees and independently verified exact-head promotion.
- R004 — Fixes role-specific model and thinking assignments and requires drift to fail closed.
- R005 — Requires durable, generation-bound `awaiting_decision` state and exactly-once answer consumption across restart.
- R006 — Defines marker-owned GitHub questions, allowlisted exact-syntax answers, and live request/head matching.
- R007 — Defines bounded events/heartbeats, lifecycle monitoring, restart continuity, and no duplicate effects.
- R008 — Requires deterministic incident replay and a merge-disabled live GitHub sandbox canary through the final human-gated parent PR.
- R009 — Requires bounded redacted evidence and prohibits secrets, credentials, raw prompts, chain-of-thought, command arguments, and unrestricted tool output in records.
- R010 — Requires replaceable narrow GitHub question/answer ports so bootstrap `gh` authentication can later become a least-privilege Shepherd App.

## Scope

### In Scope

- `shepherd supervise --config <path> --issue <N>` for one explicitly admitted validated issue.
- Continuous canonical-state reconciliation and exactly one fenced active unit.
- Foreground signal-aware process lifecycle, safe checkpointed cancellation, and restart resume.
- Deterministic failure classification, persisted bounded retries/backoff, and fail-closed escalation.
- Transactional attempt worktrees, independent verification, fresh-head checks, and in-scope commit promotion.
- Durable `awaiting_decision`, marker-owned GitHub questions, allowlisted exact answers, and exactly-once consumption.
- Concise live events plus machine-readable durable status/history.
- Required role-specific model/thinking validation and live drift detection.
- Privacy-safe bounded telemetry, comments, summaries, and incident evidence.
- Deterministic incident replay plus a real merge-disabled GitHub/GSD Pi canary.
- Parent-branch integration, required automated review coverage, and a human-ready parent PR.

### Out of Scope / Non-Goals

- GitHub label polling or automatic discovery/admission of issues.
- Multiple issue IDs, a durable multi-issue queue, or multi-issue parallelism.
- Interactive TUI, installed daemon/service, CI-hosted supervisor, or distributed coordinator.
- Generic autonomous replanning or authority expansion beyond allowlisted recovery actions.
- Automatic final parent-PR merge or bypass of any human gate.
- Generic shell, generic HTTP write, or generic SQL write capabilities.
- Replacing bootstrap GitHub authentication with a Shepherd App in V1; only the narrow-port migration seam is required.
- Connector architecture migration or changes to legacy connector cutover behavior.

## Technical Constraints

- Implementation remains inside the standalone Go module under `agent-runtime/shepherd/` and must not compile into `pm`.
- Official GSD Pi stays pinned and source-governed; the supervisor must use the repository's canonical adapter and validated headless execution path.
- One active issue and exactly one fenced canonical unit per supervisor process.
- Planning, recovery planning, coordination, independent validation, and UAT use `openai-codex/gpt-5.6-sol` with `high` thinking; implementation and delegated execution use `openai-codex/gpt-5.5` with `high` thinking. Any observed drift fails closed.
- State transitions, retry metadata, outbox effects, leases, decision generations, and answer consumption must be durable through SQLite/WAL-backed state.
- Command arguments and GitHub content are untrusted. Reject control characters, malformed syntax, stale heads/generations, path traversal, broad write scopes, and unauthorized actors.
- Events and evidence are bounded and redacted. Never store or expose secrets, credentials, raw prompts, chain-of-thought, unrestricted tool output, or unbounded command arguments.
- Heartbeat or bounded progress evidence must occur at least every 15 seconds while the supervisor is active or waiting.
- All promoted work requires independent verification, in-scope commits, and a fresh exact-head check.
- Final merge remains human-only.

## Integration Points

- Official GSD Pi subprocess — executes canonical units through validated headless commands and emits allowlisted native events for reconciliation.
- Git and transactional worktrees — isolate attempts, enforce write scope, snapshot exact heads, and promote only verified commits.
- SQLite/WAL state and outbox — persist run generations, leases, checkpoints, retry state, pending effects, and durable human requests.
- GitHub REST/`gh` integration — reads canonical issue/PR/review state, publishes marker-owned bounded questions and summaries, polls allowlisted answers, and creates or updates the parent PR through narrow replaceable ports.
- Model providers — execute role-bound work only when exact configured model and thinking observations match policy.
- Automated review routing — confirms local reviewer/verifier/security coverage for the relevant commits, with human fallback only when repository policy permits it.
- Operator terminal — starts and interrupts the foreground process and consumes concise live events and machine-readable status without becoming an ephemeral authority channel.

## Testing Requirements

Use strict test-first vertical slices in the nested Go module.

- Unit tests: table-driven tests for every run-state transition, unit-selection fence, failure classification, allowed recovery action, retry exhaustion, backoff scheduling, model/thinking validation, redaction rule, GitHub answer validator, and exit/status mapping.
- Persistence tests: SQLite/WAL transaction tests for checkpoints, leases, generations, outbox idempotency, retry metadata, decision requests, and exactly-once answer consumption across reopen/restart.
- Git tests: temporary repositories and worktrees proving failed/cancelled attempts leave canonical state unchanged; stale head, dirty tree, scope breach, dead generation, and verified promotion scenarios.
- Process tests: fake and real subprocess fixtures for dead worker, silence, heartbeat deadlines, cancellation, budget/deadline exhaustion, bounded output, and restart reconciliation.
- GitHub contract tests: narrow-port fixtures for marker ownership, duplicate delivery, edited or malformed replies, bot/unauthorized actors, stale request/head/generation, exact-syntax acceptance, and idempotent summaries.
- Integration tests: actual Shepherd command with the pinned GSD Pi subprocess and controlled repositories, including SIGINT and process restart at pre-effect, post-effect/pre-checkpoint, and awaiting-decision boundaries.
- Incident replay: deterministic Twenty and Asana scenarios covering all known failure classes and expected safe terminal states.
- Live canary: a merge-disabled GitHub sandbox issue, branch, marker-owned question, allowlisted reply, automated review coverage, and human-ready parent PR. Exercise at least one restart and prove exactly-once decision and external-effect behavior.
- Repository gates: run formatting, vet, tests, build, and the Shepherd module's verification target; run broader repository verification when shared files are touched.

## Acceptance Criteria

### Autonomous Supervisor Loop

- One `supervise` invocation accepts one validated issue and advances canonical GSD state without routine per-unit operator commands.
- At most one fenced canonical unit is active, and stale generations cannot select, write, promote, or advance work.
- The process remains alive during healthy work and durable GitHub waits, emitting bounded state or heartbeat evidence at least every 15 seconds.

### Classified Recovery

- Every known failure class maps deterministically to an allowed bounded action or safe stop.
- Retry count, backoff, generation, exact head, and selected recovery action survive restart.
- Unknown, exhausted, unsafe, scope-ambiguous, and model/thinking-drift cases fail closed and clearly identify the required authority.

### Transactional Attempts

- Failed and cancelled attempts leave the canonical issue worktree unchanged.
- Only independently verified, in-scope commits are promoted after a fresh exact-head check.
- Crash cleanup never deletes a worktree owned by a live generation and does not leave ambiguous ownership.

### Durable Human Decisions

- Shepherd creates exactly one marker-owned bounded question for a live request and resumes polling after restart.
- Only the configured allowlisted human's unedited exact-syntax answer matching the live issue/PR, unit, generation, and head advances the run.
- Duplicate, edited, malformed, bot-authored, unauthorized, stale-generation, and stale-head answers are rejected and recorded without duplicate consumption.

### Operational Continuity and Visibility

- SIGINT safely cancels and fences active work, checkpoints state, releases owned claims, and exits without duplicate effects.
- Restart reconciles before acting and resumes the same run, retry, or decision wait.
- Live events and machine-readable status/history expose bounded state, progress, failure class, recovery budget, and required next authority without secret or prompt leakage.

### Final Integration

- A deterministic replay suite passes known Twenty and Asana failure scenarios.
- A real merge-disabled GitHub sandbox canary uses the pinned official GSD Pi subprocess and reaches a verified, automated-review-covered, human-ready parent PR.
- The canary includes interruption/restart and exactly-once decision evidence.
- Shepherd does not merge the final parent PR.

## Open Questions

- What exact default retry counts, backoff schedule, and overall recovery budgets should each recoverable failure class use? — Planning should choose conservative validated defaults; configuration must remain bounded and fail closed.
- What exact JSON schema and exit-code contract should machine-readable status expose? — It must cover current state, generation, issue/unit identity, exact head, heartbeat/progress, retry budget, pending human request, and safe next authority without exposing prohibited data.
- Which specific review states satisfy “required automated review coverage” for terminal readiness? — Follow repository policy: local automated review coverage is primary; a recorded human fallback is acceptable only for the defined blocker path and exact commit range.
- Which bootstrap GitHub client should back the narrow ports in V1? — Existing `gh` authentication is acceptable if validated and bounded; policy must remain independent so a least-privilege Shepherd App can replace it later.