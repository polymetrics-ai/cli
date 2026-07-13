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
13. Replace the Shepherd Podman backend with pinned official GSD Pi running directly on the host.
    Preserve Git-tracked project policy under each issue worktree's `.gsd/`, set `GSD_STATE_DIR`
    to the delivery's protected state directory, use a delivery-specific GSD home, and verify the
    migration against the real Asana issue before deleting the container and search-sidecar assets.

## Boundaries

- GSD Pi owns workflow state. Shepherd only uses documented CLI/query/event surfaces.
- Shepherd owns external-effect authorization and never autonomously merges to `main`.
- No raw prompts, reasoning, credentials, command arguments, or tool results are persisted.
- Research tools default to read-only behavior, bounded output, and controller-owned configuration.
- The agent image has no GitHub CLI, publisher token, unrestricted curl, or host search port.
- The module lives under `agent-runtime/shepherd/` with a separate `go.mod`.
- Legacy removal and issue/PR closure are a final, separately reviewed cutover step.
- Local GSD project state is scoped by the canonical issue worktree; external GSD state and
    managed worktrees are scoped by the delivery `state_dir`. No two deliveries share either path.
14. Publish the protected decision ledger to the bound pull request after every answered GSD gate.
    Use one marker-owned, idempotently updated PR comment; preserve `human`, `shepherd`, and
    `contract` provenance; and fail the governed unit if a durable decision cannot be published.
15. Bound lifecycle event lines at a realistic size for official GSD nested-agent returns. Keep raw
    payloads out of telemetry, but allow an 8 MiB input envelope so the compact projector can retain
    the tool-end action after a healthy multi-minute subagent completes.
16. Permit a successful `execute-task` unit to leave scoped source edits, then create a local
    controller checkpoint commit before the next fenced unit. Validate every changed path against
    the immutable protected issue-context `write_scope`; retain the clean-start invariant and never
    push from the worker runtime.
17. Preserve the primary reconciliation failure when a failed unit also leaves a dirty worktree;
    report both causes so recovery distinguishes a completed scoped task from an incomplete red-test
    checkpoint.
18. Resolve runtime identity from the newest exact-worktree Pi session only. Ignore older oversized
    sessions and nested sessions, scan the selected header/metadata with the same 8 MiB bound, and
    continue rejecting any effective model other than GPT-5.6 Sol/high.

## Required skills and workflows

- `golang-how-to`, `go-engineering`, Go project-layout, design-patterns, observability,
  concurrency, context, testing, error-handling, safety, and database guidance.
- `gsd-programming-loop` with RED/GREEN/refactor evidence.
- The repo-local adapter exposes `programming-loop`; use it for the host-runtime migration and
  record RED/GREEN/refactor evidence below.
- Issue-first and parent/subissue orchestration contracts.

## Execution decisions

- Qualification: `local_critical_path` because it blocks every implementation slice.
- Architecture/workflow review: `read_only_spawned` using two non-mutating review agents.
- Production implementation: `local_critical_path` in the dedicated issue branch; no shared-checkout
  mutating subagent is permitted.
