# Reddit connector — three evidenced defect fixes (PR #3677)

## GSD setup

- PR: https://github.com/polymetrics-ai/cli/pull/3677
- Branch: `fm/cli-reddit-connector-defects-r1`
- Base: `9c977b208` (`chore(ashby): stage unvalidated parity checkpoint (#3542)`)
- GSD preflight: `scripts/gsd doctor` passed on 2026-08-04 (node v24.13.1, 69 commands registered,
  official docs / commands registry / upstream lock / Pi adapter resources all `ok`).
- GSD prompt path: `scripts/gsd prompt programming-loop init --phase cli-reddit-connector-defects-r1 --dry-run`
  was attempted first and the repo-local command registry returned
  `scripts/gsd: unknown GSD command: programming-loop`. `scripts/gsd` was therefore not used
  interactively to drive this slice, and this phase is an explicit **manual-GSD fallback** recorded
  per `.agents/agentic-delivery/references/gsd-pi-adapter.md` and
  `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`. Same fallback shape as the
  landed `xero-parity-wave04-r1` and `issue-3674-issueguard-checkpoint-links` phases.
- Orchestration decision, all cycles: `local_critical_path` — one connector-owned mutating scope in
  an already isolated worktree; parallel mutating workers would collide on the reddit bundle and its
  generated catalog/website surfaces.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-documentation`
- CLI help/docs/website parity reference
  (`.agents/agentic-delivery/references/cli-help-docs-website-parity.md`)

## Problem statement

Three defects in the already-shipped `reddit` connector, each confirmed on `main` by the
`cli-reddit-rdtcli-official-api-parity-r1` research report (report.md §4.3 and §7). All three are
independent of the still-open Reddit commercial-terms / data-retention decisions, so they are fixed
here rather than blocked behind that thread.

### Defect 1 — non-conforming User-Agent (BREAKING config change)

`streams.json` declared `base.user_agent: "polymetrics-go-cli"`. Reddit's API rules mandate
`<platform>:<app ID>:<version> (by /u/<reddit username>)` and drastically rate-limit clients that do
not conform.

The value has to carry a real operator identity, which means it must come from config at runtime.
Two engine facts constrain the fix and were both verified in-tree, not assumed:

1. `base.user_agent` is copied to `connsdk.Requester.UserAgent` verbatim and is **never**
   interpolated — there is no `{{ }}` support on that field. Putting the template there would ship
   the literal text `{{ config.reddit_username }}` on the wire.
2. `Requester.applyHeaders` (`internal/connectors/connsdk/http.go:623`) sets `User-Agent` from
   `r.UserAgent` first (lines 629-631) and then applies `r.DefaultHeaders` with `Header.Set` in a
   trailing loop (lines 638-640). `DefaultHeaders` therefore always wins over `UserAgent`.

Fix: move the value into a templated `base.headers["User-Agent"]` entry built from a new
`reddit_username` config field.

**This is a breaking config change**, and is shipped as such. `reddit_username` is added to
`spec.json`'s `required[]` on a `release_stage=ga` connector. The captain's explicit decision is to
keep it required (there are no existing reddit connector users, so backward compatibility is not a
practical concern) but to mark it honestly: commit and PR type `fix(connectors)!:` so
release-please / `CHANGELOG.md` reflect the break regardless of current user count.

### Defect 2 — api_surface.json claimed documented coverage it does not have

`api_surface.json` listed the comments stream's `GET /r/{subreddit}/comments` as ordinary covered
surface. That route is **not** among Reddit's 202 documented endpoints at
https://www.reddit.com/dev/api/. The only documented comments listing is the per-article tree
`GET [/r/subreddit]/comments/{article}`.

Migrating to the documented route is not feasible inside current engine constraints:

- It returns a **two-element top-level JSON array** (`[post_listing, comment_listing]`).
  `records.path` is a dotted-key traversal dialect only (`connsdk.RecordsAt`) with no array-index
  selection, so the response is not expressible.
- It is per-article, so covering a subreddit means an N+1 fan-out over every post, against Reddit's
  100 QPM budget.

Both are real engine/budget gaps, not oversights. Fix: keep the working bare route and make
`api_surface.json` **honest** instead — a scope caveat with citation and date, plus an explicit
`excluded` entry for the real documented endpoint.

### Defect 3 — access_token expiry undocumented

`spec.json` required a caller-supplied `access_token` and declared OAuth acquisition/refresh out of
scope, but Reddit bearer tokens expire 1 hour after issuance. Nothing told the operator that. Fix:
document the limit plainly in the `access_token` description, `docs.md`, and the 401 hint.

## Scope boundaries

### In scope — connector bundle

1. `internal/connectors/defs/reddit/streams.json` — `base.user_agent` removed; templated
   `base.headers["User-Agent"]` added; 401 and 429 hints sharpened.
2. `internal/connectors/defs/reddit/spec.json` — `reddit_username` property added and added to
   `required[]`; `access_token` description records the 1-hour expiry.
3. `internal/connectors/defs/reddit/api_surface.json` — `reviewed_at` bumped to `2026-08-03`; scope
   caveat with report citation; `excluded` entry for `[/r/subreddit]/comments/{article}`.
4. `internal/connectors/defs/reddit/docs.md` — `reddit_username` field, User-Agent format, token
   expiry, undocumented-route caveat, and the rate-limit header names.

### In scope — regression test

5. `internal/connectors/engine/reddit_user_agent_test.go` — new. Pins the outbound User-Agent,
   following the existing per-connector test precedent (`xero_operations_test.go`,
   `cmd/connectorgen/gong_api_surface_test.go`). Three cases: declaration site, wire value across
   paginated requests, and fail-closed behavior when `reddit_username` is absent.

### In scope — generated surfaces

6. `docs/connectors/catalog/all-connectors.json`, `docs/connectors/reddit/MANUAL.md`,
   `docs/connectors/reddit/SKILL.md`, `website/data/connectors.generated.json`,
   `website/lib/connectors.catalog.data.generated.json` — brought in sync with the new bundle,
   touching **only** the reddit entry. A full `pm docs generate` regenerates all ~550 connectors due
   to pre-existing unrelated drift; that churn is deliberately kept out of this connector-scoped PR.
   Verified: the catalog diff is 9 lines, all inside the reddit entry.

### In scope — GSD planning evidence

7. `.planning/phases/cli-reddit-connector-defects-r1/PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`,
   and `RUN-STATE.json` — required alongside the code because `internal/**` changed;
   `scripts/verify-gsd-workflow` fails the `gsd-workflow-evidence` gate without changed
   `.planning/**` evidence. These are documentation artifacts and carry no behavior.

### Out of scope

- **No shared-runtime changes.** Three engine limitations are recorded and worked around rather than
  fixed, because each would be a shared change outside this connector-scoped task:
  - No `{{ version }}` template namespace exists, so the version segment of the User-Agent is
    hardcoded `v1`.
  - `records.path` has no array-index selection, so the documented per-article comments route stays
    unreachable.
  - There is no engine auth mode for refresh tokens, so `access_token` stays caller-supplied.
- **No new `excluded.category` enum value.** The documented-but-unreachable endpoint is recorded as
  `out_of_scope`, following the existing YNAB-connector precedent, rather than inventing a category
  that would require editing the shared engine schema.
- No other connector's definitions, docs, or generated entries.
- No `pm` command, subcommand, flag, or help-topic change.
- No live provider calls, no credential use, no reverse ETL execution, no dependency changes.
- The open Reddit commercial-terms / data-retention decisions are untouched; all three defects here
  are independent of that thread.

## Implementation slices

1. **Red** — add `reddit_user_agent_test.go` pinning the templated header, the wire value across
   pagination, and the fail-closed path. Confirm red against the base reddit bundle.
2. **Green, defect 1** — move User-Agent into `base.headers`, add `reddit_username` to `spec.json`
   (property + `required[]`).
3. **Green, defect 2** — rewrite `api_surface.json` scope with the cited caveat and add the
   `excluded` entry for the documented per-article route.
4. **Green, defect 3** — record the 1-hour token expiry in `spec.json`, `docs.md`, and the 401 hint;
   name `X-Ratelimit-Used` / `-Remaining` / `-Reset` in the 429 hint.
5. **Generated-surface sync** — update the reddit entry only across catalog, manual, skill, and
   website data; verify the diff touches nothing else.
6. **Evidence** — record actual command output in `TDD-LEDGER.md` and `VERIFICATION.md`.

## CLI/help/docs/website parity checklist

Per `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`:

- `pm` command / subcommand / flag / help topic: **not applicable** — no CLI surface changes; this
  slice only edits a connector bundle and its generated documentation.
- Connector surface (`pm connectors inspect reddit --json`): **applicable and updated** — the spec
  gains a required `reddit_username` field, reflected in `docs/connectors/reddit/MANUAL.md`,
  `docs/connectors/reddit/SKILL.md`, `docs/connectors/catalog/all-connectors.json`, and both
  generated website data files.
- Website docs under `website/**`: **applicable and updated**, reddit entry only.

## Verification checklist

- `gofmt -l internal/connectors/engine/reddit_user_agent_test.go`
- `go vet ./internal/connectors/engine/...`
- `go test ./internal/connectors/engine/... -count=1`
- `go build ./cmd/pm`
- Base-bundle red run (reddit `streams.json` + `spec.json` from `9c977b208`, current test)
- `scripts/verify-gsd-workflow`
- Catalog diff scoped to the reddit entry

## Commit checkpoint plan

Implementation commit for the three defect fixes, a `fix(connectors)!:` commit carrying the breaking
`reddit_username` requirement plus catalog sync and the regression test, a docs-markup fix, and this
planning evidence. Never push to `main`; merge stays human-gated.
