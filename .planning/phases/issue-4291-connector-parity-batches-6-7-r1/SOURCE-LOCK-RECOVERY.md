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

## Remaining suspect connectors

`copper`, `freshdesk`, `intercom`, `iterable`, `klaviyo`, `square`, and `service-now` remain in the recovery queue. Each will receive a comprehensive provider-surface remap: source lock, API surface, and every disposition row are regenerated from the authoritative source rather than preserving the legacy API-surface boundary. ServiceNow is dynamic-schema: its fixed platform surface must be pinned separately from its instance-dependent schema basis. Any source that genuinely has a small complete surface will carry the source's exact count plus an explicit confidence/basis; a small number alone will never be presented as complete coverage.
