# ADR: Partial adoption of pi-workflow-engine 0.12.0

- Status: accepted for issue #490
- Date: 2026-07-23
- Parent architecture: #471 Shepherd

## Context

Shepherd needs Pi `0.80.10` compatibility and an event-agnostic AgentSession completion pattern. The exact locally installed `pi-workflow-engine@0.12.0` provides bounded fan-out, typed workflow handoffs, cancellation/budgets/progress, and isolated patch experiments for development and review.

Shepherd's production port additionally requires caller-claimed persistent cwd/workspace authority, per-session scoped host tools, exact binding metadata, exhaustive typed failure behavior, caller-owned abort, join-before-release, and durable receipts tied to Shepherd's state/effect journals.

## Decision

Adopt `pi-workflow-engine@0.12.0` as the project-local development/review orchestrator only. Keep the existing `ProductionAgentSessionPort` implementation and reuse the engine's documented public Pi pattern conceptually: await prompt completion, treat custom typed terminal capture as result authority, and use raw events only for bounded non-authoritative progress.

Do not deep-import engine internals. Do not treat `.pi/.workflow-runs`, engine run IDs, resume journals, patches, disposable worktrees, retries, or background state as Shepherd evidence. Do not grant engine workflows Git/GitHub publication, human-decision, integration, or merge authority. Shepherd production neither imports the package nor routes `pm-shepherd` execution to workflow-engine commands or built-ins. A developer may explicitly invoke the package's built-in workflows with the normal authority of that Pi coding-agent session, including their documented tools; that developer-tool use is outside Shepherd production and is never Shepherd verification evidence or durable authority. Shepherd review evidence uses only host-bound inline workflows: the host captures and bounds exact Git material before dispatch, review agents receive only that material with `tools: []`, no dynamic tool hints, and no skills, and the host revalidates the range afterward. Those bounded review runs do not receive GitHub mutation authority.

## Why production embedding is rejected

The documented package surface does not provide all of the following as a stable host API:

1. acknowledged caller-supplied cwd/persistent workspace claim;
2. per-session Shepherd `ToolDefinition[]`/host capability injection with an authority receipt;
3. Shepherd binding metadata in the result;
4. an exhaustive typed failure contract;
5. a public caller-owned cancellation operation; and
6. typed `abort(runId)` plus `join(runId)` receipts.

The package manifest also has wildcard Pi peer dependencies and no package-root host SDK export. Deep-importing `.pi/extensions/.../src/**` would violate the issue boundary and create upgrade risk.

## Consequences

- Shepherd durability, worktrees, effects, recovery, retries, reviews, approvals, and no-main-merge policy remain unchanged.
- Workflow-engine is immediately useful for bounded read-only analysis and the one independent exact-head review through the restricted inline policy above.
- `.pi/.workflow-runs/` is ignored local state and cannot dirty Shepherd's canonical target evidence.
- CI installs the exact artifact, verifies lock integrity/public entry provenance, and exercises isolated plus co-loaded offline RPC command discovery.
- Production remains on public Pi SDK APIs behind its existing port.
- Reconsider production embedding only after upstream documents and exports every missing authority primitive in a separately reviewed issue.

## Provenance

Installed package version `0.12.0`; npm tarball integrity `sha512-DX+e2U03raK8o8YbwnDUcAQSKNZm0v1J6jWS+bk2j2kEFihLmZCf0sUlrHWou1kWC3Zw+CA4HCgqpjLWlmtcRg==`. Evidence is from `.pi/npm/package-lock.json` and the installed manifest; local `.pi/npm` remains ignored cache, not durable Shepherd state.
