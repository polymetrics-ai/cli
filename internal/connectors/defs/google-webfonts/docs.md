# Overview

Reads Google Web Fonts families (default, popular, trending, newest, and alphabetical views) through
the Google Fonts Developer API. Read-only.

Readable streams: `webfonts`, `popular_fonts`, `trending_fonts`, `newest_fonts`, `alpha_fonts`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.google.com/fonts/docs/developer_api.

## Auth setup

Connection fields:

- `alt` (optional, string); Optional passthrough: alternate response representation (API-standard
  `alt` query parameter, e.g. json).
- `api_key` (required, secret, string).
- `base_url` (optional, string); default `https://www.googleapis.com/webfonts/v1`; format `uri`;
  Google Fonts Developer API base URL.
- `capability` (optional, string); Optional passthrough filter: restrict results to families
  supporting this comma-separated list of capabilities (e.g. WOFF2, VF).
- `category` (optional, string); Optional passthrough filter: restrict results to this
  comma-separated list of font categories (e.g. serif, sans-serif).
- `family` (optional, string); Optional passthrough filter: restrict results to a comma-separated
  list of font family names.
- `pretty_print` (optional, string); Optional passthrough: API-standard `prettyPrint` query
  parameter (sent verbatim as prettyPrint).
- `subset` (optional, string); Optional passthrough filter: restrict results to families supporting
  this comma-separated list of subsets.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://www.googleapis.com/webfonts/v1`.

Authentication behavior:

- API key authentication in query parameter `key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/webfonts`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `pageToken`; next token from
`nextPageToken`; maximum 100 page(s).

- `webfonts`: GET `/webfonts` - records path `items`; query `alt` from template `{{ config.alt }}`,
  omitted when absent; `capability` from template `{{ config.capability }}`, omitted when absent;
  `category` from template `{{ config.category }}`, omitted when absent; `family` from template `{{
  config.family }}`, omitted when absent; `prettyPrint` from template `{{ config.pretty_print }}`,
  omitted when absent; `subset` from template `{{ config.subset }}`, omitted when absent; cursor
  pagination; cursor parameter `pageToken`; next token from `nextPageToken`; maximum 100 page(s);
  computed output fields `subset_count`, `variant_count`.
- `popular_fonts`: GET `/webfonts` - records path `items`; query `alt` from template `{{ config.alt
  }}`, omitted when absent; `capability` from template `{{ config.capability }}`, omitted when
  absent; `category` from template `{{ config.category }}`, omitted when absent; `family` from
  template `{{ config.family }}`, omitted when absent; `prettyPrint` from template `{{
  config.pretty_print }}`, omitted when absent; `sort`=`popularity`; `subset` from template `{{
  config.subset }}`, omitted when absent; cursor pagination; cursor parameter `pageToken`; next
  token from `nextPageToken`; maximum 100 page(s); computed output fields `subset_count`,
  `variant_count`.
- `trending_fonts`: GET `/webfonts` - records path `items`; query `alt` from template `{{ config.alt
  }}`, omitted when absent; `capability` from template `{{ config.capability }}`, omitted when
  absent; `category` from template `{{ config.category }}`, omitted when absent; `family` from
  template `{{ config.family }}`, omitted when absent; `prettyPrint` from template `{{
  config.pretty_print }}`, omitted when absent; `sort`=`trending`; `subset` from template `{{
  config.subset }}`, omitted when absent; cursor pagination; cursor parameter `pageToken`; next
  token from `nextPageToken`; maximum 100 page(s); computed output fields `subset_count`,
  `variant_count`.
- `newest_fonts`: GET `/webfonts` - records path `items`; query `alt` from template `{{ config.alt
  }}`, omitted when absent; `capability` from template `{{ config.capability }}`, omitted when
  absent; `category` from template `{{ config.category }}`, omitted when absent; `family` from
  template `{{ config.family }}`, omitted when absent; `prettyPrint` from template `{{
  config.pretty_print }}`, omitted when absent; `sort`=`date`; `subset` from template `{{
  config.subset }}`, omitted when absent; cursor pagination; cursor parameter `pageToken`; next
  token from `nextPageToken`; maximum 100 page(s); computed output fields `subset_count`,
  `variant_count`.
- `alpha_fonts`: GET `/webfonts` - records path `items`; query `alt` from template `{{ config.alt
  }}`, omitted when absent; `capability` from template `{{ config.capability }}`, omitted when
  absent; `category` from template `{{ config.category }}`, omitted when absent; `family` from
  template `{{ config.family }}`, omitted when absent; `prettyPrint` from template `{{
  config.pretty_print }}`, omitted when absent; `sort`=`alpha`; `subset` from template `{{
  config.subset }}`, omitted when absent; cursor pagination; cursor parameter `pageToken`; next
  token from `nextPageToken`; maximum 100 page(s); computed output fields `subset_count`,
  `variant_count`.

## Write actions & risks

This connector is read-only. Read behavior: external Google Fonts Developer API read of public font
metadata.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
