# PLAN — Google Search Console parity wave03 (#3038–#3045)

## GSD activation

- Requested command path: `scripts/gsd prompt programming-loop init --phase issue-3038-google-search-console-parity-wave03 --dry-run`.
- Adapter result: `scripts/gsd` returned `unknown GSD command: programming-loop` even though `scripts/gsd doctor` passed. Manual-GSD fallback is recorded here and follows `.pi/prompts/pm-gsd-loop.md` plus `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`.
- Orchestration decision: `local_critical_path` — firstmate already assigned this single connector wave in an isolated worktree; no mutating subagents spawned because the write scope is one connector plus generated docs/catalog surfaces.
- Required skills loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`.
- Required repo references read: `AGENTS.md`, required skill routing, GSD adapter, CLI/help/docs/website parity, issue/parent contracts, parent orchestration references, connector migration handoff/conventions/design, universal runtime loop, project planning files, and issue bodies #3038–#3045 via `gh-axi`.

## Objective

Complete documented Google Search Console connector parity for parent #3038 and children #3039–#3045 without live provider calls, credentials, certification claims, pushes, PR updates, no-mistakes runs, VPS, or Thaalam changes.

## Current audited baseline

- Official source re-audit saved in `research/official-source-audit.md` and `research/official-operation-inventory.json`.
- The 2026-07-29 discovery document still has exactly 11 operations: 4 ETL reads, 4 reverse ETL writes, 3 bounded direct/provider-search/query reads, 0 binary, 0 CDC, 0 excluded/not-applicable.
- Existing bundle has 9 streams, 4 write actions, 7 fixture rows, no `operations.json`, no `cli_surface.json`, and no `certification.json`.

## Slices

1. **Red contracts before production edits**
   - Add a connector-owned conformance contract test that fails on the current bundle: missing `operations.json`, missing implemented direct-read commands for Search Analytics, URL Inspection, Mobile Friendly Test, incomplete official operation coverage, missing stream fixtures for the GET detail streams, missing certification metadata, and missing destructive redaction.
   - Run the focused test and record red evidence in `TDD-LEDGER.md`.
2. **Operation/direct-read/CLI metadata**
   - Add `operations.json` with three bounded `rest_read` POST operations: `google-search-console.searchanalytics_query`, `google-search-console.urlinspection_index_inspect`, and `google-search-console.mobile_friendly_test_run`.
   - Add `cli_surface.json` with implemented, credential-safe provider commands for the 4 ETL streams, 3 direct reads, and 4 reverse actions; no raw method/path/body or generic query passthrough.
   - Keep provider query/search distinct from warehouse `pm query`; keep `metadata.capabilities.query=false` unless a shared foundation explicitly changes it.
3. **Ledger and safety**
   - Refresh `api_surface.json` from the official discovery inventory, using de-duplicated official operations and explicit direct-read coverage. Because the bundle still intentionally exposes five dimension-specific Search Analytics ETL streams for the one Search Analytics official operation, keep truthful surface rows for validator completeness and document the de-duplication in scope/docs.
   - Add `operation_ledger_version: 1` if compatible with validator mode; do not use legacy `excluded` in ledger mode.
   - Add destructive redaction (`site_url`, `feedpath`) and additional-property closed write schemas where applicable.
4. **Fixtures and certification metadata**
   - Add sanitized fixtures for `site_details` and `sitemap_details` so all four official GET reads have stream fixtures.
   - Add `certification.json` with fixture-only safe defaults/direct-read candidates/write pairings where schema-valid; do not claim live certification.
5. **Docs, generated surfaces, and website/catalog**
   - Update `docs.md` with exact official counts, direct-read commands, Search Analytics stream-vs-operation de-duplication, fixture-only uncertified status, and safety notes.
   - Regenerate `docs/connectors/**`, connector catalog docs, and website generated connector data after bundle changes.
   - Update CLI help/golden expectations if generated dynamic help changed.
6. **Issue addendum**
   - Through `gh-axi`, idempotently append the captain-policy addendum marker to #3038–#3045 with truthful post-change counts and no certification claim.
7. **Verification and commit**
   - Required gates: `go run ./cmd/connectorgen validate internal/connectors/defs/google-search-console`, `go test ./internal/connectors/conformance -run 'TestConformance/google-search-console' -count=1`, `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1`, `go build ./cmd/pm`, `make connector-boundary`, `make verify`, `git diff --check`.
   - Commit cleanly on `fm/cli-google-search-console-parity-wave03-r1`; do not push or open/update PR.

## Completion status

- Slices 1–7 completed locally on `fm/cli-google-search-console-parity-wave03-r1`.
- Required gates passed and recorded in `VERIFICATION.md`.
- Issue addendum marker appended once to #3038–#3045 through `gh-axi`.
- No push, PR, merge, live provider call, credential request, or certification claim was made.

## Safety boundaries

- No secrets, credential prompts, live provider operations, live writes, certification claims, new dependencies, generic raw API/write/query/shell tooling, or quality-gate reductions.
- Reverse ETL remains plan → preview → explicit approval → execute.
- Destructive/delete actions must be typed, closed-schema, redacted, idempotent where supported, and `confirm: "destructive"`.
