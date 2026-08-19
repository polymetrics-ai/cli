# Source-lock recovery — issue #4291

PR #4296 is held while the provider-source denominator is repaired. `declared_percent` is retired: it only measured the legacy crosswalk against itself.

## Salesloft — first green slice

| Measure | Prior lock | Recovered lock |
| --- | ---: | ---: |
| Source artifact | 84,498-byte documentation index | 315-page complete rendered API reference |
| `counts.total` / `operations_found` | 12 | **211** |
| GET | 10 | 120 |
| POST | 2 | 50 |
| PUT | 0 | 23 |
| DELETE | 0 | 18 |

The recovered lock crawls every `https://developers.salesloft.com/docs/api/` page listed by the provider's public sitemap, pins the aggregate bytes/SHA-256, and cites the exact endpoint page for each of the 211 operation records. Its confidence is `complete_rendered_reference` with an explicit basis.

The prior twelve `api_surface.json` crosswalk rows were discarded. The regenerated `api_surface.json` and declaration-disposition ledger each contain all 211 cited provider operations. `declared_percent` is absent; the map reports `operations_found` plus the complete-rendered-reference confidence basis. No operation is enabled: Salesloft has no runnable CLI command or typed write action binding.

Salesloft's regenerated six-class breakdown is 115 direct reads, 91 direct writes (including all 18 deletes), and 5 ETL stream operations. Every direct write records reverse-ETL only as an attribute, with `generic-typed-destination-executor` as its destination-executor foundation gap; none is classified as reverse ETL.

## Official-spec slices

The same complete-remap invariant is now green for these provider-published machine-readable specifications:

| Connector | Legacy rows | Recovered provider operations | Exact source |
| --- | ---: | ---: | --- |
| Iterable | 4 | **148** | `https://api.iterable.com/api-docs` (Swagger 2.0) |
| Klaviyo | 9 | **345** | `https://raw.githubusercontent.com/klaviyo/openapi/main/openapi/stable.json` (OpenAPI 3.0.2 full GA specification) |
| Intercom | 10 | **231** | `https://raw.githubusercontent.com/intercom/Intercom-OpenAPI/main/descriptions/2.16/api.intercom.io.yaml` (OpenAPI 3.0.1) |
| Outreach | 259 | **259** | official OpenAPI 3.0.3 plus the provider’s fixed generic custom-object route reference |
| Customer.io | 159 | **166** | `https://docs.customer.io/files/journeys-app.json` (OpenAPI 3.1.0) |
| Gorgias | 114 | **114** | official OpenAPI 3.1.0 API-registry specification |
| Chatwoot | 148 | **148** | official OpenAPI 3.1.0 `swagger/swagger.json` specification |

For each slice, the source-lock count, API-surface rows, and disposition rows reconcile exactly; `declared_percent` was removed. Existing stream coverage was retained only where its method/path occurs in the authoritative artifact, and no source operation was reported enabled without a runnable CLI command or typed write action.

Outreach is an explicit count-unchanged audit result, not an unverified carry-forward: the current public OpenAPI has 253 static operations, and the provider documents six fixed generic custom-object routes separately because the per-account object schemas are dynamic. The combined complete fixed surface is therefore 259. All 163 enabled rows have an exact typed write-action binding; the 96 stream-only rows remain ETL and declaration-pending. Customer.io increases from 159 to 166 source operations; its 10 enabled rows each retain an exact typed write-action binding.

Gorgias and Chatwoot are explicit count-unchanged audit results from their actual current OpenAPI sources, not carry-forwards from documentation landings. Gorgias reconciles to 114 source operations with 103 exact command/write bindings; Chatwoot reconciles to 148, with 94 exact command/write bindings. Their remaining source rows stay declaration-pending.

## Square — partial crawl superseded by complete machine spec

The rendered crawl records its static-card extraction defect separately: all 40 persisted group pages were fetched, seven initially yielded no static cards, and their browser-rendered cards were checkpointed. It remains `coverage_confidence: partial` with `operations_found: null`; its preliminary 255-operation de-duplication was discarded and never became a source lock.

Square publishes an authoritative OpenAPI 3.0.0 document at `https://raw.githubusercontent.com/square/connect-api-specification/master/api.json`. That 3,279,392-byte source pins **334** REST operations (121 GET, 150 POST, 36 PUT, 27 DELETE), so the complete source lock, API surface, and disposition ledger are regenerated **334/334/334** from the machine-readable source. The four existing ETL stream bindings remain declaration-pending; no Square operation is enabled without a runnable command or typed write action.

Freshdesk’s complete rendered-reference pass is also green: all 171 endpoint sections in the provider’s single 3.2MB reference normalize to **170** unique HTTP method/path operations (78 GET, 39 POST, 30 PUT, 22 DELETE, 1 PATCH), replacing the legacy 10-row boundary. Its full source lock, API surface, and disposition ledger reconcile at 170 rows.

Copper’s provider-published MkDocs `search_index.json` contains its complete **637-document** rendered reference corpus. The persisted, resumable parse completed all 637 documents and found **89** unique declaration-form REST operations (32 GET, 35 POST, 11 PUT, 11 DELETE), replacing the legacy five synthetic `HOOK` labels. Its five ETL streams are now bound to the provider-documented `POST /v1/<resource>/search` operations, with exact routing evidence in `internal/connectors/native/copper/streams.go:5-25`; no source operation is enabled without a runnable command or typed write action.

## Zoho Bigin — complete rendered-reference recovery

Bigin's 8KB landing page is not a usable operation source. Its provider sitemap enumerates **98** pages under `/developer/docs/apis/v2/`; the resumable crawler fetched every page and persisted its URL, response bytes, SHA-256, and extracted operations before advancing the checkpoint. It also includes the separate documented `/bigin/bulk/v2` API family, which a base-v2-only parser would otherwise omit. Regional host replicas, query variants, and illustrative concrete-module examples are normalized to their documented endpoint templates.

The recovered denominator is **75** operations: 32 GET, 22 POST, 9 PUT, 1 PATCH, and 11 DELETE. The regenerated source lock, `api_surface.json`, and disposition ledger reconcile at 75 rows with `counts.total` and `operations_found.total` both 75 and `coverage_confidence: complete_rendered_reference`. The legacy 50-row surface is discarded. Six pre-existing exact typed actions remain enabled `direct_write`; the other 37 direct writes are declaration-pending but retain the neutral-destination foundation gap as an attribute, never a `reverse_etl` primary class.

## ServiceNow — fixed platform surface, dynamic target schema

ServiceNow is intentionally not assigned a fabricated count of instance tables. The official Table API publishes **six** fixed method/path templates: list, get-by-`sys_id`, insert, replace, partial update, and delete. Its target `table_name`, dictionary fields, ACLs, and customer-defined tables are determined by the configured instance, not the public platform reference. The recovered source lock therefore records `counts.total: 6` and `operations_found.total: 6` only for the fixed platform templates, plus a `dynamic_schema` basis with `instance_operation_total: null`.

The legacy 22 rows were three selected table instantiations, not a truthful platform denominator. The regenerated ledger is six source rows (two direct writes with exact existing typed actions, four declaration-pending), and six extra `api_surface` coverage projections preserve the existing incident/user/group streams and typed actions without increasing the six-template count. Every one of the four direct-write templates records `generic-typed-destination-executor` only as a reverse-ETL attribute; no transport binding is declared.

## ActiveCampaign — embedded machine-readable reference recovery

ActiveCampaign's reference renderer embeds a complete public API v3 OpenAPI 3.1.0 document in the official `list-all-contacts` reference response. The 1,867,455-byte provider artifact contains **186** paths and **296** operations: 139 GET, 58 POST, 44 PUT, 5 PATCH, and 50 DELETE. It replaces the legacy 61-row documentation-index boundary with a complete machine-readable 296-operation source lock, API surface, and disposition ledger. No operation becomes enabled solely through the remap: all 157 direct writes are `direct_write`, are declaration-pending until source-backed typed actions exist, and carry the neutral destination executor only as their reverse-ETL foundation gap.

## Remaining full-batch audit

All 20 owned connectors are in the recovery audit—not only the eight whose initial counts were visibly implausible. Each will receive a comprehensive provider-surface review: source lock, API surface, and every disposition row are regenerated from the authoritative source when the legacy source understates it. ServiceNow is dynamic-schema: its fixed platform surface must be pinned separately from its instance-dependent schema basis. A connector whose complete source proves the legacy count correct is an explicit no-change result, with source count and confidence basis recorded. A small number alone will never be presented as complete coverage.
