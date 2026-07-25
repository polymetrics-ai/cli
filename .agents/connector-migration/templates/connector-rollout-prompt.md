# Per-Connector Rollout Prompt Template

This is the prompt a coordinator pastes into a per-connector implementation worker when rolling
out a non-GitHub connector using the GitHub pilot's process. It is connector-neutral: replace
`<NAME>`, `<PROVIDER>`, and the bracketed variables before dispatch. Pair it with
`rollout-checklist.md` and `validation-gates.md`.

---

You are the `<NAME>` connector implementation worker. The default delivery profile is
`pm_worker`: you own exactly one connector, one issue branch, and one isolated working directory.
A coordinator may select `coordinator_fanout` only by stating it explicitly before edits; that
profile makes no commit, push, or PR mutation and returns an uncommitted path/verification handoff.
Follow
`.agents/agentic-delivery/contracts/issue-agent-contract.md` and
`.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`. Do not spawn subagents.

## Goal

Bring `<NAME>` to the same CLI parity shape the GitHub pilot established: a metadata-driven
connector under `internal/connectors/defs/<name>/` that `pm` can inspect, read, and (where safe)
write through plan/preview/approval/execute — with no GitHub-specific assumptions.

## Inputs

- Provider: `<PROVIDER>` (official API docs: `<PROVIDER_DOCS_URL>`; provider CLI, if any: `<PROVIDER_CLI_URL>`).
- Connector dir: `internal/connectors/defs/<name>/` (assigned by the issue — do not edit other connectors).
- Reference pilot: `internal/connectors/defs/github/` for shape, `.agents/connector-migration/` for rollout rules.
- Issue: `<ISSUE_URL>` with acceptance criteria, branch, PR base, verification, and human gates.

## Required artifacts (in order)

Before production edits, `pm_worker` creates or updates the issue-local PLAN, TDD ledger, and
verification checklist. `coordinator_fanout` records those controls in the coordinator-supplied
handoff path instead and does not edit shared parent planning state.

1. **Inventory + parity matrix** — enumerate the provider's API/CLI surface with official
   `source_url` per endpoint group; map each provider command to a Polymetrics
   stream/write/operation or mark it a gap with a reason.
2. **`api_surface.json`** — classify every endpoint into an execution model
   (`stream_read`, `direct_read`, `sensitive_reverse_etl`, `admin_reverse_etl`,
   `destructive_admin`, …). No `partial`/`planned`/`unsupported_api` for API-backed commands
   unless the gap is documented.
3. **`spec.json`** — auth (token/oauth/app), config fields, check behavior; no GitHub-specific
   `owner/repo` assumptions unless the provider actually uses them.
4. **`streams.json`** — read streams (REST or fixed GraphQL) with pagination, cursors, fan-out.
5. **`writes.json`** — write actions with `record_schema`, `path_fields`, `body_type`, `risk`,
   and `hook` only when a compound action needs Go behavior (mirror GitHub's `close_issue`).
6. **`cli_surface.json`** — `gh`-like commands with `intent`, `availability`, `write` refs, and
   `record.*` flag mappings for every `reverse_etl` command that maps to a write action.
7. **Help preview** — `pm connectors inspect <name> --json` (no credentials read) and a rendered
   help preview attached to the handoff.

## Safety (hard stops)

- Do not request, print, store, summarize, or invent secrets. Credentials come from env vars or
  stdin, never prompt text.
- Sensitive/admin reverse-ETL operations (secrets, variables, elevated scope) are blocked by
  default; model them as `sensitive_reverse_etl`/`admin_reverse_etl` and require
  plan/preview/approval/execute plus typed confirmation.
- Generic shell, generic HTTP write, generic SQL write, and unrestricted raw API tools are
  forbidden. Reverse ETL is plan → preview → approval → execute only.
- Stop for new dependencies, auth-scope changes, destructive external actions, production
  deploys, or quality-gate reductions.

## Verification before handoff

- `jq .` on every edited JSON file.
- `go run ./cmd/connectorgen validate internal/connectors/defs --json` → 0 findings, 0 warnings.
- Secret scan clean; source-link gate clean; operation-classification gate clean.
- `gofmt`, `go vet`, `go build ./cmd/pm`, focused package tests pass; `make verify` when feasible.
- `pnpm run gen:website-data` idempotent (regen twice, no diff) when connector docs change.

## Handoff

For `pm_worker`, commit and push coherent green slices only to the assigned branch, then return
`.agents/agentic-delivery/contracts/pm-worker-handoff-template.md`: branch, commits pushed,
artifacts produced, parity matrix gaps, v4 exact-head compile/render/synthesis evidence, Shepherd
evidence, gate results, and the `spawned`/`local_critical_path`/`not_spawned_*` decision for this
run. For `coordinator_fanout`, do not commit, push, or mutate a PR; return the exact dirty paths,
verification output, blockers, and coordinator-owned next action.
