# Plan: Zendesk CLI Surface Metadata

Parent issue: #156
Sub-issue: #157
Parent branch: `feat/156-zendesk-cli-parity`
Sub-issue branch: `feat/157-zendesk-cli-surface-metadata` (create after parent PR opens)

## GSD Command Path

- Planning inherited from: `scripts/gsd prompt plan-phase issue-156-zendesk-cli-parity --skip-research`.
- Execution prompt generated: `scripts/gsd prompt execute-phase issue-157-zendesk-cli-surface-metadata --plan 1`.
- Manual programming-loop fallback: `scripts/gsd prompt programming-loop init --phase issue-156-zendesk-cli-parity --dry-run` failed with `unknown GSD command: programming-loop`; follow the universal manual GSD/TDD loop.

## Required Skills Loaded

`gsd-core`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-lint`.

## Objective

Create the initial safe Zendesk connector bundle metadata and CLI/API surface inventory from the official Zendesk OAS without enabling unsafe runtime dispatch or credentialed checks.

## Inputs

- Official source: `https://developer.zendesk.com/zendesk/oas.yaml`.
- Baseline: 617 operations across 429 paths; methods GET=320, PUT=89, POST=110, DELETE=85, PATCH=13.
- Current repo state: no `internal/connectors/defs/zendesk/` bundle; related product-specific bundles (`zendesk-support`, `zendesk-chat`, `zendesk-sunshine`, `zendesk-talk`) already exist and must not be modified in this slice.

## Implementation Steps

1. Create the sub-issue branch from the parent branch after parent PR creation.
2. Capture red evidence that `internal/connectors/defs/zendesk/` and embedded runtime metadata are absent.
3. Download/parse the official OAS into a temporary file only; do not commit the raw OAS.
4. Generate/author `internal/connectors/defs/zendesk/` with:
   - `metadata.json` — identity, capabilities, and risk text for the umbrella Zendesk API metadata bundle.
   - `spec.json` — safe config/secret schema using `base_url`, `access_token`, `api_token`, and `email` patterns consistent with Zendesk auth precedence.
   - `streams.json` — minimal HTTP base/check scaffold with no streams enabled yet.
   - `api_surface.json` — every official operation listed exactly once as a blocked-by-default operation-ledger row with candidate model and source URL/notes.
   - `cli_surface.json` — provider/API-inspired command inventory with safe intents (`etl`, `direct_read`, `reverse_etl`, `docs_only`, or unsupported statuses) and no raw generic HTTP write command.
   - `docs.md` — required headings and honest metadata-only known limit.
5. Add focused test coverage if needed so `engine.Load(defs.FS, "zendesk")` loads the embedded metadata and CLI surface.
6. Validate with focused tests and `connectorgen validate`.

## Safety Rules

- No secrets in prompts, docs, fixtures, examples, or logs.
- No credentialed Zendesk checks.
- No raw generic HTTP write/direct-write/raw API command is implemented.
- All write-like operations remain blocked or planned until reverse-ETL schemas and approval gates are implemented in later issues.
- No new dependencies; use stdlib or existing tooling only.

## CLI Help / Docs / Website Parity

This slice is metadata-only and does not add runtime `pm zendesk` dispatch. Mark runtime help/docs/website parity as not applicable until #158, but ensure `cli_surface.json` metadata is safe for future help rendering.

## Verification

Targeted:

```bash
test -d internal/connectors/defs/zendesk
go test ./internal/connectors/engine -run 'CLISurface|Definition|Zendesk' -count=1
go test ./cmd/connectorgen -run 'CLISurface|Surface' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs
```

Before handoff:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```
