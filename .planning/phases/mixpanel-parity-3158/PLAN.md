# Mixpanel connector parity plan (#3158)

## Scope

Implement documented Mixpanel official API parity for parent #3158 and subissues #3159-#3165 on branch `fm/cli-mixpanel-parity-wave04-r1`.

Allowed production paths for this slice:

- `internal/connectors/defs/mixpanel/**`
- Mixpanel connector-owned fixtures under `internal/connectors/defs/mixpanel/fixtures/**`
- Mixpanel connector-owned docs/generated metadata surfaces: `docs.md`, `api_surface.json`, `operations.json`, `writes.json`, `cli_surface.json`, `certification.json`
- Owned generated connector docs/catalog/help surfaces needed for CLI/docs/website parity.

Do not edit shared runtime/engine behavior unless a verification failure proves the existing declarative contract cannot safely express the official documented surface. If a shared foundation is missing, record it as blocked/planned rather than inventing a broad escape hatch.

## GSD command path

- `scripts/gsd doctor`: passed earlier in this branch setup.
- `scripts/gsd list`: passed earlier.
- GSD quick prompt generated for Mixpanel parity; manual GSD tracking in this phase directory records plan/TDD/verification before production edits.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-security`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-error-handling`
- `golang-safety`
- `golang-context`
- `golang-concurrency`
- `golang-documentation`
- CLI help/docs/website parity reference

## Source inventory

Official Mixpanel source set from #3158:

- Overview: `https://docs.mixpanel.com/reference/overview`
- OpenAPI YAML files: ingestion, identity, query, export, lexicon-schemas, data-pipelines, service-accounts, annotations, gdpr, warehouse-connectors, feature-flags, feature-flags-management, experiments.
- Current local evidence: `/tmp/mixpanel-openapi/ops.json`, `/tmp/mixpanel-openapi/ledger.txt`, `/tmp/mixpanel-openapi/auth-summary.txt`, `/tmp/mixpanel-openapi/write-body-summary.txt`.
- Official operation count: 105 rows. Official lane counts: `etl_read=24`, `reverse_etl_write=61`, `direct_read_query_search=18`, `binary_file=1`, `cdc_changefeed=1`, `excluded_not_applicable=0`.

## Implementation slices

1. **Operation ledger parity (#3159)**
   - Regenerate `api_surface.json` with all 105 official operations exactly once using `operation_ledger_version: 1`.
   - Keep unsupported executor gaps as truthful blocked/planned operation rows, not silent omissions or unsafe exclusions.

2. **ETL and changefeed (#3160)**
   - Declare documented list/detail ETL streams where the declarative HTTP reader can execute fixed, bounded JSON responses.
   - Cover the Mixpanel activity stream as a bounded stream while keeping any unsupported CDC-specific claim truthful.
   - Add sanitized replay fixtures and schemas for every declared stream.

3. **Direct/binary (#3161)**
   - Declare bounded operation direct-read metadata and CLI commands for safe JSON GET query endpoints.
   - Keep JQL/arbitrary script, x-www-form-urlencoded query POSTs, and raw export/binary download blocked/planned unless an existing safe executor supports them.

4. **Reverse ETL writes (#3162)**
   - Declare typed reverse-ETL writes for documented JSON/form/idempotent mutation endpoints supported by the declarative write engine.
   - Use closed record schemas, path/body field allow-lists, redaction metadata, and destructive confirmation for DELETE/destructive/admin actions.
   - Keep unsupported raw CSV lookup table replacement blocked/planned.

5. **CLI/config/docs parity (#3163)**
   - Generate Mixpanel `cli_surface.json` with implemented stream/direct commands and planned/blocked entries for unsupported official operations.
   - Update connector docs/manual/skill and generated connector catalogs/help/golden transcripts as applicable.

6. **Fixtures/Guard (#3164)**
   - Fixture-only, secret-free replay for every declared stream and write action.
   - Run focused connectorgen validation/conformance and CLI tests.

7. **Certification/release truthfulness (#3165)**
   - Keep certification fixture-only/uncertified; no live provider calls or credentials.
   - Append captain-policy destructive-operation addendum idempotently to #3158-#3165 with actual local counts.

## Non-goals and safety gates

- No live Mixpanel credentials, provider calls, provider writes, certification, VPS/Thaalam, pushes, PRs, merges, or `/no-mistakes`.
- No generic HTTP method/path/body/query/script/file/shell passthroughs.
- Reverse ETL remains plan → preview → explicit approval → execute.
