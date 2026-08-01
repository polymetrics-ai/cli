# GSD Plan — issue #3215 BambooHR parity final wave

## Scope

Implement complete current official BambooHR (`bamboo-hr`) documented API parity using connector-local definition files, fixtures, CLI metadata, dedicated tests, docs, website generated connector data, and generated BambooHR surfaces only.

Authoritative sources:
- Parent #3215 and subissues #3216-#3222 via `gh-axi`.
- BambooHR docs index: `https://documentation.bamboohr.com/llms.txt`.
- BambooHR OpenAPI: `https://openapi.bamboohr.io/main/latest/docs/openapi/public-openapi.yaml` (OpenAPI 3.1.0, info.version 1.0), plus the top-level OpenAPI `webhooks` object.

## GSD path

- Ran `scripts/gsd doctor`: ok.
- Attempted required `scripts/gsd prompt programming-loop init --phase issue-3215-bamboohr-parity --dry-run`: unavailable (`unknown GSD command: programming-loop`).
- Fallback: used `scripts/gsd prompt gsd-quick "Implement BambooHR connector parity from issue 3215"` and this manual GSD loop with PLAN/TDD/VERIFICATION/RUN-STATE artifacts before production edits.

## Required skills loaded

- `golang-how-to`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-testing`
- `golang-cli`
- `golang-documentation`
- `golang-context`
- `golang-concurrency`
- `golang-lint`
- GSD references: required-skills routing, gsd-pi-adapter, CLI help/docs/website parity
- Connector references: `docs/migration/HANDOFF-CODEX.md`, `docs/migration/conventions.md`, `docs/architecture/connector-architecture-v2-design.md`

## Safety constraints

- No live BambooHR provider calls and no credentials.
- No provider writes outside fixture replay.
- No generic HTTP method/path/body/query passthrough.
- No shared runtime behavior edits.
- Destructive operations are named typed writes with closed schemas, risk text, `confirm: destructive`, redaction where path/body fields can expose sensitive identifiers, and documented idempotent missing-status handling only where safe.
- Binary/file/multipart/XML/inbound webhook operations remain blocked with exact source/rationale where the existing connector-local runtime cannot safely execute them.
- Reverse ETL remains plan → preview → explicit approval → execute.

## Implementation slices

1. **Red test / inventory lock** — done
   - Added `cmd/connectorgen/bamboo_hr_surface_test.go` to lock the current official 316-operation inventory, operation-ledger v1, no legacy `excluded` rows, duplicate-free coverage, exact implemented counts, and blocker models.
   - Red state: failed on `operation_ledger_version = 0` before production definition edits.

2. **Official inventory + ledger refresh** — done
   - Replaced `api_surface.json` with operation-ledger v1 rows for all 316 current OpenAPI operations (310 path operations + 6 top-level webhooks).
   - Removed 47 stale local endpoint rows and added 23 missing official rows.
   - Added typed blocked rows with source/rationale for unsupported binary, multipart/file, login, and inbound-webhook operations.

3. **Read/direct coverage** — done
   - Expanded fixture-backed streams to 138 JSON-compatible read operations.
   - Added 9 bounded fixed-target JSON direct-read command surfaces and 6 `operations.json` specs for POST direct reads.
   - Direct-read POST body schemas are root-closed to avoid arbitrary request-body escape hatches.

4. **Write coverage** — done
   - Expanded typed reverse-ETL writes to 149 JSON/no-body mutations with synthetic fixtures.
   - Kept destructive/remove/delete/clear operations approval-gated with `confirm: destructive` and risk text.
   - Removed stale write fixtures/actions whose official endpoints are no longer current.

5. **Docs/CLI/certification evidence** — done
   - Updated BambooHR `docs.md`, generated connector manual/SKILL, dynamic CLI golden transcript, website generated connector data/catalog, `cli_surface.json`, `operations.json`, `spec.json`, `metadata.json`, and fixture-only `certification.json`.
   - Certification remains fixture-only/uncertified: no credentials, no provider calls, no live certification claim.

6. **Verification and issue update** — done locally before commit
   - Focused inventory, conformance, CLI/golden, docs/help, website-generation, connector-boundary, diff-check, and `make verify` gates passed on the final run.
   - Parent/subissue GitHub status was updated via `gh-axi` after local verification.
   - Final step: local clean commit only; no push/PR/merge.
