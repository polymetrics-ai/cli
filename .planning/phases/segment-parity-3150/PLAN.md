# Segment connector official API parity (#3150-#3157)

## GSD activation

- Adapter: `scripts/gsd doctor` already passed in this session.
- Prompt loaded: `scripts/gsd prompt gsd-plan-phase "Segment connector parity #3150-#3157" --skip-research` and `scripts/gsd prompt gsd-execute-phase "Segment connector parity"`.
- Manual fallback: repo-local GSD has no `programming-loop` prompt name in `scripts/gsd list`; this plan records the manual GSD/TDD loop before production edits.
- Required skills used: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-lint`, `context-mode`.

## Scope

Implement definition-owned Segment Public API parity from the official Redocly OpenAPI surface (OpenAPI 3.0.3, Segment Public API 73.0.8, 125 paths, 197 operations):

- #3151 ledger: every official operation appears exactly once in `api_surface.json` and `operations.json`.
- #3152 ETL/CDC: fixture-backed declarative streams for 79 ETL GET operations plus `audit_events` CDC/changefeed.
- #3153 direct/binary: 19 bounded direct/provider query commands and a blocked binary-download operation row for `createDownload`.
- #3154 reverse ETL: 96 named write actions with fixed method/path, schemas, risk text, plan/preview/approval execution through existing reverse ETL gates.
- #3155 CLI/config/help: definition-owned `cli_surface.json`, runtime help, generated docs/skills/manual parity.
- #3156 fixtures/guard: sanitized fixtures, conformance, Connector Guard, and no shared guardrail weakening.
- #3157 certification/release: fixture-only evidence remains uncertified; no live provider calls or credentials.

## Constraints

- Do not push, open/update PRs, merge, or invoke `/no-mistakes`.
- Do not run live Segment calls, credentials, secrets, VPS, certification, or Thaalam changes.
- Keep production changes connector-local except owned tests, docs, generated help/manual artifacts, planning artifacts, and a behavior-preserving bundleregistry cache needed to keep the enlarged declarative bundle set within local test timeouts.
- No generic HTTP method/path/body/query, shell, file, SQL, arbitrary event/body, or passthrough escape hatch.
- Reverse ETL remains plan -> preview -> explicit approval -> execute; deletes require destructive confirmation.

## Implementation slices

1. Generate Segment definitions from the audited OpenAPI operation inventory.
   - Replace stale v73.0.0/187-operation ledger with v73.0.8/197-operation ledger.
   - Remove stale `/workspaces` coverage in favor of official `/` and `/spaces` operations.
   - Keep auth as Bearer `api_token` plus `base_url` default.
2. Add executable surfaces.
   - Streams: 80 fixture-backed reads (79 ETL + `audit_events`) using Segment cursor query shape (`pagination.count`, `pagination.cursor`) when list-like.
   - Direct reads: 19 fixed-target GET/POST operation commands with JSON redaction and max-byte bounds.
   - Writes: 96 fixed-target write actions, path fields declared, record schemas generated from path/request schemas, DELETE actions destructive-confirmed.
   - Binary: `createDownload` remains a typed blocked binary operation row; no local download executor in this slice.
3. Add focused coverage.
   - Segment count/coverage test for 197 official operations and lane allocations.
   - Segment conformance and connectorgen validation.
   - Focused CLI/help tests via existing golden transcripts and direct Segment help/inspect commands.
4. Regenerate docs/skills/manual/help surfaces.
5. Run gates and commit locally only.

## Verification checklist

See `VERIFICATION.md`; all required local gates passed before the final commit.

## Completion notes

- Segment now tracks the official v73.0.8 OpenAPI surface: 197 API endpoints, 197 operation rows, 80 fixture-backed streams, 96 fixed write actions, 19 bounded direct reads, 1 blocked binary/file workflow, and 1 disallowed echo/test endpoint.
- The generated CLI surface intentionally documents representative stream/write groups plus every fixed direct-read operation (28 commands total) while full operation parity is enforced in `api_surface.json`/`operations.json` and docs.
- Reverse ETL writes remain fixed-action only and continue through the shared plan/preview/approval flow; no raw method/path/body escape hatch was added.
