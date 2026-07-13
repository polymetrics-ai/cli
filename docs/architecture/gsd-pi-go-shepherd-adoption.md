# GSD Pi and Go Shepherd adoption

Status: accepted for governed integration; rejected for unmanaged direct adoption.

## Decision

Pin `@opengsd/gsd-pi@1.11.0` and use only its documented `headless`, `headless query`, JSON event,
supervised-response, and stop/resume surfaces. Generic recovery is deliberately excluded because it
can delete authoritative hierarchy rows. A standalone Go process supervises those
interfaces and owns policy, external-effect authority, exact-head ratification, bounded liveness,
and privacy-safe telemetry.

GSD Pi's SQLite database remains private workflow state. Shepherd has a separate SQLite/WAL database
for grants, leases, approvals, attestations, and its outbox. Analytics are optional sinks, never
controller state.

## Qualification observations

- Version 1.11.0 and the Codex `gpt-5.6-sol` model catalog entry were confirmed locally.
- Filtered headless lifecycle events and `headless query` are machine-readable.
- A clean-fixture `new-milestone` run returned while the milestone remained in `pre-planning`.
- `plan` is not a documented headless command and was interpreted as a quick task.
- Milestone discussion requires a real human depth confirmation; answer defaults do not bypass it.
- Unfiltered message events can contain large encrypted runtime payloads. Shepherd therefore uses a
  strict lifecycle allowlist and never stores full upstream event objects.
- A governed intake canary produced native tool events and supervisor heartbeats without gaps over
  15 seconds, then returned early with a pending discussion unit. Query reconciliation correctly
  classified the apparent success as blocked.
- The first actual coordinator session used the requested Sol model but inherited thinking `off`.
  Shepherd now validates the controlled runtime settings at admission. A subsequent disposable
  governed session recorded `openai-codex/gpt-5.6-sol` with `thinkingLevel: high` and forwarded the
  native depth confirmation as a real human gate.

## Consequences

- `headless query` is treated as a fenced reconciliation transition, not a read-only API; a native
  `skip` result blocks the controller for an explicit next unit.
- Human questions are surfaced through supervised mode and remain auditable gates.
- Model identity is checked from observed events. Missing or different validator identity fails
  closed; there is no silent downgrade.
- The existing shell controllers and repo-local GSD adapter remain until deterministic replay and a
  merge-disabled canary pass.
