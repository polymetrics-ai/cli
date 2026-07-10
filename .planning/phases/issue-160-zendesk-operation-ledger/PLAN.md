# Plan: Zendesk Operation Ledger

Parent issue: #156
Sub-issue: #160
Parent branch: `feat/156-zendesk-cli-parity`
Stack base: `feat/157-zendesk-cli-surface-metadata` until #157 review coverage/integration is resolved.
Sub-issue branch: `feat/160-zendesk-operation-ledger`

## GSD Command Path

- GSD health: `scripts/gsd doctor`, `scripts/gsd verify-pi`, `scripts/gsd list --json` passed in this session.
- Planning prompt: `scripts/gsd prompt plan-phase issue-156-zendesk-complete-implementation --skip-research`.
- Execution prompt: `scripts/gsd prompt execute-phase issue-160-zendesk-operation-ledger --plan 1`.
- Programming loop prompt attempted: `scripts/gsd prompt programming-loop init --phase issue-156-zendesk-complete-implementation --dry-run`; adapter returned `unknown GSD command: programming-loop`, so the manual GSD/TDD fallback is active and recorded here.

## Required Skills Loaded

`gsd-core`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-documentation`.

## Objective

Account every official Zendesk OAS operation in durable runtime-readable metadata so later stream, direct-read, binary, and reverse-ETL lanes can enable exact typed commands without raw generic HTTP/write escape hatches.

## Scope

- Add `internal/connectors/defs/zendesk/operations.json` containing all 617 official operations exactly once.
- Update `api_surface.json` coverage from candidate-only rows to exact ledger mapping:
  - safe structured reads covered by `direct_reads` and/or later streams;
  - binary/file reads covered by binary operation ids and explicit size policy;
  - mutating operations covered by typed reverse-ETL operation ids but still blocked from execution until write schemas/policy land;
  - deprecated operations explicitly excluded as deprecated.
- Keep executable behavior conservative: no credentialed Zendesk checks, no reverse ETL execution, no generic raw HTTP writes.

## Non-goals

- Do not add new dependencies.
- Do not fetch or store credentials.
- Do not execute Zendesk requests.
- Do not mark writes executable in `cli_surface.json`; #163 owns typed write action enablement.
- Do not implement broad raw API dispatch.

## TDD / Red-Green Plan

1. Add a red test proving the embedded Zendesk bundle must have an operation ledger with exactly 617 operations and that every `api_surface` endpoint is covered or explicitly excluded.
2. Generate/author `operations.json` from the official OAS already cached in `/tmp/zendesk-oas.yaml`, re-fetching to `/tmp` only if missing.
3. Update `api_surface.json` coverage fields to reference operation ids/exclusions while preserving blocked-by-default safety text.
4. Run focused tests, schema validation, and connector validation.
5. Commit the green operation-ledger slice before moving to direct-read/stream/write enablement.

## Safety Rules

- All Zendesk operations remain either metadata-only or routed through existing connector guardrails.
- Mutating operations must carry risk, approval, mutation class, and operation metadata; execution remains blocked until #163.
- Binary operations must declare bounded `max_bytes`, no overwrite, and no archive extraction.
- Deprecated operations must remain excluded and non-executable.
- No secrets, no credentialed connector checks, and no broad generated-file rewrites outside this issue scope.

## CLI Help / Docs / Website Parity

This issue updates connector metadata but does not add user-facing runtime help behavior. Docs/catalog regeneration is limited to connector generated artifacts if validation requires it. #158 owns complete help/docs/website parity.

## Verification

Targeted:

```bash
go test ./internal/connectors/engine -run 'Zendesk|Operation' -count=1
go test ./cmd/connectorgen -run 'Operations|Surface|Zendesk' -count=1
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
