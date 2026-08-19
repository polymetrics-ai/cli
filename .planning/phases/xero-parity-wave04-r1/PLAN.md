# Xero connector parity wave04 plan (#3102-#3109)

## Scope

Implement connector-local Xero Accounting API parity evidence for parent issue #3102 and subissues #3103-#3109 on branch `fm/cli-xero-parity-wave04-r1`.

Allowed production paths for this slice:

- `internal/connectors/defs/xero/**`
- Xero-owned conformance/validation tests and fixtures
- Xero-generated connector docs/manuals/catalog/website data produced from the bundle
- Issue-local GSD artifacts under `.planning/phases/xero-parity-wave04-r1/**`

Do not edit shared runtime behavior, add dependencies, run live provider calls, request credentials, certify live behavior, push, open/update a PR, merge, or touch VPS/Thaalam/Herdr lifecycle surfaces.

## GSD command path and fallback

- `scripts/gsd doctor`: passed.
- `scripts/gsd list`: passed; adapter listed 69 commands.
- Required implementation command attempted: `scripts/gsd prompt programming-loop init --phase xero-parity-wave04-r1 --dry-run`.
- Result: `scripts/gsd: unknown GSD command: programming-loop`.
- Manual GSD fallback is active for this phase using `.pi/prompts/pm-gsd-loop.md`, `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`, `docs/plans/universal-programming-loop-prd.md`, and `docs/prompts/universal-programming-loop-prompts.md`.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-testing`
- `golang-context`
- `golang-concurrency`
- `golang-cli`
- `golang-documentation`
- `vercel-react-best-practices` and `vercel-composition-patterns` for generated website data parity
- CLI help/docs/website parity reference
- context-mode for large-output processing

## Source inventory

Official source for this slice is the provider-owned Xero Accounting OpenAPI document only:

- `https://raw.githubusercontent.com/XeroAPI/Xero-OpenAPI/master/xero_accounting.yaml`
- Supporting provenance links named by #3102: manifest, webhooks, and master commit metadata.

Local re-audit source snapshot (fetched without credentials or provider API calls): OpenAPI `3.0.0`, `info.version=16.1.0`, 235 HTTP operations (`GET=126`, `PUT=53`, `POST=46`, `DELETE=10`). Operation-lane classification is deterministic from the official accounting paths: reports are direct/provider reads (11), attachment and PDF endpoints are binary/file lane (59), remaining GETs are ETL reads (78), and remaining mutations are reverse-ETL writes (87). This intentionally corrects the landed r2 table's attachment-only binary count by treating the four official `/pdf` GET operations as file-returning binary operations.

## Implementation slices

1. **Operation ledger parity**
   - Replace Xero's stale `/api.xro/2.0`-prefixed `api_surface.json` rows with the exact 235 official Accounting API paths.
   - Add `operation_ledger_version: 1` and source-linked notes.
   - Ensure each official operation is partitioned once by lane: stream, write, direct read, binary metadata, or blocked shared-runtime dependency.

2. **Reverse ETL parity**
   - Add the two missing official BankTransfers destructive POST write actions.
   - Close Xero write record schemas (`additionalProperties: false`) and keep destructive actions gated by `confirm: "destructive"`.
   - Generate sanitized write fixtures for every executable write action.

3. **Direct and binary operation metadata**
   - Add bounded `operations.json` entries for all report direct reads and all attachment binary/file operations.
   - Add implemented provider-style CLI metadata for bounded report direct reads only; binary download/upload commands remain blocked unless the shared binary executor exists.

4. **Fixtures and conformance coverage**
   - Generate sanitized stream fixtures for all Xero streams.
   - Preserve bounded pagination fixtures for paginated streams.
   - Add validation tests that assert official counts, fixture coverage, and no legacy exclusions.

5. **Docs, generated surfaces, and certification truthfulness**
   - Update `docs.md`, add a fixture-only `certification.json`, regenerate connector manuals/catalogs and website generated data.
   - Report fixture-only evidence as uncertified; no live provider certification is claimed.

6. **Issue addendum**
   - Append the idempotent captain-policy addendum to #3102-#3109 using `gh-axi` after local counts are known.

## Implementation summary

- `api_surface.json` now contains exactly 235 unprefixed official Accounting API operation rows with `operation_ledger_version: 1` and no legacy `excluded` rows.
- Xero writes now cover all 87 non-attachment mutations, including the two missing destructive BankTransfers status-delete POST operations.
- `operations.json` records 70 bounded direct/binary/file operations: 11 report reads, 11 attachment metadata reads, 26 binary/PDF downloads, and 22 attachment uploads.
- `cli_surface.json` exposes implemented bounded report direct-read commands and marks binary/file transfer commands as planned/blocked on the shared executor.
- Synthetic fixtures cover 100 stream fixture directories and 87 write fixtures; no live Xero credential or provider call was used.

## Orchestration decision

`local_critical_path`: this task assigns one coupled connector-owned tree and issue bundle to one isolated worktree. Mutating subagents were not spawned because all production edits converge on `internal/connectors/defs/xero/**` plus generated surfaces, and parallel mutating workers would collide. Read-only recon is performed with local context-mode tooling.

## Human gates / non-goals

- No secrets, credential requests, live Xero calls, provider writes, certification claims, VPS/Thaalam/Herdr work, pushes, PRs, merges, or new dependencies.
- No raw SQL/query, arbitrary GraphQL, generic HTTP method/path/body, shell, file, or passthrough escape hatches.
- Reverse ETL remains plan → preview → explicit approval → execute.
- Binary transfer execution is not added in shared runtime code; any missing executor remains a documented shared-runtime dependency.
