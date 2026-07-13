# Project Description

Adopt pinned GSD Pi as the native local workflow engine and govern it with a standalone Go Shepherd.

# Why This Milestone

The prior shell loop accumulated duplicated state, silent intervals, stale-head races, and ambiguous
authority. This milestone creates a typed, observable, fail-closed replacement without coupling it
to the Polymetrics CLI binary.

# User-Visible Outcome

One Go command starts an issue-backed supervised GSD milestone, reports activity at least every 15
seconds, preserves real human gates, and reaches a merge-disabled exact-head handoff.

# Completion Class

Behavioral infrastructure change with local and CI verification plus a final human gate.

# Final Integrated Acceptance

Nested Go tests and race tests pass, root module isolation is proven, GSD query/event integration is
qualified, incident replays pass, and no old controller is removed before the canary succeeds.

# Architectural Decisions

GSD Pi owns workflow state. Shepherd owns governance, effects, ratification, liveness, and normalized
telemetry. No autonomous merge capability exists. Controller state uses separate SQLite/WAL;
activity uses append-only JSONL segments.

# Error Handling Strategy

Unknown commands, events, query shapes, models, state transitions, generations, scopes, and moved
heads fail closed with typed terminal results.

# Risks and Unknowns

Pinned GSD Pi 1.11.0 can return early from new-milestone and requires real human depth confirmation.
Shepherd reconciles every terminal event with headless query and forwards supervised questions.

# Existing Codebase / Prior Art

The old shell controllers and repo-local Pi adapter remain as rollback references until cutover.

# Relevant Requirements

Issues #372 through #379 and the repository issue-first, GSD, skill-routing, TDD, review, and human
gate contracts.

# Scope

Project-local GSD policy resources, standalone `agent-runtime/shepherd`, deterministic replay,
merge-disabled canary, and post-approval legacy removal.

# Technical Constraints

Node 24, pinned `@opengsd/gsd-pi@1.11.0`, Go 1.25.4/toolchain 1.25.12, no root Go dependency, no
raw prompts/reasoning/secrets/tool output in telemetry, and no GSD database access.

# Integration Points

GSD headless JSON events, supervised stdin responses, headless query, Git read-only identity, Go
SQLite authority store, append-only activity spool, and future idempotent GitHub publishers.

# Testing Requirements

Unit, race, process cancellation, query reconciliation, module boundary, privacy, authority, and
named Twenty incident replay tests.

# Acceptance Criteria

All criteria in issue #372 are enforced or explicitly remain gated behind the canary.

# Open Questions

Legacy deletion inventory and superseded record closure require explicit review after the canary.

