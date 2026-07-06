# Overview

Reads Project Gutenberg books from the free, public Gutendex JSON API (books, popular, latest, and
English-language views). Read-only; no credentials required.

Readable streams: `books`, `popular_books`, `latest_books`, `english_books`.

This connector is read-only; no write actions are declared.

Service API documentation: https://gutendex.com/.

## Auth setup

Connection fields:

- `author_year_end` (optional, string); Optional upper bound on author birth year.
- `author_year_start` (optional, string); Optional lower bound on author birth year.
- `base_url` (optional, string); default `https://gutendex.com`; format `uri`; Gutendex API base URL
  override for tests or proxies.
- `copyright` (optional, string); Optional copyright filter (true/false/null, per Gutendex's own
  values).
- `ids` (optional, string); Optional comma-separated Gutenberg book ids to filter to.
- `languages` (optional, string); Optional comma-separated language codes to filter books by (e.g.
  "en,fr").
- `mode` (optional, string).
- `search` (optional, string); Optional search terms, forwarded to Gutendex's search query
  parameter.
- `sort` (optional, string); Optional sort override; the
  books/popular_books/latest_books/english_books streams already set this per-view and take
  precedence.
- `topic` (optional, string); Optional topic/subject filter.

Default configuration values: `base_url=https://gutendex.com`.

Authentication is handled by the connector-specific implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/books/`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `next`; next URLs stay
on the configured API host.

- `books`: GET `/books/` - records path `results`; query `author_year_end` from template `{{
  config.author_year_end }}`, omitted when absent; `author_year_start` from template `{{
  config.author_year_start }}`, omitted when absent; `copyright` from template `{{ config.copyright
  }}`, omitted when absent; `ids` from template `{{ config.ids }}`, omitted when absent; `languages`
  from template `{{ config.languages }}`, omitted when absent; `search` from template `{{
  config.search }}`, omitted when absent; `sort` from template `{{ config.sort }}`, omitted when
  absent; `topic` from template `{{ config.topic }}`, omitted when absent; follows a next-page URL
  from the response body; URL path `next`; next URLs stay on the configured API host; computed
  output fields `bookshelves`, `languages`, `subjects`.
- `popular_books`: GET `/books/` - records path `results`; query `author_year_end` from template `{{
  config.author_year_end }}`, omitted when absent; `author_year_start` from template `{{
  config.author_year_start }}`, omitted when absent; `copyright` from template `{{ config.copyright
  }}`, omitted when absent; `ids` from template `{{ config.ids }}`, omitted when absent; `languages`
  from template `{{ config.languages }}`, omitted when absent; `search` from template `{{
  config.search }}`, omitted when absent; `sort`=`popular`; `topic` from template `{{ config.topic
  }}`, omitted when absent; follows a next-page URL from the response body; URL path `next`; next
  URLs stay on the configured API host; computed output fields `bookshelves`, `languages`,
  `subjects`.
- `latest_books`: GET `/books/` - records path `results`; query `author_year_end` from template `{{
  config.author_year_end }}`, omitted when absent; `author_year_start` from template `{{
  config.author_year_start }}`, omitted when absent; `copyright` from template `{{ config.copyright
  }}`, omitted when absent; `ids` from template `{{ config.ids }}`, omitted when absent; `languages`
  from template `{{ config.languages }}`, omitted when absent; `search` from template `{{
  config.search }}`, omitted when absent; `sort`=`descending`; `topic` from template `{{
  config.topic }}`, omitted when absent; follows a next-page URL from the response body; URL path
  `next`; next URLs stay on the configured API host; computed output fields `bookshelves`,
  `languages`, `subjects`.
- `english_books`: GET `/books/` - records path `results`; query `author_year_end` from template `{{
  config.author_year_end }}`, omitted when absent; `author_year_start` from template `{{
  config.author_year_start }}`, omitted when absent; `copyright` from template `{{ config.copyright
  }}`, omitted when absent; `ids` from template `{{ config.ids }}`, omitted when absent;
  `languages`=`en`; `search` from template `{{ config.search }}`, omitted when absent; `sort` from
  template `{{ config.sort }}`, omitted when absent; `topic` from template `{{ config.topic }}`,
  omitted when absent; follows a next-page URL from the response body; URL path `next`; next URLs
  stay on the configured API host; computed output fields `bookshelves`, `languages`, `subjects`.

## Write actions & risks

This connector is read-only. Read behavior: external read of the public, unauthenticated Gutendex
book catalog.

## Known limits

- Batch defaults: read_page_size=32.
- API coverage includes 4 stream-backed endpoint group(s).
