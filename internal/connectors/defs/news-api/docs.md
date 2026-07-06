# Overview

Reads articles and news sources from the News API (newsapi.org): the everything search, top
headlines, and the sources directory.

Readable streams: `everything`, `top_headlines`, `sources`.

This connector is read-only; no write actions are declared.

Service API documentation: https://newsapi.org/docs.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); News API (newsapi.org) API key. Sent as the X-Api-Key header
  on every request; never logged.
- `base_url` (optional, string); default `https://newsapi.org/v2`; format `uri`; News API base URL
  override; defaults to production.
- `category` (optional, string); News category for the top-headlines/sources streams: business,
  entertainment, general, health, science, sports, technology.
- `country` (optional, string); 2-letter ISO 3166-1 country code for the top-headlines/sources
  streams.
- `domains` (optional, string); Comma-separated domains to restrict the everything search to.
- `end_date` (optional, string); format `date-time`; Newest article publish date/time (`to`) for the
  everything search.
- `exclude_domains` (optional, string); Comma-separated domains to exclude from the everything
  search.
- `language` (optional, string); 2-letter ISO-639-1 language code to restrict the
  everything/top-headlines/sources streams to.
- `search_in` (optional, string); Comma-separated fields to restrict the everything search to
  (`searchIn`): title, description, content.
- `search_query` (optional, string); Free-text query (`q` param) for the everything and
  top-headlines streams.
- `sort_by` (optional, string); Sort order for the everything search: relevancy, popularity, or
  publishedAt.
- `sources` (optional, string); Comma-separated News API source identifiers to restrict the
  everything/top-headlines streams to.
- `start_date` (optional, string); format `date-time`; Oldest article publish date/time (`from`) for
  the everything search. Also used as the incremental sync start when no cursor state exists yet.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://newsapi.org/v2`.

Authentication behavior:

- API key authentication in `X-Api-Key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/top-headlines/sources`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `sources`; page_number: `everything`, `top_headlines`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `everything`: GET `/everything` - records path `articles`; query `domains` from template `{{
  config.domains }}`, omitted when absent; `excludeDomains` from template `{{ config.exclude_domains
  }}`, omitted when absent; `language` from template `{{ config.language }}`, omitted when absent;
  `q` from template `{{ config.search_query }}`, omitted when absent; `searchIn` from template `{{
  config.search_in }}`, omitted when absent; `sortBy` from template `{{ config.sort_by }}`, omitted
  when absent; `sources` from template `{{ config.sources }}`, omitted when absent; `to` from
  template `{{ config.end_date }}`, omitted when absent; page-number pagination; page parameter
  `page`; size parameter `pageSize`; starts at 1; page size 100; incremental cursor `published_at`;
  sent as `from`; formatted as `rfc3339`; initial lower bound from `start_date`; computed output
  fields `published_at`, `source_id`, `source_name`, `url_to_image`.
- `top_headlines`: GET `/top-headlines` - records path `articles`; query `category` from template
  `{{ config.category }}`, omitted when absent; `country` from template `{{ config.country }}`,
  omitted when absent; `language` from template `{{ config.language }}`, omitted when absent; `q`
  from template `{{ config.search_query }}`, omitted when absent; `sources` from template `{{
  config.sources }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `pageSize`; starts at 1; page size 100; computed output fields `published_at`,
  `source_id`, `source_name`, `url_to_image`.
- `sources`: GET `/top-headlines/sources` - records path `sources`; query `category` from template
  `{{ config.category }}`, omitted when absent; `country` from template `{{ config.country }}`,
  omitted when absent; `language` from template `{{ config.language }}`, omitted when absent.

## Write actions & risks

This connector is read-only. Read behavior: external News API read of article and source metadata.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s).
