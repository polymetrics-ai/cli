# Issue #180 — Freshchat CLI parity parent roadmap

Parent issue: https://github.com/polymetrics-ai/cli/issues/180
Parent branch: `feat/180-freshchat-cli-parity`
Default branch: `main`
Connector: `freshchat`
Definition scope: `internal/connectors/defs/freshchat/`

## GSD command path

- `scripts/gsd doctor` — pass (2026-07-09).
- `scripts/gsd verify-pi` — pass (2026-07-09).
- `scripts/gsd list --json` — pass (2026-07-09).
- `scripts/gsd prompt plan-phase issue-180-freshchat-cli-parity --skip-research` — generated and followed.
- `scripts/gsd prompt programming-loop init --phase issue-180-freshchat-cli-parity --dry-run` — unavailable: `scripts/gsd: unknown GSD command: programming-loop`.

Manual-GSD fallback is active only for the missing `programming-loop` registry entry. The repository-local Pi prompt `/pm-gsd-loop` and `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` are the fallback programming-loop policy for implementation slices.

## Required skills loaded

- GSD: `gsd-core`.
- Go/CLI/connector: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-lint`.
- Repo references: `required-skills-routing.md`, `gsd-pi-adapter.md`, `cli-help-docs-website-parity.md`, connector migration conventions/design docs, parent/subissue contracts, CodeRabbit and automated-review routing loops.

## Parent objective

Bring Freshchat to full connector CLI parity using connector-architecture-v2 bundles and the command-surface metadata path. Every official operation must be classified as one of:

1. stream-backed durable read;
2. bounded direct read/read query;
3. typed reverse-ETL write;
4. bounded binary/file policy; or
5. blocked with an exact duplicate/deprecated/disallowed/auth-internal/product-scope reason.

No generic HTTP write, shell, SQL write, raw mutation, credential prompt, or credentialed live connector check is allowed.

## Current baseline

- Current Freshchat bundle has 18 streams and 13 write actions.
- Current `api_surface.json` has 34 official endpoint rows.
- Official source fetched for planning from `https://developers.freshchat.com/api/`; sanitized endpoint extraction found 34 operation rows after discarding a documentation typo/example (`GET /metric`) and retaining `GET /users/{user_id}/conversations` from the official navigation/body.
- The official docs page contains secret-shaped example Authorization values; raw docs were not committed and must not be copied into artifacts, fixtures, docs, comments, or logs.

## Sub-issue lanes

| Issue | Lane | Dependency | Write scope | Planned branch |
| ---: | --- | --- | --- | --- |
| #181 | CLI surface metadata | parent PR open | `internal/connectors/defs/freshchat/cli_surface.json`, narrowly-scoped tests, Freshchat docs/api metadata if needed | `feat/181-freshchat-cli-surface-metadata` |
| #182 | Help renderer/docs | #181 | help/docs renderer and docs surfaces | TBD |
| #183 | Stream runner | #181/#184 | Freshchat streams/schemas/fixtures/runner tests | TBD |
| #184 | Operation ledger | #181 | Freshchat `api_surface.json`, operation ledger artifacts | TBD |
| #185 | Direct reads | #181/#184 | direct-read engine/metadata/tests | TBD |
| #186 | Advanced query/binary engine | #184/#185 | engine/query/binary policy metadata/tests | TBD |
| #187 | Sensitive/admin policy | #184 | Freshchat writes/operations policy/tests | TBD |

## Execution plan

1. Parent setup checkpoint
   - Create this parent plan/TDD/verification/orchestration state.
   - Commit and push the parent planning checkpoint on `feat/180-freshchat-cli-parity`.
   - Open a draft parent PR to `main` with `Refs #180`.
2. Issue #181 local critical path
   - Because the current Pi harness exposes no `subagent` tool, record `not_spawned_runtime_capability_missing` and execute #181 inline or on an isolated branch from the parent branch.
   - Add a red test proving Freshchat lacks embedded command-surface metadata.
   - Add validated Freshchat `cli_surface.json` that maps existing safe streams/writes into app intents without raw API/write escape hatches.
   - Run focused validation and tests.
   - Commit/push #181 branch and open a stacked PR to the parent branch with `Refs #181` and `Refs #180`.
3. Later lanes
   - Do not implement broad direct reads, binary, admin/destructive expansion, or help renderer changes until their sub-issue plan/TDD ledgers exist.

## TDD slices

- #181 red: Freshchat bundle should expose non-nil command surface when loaded from embedded defs; fails before `cli_surface.json` exists.
- #181 green: add `cli_surface.json` and validate with `go test ./internal/connectors/engine -run TestBundleLoadEmbeddedFreshchatCLISurface` and `go run ./cmd/connectorgen validate internal/connectors/defs`.
- Follow-up red/green slices for #182-#187 must be recorded in their issue phase directories before production edits.

## CLI help/docs/website parity checklist

Applies to all CLI-visible lanes. For #181, command metadata becomes visible through the existing connector command runner/help path, but no help renderer/docs generator changes are planned in this first slice.

- [ ] `pm help <topic>` checked where the topic exists or marked blocked/not applicable.
- [ ] `pm freshchat` or connector command dispatch checked after metadata lands, without credentials.
- [ ] `pm freshchat <command> --help` checked if the help renderer supports it, otherwise deferred to #182.
- [ ] `docs/cli/**` updates handled in #182 or explicitly deferred.
- [ ] `website/**` updates handled in #182 or explicitly deferred.
- [ ] Generated help/manual artifacts handled in #182 or explicitly deferred.
- [ ] Tests cover metadata parsing/validation in #181.

## Safety gates

- No secrets in prompts, artifacts, logs, fixtures, or committed docs.
- No credentialed Freshchat checks unless explicitly requested.
- No reverse ETL execution; only plan/preview metadata and tests are allowed.
- Destructive/admin actions require typed confirmation policy and remain blocked until #187.
- No new dependencies.
- No push to `main`; parent PR merge is human-gated.

## Spawn decision

2026-07-09T19:37:13Z — `not_spawned_runtime_capability_missing` for parallel sub-issue workers. Evidence: current tool surface exposes `read`, `bash`, `edit`, and `write`; no Pi `subagent` tool is available in this harness. Coordinator will take local critical-path action for #181 after parent PR setup.
