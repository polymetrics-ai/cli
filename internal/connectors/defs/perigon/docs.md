# Overview

Reads Perigon news articles, story clusters, journalists, sources, companies, people, and topics
through the Perigon REST API.

Readable streams: `articles`, `stories`, `journalists`, `sources`, `companies`, `people`, `topics`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.goperigon.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Perigon API key, sent as the apiKey query parameter. Never
  logged.
- `base_url` (optional, string); default `https://api.perigon.io`; format `uri`; Perigon API base
  URL override for tests or proxies.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-100).
- `query` (optional, string); Optional search query string sent as the articles stream's q
  parameter.
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; only articles published
  at or after this time are read (articles stream only).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.perigon.io`, `max_pages=0`, `page_size=100`.

Authentication behavior:

- API key authentication in query parameter `apiKey` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v1/articles/all` with query `size`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `size`; starts at
1; page size 100.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `articles`: GET `/v1/articles/all` - records path `articles`; query `from` from template `{{
  incremental.lower_bound }}`, omitted when absent; `q` from template `{{ config.query }}`, omitted
  when absent; page-number pagination; page parameter `page`; size parameter `size`; starts at 1;
  page size 100; incremental cursor `pub_date`; formatted as `rfc3339`; initial lower bound from
  `start_date`; computed output fields `article_id`, `pub_date`.
- `stories`: GET `/v1/stories/all` - records path `stories`; page-number pagination; page parameter
  `page`; size parameter `size`; starts at 1; page size 100; emits passthrough records.
- `journalists`: GET `/v1/journalists/all` - records path `results`; query `q` from template `{{
  config.query }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `size`; starts at 1; page size 100; emits passthrough records.
- `sources`: GET `/v1/sources/all` - records path `results`; query `q` from template `{{
  config.query }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `size`; starts at 1; page size 100; emits passthrough records.
- `companies`: GET `/v1/companies/all` - records path `results`; query `q` from template `{{
  config.query }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `size`; starts at 1; page size 100; emits passthrough records.
- `people`: GET `/v1/people/all` - records path `results`; query `q` from template `{{ config.query
  }}`, omitted when absent; page-number pagination; page parameter `page`; size parameter `size`;
  starts at 1; page size 100; emits passthrough records.
- `topics`: GET `/v1/topics/all` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `size`; starts at 1; page size 100; emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external Perigon API read of public news article, story,
journalist, source, company, people, and topic data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 7 stream-backed endpoint group(s).
