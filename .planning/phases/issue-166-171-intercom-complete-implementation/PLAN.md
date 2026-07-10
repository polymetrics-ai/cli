# GSD Plan: Intercom Complete CLI Implementation (#166-#171)

Parent: #164 / PR #220  
Branch: `feat/166-171-intercom-complete-implementation` from parent branch `feat/164-intercom-cli-parity`  
Stacked PR: https://github.com/polymetrics-ai/cli/pull/257

## GSD command path

- `scripts/gsd doctor` passed in this session.
- Required command `scripts/gsd prompt programming-loop init --phase issue-164-intercom-complete-implementation --dry-run` failed with `unknown GSD command: programming-loop`.
- Manual GSD fallback is active, using `scripts/gsd prompt quick --full ...` generated at `/tmp/gsd-intercom-complete-prompt.md`.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-cli`, `golang-spf13-cobra`, `golang-spf13-viper`
- `golang-testing`
- `golang-security`, `golang-safety`, `golang-error-handling`
- `golang-design-patterns`, `golang-structs-interfaces`
- `golang-context`, `golang-concurrency`
- `golang-documentation`
- CLI help/docs/website parity reference

## Objective

Complete Intercom connector CLI parity for the official Intercom OpenAPI 2.14 surface: every official operation must be covered by an executable ETL stream, safe direct read, typed reverse-ETL write behind plan/preview/approval/confirmation gates, bounded binary/file policy, duplicate/deprecated/disallowed block, or product-scope block.

## Safety constraints

- No secrets, no credentialed live Intercom checks, and no new dependencies.
- No raw generic HTTP write, SQL write, shell, or arbitrary GraphQL escape hatch.
- Reverse ETL remains plan -> preview -> approval token -> execute, with destructive/admin operations requiring typed confirmation where declared.
- Binary/file operations must be bounded and metadata-first; unsafe downloads remain blocked unless a bounded policy is implemented and tested.
- All command args, path parameters, query params, and filesystem paths are untrusted.

## Slices

1. **Red tests / contracts**
   - Update Intercom API surface tests to require full coverage accounting instead of blocked metadata-only rows.
   - Add command runner/direct-read tests for generic Intercom JSON direct-read policy and path/query flag mapping.
   - Add CLI help tests for bare connector namespace (`pm intercom`) and command help (`pm intercom contact list --help` / blocked invalid command behavior).
   - Add write-plan tests for representative create/update/delete Intercom commands proving no external execution before approval.

2. **Help renderer (#166)**
   - Render connector namespace help from `cli_surface.json` when a connector command is invoked with no path or help flags.
   - Ensure blocked/invalid commands still return usage/blocked errors.
   - Update docs/website parity or record not-applicable generated connector-doc path.

3. **Stream runner (#167)**
   - Expand `streams.json` with official Intercom collection/read-query candidates that can run via bounded pagination.
   - Add schemas or passthrough where official shapes are broad.
   - Account stream coverage in `api_surface.json`.

4. **Operation ledger (#168)**
   - Convert metadata-only rows into covered_by stream/direct_read/write/direct_reads where implemented.
   - Leave only true duplicate/deprecated/disallowed/product-scope rows blocked with reasons.

5. **Direct reads (#169)**
   - Add generic JSON direct-read output policy for provider object/list reads with bounded max bytes and no absolute URLs.
   - Implement Intercom read commands for official GET item/search endpoints not modeled as streams.

6. **Advanced query/binary (#170)**
   - Support Intercom search/query streams/direct reads with declared POST body/query parameter actions when safe.
   - Implement bounded metadata/file policy for binary/export endpoints or keep unsafe downloads explicitly blocked with product-scope reason.

7. **Sensitive/admin policy (#171)**
   - Add typed write actions for official POST/PUT/DELETE mutations with path fields, JSON/body field schemas, risk text, and destructive confirmation where needed.
   - Add CLI metadata for each write command; verify `pm intercom ... --preview --json` only creates/stages a plan.

## Implementation status (2026-07-10)

- Complete Intercom API surface now has 149 CLI command entries covering all 149 official operations.
- Stream coverage expanded to 31 stream/direct collection operations with passthrough schemas where official response shapes are broad.
- Direct-read coverage uses bounded `json_response`, `text_response`, and `binary_metadata` policies rather than raw file writes.
- Reverse ETL coverage uses 77 typed write actions with schemas, risk text, approval requirements, and destructive/admin confirmation metadata.
- Runtime help parity is implemented for `pm intercom`, `pm help intercom`, and `pm intercom <resource> <action> --help`.
- Docs/website parity added: `docs/cli/intercom.md` and `website/content/docs/intercom-cli-surface.mdx`.
- Local verification passed; see `TDD-LEDGER.md` and `VERIFICATION.md`.

## Commit checkpoints

- [x] Plan checkpoint before production edits.
- [x] Red-test checkpoint when feasible.
- [x] Green implementation checkpoint.
- [x] Docs/help parity checkpoint.
- [x] Verification/review-route checkpoint: stacked PR #257 opened against parent branch; CI/review routing pending.

## Human gates

- Any new dependency, token-scope change, credentialed/live Intercom run, destructive external action, reverse ETL execution, or parent PR merge to `main`.
