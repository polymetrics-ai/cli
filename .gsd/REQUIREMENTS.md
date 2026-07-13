# Requirements

This file is the explicit capability and coverage contract for the project.

## Active

### R001 — One standalone Go command starts and supervises an issue-backed GSD milestone through a merge-disabled exact-head handoff.
- Class: primary-user-loop
- Status: active
- Description: One standalone Go command starts and supervises an issue-backed GSD milestone through a merge-disabled exact-head handoff.
- Why it matters: This is the user-visible replacement for the duplicated shell controller loop.
- Source: spec
- Primary owning slice: M001-k9bwxs/S04
- Supporting slices: M001-k9bwxs/S05, M001-k9bwxs/S07
- Validation: mapped
- Notes: The handoff remains human-gated and does not merge.

### R002 — Shepherd admits only pinned @opengsd/gsd-pi 1.11.0 and qualifies its documented headless commands, JSON events, supervised responses, and read-only query surface.
- Class: integration
- Status: active
- Description: Shepherd admits only pinned @opengsd/gsd-pi 1.11.0 and qualifies its documented headless commands, JSON events, supervised responses, and read-only query surface.
- Why it matters: Pinned behavior is the only safe integration contract for the governance runtime.
- Source: spec
- Primary owning slice: M001-k9bwxs/S01
- Supporting slices: M001-k9bwxs/S04
- Validation: mapped
- Notes: Early new-milestone return and real human depth confirmation are explicit qualification cases.

### R003 — Every dispatched unit enforces repository issue-first, required-skill, TDD, exact-head review, automated-review routing, and human-gate contracts.
- Class: compliance/security
- Status: active
- Description: Every dispatched unit enforces repository issue-first, required-skill, TDD, exact-head review, automated-review routing, and human-gate contracts.
- Why it matters: Native GSD workflow state must not weaken Polymetrics delivery governance.
- Source: spec
- Primary owning slice: M001-k9bwxs/S02
- Supporting slices: M001-k9bwxs/S05, M001-k9bwxs/S07
- Validation: mapped
- Notes: Missing objective, outputs, tools, boundaries, or evidence fails closed.

### R004 — Unknown commands, events, query shapes, models, state transitions, generations, scopes, and moved heads produce typed fail-closed terminal results.
- Class: failure-visibility
- Status: active
- Description: Unknown commands, events, query shapes, models, state transitions, generations, scopes, and moved heads produce typed fail-closed terminal results.
- Why it matters: Ambiguous authority or workflow interpretation caused silent and unsafe prior-controller behavior.
- Source: spec
- Primary owning slice: M001-k9bwxs/S03
- Supporting slices: M001-k9bwxs/S04, M001-k9bwxs/S05
- Validation: mapped
- Notes: No permissive fallback for unrecognized authority-bearing input.

### R005 — Controller authority is stored in a separate local SQLite database using WAL, durable settings, fencing records, grants, leases, approvals, attestations, and an outbox.
- Class: continuity
- Status: active
- Description: Controller authority is stored in a separate local SQLite database using WAL, durable settings, fencing records, grants, leases, approvals, attestations, and an outbox.
- Why it matters: Governance state must survive process interruption without reading or competing with GSD workflow state.
- Source: spec
- Primary owning slice: M001-k9bwxs/S03
- Supporting slices: M001-k9bwxs/S05
- Validation: mapped
- Notes: The store is local controller authority, not a distributed lock service.

### R006 — A supervised run reports normalized activity at least every 15 seconds and reconciles every terminal event or process exit with a fresh headless query.
- Class: operability
- Status: active
- Description: A supervised run reports normalized activity at least every 15 seconds and reconciles every terminal event or process exit with a fresh headless query.
- Why it matters: Operators need bounded silence and authoritative terminal classification.
- Source: spec
- Primary owning slice: M001-k9bwxs/S04
- Supporting slices: M001-k9bwxs/S06, M001-k9bwxs/S07
- Validation: mapped
- Notes: Stream events indicate activity; headless query determines reconciled workflow state.

## Validated

## Deferred

## Out of Scope

## Traceability

| ID | Class | Status | Primary owner | Supporting | Proof |
|---|---|---|---|---|---|
| R001 | primary-user-loop | active | M001-k9bwxs/S04 | M001-k9bwxs/S05, M001-k9bwxs/S07 | mapped |
| R002 | integration | active | M001-k9bwxs/S01 | M001-k9bwxs/S04 | mapped |
| R003 | compliance/security | active | M001-k9bwxs/S02 | M001-k9bwxs/S05, M001-k9bwxs/S07 | mapped |
| R004 | failure-visibility | active | M001-k9bwxs/S03 | M001-k9bwxs/S04, M001-k9bwxs/S05 | mapped |
| R005 | continuity | active | M001-k9bwxs/S03 | M001-k9bwxs/S05 | mapped |
| R006 | operability | active | M001-k9bwxs/S04 | M001-k9bwxs/S06, M001-k9bwxs/S07 | mapped |

## Coverage Summary

- Active requirements: 6
- Mapped to slices: 6
- Validated: 0
- Unmapped active requirements: 0
