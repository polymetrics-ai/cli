# Overview

Reads latest, crypto, and archived news articles plus available news sources from the NewsData.io
REST API.

Readable streams: `latest`, `crypto`, `archive`, `sources`.

This connector is read-only; no write actions are declared.

Service API documentation: https://newsdata.io/documentation.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); NewsData.io API key, sent as the 'apikey' query parameter.
  Never logged.
- `base_url` (optional, string); default `https://newsdata.io/api/1`; format `uri`; NewsData.io API
  base URL override for tests or proxies.
- `categories` (optional, string); Optional comma-separated category filter, sent as the 'category'
  query parameter.
- `countries` (optional, string); Optional comma-separated country filter, sent as the 'country'
  query parameter.
- `domains` (optional, string); Optional comma-separated domain filter, sent as the 'domain' query
  parameter.
- `end_date` (optional, string); Optional upper-bound date (YYYY-MM-DD), sent as 'to_date'.
- `languages` (optional, string); Optional comma-separated language filter, sent as the 'language'
  query parameter.
- `mode` (optional, string).
- `page_size` (optional, string); default `10`; Records per page (1-50), sent as the 'size' query
  parameter on article streams.
- `search_query` (optional, string); Optional search keywords, sent as the 'q' query parameter.
- `start_date` (optional, string); Optional lower-bound date (YYYY-MM-DD), sent as 'from_date'.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://newsdata.io/api/1`, `page_size=10`.

Authentication behavior:

- API key authentication in query parameter `apikey` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/latest` with query `size`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `page`; next token from `nextPage`; maximum
10 page(s).

Pagination by stream: cursor: `latest`, `crypto`, `archive`; none: `sources`.

- `latest`: GET `/latest` - records path `results`; query `category` from template `{{
  config.categories }}`, omitted when absent; `country` from template `{{ config.countries }}`,
  omitted when absent; `domain` from template `{{ config.domains }}`, omitted when absent;
  `from_date` from template `{{ config.start_date }}`, omitted when absent; `language` from template
  `{{ config.languages }}`, omitted when absent; `q` from template `{{ config.search_query }}`,
  omitted when absent; `size` from template `{{ config.page_size }}`, default `10`; `to_date` from
  template `{{ config.end_date }}`, omitted when absent; cursor pagination; cursor parameter `page`;
  next token from `nextPage`; maximum 10 page(s).
- `crypto`: GET `/crypto` - records path `results`; query `category` from template `{{
  config.categories }}`, omitted when absent; `country` from template `{{ config.countries }}`,
  omitted when absent; `domain` from template `{{ config.domains }}`, omitted when absent;
  `from_date` from template `{{ config.start_date }}`, omitted when absent; `language` from template
  `{{ config.languages }}`, omitted when absent; `q` from template `{{ config.search_query }}`,
  omitted when absent; `size` from template `{{ config.page_size }}`, default `10`; `to_date` from
  template `{{ config.end_date }}`, omitted when absent; cursor pagination; cursor parameter `page`;
  next token from `nextPage`; maximum 10 page(s).
- `archive`: GET `/archive` - records path `results`; query `category` from template `{{
  config.categories }}`, omitted when absent; `country` from template `{{ config.countries }}`,
  omitted when absent; `domain` from template `{{ config.domains }}`, omitted when absent;
  `from_date` from template `{{ config.start_date }}`, omitted when absent; `language` from template
  `{{ config.languages }}`, omitted when absent; `q` from template `{{ config.search_query }}`,
  omitted when absent; `size` from template `{{ config.page_size }}`, default `10`; `to_date` from
  template `{{ config.end_date }}`, omitted when absent; cursor pagination; cursor parameter `page`;
  next token from `nextPage`; maximum 10 page(s).
- `sources`: GET `/sources` - records path `results`; query `category` from template `{{
  config.categories }}`, omitted when absent; `country` from template `{{ config.countries }}`,
  omitted when absent; `domain` from template `{{ config.domains }}`, omitted when absent;
  `from_date` from template `{{ config.start_date }}`, omitted when absent; `language` from template
  `{{ config.languages }}`, omitted when absent; `q` from template `{{ config.search_query }}`,
  omitted when absent; `to_date` from template `{{ config.end_date }}`, omitted when absent.

## Write actions & risks

This connector is read-only. Read behavior: external NewsData.io API read of news articles and
sources.

## Known limits

- Batch defaults: read_page_size=10.
- API coverage includes 4 stream-backed endpoint group(s).
