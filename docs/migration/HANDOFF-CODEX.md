# Parallel handoff — continuing connector-architecture-v2 across Codex + Claude sessions

Repo: `polymetrics-ai/cli`, branch `connector-architecture-v2` (open as **PR #27** → `main`).
This document lets fresh **Codex** or **Claude** sessions pick up disjoint workstreams in parallel,
then hand results back for validation by Claude. Read `docs/migration/conventions.md` and
`docs/architecture/connector-architecture-v2-design.md` before touching code.

## State (2026-07-04)

- **On the new architecture: 523 / 557** — 517 JSON bundles (`internal/connectors/defs/`) + 6 Tier-3
  natives (`internal/connectors/native/`: postgres, dynamodb, amazon-sqs, bing-ads, tally-prime,
  faker) + 47 hook packages. 31 typed-blocked (`docs/migration/quarantine.json`).
- **Full API surface (Pass B / wave 5): ~240 / 517 done**, ~275 still at migration parity. A Pass B
  fan-out is **actively running in the primary Claude session** and owns those 275 `defs/` dirs.
- **Certify harness complete** (`internal/connectors/certify/`, 21 stages + batch + `pm connectors
  certify` CLI). **CDC**: engine-ready but postgres `ReadCDC` is a documented stub (needs the
  human-gated `pglogrepl` dependency); no connector advertises `cdc:true`.
- 95 fully-expanded connectors already committed/pushed to PR #27; more land as phased commits.

## Hard rule: avoid working-tree collisions

The primary session's Pass B run is writing `internal/connectors/defs/<parity-connector>/` dirs.
Two processes in ONE working tree corrupt each other (git index contention; the `//go:embed all:*`
in `defs/defs.go` breaks if any `defs/` dir is momentarily empty). Therefore every parallel session
MUST: (1) work in its **own git worktree** — `git worktree add ../wtX connector-architecture-v2`
then branch — and (2) take a workstream whose files are **disjoint** from `defs/` parity expansion.
Commit locally; **do NOT push** — the primary Claude session owns pushes (history must route through
a secret-scrubbed clone; see "Pushing" below). Hand branches back for Claude to merge + validate.

## Workstreams (disjoint — run in parallel)

### A. Wave 6 convergence PREP (Codex) — files: core `internal/connectors/*.go`, `cmd/`, `internal/cli/`, docs
Build everything up to (but NOT including) any deletion or the registry flip switch:
- `Registry.CatalogEntries()` view + a `cmd/cataloggen` (or extend connectorgen) that generates the
  catalog FROM the loaded bundles; wire `NewRegistry` to serve bundles via `engine.LoadAll(defs.FS)`
  behind a flag so it can be flipped without deleting legacy yet.
- Naming clean-break sweep script (bare names only; grep for `source-`/`destination-`).
- A legacy-deletion **manifest** (list of paths: `slug.go`, `catalog_data.json`, `native_port.go`,
  `native_conformance.go`, `manifest.go` structs, `registryset/`, per-connector legacy `*.go`,
  `cmd/registrygen`, `cmd/pm-cataloggen`) — as a documented script, NOT executed.
- Capture before/after `go build ./cmd/pm` binary size.
HUMAN GATE: do not delete anything or flip the production default. Deliver a branch + a written
cutover plan for approval.

### B. Quarantine 31 + CDC (Codex) — files: `internal/connectors/hooks/<name>/`, `native/<name>/`, their `defs/`
- Un-block the 31 in `docs/migration/quarantine.json` (mostly AUTH_COMPLEX → Tier-2 OAuth-refresh
  hooks copying `internal/connectors/hooks/gmail/hooks.go`; a few NON_REST → Tier-3 native). These
  `defs/` dirs are NOT in the Pass B roster, so they don't collide.
- CDC decoder: implement the postgres pgoutput Insert/Update/Delete → `connectors.CDCEvent` decoder
  with unit tests against captured/synthetic messages (NO live DB, NO new dependency for the pure
  decoder). Leave the live `START_REPLICATION` wiring behind the `pglogrepl` human gate.

### C. Expansion review (Codex reviewer agent or Claude) — read-only
Once a batch of Pass B connectors is committed, run the `connector-reviewer` agent over a 15–20%
sample vs legacy + real API docs; repair fails.

## Launching Codex

Agents are defined in `.codex/agents/passb-expander.toml` and `.codex/agents/connector-reviewer.toml`
(project-scoped; Codex auto-loads them). Non-interactive, headless:

```bash
# in a dedicated worktree for the workstream
git worktree add ../wt-wave6 connector-architecture-v2 && cd ../wt-wave6
git switch -c wave6-prep

CODEX_API_KEY=... codex exec --sandbox workspace-write \
  "Execute workstream A (Wave 6 convergence PREP) from docs/migration/HANDOFF-CODEX.md. Read that
   file + docs/migration/conventions.md + docs/architecture/connector-architecture-v2-design.md
   first. Build up to the human gate; do NOT delete legacy or flip the default; commit to this
   branch; do not push. End with a cutover plan."
```

For the expansion/quarantine agents, spawn per-connector so Codex parallelises (default
`agents.max_threads = 6`): ask the session to "spawn the passb-expander agent for each of: <list>".

## Pushing (Claude/coordinator only)

The branch history was scrubbed of two fake Stripe-format fixture keys via `git filter-repo` in a
clone at `<scratchpad>/scrub-clone` (origin points there). Direct pushes from the working repo's
un-scrubbed history are blocked by GitHub secret-scanning. Protocol: fetch the feature branch into
`scrub-clone`, `git cherry-pick 859b8e0e..FETCH_HEAD` onto scrubbed HEAD (skip empties), verify
`git grep 'sk_test_deadbeef\|51Hxxxx'` is empty, then `git push origin connector-architecture-v2`.

## GSD loop

These are GSD phases. Per phase: PLAN → EXECUTE test-first → gate (`go build ./...` &&
`connectorgen validate` 0 findings && conformance PASS && `make lint`) → VERIFY. Codex reads
`AGENTS.md`; the `.planning/` phase artifacts and `docs/migration/*` are the shared source of truth.
