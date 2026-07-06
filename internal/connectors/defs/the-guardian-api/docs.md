# Overview

Reads Guardian content search results through the Guardian Open Platform Content API.

Readable streams: `search`, `tags`, `sections`, `editions`, `content`.

This connector is read-only; no write actions are declared.

Service API documentation: https://open-platform.theguardian.com/documentation/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Guardian Open Platform API key, sent as the api-key query
  parameter. Never logged.
- `base_url` (optional, string); default `https://content.guardianapis.com`; format `uri`; Guardian
  Content API base URL override for tests or proxies.
- `content_id` (optional, string); Guardian content item path (e.g.
  'world/2024/jan/01/example-article') used by the content stream's single-item lookup; required
  only when the content stream is selected.
- `query` (optional, string); Optional free-text search query (Guardian's q parameter); when unset,
  the search stream returns Guardian's default (unfiltered) content listing.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://content.guardianapis.com`.

Authentication behavior:

- API key authentication in query parameter `api-key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/search`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `page-size`;
starts at 1; page size 50.

Pagination by stream: none: `sections`, `editions`, `content`; page_number: `search`, `tags`.

- `search`: GET `/search` - records path `response.results`; query `q` from template `{{
  config.query }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `page-size`; starts at 1; page size 50; computed output fields `published_at`, `title`.
- `tags`: GET `/tags` - records path `response.results`; query `q` from template `{{ config.query
  }}`, omitted when absent; page-number pagination; page parameter `page`; size parameter
  `page-size`; starts at 1; page size 50.
- `sections`: GET `/sections` - records path `response.results`; query `q` from template `{{
  config.query }}`, omitted when absent.
- `editions`: GET `/editions` - records path `response.results`; query `q` from template `{{
  config.query }}`, omitted when absent.
- `content`: GET `/{{ config.content_id }}` - single-object response; records path
  `response.content`; computed output fields `published_at`, `title`.

## Write actions & risks

This connector is read-only. Read behavior: external Guardian Open Platform API read of published
content search results.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=1.
