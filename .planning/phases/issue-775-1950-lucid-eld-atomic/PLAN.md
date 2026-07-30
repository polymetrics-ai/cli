# GSD Plan — Issue #1950 Lucid ELD Atomic Pilot Bundle

## Phase identity

- Phase path: `.planning/phases/issue-775-1950-lucid-eld-atomic/`
- Child issue / parent: #1950 / #775
- Branch: `feat/1950-lucid-eld-operation-ledger`
- PR: #3166, base `feat/775-lucid-eld-full-parity`
- Execution decision: `local_critical_path` (isolated worker cwd already assigned; recursive subagents blocked for worker role)

## Corrective scope authorized by parent PM orchestrator

#1950 originally delivered only `internal/connectors/defs/lucid-eld/api_surface.json`. CI proved that an api-surface-only connector-def directory is not an independently green bundle. This corrective cycle treats #1950 as the atomic pilot/bootstrap child for Lucid ELD and pulls forward only the minimum necessary acceptance criteria from sibling issues so `internal/connectors/defs/lucid-eld/` is a Tier-1, conformance-valid declarative bundle.

Allowed production scope:

- `internal/connectors/defs/lucid-eld/**`
- Required generated/help parity artifacts for the new connector: `docs/cli/connectors.md`, `docs/connectors/{README.md,catalog/**,lucid-eld/**}`, internal golden transcripts/count tests.
- Generic `cmd/connectorgen` single-bundle validation-path repair required by the exact local gate.
- No provider-specific Go, hooks, natives, new dependencies, live credentialed calls, or `writes.json`.

Allowed planning scope:

- `.planning/phases/issue-775-1950-lucid-eld-atomic/**`
- Preserve/move old `.planning/issue-775/1950/**` evidence and planning files into this phase path.
- Do not edit parent `.planning/issue-775/STATE.md` or parent branch artifacts.

Pulled-forward scope (siblings remain open):

- #1951: minimum spec/auth + CLI surface metadata needed for validation.
- #1952: three streams (`drivers`, `vehicles`, `vehicle_location_history`) with passthrough schemas/fixtures.
- #1953: five direct reads via `operations.json` + `cli_surface.json`.
- #1954: confirmed no writes; omit `writes.json`, set `metadata.capabilities.write=false`.
- #1955: required connector `docs.md` and partial certification/default-source decision; deeper docs/certification audit remains open.

## GSD mode and fallback

- `scripts/gsd doctor` passed on 2026-07-30.
- Required adapter command attempted: `scripts/gsd prompt programming-loop init --phase issue-775-1950-lucid-eld-atomic --dry-run`.
- Exact fallback result: `scripts/gsd: unknown GSD command: programming-loop`.
- Manual GSD fallback active per `.agents/agentic-delivery/references/gsd-pi-adapter.md` and issue-agent contract.
- Red evidence is the already-failing CI/current branch state: `connector-boundary` fails on missing `metadata.json`; `gsd-workflow-evidence` fails because evidence lived under unrecognized `.planning/issue-775/1950/**`; PR verify fails as a consequence of incomplete bundle. Re-confirmed local red commands before production edits.

## Required reading and skills loaded

Required docs read: `AGENTS.md`, issue #1950 body/AC, `.agents/agentic-delivery/contracts/issue-agent-contract.md`, `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`, `.agents/agentic-delivery/references/required-skills-routing.md`, `.agents/agentic-delivery/references/gsd-pi-adapter.md`, `docs/migration/HANDOFF-CODEX.md`, full `docs/migration/conventions.md`, `docs/architecture/connector-architecture-v2-design.md`, `docs/migration/connector-boundary-guard.md`, `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`, `.planning/{config.json,PROJECT.md,ROADMAP.md,STATE.md}`, `docs/plans/universal-programming-loop-prd.md`, `docs/prompts/universal-programming-loop-prompts.md`, and existing phase artifacts.

Skills loaded: `gsd-core`, `caveman`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-spf13-viper`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-lint`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-documentation`.

Note: `.pi/skills/go-implementation/SKILL.md` is not present in this checkout; required Go implementation routing is satisfied by the repo routing reference plus loaded global Go skills.

## Authoritative technical ground truth

- Official source: DriveHOS/Lucid ELD Partner API v2 Swagger/OpenAPI 2.0, `https://api.drivehos.app/partner/swagger/doc.json`, fetched 2026-07-30, SHA-256 `1b3756f4c69c9133e24754a856d2fe9ec2b08768edd5dec25b899f564ddb7ec4`.
- Auth: required headers `X-API-Provider-Key` and `X-API-Company-Key`, both credential-shaped and marked `x-secret: true` in `spec.json`.
- Response envelope: `data`, `description`, `next_page_token`, `page`, `size`, `status_code`, `total_elements`, `total_pages`.
- Official API documents 8 GET operations, zero POST/PUT/PATCH/DELETE, zero webhooks, zero reports, zero binary/media operations.
- WithTerminal evidence is capability-only, not DriveHOS wire-schema evidence. Do not use Terminal normalized model fields as Lucid ELD schema properties.

## Slice plan

### Slice 1 — Planning relocation and manual-red evidence

- Move `.planning/issue-775/1950/**` to `.planning/phases/issue-775-1950-lucid-eld-atomic/**`.
- Trim large verbatim remote-doc snippets while retaining URLs, fetched dates, SHA-256 hashes, endpoint table, and concise capability reconciliation.
- Update `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `SUMMARY.md`, `PROMPTS.md`, `RUN-STATE.json` before production edits.
- Record red evidence: `scripts/verify-gsd-workflow feat/775-lucid-eld-full-parity` exit 1; `make connector-boundary` exit 2.
- Commit checkpoint: planning relocation/fallback evidence after diff check.

### Slice 2 — Atomic Tier-1 bundle

- Add `metadata.json`, `spec.json`, `streams.json`, `operations.json`, `cli_surface.json`, three schemas, stream fixtures, `fixtures/check.json`, and `docs.md` under `internal/connectors/defs/lucid-eld/`.
- Preserve existing `api_surface.json` classifications unchanged.
- Omit `writes.json`.
- Use `projection: "passthrough"` and intentionally open object schemas for all three streams because official `data` is untyped `{}` and no sample/live response exists.
- Synthetic fixtures exercise envelope/pagination plumbing only and must not claim real provider field truth.

### Slice 2a — Generic validation-path repair if required

- If the required command `go run ./cmd/connectorgen validate internal/connectors/defs/lucid-eld` treats `schemas/` and `fixtures/` as sibling connectors, repair `cmd/connectorgen` generically so a directory containing `metadata.json` is validated as one bundle rooted at its parent/name.
- This is not provider-specific connector logic; red evidence is the exact failing command output in `TDD-LEDGER.md`.
- Add/adjust focused `cmd/connectorgen` test coverage before claiming green.

### Slice 3 — Focused green gates

Run and record:

```bash
go run ./cmd/connectorgen validate internal/connectors/defs/lucid-eld
go test ./internal/connectors/conformance -run 'TestConformance/lucid-eld' -count=1
go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1
go vet ./internal/connectors/... ./internal/cli/...
go build ./cmd/pm
make connector-boundary
scripts/verify-gsd-workflow feat/775-lucid-eld-full-parity
git diff --check feat/775-lucid-eld-full-parity...HEAD
gofmt -l cmd internal
```

Also run `make verify`; record exact pass/fail and name any unrelated baseline failures precisely.

### Slice 4 — PR update, commit, push, handoff

- Commit coherent green slices (planning move, bundle completion, verification/PR update as needed).
- Push only `feat/1950-lucid-eld-operation-ledger`.
- Update PR #3166 title/body with `Refs #1950` / `Refs #775`, atomic-pilot scope decision, manual-GSD fallback, loaded skills, safety notes, and exact verification.
- Do not request `@claude review`; orchestrator owns review routing.

## Acceptance checklist

- [x] Required reading completed before edits.
- [x] Required skills loaded and recorded.
- [x] Manual-GSD fallback recorded with exact command/result.
- [x] Red CI/local evidence captured before production edits.
- [x] Planning evidence relocated to `.planning/phases/issue-775-1950-lucid-eld-atomic/**` and old child phase path removed.
- [x] Lucid ELD bundle is a complete Tier-1 declarative bundle.
- [x] All 8 official GET operations executable via stream or direct-read metadata.
- [x] `writes.json` absent; write capability false.
- [x] Schemas avoid invented properties and stream projection is passthrough.
- [x] Synthetic fixture caveat documented.
- [x] Required verification commands pass; `make verify` passed.
- [ ] PR #3166 updated and branch pushed.
