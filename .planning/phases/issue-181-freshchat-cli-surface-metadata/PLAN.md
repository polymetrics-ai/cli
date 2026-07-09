# Issue #181 — Freshchat CLI surface metadata

Issue: https://github.com/polymetrics-ai/cli/issues/181
Parent: https://github.com/polymetrics-ai/cli/issues/180
Planned branch: `feat/181-freshchat-cli-surface-metadata`
Base branch: `feat/180-freshchat-cli-parity`

## GSD command path

- `scripts/gsd prompt plan-phase issue-181-freshchat-cli-surface-metadata --skip-research` — generated and followed.
- `scripts/gsd prompt programming-loop init --phase issue-181-freshchat-cli-surface-metadata --dry-run` — unavailable because `programming-loop` is not registered in `scripts/gsd list --json`.
- Manual-GSD fallback: use `.pi/prompts/pm-gsd-loop.md` and `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`.

## Required skills loaded

`gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-lint`.

## Scope

Create Freshchat command-surface metadata that maps safe provider/app intents onto the existing connector streams and write actions. This issue does not add new direct-read execution, binary transfer, advanced query/body support, or sensitive/admin execution.

Allowed production write scope:

- `internal/connectors/defs/freshchat/cli_surface.json`.
- Narrow tests that prove Freshchat CLI surface metadata loads/validates (expected: `internal/connectors/engine/bundle_test.go` or equivalent).
- Freshchat docs/API metadata only if needed to keep validation honest.

Out of scope:

- Credentialed Freshchat checks.
- Runtime execution of Freshchat writes.
- New engine direct-read output policies.
- Generic raw HTTP, shell, SQL, or arbitrary mutation tools.
- Help renderer/docs/website implementation (tracked by #182).

## Official source baseline

Official Freshchat docs fetched from `https://developers.freshchat.com/api/` for planning. Sanitized extraction found 34 documented operation rows matching the current bundle baseline, with one docs typo/example (`GET /metric`) ignored and `GET /users/{user_id}/conversations` retained from the official nav/body.

See `traces/official-surface-2026-07-09.md` for the sanitized operation list. Do not commit raw Freshchat docs HTML; it contains secret-shaped Authorization examples.

## Implementation plan

1. Red
   - Add a focused engine bundle test requiring embedded Freshchat command-surface metadata.
   - Run `go test ./internal/connectors/engine -run TestBundleLoadEmbeddedFreshchatCLISurface` and capture failure.
2. Green
   - Add `internal/connectors/defs/freshchat/cli_surface.json`.
   - Map existing stream-backed commands to `intent: etl` and `availability: implemented` only where commandrunner can execute safely using existing stream/config semantics.
   - Map existing reverse-ETL write actions to `intent: reverse_etl`; use `implemented` when required fields can be safely mapped to primitive flags, and `partial` when the write requires structured object/array payload fields that the CLI flag mapper does not express yet.
   - Mark file/image upload commands as `direct_write` + `unsupported_api` or `unsafe_or_disallowed` (not executable) because they require bounded binary/multipart policy in #186.
   - Mark `/users/fetch` as `direct_read` + `planned` or `unsupported_api` because safe POST-body direct-read support belongs to #185/#186.
   - Include `api_surface` references only when they match existing `api_surface.json` executable rows.
3. Refactor/verify
   - `gofmt -w` edited Go test file.
   - `go test ./internal/connectors/engine -run TestBundleLoadEmbeddedFreshchatCLISurface`.
   - `go run ./cmd/connectorgen validate internal/connectors/defs`.
   - Focused CLI/commandrunner tests if commandrunner behavior changes (not expected in #181).

## CLI help/docs/website parity

This slice adds metadata consumed by existing connector command handling. It does not implement the Freshchat help renderer/docs parity surface; that remains #182.

Checklist for #181:

- [ ] `cli_surface.json` validates and is embedded.
- [ ] No help renderer files changed.
- [ ] No `docs/cli/**` or `website/**` changes unless existing tests demand them.
- [ ] PR body records #182 as the help/docs parity follow-up.

## Safety

- No secrets in examples; use placeholders such as `user_123`, `conv_123`, and dates.
- No credentialed commands.
- No reverse ETL execution.
- No binary upload/download execution.
- No new dependencies.
