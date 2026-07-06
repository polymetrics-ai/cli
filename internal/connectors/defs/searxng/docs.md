# Overview

Reads web and Reddit search results from a SearXNG metasearch instance's JSON API (format=json).
Read-only. Requires base_url; no credentials by default.

Readable streams: `search`, `reddit`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.searxng.org.

## Auth setup

Connection fields:

- `api_key` (optional, secret, string); Optional Bearer token for SearXNG instances behind an auth
  proxy, applied via streams.json base.auth's when-gated bearer spec (falls back to no auth when
  unset). Public instances are open and need no credentials.
- `base_url` (required, string); format `uri`; Your SearXNG instance's base URL (e.g.
  https://searx.example.com). No default: a SearXNG instance must be named explicitly.
- `query` (optional, string); Search terms for the 'search' stream (required for that stream) and
  appended to the reddit-scoped query for the 'reddit' stream.

Secret fields are redacted in logs and write previews: `api_key`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key` when `{{ secrets.api_key }}`.
- No authentication.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/search`.

## Streams notes

Default pagination: page-number pagination; page parameter `pageno`; starts at 1; page size 10;
maximum 1 page(s).

- `search`: GET `/search` - records path `results`; query `format`=`json`; `q`=`{{ config.query }}`;
  page-number pagination; page parameter `pageno`; starts at 1; page size 10; maximum 1 page(s);
  computed output fields `engines`, `published_date`, `stream`.
- `reddit`: GET `/search` - records path `results`; query `format`=`json`; `q`=`site:reddit.com {{
  config.query }}`; page-number pagination; page parameter `pageno`; starts at 1; page size 10;
  maximum 1 page(s); computed output fields `engines`, `published_date`, `stream`.

## Write actions & risks

This connector is read-only. Read behavior: external SearXNG instance read of web/Reddit search
results.

## Known limits

- Batch defaults: read_page_size=10.
- API coverage includes 2 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=2.
