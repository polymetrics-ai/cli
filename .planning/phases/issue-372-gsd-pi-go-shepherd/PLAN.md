# Issue 372: GSD Pi + Go Shepherd

## Objective

Adopt pinned `@opengsd/gsd-pi@1.11.0` as the workflow engine and add a standalone Go governance
module without coupling it to the Polymetrics CLI module.

## Delivery slices

1. Qualify supported GSD Pi headless interfaces and record incompatibilities (#373).
2. Encode the Polymetrics issue-first/TDD/review/human-gate overlay (#374).
3. Build the typed Go governance domain and SQLite authority store (#375).
4. Add bounded headless execution, allowlisted event projection, liveness, and query reconciliation
   (#376).
5. Add capability grants, fencing, exact-head ratification, and an idempotent GitHub outbox (#377).
6. Add redacted local telemetry and analytics export boundaries (#378).
7. Replay known incidents and run a merge-disabled canary before legacy removal (#379).
8. Qualify the pinned Podman image: prove it builds without UID collisions, reports the exact GSD
   version, runs as the intended non-root identity, and preserves task-isolated planning state.
9. Install the approved agent development surface in the governed image: the repository's pinned Go
   toolchain, Make, Git, CA trust, jq, ripgrep, a bounded read-only fetch command, and pinned
   `agent-browser`; explicitly exclude GitHub/publisher CLIs and unrestricted HTTP mutation tools.
10. Provision trusted Context7 HTTP MCP configuration from the controller, not from worker-controlled
    repository state, and keep API credentials optional and outside version control.
11. Run SearXNG as a separately pinned, private-network sidecar with JSON search enabled and no host
    port; do not bake a search service into the worker image.
12. Make official GSD Pi the documented execution runtime locally and in Podman. Retain
    `scripts/gsd` only as a deterministic compatibility prompt renderer, and merge a tested
    repo-local `programming-loop` command overlay into both the shell renderer and Pi aliases.

## Boundaries

- GSD Pi owns workflow state. Shepherd only uses documented CLI/query/event surfaces.
- Shepherd owns external-effect authorization and never autonomously merges to `main`.
- No raw prompts, reasoning, credentials, command arguments, or tool results are persisted.
- Research tools default to read-only behavior, bounded output, and controller-owned configuration.
- The agent image has no GitHub CLI, publisher token, unrestricted curl, or host search port.
- The module lives under `agent-runtime/shepherd/` with a separate `go.mod`.
- Legacy removal and issue/PR closure are a final, separately reviewed cutover step.

## Required skills and workflows

- `golang-how-to`, `go-engineering`, Go project-layout, design-patterns, observability,
  concurrency, context, testing, error-handling, safety, and database guidance.
- `gsd-programming-loop` with RED/GREEN/refactor evidence.
- The repo-local adapter currently lacks `programming-loop`; use and record the permitted
  manual-GSD fallback for the Podman qualification correction.
- Issue-first and parent/subissue orchestration contracts.

## Execution decisions

- Qualification: `local_critical_path` because it blocks every implementation slice.
- Architecture/workflow review: `read_only_spawned` using two non-mutating review agents.
- Production implementation: `local_critical_path` in the dedicated issue branch; no shared-checkout
  mutating subagent is permitted.
