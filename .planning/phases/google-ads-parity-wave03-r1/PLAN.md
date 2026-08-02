# Google Ads connector parity wave03-r1 plan (#3021-#3028)

## Scope

Implement the Google Ads (`google-ads`) connector parity slice on branch `fm/cli-google-ads-parity-wave03-r1` in the disposable worktree.

Allowed paths for this slice:

- `internal/connectors/defs/google-ads/**`
- connector-owned Google Ads hook tests/fixtures under `internal/connectors/hooks/google-ads/**` when stream-hook evidence must stay aligned with version/path changes
- generated connector docs/catalog/skills/website data that reflect the Google Ads definition
- `.planning/phases/google-ads-parity-wave03-r1/**` planning, TDD, source-audit, and verification artifacts
- small generic validation-tool compatibility only if required to make the user-mandated `connectorgen validate internal/connectors/defs/google-ads` gate runnable; no provider-specific shared runtime behavior

Do not add dependencies, request credentials, run live provider APIs, execute provider writes, certify live behavior, push, open/update PRs, invoke `/no-mistakes`, or edit shared connector runtime semantics.

## GSD command path and fallback

- Ran `scripts/gsd doctor`: passed.
- Ran `scripts/gsd list`: passed (69 commands listed).
- Attempted required implementation command `scripts/gsd prompt programming-loop init --phase google-ads-parity-wave03-r1 --dry-run`: failed with `scripts/gsd: unknown GSD command: programming-loop`.
- Ran available adapter path `scripts/gsd prompt quick --full "Implement Google Ads connector parity wave03-r1 for issues 3021-3028 with fixture-only validation"` and follow this prompt inline.
- Manual GSD programming-loop fallback is active for this phase because the repo-local adapter has no `programming-loop` command entry.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-context`
- `golang-concurrency`
- `golang-documentation`
- CLI help/docs/website parity reference

## Official source audit baseline

Public, credential-free sources re-audited before production edits:

- Discovery: `https://googleads.googleapis.com/$discovery/rest?version=v22`
  - observed `kind=discovery#restDescription`, `version=v22`, `revision=20260721`, `rootUrl=https://googleads.googleapis.com/`
  - observed raw discovery methods: 163 (`POST=151`, `GET=11`, `DELETE=1`)
  - observed schemas: 1363
- Parent issue source link: `https://developers.google.com/google-ads/api/reference/rpc/v22/overview`

Initial local connector state:

- 3 streams: `accessible_customers`, `campaigns`, `ad_groups`
- 0 write actions
- no `operations.json` or `cli_surface.json`
- `metadata.capabilities.write=false`
- `api_surface.json` is coarse and version-skewed to v24, with 12 grouped rows rather than exact v22 method rows

Initial pre-edit gate finding:

- `go run ./cmd/connectorgen validate internal/connectors/defs/google-ads` currently fails before code changes because `connectorgen validate` treats child directories (`fixtures`, `schemas`) as connector roots. This will be handled either by a generic validation-path compatibility fix or by recording the exact blocker if editing shared tooling is disallowed.

## Implementation slices

1. **Source-audit and operation ledger**
   - Preserve a generated source audit artifact with raw v22 method counts, path-variable evidence, classification rules, and post-change counts.
   - Replace the coarse `api_surface.json` with `operation_ledger_version: 1` rows, partitioning each v22 discovery method or connector-owned stream row exactly once.
   - Use blocked operation rows for reserved-expansion resource-name methods that cannot be represented truthfully by the current connector-local path contract.

2. **Version and base URL alignment**
   - Align `spec.json`, `streams.json`, docs, fixtures, and Google Ads hook tests with v22 discovery (`base_url=https://googleads.googleapis.com`, request paths under `/v22...`).
   - Preserve secret redaction for `access_token` and `developer_token`; keep optional `login_customer_id` omitted when unset.

3. **Executable direct/provider-search operations**
   - Add definition-owned `operations.json` and `cli_surface.json` entries for fixed, bounded, typed direct reads whose paths only require `customer_id` or no path variables.
   - Use `output_policy: json_redacted`, positive `max_bytes`, fixed endpoint paths, and operation body schemas. Do not add a raw method/path/body/query escape hatch.
   - Keep arbitrary `googleAds:search` / `searchStream` GAQL beyond the fixed `campaigns`/`ad_groups` streams blocked/planned unless represented by a fixed typed operation.

4. **Executable reverse/write actions**
   - Add write actions for v22 write/mutate methods whose path can be represented by existing connector-local templating (`customer_id` or no path variables).
   - Use required closed top-level schemas, `path_fields` for record path variables, `redact_fields` where path/body values can be sensitive, risk text, and destructive confirmation for destructive/admin mutate/remove/delete families.
   - Keep resource-name reserved expansion methods blocked/planned when the engine would URL-encode slashes and therefore send an untruthful path.

5. **Fixtures and tests**
   - Add sanitized fixture-only conformance write fixtures for every executable write action.
   - Update Google Ads hook tests for v22 paths and add coverage for auth headers/path alignment if needed.
   - No live provider calls, no credentials, no certification.

6. **Docs, generated surfaces, and issue addendum**
   - Regenerate connector manual/skill/catalog/docs/website data using local generators.
   - Update `docs.md` with v22 source evidence, write/direct-read truth, blocked reasons, fixture-only uncertified status, and reverse ETL safety.
   - Append the established captain-policy addendum idempotently to #3021-#3028 with actual post-change counts and no certification claims.

## Human gates / non-goals

- No new dependencies.
- No credentialed Google Ads checks.
- No provider writes or live certification.
- No generic provider query, generic HTTP write, generic SQL write, or raw shell surfaces.
- Reverse ETL remains plan -> preview -> explicit approval -> execute.
- No PR, push, merge, `/no-mistakes`, VPS, Herdr, or Thaalam changes.

## Orchestration decision

`local_critical_path`: this worker owns one connector tree in one isolated worktree. Mutating subagents were not spawned because additional mutating workers would collide on the same connector-owned JSON/docs/generated files. Read-only subagents may be used only for review/recon.

## Completion notes

Implemented and locally verified on 2026-07-31:

- Official source audit preserved in `SOURCE-AUDIT.md` / `SOURCE-AUDIT.json`.
- Google Ads aligned to public v22 discovery paths with fixture-only validation.
- Added 21 fixed direct-read operations, 7 guarded write actions, and 7 sanitized write fixtures.
- Replaced coarse legacy `api_surface.json` with v22 operation-ledger rows: 164 local rows (3 streams, 21 direct reads, 7 writes, 133 blocked/planned). The local row count is one greater than raw discovery because one official `customers.googleAds.search` method backs two fixed connector streams.
- Added a generic `connectorgen validate` path compatibility fix so the required single-connector validation gate checks one bundle instead of treating `fixtures/` and `schemas/` as sibling connectors.
- Regenerated Google Ads connector manual/skill, connector catalogs, website connector data, and CLI golden transcripts.
