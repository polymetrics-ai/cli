# Overview

Reads latest news, cryptocurrency news, and news sources from the NewsData.io REST API.

Readable streams: `latest`, `crypto`, `sources`.

This connector is read-only; no write actions are declared.

Service API documentation: https://newsdata.io/documentation.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); NewsData.io API key. Sent as the `apikey` query parameter on
  every request; never logged.
- `base_url` (optional, string); default `https://newsdata.io/api/1`; format `uri`; NewsData.io API
  base URL override; defaults to production.
- `category` (optional, string); Comma-separated category filter forwarded as the `category` query
  parameter.
- `country` (optional, string); Comma-separated country filter forwarded as the `country` query
  parameter.
- `domain` (optional, string); Comma-separated domain filter forwarded as the `domain` query
  parameter.
- `language` (optional, string); Comma-separated language filter forwarded as the `language` query
  parameter.
- `query` (optional, string); Free-text search query forwarded as the `q` query parameter.
- `query_in_title` (optional, string); Free-text search query restricted to article titles,
  forwarded as the `qInTitle` query parameter.
- `size` (optional, string); Page size forwarded as the `size` query parameter (NewsData.io accepts
  this as a string-shaped value in its own docs).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://newsdata.io/api/1`.

Authentication behavior:

- API key authentication in query parameter `apikey` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/latest`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `page`; next token from `nextPage`; maximum
5 page(s).

Pagination by stream: cursor: `latest`, `crypto`; none: `sources`.

- `latest`: GET `/latest` - records path `results`; query `category` from template `{{
  config.category }}`, omitted when absent; `country` from template `{{ config.country }}`, omitted
  when absent; `domain` from template `{{ config.domain }}`, omitted when absent; `language` from
  template `{{ config.language }}`, omitted when absent; `q` from template `{{ config.query }}`,
  omitted when absent; `qInTitle` from template `{{ config.query_in_title }}`, omitted when absent;
  `size` from template `{{ config.size }}`, omitted when absent; cursor pagination; cursor parameter
  `page`; next token from `nextPage`; maximum 5 page(s).
- `crypto`: GET `/crypto` - records path `results`; query `category` from template `{{
  config.category }}`, omitted when absent; `country` from template `{{ config.country }}`, omitted
  when absent; `domain` from template `{{ config.domain }}`, omitted when absent; `language` from
  template `{{ config.language }}`, omitted when absent; `q` from template `{{ config.query }}`,
  omitted when absent; `qInTitle` from template `{{ config.query_in_title }}`, omitted when absent;
  `size` from template `{{ config.size }}`, omitted when absent; cursor pagination; cursor parameter
  `page`; next token from `nextPage`; maximum 5 page(s).
- `sources`: GET `/sources` - records path `results`; query `category` from template `{{
  config.category }}`, omitted when absent; `country` from template `{{ config.country }}`, omitted
  when absent; `domain` from template `{{ config.domain }}`, omitted when absent; `language` from
  template `{{ config.language }}`, omitted when absent; `q` from template `{{ config.query }}`,
  omitted when absent; `qInTitle` from template `{{ config.query_in_title }}`, omitted when absent;
  `size` from template `{{ config.size }}`, omitted when absent.

## Write actions & risks

This connector is read-only. Read behavior: external NewsData.io API read of article and source
metadata.

## Known limits

- Batch defaults: read_page_size=10.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=1.
