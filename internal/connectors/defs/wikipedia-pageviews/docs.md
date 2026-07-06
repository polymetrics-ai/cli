# Overview

Reads Wikimedia pageview metrics for articles and top-article reports through the public Wikimedia
REST API.

Readable streams: `pageviews`, `top_articles`.

This connector is read-only; no write actions are declared.

Service API documentation: https://wikitech.wikimedia.org/wiki/Analytics/AQS/Pageviews.

## Auth setup

Connection fields:

- `access` (required, string); Access method filter for the pageviews endpoint, e.g. all-access,
  desktop, mobile-app, mobile-web.
- `agent` (required, string); Agent type filter for the pageviews endpoint, e.g. all-agents, user,
  spider, automated.
- `article` (required, string); Article title (as it appears in the Wikimedia URL, e.g.
  Ada_Lovelace) for the per-article pageviews stream.
- `base_url` (optional, string); default `https://wikimedia.org`; format `uri`; Wikimedia REST API
  base URL override for tests or proxies.
- `country` (optional, string); ISO country code for the top_articles stream, e.g. US. Required only
  when reading the top_articles stream.
- `day` (optional, string); 2-digit day for the top_articles stream, e.g. 01. Required only when
  reading the top_articles stream.
- `end` (required, string); End date/time for the pageviews stream's requested range, YYYYMMDD or
  YYYYMMDDHH.
- `month` (optional, string); 2-digit month for the top_articles stream, e.g. 01. Required only when
  reading the top_articles stream.
- `project` (required, string); Wikimedia project domain, e.g. en.wikipedia.org.
- `start` (required, string); Start date/time for the pageviews stream's requested range, YYYYMMDD
  or YYYYMMDDHH.
- `year` (optional, string); 4-digit year for the top_articles stream, e.g. 2026.

Default configuration values: `base_url=https://wikimedia.org`.

Authentication behavior:

- No authentication.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `api/rest_v1/metrics/pageviews/per-article/{{ config.project }}/{{
config.access }}/{{ config.agent }}/{{ config.article }}/daily/{{ config.start }}/{{ config.end }}`.

## Streams notes

Default pagination: single request; no pagination.

- `pageviews`: GET `api/rest_v1/metrics/pageviews/per-article/{{ config.project }}/{{ config.access
  }}/{{ config.agent }}/{{ config.article }}/daily/{{ config.start }}/{{ config.end }}` - records
  path `items`; computed output fields `id`; emits passthrough records.
- `top_articles`: GET `api/rest_v1/metrics/pageviews/top-per-country/{{ config.project }}/{{
  config.country }}/{{ config.access }}/{{ config.year }}/{{ config.month }}/{{ config.day }}` -
  records path `items`; computed output fields `id`; emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external Wikimedia public API read of aggregate pageview
metrics; no authentication, no PII.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 2 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=3.
