# Overview

Reads GNews articles from the keyword search and top-headlines endpoints of the GNews REST API.
Read-only.

Readable streams: `search`, `top_headlines`.

This connector is read-only; no write actions are declared.

Service API documentation: https://gnews.io/docs/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); GNews API key. Sent as the 'apikey' query parameter on every
  request; never logged.
- `base_url` (optional, string); default `https://gnews.io/api/v4`; format `uri`; GNews API base URL
  override for tests or proxies.
- `country` (optional, string); Optional GNews 'country' filter (e.g. 'us').
- `end_date` (optional, string); format `date-time`; RFC3339 upper bound sent as the 'to' filter.
- `in` (optional, string); Optional GNews 'in' filter restricting which article fields are searched
  (e.g. 'title,description').
- `language` (optional, string); Optional GNews 'lang' filter (e.g. 'en').
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `nullable` (optional, string); Optional GNews 'nullable' filter naming fields allowed to be null
  in the response.
- `page_size` (optional, string); default `10`; Records per page (1-100), sent as GNews's 'max'
  parameter.
- `query` (optional, string); Also used as the 'top_headlines' stream's query when
  top_headlines_query is unset.
- `sortby` (optional, string); Optional GNews 'sortby' filter ('publishedAt' or 'relevance').
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound for the 'from' filter;
  only articles published at or after this time are read on a fresh sync.
- `top_headlines_query` (optional, string); Optional keyword query for the 'top_headlines' stream,
  overriding the shared 'query' value for that stream only.
- `top_headlines_topic` (optional, string); Optional GNews topic filter (e.g. 'technology',
  'business') for the 'top_headlines' stream.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://gnews.io/api/v4`, `max_pages=0`, `page_size=10`.

Authentication behavior:

- API key authentication in query parameter `apikey` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/search`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `max`; starts at
1; page size 10.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `search`: GET `/search` - records path `articles`; query `country` from template `{{
  config.country }}`, omitted when absent; `from` from template `{{ incremental.lower_bound }}`,
  omitted when absent; `in` from template `{{ config.in }}`, omitted when absent; `lang` from
  template `{{ config.language }}`, omitted when absent; `nullable` from template `{{
  config.nullable }}`, omitted when absent; `q` from template `{{ config.query }}`, default `news`;
  `sortby` from template `{{ config.sortby }}`, omitted when absent; `to` from template `{{
  config.end_date }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `max`; starts at 1; page size 10; incremental cursor `published_at`; sent as `from`;
  formatted as `rfc3339`; initial lower bound from `start_date`; computed output fields
  `published_at`, `source_country`, `source_id`, `source_name`, `source_url`.
- `top_headlines`: GET `/top-headlines` - records path `articles`; query `country` from template `{{
  config.country }}`, omitted when absent; `from` from template `{{ incremental.lower_bound }}`,
  omitted when absent; `in` from template `{{ config.in }}`, omitted when absent; `lang` from
  template `{{ config.language }}`, omitted when absent; `nullable` from template `{{
  config.nullable }}`, omitted when absent; `q` from template `{{ config.top_headlines_query }}`,
  omitted when absent; `sortby` from template `{{ config.sortby }}`, omitted when absent; `to` from
  template `{{ config.end_date }}`, omitted when absent; `topic` from template `{{
  config.top_headlines_topic }}`, omitted when absent; page-number pagination; page parameter
  `page`; size parameter `max`; starts at 1; page size 10; incremental cursor `published_at`; sent
  as `from`; formatted as `rfc3339`; initial lower bound from `start_date`; computed output fields
  `published_at`, `source_country`, `source_id`, `source_name`, `source_url`.

## Write actions & risks

This connector is read-only. Read behavior: external GNews API read of news article search results.

## Known limits

- Batch defaults: read_page_size=10.
- API coverage includes 2 stream-backed endpoint group(s).
