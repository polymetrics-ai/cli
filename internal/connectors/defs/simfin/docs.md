# Overview

SimFin is a read-only declarative HTTP connector for the SimFin REST API v3. Pass B expands the legacy companies, statements, and markets streams with the current ReadMe-documented company general data, financial statement, price, share, filing-index, changed-company, and database-change endpoints. The legacy Go package remains the record-shape authority until the registry cutover.

## Auth setup

Provide the SimFin API key via the `api_key` secret. The legacy connector sends it as the `api-key` query parameter, and this bundle preserves that behavior. Current SimFin ReadMe v3 docs show the same key in an Authorization header, so the bundle also sends a matching header for the newer endpoints without exposing the secret in fixtures or docs. `base_url` defaults to `https://backend.simfin.com`.

## Streams notes

The legacy streams `companies`, `statements`, and `markets` keep the legacy `data` envelope, one-based `page`/`limit` pagination, primary key `id`, and statement `updated_at` computed from `fiscalPeriod`. Those choices preserve the records emitted by `internal/connectors/simfin`.

The added ReadMe v3 streams are `company_general_compact`, `company_general_verbose`, `company_statements_compact`, `company_statements_verbose`, `company_prices_compact`, `company_prices_verbose`, `common_shares_outstanding`, `weighted_shares_outstanding`, `filings_by_company`, `filings`, `changed_companies`, and `data_change_log`. SimFin's current OpenAPI blocks do not publish concrete response schemas for most of these endpoints, so these streams use `projection: passthrough` with permissive schemas and recorded synthetic fixtures. The compact endpoints retain their nested compact payloads instead of flattening `columns`/`data` arrays, because the engine has no documented columnar response flattener.

Optional filters in `spec.json` are wired only where the docs publish matching query parameters. `statements` defaults to all documented statement families when callers do not provide it. `filings` uses documented zero-based `page`/`per-page` pagination, and `data_change_log` uses documented zero-based `pageNumber`/`itemsPerPage` pagination.

## Write actions & risks

None. SimFin's documented v3 surface in this bundle is read-only, `capabilities.write` is false, and the connector ships no `writes.json`.

## Known limits

- The historical `https://simfin.com/api/v3/documentation/` URL now redirects to a 404; the reviewed official docs source is `https://simfin.readme.io/llms.txt` and its linked v3 API reference pages.
- Current SimFin OpenAPI pages mostly omit field-level response schemas. New Pass B streams therefore use passthrough schemas and fixtures that preserve documented envelope families rather than guessing a closed column set.
- Compact company, statement, and price endpoints are exposed as one record per top-level company payload. The nested compact `columns`/`data` arrays are not expanded into one row per measurement because that would require a columnar-response mapper not present in the declarative engine.
- `GET /api/v3/filings/get-report/{filingType}/{filingIdentifier}` is excluded as `binary_payload` because it returns the raw filing document as HTML/PDF or encoded file content, not a JSON data stream.
- Legacy `page_size` and `max_pages` config overrides remain unwired for the legacy streams because the engine pagination block is static. The bundle keeps the legacy default page size of 100 and documents the narrowing here instead of declaring dead config.
