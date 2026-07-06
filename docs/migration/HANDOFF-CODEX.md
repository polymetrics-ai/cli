# Parallel handoff — continuing connector-architecture-v2 across Codex + Claude sessions

Repo: `polymetrics-ai/cli`, branch `connector-architecture-v2` (open as **PR #27** → `main`).
This document lets fresh **Codex** or **Claude** sessions pick up disjoint workstreams in parallel,
then hand results back for validation. Read `docs/migration/conventions.md` and
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
Commit regularly after green verification, push only issue/PR branches, and never push to `main`.
Historical scrubbed-history warnings are recorded in "Pushing and PR creation" below; they do not
block verified issue-branch pushes or PR creation.

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
Once a batch of Pass B connectors is committed, run the `connector-reviewer` agent spec over a
15-20% sample vs legacy + real API docs; repair fails.

## Launching Agents

Agent specs are runner-neutral YAML files:

- `.agents/connector-migration/agents/implementation/passb-expander.agent.yaml`
- `.agents/connector-migration/agents/review/connector-reviewer.agent.yaml`

Translate the relevant YAML spec into the runner prompt for Codex, Claude, OpenCode, or another
agent runtime. Non-interactive Codex example:

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

For expansion/review agents, spawn one task per connector so the runtime can parallelize safely.
Use the `passb-expander` spec for implementation work and the `connector-reviewer` spec for
read-only review.

## Pushing and PR creation

The branch history was scrubbed of two fake Stripe-format fixture keys via `git filter-repo` in a
clone at `<scratchpad>/scrub-clone` during the connector-architecture-v2 migration. That history
note is not a blanket ban on autonomous delivery. Agents should push committed, verified issue/PR
branches and open linked PRs so CI and review automation can run.

Required push rules:

- Never push directly to `main`.
- Push only the active issue/PR branch or the parent integration branch named by the issue plan.
- Run the issue's local verification before pushing a green slice.
- Do not push commits that contain real secrets, private keys, authorization headers, or
  secret-looking fixtures.
- For stacked work, create or confirm the parent PR from the parent branch to `main` before
  treating sub-issues as executable.
- Do not treat a green CodeRabbit status as review completion when CodeRabbit also reports that the
  review was skipped or disabled for a non-default base branch. The sub-PR must have CodeRabbit
  review records, or the parent PR to `main` must review the integrated commit range.
- Stop for the human gates in `AGENTS.md`: auth scope changes, secrets, dependencies, destructive
  external actions, production deploys, quality gate reductions, generic write tools, or parent PR
  merge to `main`.

## GSD loop

These are GSD phases. Per phase: PLAN → EXECUTE test-first → gate (`go build ./...` &&
`connectorgen validate` 0 findings && conformance PASS && `make lint`) → VERIFY. Codex reads
`AGENTS.md`; the `.planning/` phase artifacts and `docs/migration/*` are the shared source of truth.
