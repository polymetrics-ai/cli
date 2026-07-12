# Plan: Freshdesk CLI Surface Metadata

Parent issue: #172
Sub-issue: #173
Branch: `feat/172-freshdesk-cli-parity` (inline parent critical path until a stacked sub-branch is warranted)
Connector: `freshdesk`

## GSD / Skills

- GSD adapter: `scripts/gsd doctor`, `scripts/gsd verify-pi`, and `scripts/gsd list --json` passed.
- GSD programming-loop fallback: `scripts/gsd prompt programming-loop init --phase issue-172-freshdesk-cli-parity --dry-run` is not registered; following `.pi/prompts/pm-gsd-loop.md` and `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` manually.
- Required skills loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-documentation`.
- CLI parity reference loaded: `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`.

## Scope

Create/refresh Freshdesk metadata and CLI/API surface inventory. This lane may edit:

- `internal/connectors/defs/freshdesk/metadata.json`
- `internal/connectors/defs/freshdesk/api_surface.json`
- `internal/connectors/defs/freshdesk/cli_surface.json`
- `internal/connectors/defs/freshdesk/docs.md` only for metadata truthfulness and no-overclaiming
- `.planning/phases/issue-173-freshdesk-cli-surface-metadata/**`

Out of scope for this lane: implementing new stream runner behavior, direct-read execution, writes, binary download execution, or destructive/admin execution.

## Steps

1. Parse the public Freshdesk API reference without credentials and produce an operation inventory matching the 170-operation parent baseline.
2. Replace the legacy 10-entry `api_surface.json` with an honest full-surface ledger. Existing implemented streams stay `covered_by.stream`; unimplemented operations receive precise `operation` blocked metadata or safe exclusions only when duplicate/deprecated/disallowed/out-of-scope.
3. Add `cli_surface.json` with provider-inspired Freshdesk command groups. Implemented commands map only to current streams; unimplemented commands are `planned`, `excluded`, `unsafe_or_disallowed`, or `unsupported_api` with notes.
4. Update metadata risk/capability language only where the current metadata overclaims or under-documents CLI parity state.
5. Validate JSON, bundle loading, and static rules.

## Red Evidence Already Captured

- Current `api_surface.json` has 10 endpoint entries, expected 170.
- `internal/connectors/defs/freshdesk/cli_surface.json` is absent.

## Green Definition

- `api_surface.json` endpoint count is 170 and every entry has exactly one coverage/classification path.
- `cli_surface.json` exists and references only existing streams, writes, operations, or valid API rows.
- `engine.Load(defs.FS, "freshdesk")` exposes non-nil `CLISurface` after embed.
- `go run ./cmd/connectorgen validate internal/connectors/defs` passes.

## Verification Checklist

- [ ] JSON parse/count check for Freshdesk `api_surface.json` and `cli_surface.json`.
- [ ] `go test ./internal/connectors/engine -run 'CLISurface|Freshdesk'` as applicable.
- [ ] `go test ./cmd/connectorgen -run CLISurface`.
- [ ] `go test ./cmd/connectorgen ./internal/connectors/engine`.
- [ ] `go test ./internal/connectors/conformance -run 'TestConformance/freshdesk'`.
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs`.
- [ ] If docs/help behavior changes later, run `pm help <topic>`, `pm <namespace>`, `pm <command> --help`, and docs/website grep/generator checks.

## Safety

- No secrets or credentialed checks.
- No raw generic HTTP write or direct write escape hatch.
- No reverse ETL execution.
- No new dependencies.
- Do not classify sensitive/admin/destructive endpoints as impossible merely because they are risky; ledger them for later gated reverse-ETL handling.
