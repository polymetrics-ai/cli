# Overview

Reads Serpstat SEO domain keyword, competitor, and top-URL data through the Serpstat
JSON-RPC-over-HTTP API. Read-only.

Readable streams: `domain_keywords`, `domain_competitors`, `domain_urls`.

This connector is read-only; no write actions are declared.

Service API documentation: https://api.serpstat.com/api-docs.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Serpstat API token, sent as the 'token' query-string
  parameter on every JSON-RPC request (Serpstat's own auth convention -- never a header). Never
  logged.
- `base_url` (optional, string); default `https://api.serpstat.com/v4`; format `uri`; Serpstat
  JSON-RPC-over-HTTP base URL override for tests or proxies.
- `domain` (optional, string); default `serpstat.com`; Domain to query for the
  domain_keywords/domain_competitors JSON-RPC procedures.
- `page_size` (optional, string); default `10`; Records per JSON-RPC page (1-1000, 'size' param).
  Pagination stops on the first page returning fewer records than this.
- `pages_to_fetch` (optional, string); default `1`; 0 means unbounded (fetch until a short page).
- `region_id` (optional, string); default `g_us`; Serpstat search-engine/region identifier ('se'
  JSON-RPC param).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.serpstat.com/v4`, `domain=serpstat.com`,
`page_size=10`, `pages_to_fetch=1`, `region_id=g_us`.

Authentication behavior:

- API key authentication in query parameter `token` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks use a connector-managed request.

## Streams notes

Default pagination: single request; no pagination.

Paging state (the page number) is carried inside the JSON-RPC request body rather than in
query-string parameters. All streams are non-incremental: each read starts from the first page
and fetches the full configured page range every time.

- `domain_keywords`: POST connector-managed request path - records path `result.data`.
- `domain_competitors`: POST connector-managed request path - records path `result.data`.
- `domain_urls`: POST connector-managed request path - records path `result.data`.

## Write actions & risks

This connector is read-only. Read behavior: external Serpstat API read of domain
keyword/competitor/top-URL SEO metrics.

## Known limits

- Batch defaults: read_page_size=10.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=1, out_of_scope=12.
