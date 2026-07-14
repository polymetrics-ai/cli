# Project

## What This Is

Polymetrics is a Go-only CLI monolith with a standalone nested Go Shepherd module under `agent-runtime/shepherd/`. Shepherd governs pinned official GSD Pi issue delivery without compiling into `pm`. The branch already has typed run state, official GSD Pi headless execution, bounded native events and heartbeats, canonical unit fencing, exact model and thinking validation, live write-scope cancellation, local checkpointing, durable decision provenance, and idempotent PR decision summaries. The remaining gap is an operator-driven command loop: an operator still selects units, classifies failures, resumes retries, and polls human decisions.

## Core Value

One validated issue can progress safely from canonical GSD intake to a final human-gated parent PR with a single supervisor command, while preserving exact authority, isolation, independent validation, restart safety, and an explicit no-merge boundary.

## Project Shape

- **Complexity:** complex
- **Why:** The supervisor crosses process, filesystem, Git worktree, SQLite, GitHub, agent-model, restart, and human-authority boundaries with exactly-once and fail-closed requirements.
- **Web stack:** not a web UI; standalone Go CLI plus official GSD Pi subprocess and GitHub REST integration

## Current State

The nested Shepherd module pins and governs `@opengsd/gsd-pi@1.11.0`. Existing packages cover typed authority and run states, validated dispatch contracts, SQLite/WAL leases and outbox state, allowlisted native event parsing, headless execution and query reconciliation, session-bound model observations, exact-head Git snapshots and ratification, decision provenance, marker-owned PR decision summaries, privacy-safe telemetry, and Twenty/Asana incident guards. Existing commands remain operator-oriented and there is no continuous `supervise` policy loop, transactional attempt worktree manager, resumable pending-decision broker, or GitHub answer reader.

## Architecture / Key Patterns

- Standalone nested module: `agent-runtime/shepherd/`; root `go list ./...` and `pm` must not include it.
- Official GSD Pi remains the workflow engine; Shepherd uses only documented headless, query, JSON event, supervised-response, and stop/resume surfaces.
- GSD Pi SQLite is private workflow state; Shepherd owns separate SQLite/WAL controller authority.
- Canonical GSD snapshots and exact Git heads fence every unit; unknown state or model drift fails closed.
- Supervisor policy is deterministic and side-effect free; process, Git, persistence, clock, and GitHub operations sit behind narrow ports.
- External effects use durable idempotency and generation binding; timing alone never provides exactly-once behavior.
- Logs, comments, and summaries are bounded and redacted.

## Capability Contract

See `.gsd/REQUIREMENTS.md` for the explicit capability contract, requirement status, and coverage mapping.

## Milestone Sequence

- [ ] M001: Autonomous Shepherd Supervisor V1 — Deliver the one-command restart-safe issue-to-human-gated-parent-PR supervisor with typed recovery, transactional worktrees, exact model routing, durable GitHub decisions, and a merge-disabled canary.
