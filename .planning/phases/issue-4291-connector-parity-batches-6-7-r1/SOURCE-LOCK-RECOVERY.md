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

For each slice, the source-lock count, API-surface rows, and disposition rows reconcile exactly; `declared_percent` was removed. Existing stream coverage was retained only where its method/path occurs in the authoritative artifact, and no source operation was reported enabled without a runnable CLI command or typed write action.

Freshdesk’s complete rendered-reference pass is also green: all 171 endpoint sections in the provider’s single 3.2MB reference normalize to **170** unique HTTP method/path operations (78 GET, 39 POST, 30 PUT, 22 DELETE, 1 PATCH), replacing the legacy 10-row boundary. Its full source lock, API surface, and disposition ledger reconcile at 170 rows.

Copper’s provider-published MkDocs `search_index.json` contains its complete **637-document** rendered reference corpus. The persisted, resumable parse completed all 637 documents and found **89** unique declaration-form REST operations (32 GET, 35 POST, 11 PUT, 11 DELETE), replacing the legacy five synthetic `HOOK` labels. Its five ETL streams are now bound to the provider-documented `POST /v1/<resource>/search` operations, with exact routing evidence in `internal/connectors/native/copper/streams.go:5-25`; no source operation is enabled without a runnable command or typed write action.

## Remaining full-batch audit

All 20 owned connectors are in the recovery audit—not only the eight whose initial counts were visibly implausible. Each will receive a comprehensive provider-surface review: source lock, API surface, and every disposition row are regenerated from the authoritative source when the legacy source understates it. ServiceNow is dynamic-schema: its fixed platform surface must be pinned separately from its instance-dependent schema basis. A connector whose complete source proves the legacy count correct is an explicit no-change result, with source count and confidence basis recorded. A small number alone will never be presented as complete coverage.
