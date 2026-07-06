# Overview

Reads GIFs, stickers, and clips from the Giphy search and trending REST endpoints. Read-only.

Readable streams: `gif_search`, `sticker_search`, `clip_search`, `trending_gifs`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.giphy.com/docs/api/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Giphy API key. Sent as the api_key query parameter on every
  request; never logged.
- `base_url` (optional, string); default `https://api.giphy.com/v1`; format `uri`; Giphy API base
  URL override for tests or proxies.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `25`; Records per page (1-50; Giphy caps limit at 50).
- `query_for_clips` (optional, string); Search terms for the clip_search stream (required for that
  stream).
- `query_for_gif` (optional, string); Search terms for the gif_search stream (required for that
  stream).
- `query_for_stickers` (optional, string); Search terms for the sticker_search stream (required for
  that stream).
- `rating` (optional, string); Optional content rating filter (y, g, pg, pg-13, r) applied to every
  search/trending request.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.giphy.com/v1`, `max_pages=0`, `page_size=25`.

Authentication behavior:

- API key authentication in query parameter `api_key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/gifs/trending` with query `limit`=`1`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 25.

- `gif_search`: GET `/gifs/search` - records path `data`; query `q`=`{{ config.query_for_gif }}`;
  `rating` from template `{{ config.rating }}`, omitted when absent; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 25.
- `sticker_search`: GET `/stickers/search` - records path `data`; query `q`=`{{
  config.query_for_stickers }}`; `rating` from template `{{ config.rating }}`, omitted when absent;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 25.
- `clip_search`: GET `/clips/search` - records path `data`; query `q`=`{{ config.query_for_clips
  }}`; `rating` from template `{{ config.rating }}`, omitted when absent; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 25.
- `trending_gifs`: GET `/gifs/trending` - records path `data`; query `rating` from template `{{
  config.rating }}`, omitted when absent; offset/limit pagination; offset parameter `offset`; limit
  parameter `limit`; page size 25.

## Write actions & risks

This connector is read-only. Read behavior: external Giphy API read of public media search/trending
results.

## Known limits

- Batch defaults: read_page_size=25.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=5.
