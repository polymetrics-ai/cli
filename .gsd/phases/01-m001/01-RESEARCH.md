# M001 Research: Autonomous Shepherd Supervisor V1

## Research framing

This milestone is a composition and crash-consistency milestone over an already substantial standalone Go runtime, not a greenfield agent runtime. The roadmap should prove authority fencing and restart-safe orchestration before adding broad recovery policy or live GitHub acceptance.

Required policy reviewed: `AGENTS.md`, `polymetrics-issue-delivery`, required skill routing, issue-agent contract, GSD Pi adapter guidance, universal runtime loop, and runtime/RLM integration guidance. This research unit is read-only; no behavior changed, so a RED checkpoint is not applicable.

## Existing foundation to reuse

- `internal/domain`: typed run states and human-decision/capability concepts.
- `internal/store`: SQLite/WAL-backed deliveries, delivery runs, leases with epochs, and idempotent effect records.
- `internal/gsd`: pinned official GSD Pi subprocess runner, allowlisted native-event parsing, 15-second heartbeat default, model/thinking governance, canonical query decoding, and reconciliation that rejects successful no-progress/mutating-skip outcomes.
- `internal/git`: exact-head and dirty-tree inspection with special handling for GSD-owned control-plane paths.
- `internal/contract`: bounded validated issue context and explicit dispatch contracts carrying objective, output format, tool guidance, tools, boundaries, and write scope.
- `internal/authority`: exact-head, generation, attempt, state-version, evidence, validator, and expiry-bound ratification.
- `internal/decision` and `internal/github`: durable append-only decision records and marker-owned idempotent PR decision summaries behind a runner interface.
- `internal/telemetry`: bounded allowlisted activity events and heartbeat-gap evaluation.
- `internal/replay`: Twenty/Asana-inspired incident guard fixtures already cover portions of stale authority, event integrity, and other safety invariants.
- `cmd/shepherd`: existing explicit run/resume lifecycle, signal handling, write-scope monitoring, activity emission, configuration validation, and narrow GitHub publication behavior.

These primitives should be composed behind new narrow supervisor ports rather than duplicated in `cmd/shepherd`.

## Confirmed gaps

Targeted source analysis found no current `supervise` command. There is also no explicit durable `awaiting_decision` state, persisted retry/backoff schedule or failure-class policy, transactional attempt-worktree abstraction, parent-PR readiness coordinator, or `status --json` contract. Existing retry behavior is intake-specific rather than a supervisor recovery state machine. Existing GitHub support publishes decision summaries to a configured PR; it does not yet implement issue/PR question publication plus answer polling/authentication.

## What should be proven first

First prove a deterministic, side-effect-free supervisor decision function over canonical persisted state: given delivery/run generation, lease epoch, exact head, GSD snapshot, active attempt, retry metadata, and pending decision, it returns exactly one allowed next action or a fail-closed wait/stop. Then prove the durable claim/transition transaction and restart reconciliation around that decision. This retires dual-worker, duplicate-unit, and crash-boundary risks before introducing real subprocess, Git, or GitHub effects.

The first vertical tracer should drive a fake canonical GSD source through one fenced no-op unit, persist before/after state, restart at each boundary, and demonstrate that only one generation/lease can advance. It should not begin with the full CLI or live GitHub canary.

## Boundary contracts that matter

1. **Canonical-state port:** bounded query snapshot in; typed unit identity and blockers out. Never infer advancement solely from process exit.
2. **Supervisor planner:** pure reconciliation input to one typed action (`start_unit`, `resume_unit`, `retry_at`, `await_decision`, `publish_effect`, `terminal_success`, `terminal_failure`). Unknown combinations fail closed.
3. **Attempt executor:** generation/lease/unit/head-bound request; typed result with evidence references only. It must not own policy decisions.
4. **Transactional workspace/promoter:** isolated attempt workspace creation, cancellation cleanup, independent verification, fresh exact-head/write-scope recheck, then atomic in-scope promotion.
5. **Recovery policy:** exhaustive known failure-class to bounded action mapping; attempts and next eligible time persisted before execution; clock and backoff injected for deterministic tests.
6. **Human-decision port:** marker-owned request keyed by issue/PR, unit, generation, and exact head; bounded polling; allowlisted immutable exact-syntax answer; transactional consume marker.
7. **External-effect outbox:** intent persisted before GitHub/comment/promotion execution, with stable idempotency key and reconciliation of unknown outcomes after restart.
8. **Status/history projection:** machine-readable schema derived from durable state, not ephemeral CLI memory; bounded/redacted human output is a projection of the same events.
9. **Terminal gate:** human-ready parent PR is success; merge is not representable as an allowed Shepherd capability.

## Preliminary slice strategy

1. Supervisor policy/state machine, fencing, restart reconciliation, status, and bounded liveness.
2. Failure taxonomy with persisted bounded recovery and backoff.
3. Transactional attempt worktrees and verified exact-head promotion.
4. Exact model/thinking routing for GSD Pi execution, recovery planning, validation, and UAT.
5. Durable `awaiting_decision` state and exactly-once decision broker.
6. GitHub question publisher/reply reader behind narrow ports and outbox idempotency.
7. Deterministic Twenty/Asana incident replay plus the merge-disabled restart canary through a human-ready parent PR.

The roadmap should preserve this risk order: fencing before effects, promotion before recovery breadth, deterministic ports before live GitHub.

## Requirement baseline

R001–R010 are active and collectively cover the primary loop, failure visibility, continuity, model governance, durable human waits, GitHub decision authenticity, operational liveness, canary launchability, redaction, and replaceable GitHub ports. R011–R012 are deferred notification/identity capabilities. R013 is the explicit authority anti-feature covering parent-PR merge, secret/auth-scope changes, destructive production effects, and premature legacy-controller removal.

At this stage no silent scope expansion is recommended. Candidate requirement analysis will be refined after examining persistence and CLI seams.
