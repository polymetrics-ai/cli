# Overview

Reads channels, videos, playlists, playlist items, comment threads, search results, video
categories, and i18n region/language reference data through the YouTube Data API.

Readable streams: `channels`, `videos`, `playlists`, `playlist_items`, `comment_threads`, `search`,
`video_categories`, `i18n_regions`, `i18n_languages`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.google.com/youtube/v3.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); YouTube Data API key, sent as the 'key' query parameter on
  every request. Never logged.
- `base_url` (optional, string); default `https://www.googleapis.com/youtube/v3`; format `uri`;
  YouTube Data API root. Also usable as a base URL override for tests/proxies.
- `channel_ids` (optional, string); Comma-separated channel IDs, sent as the 'id' query parameter
  for the 'channels' stream. Omitted entirely when unset.
- `ids` (optional, string); Comma-separated video/playlist IDs, sent as the 'id' query parameter for
  the 'videos' and 'playlists' streams. Omitted entirely when unset.
- `mode` (optional, string).
- `playlist_ids` (optional, string); Comma-separated playlist IDs to fan out over for the
  'playlist_items' stream (one request sequence per id, stamped onto every emitted record's
  playlist_id field). Required for that stream to emit anything; a full-refresh read with this unset
  simply yields zero playlist_items records.
- `region_code` (optional, string); default `US`; ISO 3166-1 alpha-2 region code for the
  'video_categories' stream (YouTube's video category list is region-scoped; there is no all-regions
  listing).
- `search_query` (optional, string); Search term for the 'search' stream's 'q' query parameter.
  Omitted entirely when unset (an unfiltered public-content search).
- `video_ids` (optional, string); Comma-separated video IDs to fan out over for the
  'comment_threads' stream (one request sequence per id, stamped onto every emitted record's
  video_id field). Required for that stream to emit anything; a full-refresh read with this unset
  simply yields zero comment_threads records.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://www.googleapis.com/youtube/v3`, `region_code=US`.

Authentication behavior:

- API key authentication in query parameter `key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: cursor: `playlist_items`, `comment_threads`, `search`; none: `channels`,
`videos`, `playlists`, `video_categories`, `i18n_regions`, `i18n_languages`.

- `channels`: GET `/channels` - records path `items`; query `id` from template `{{
  config.channel_ids }}`, omitted when absent; `part`=`snippet,statistics`; computed output fields
  `title`, `view_count`.
- `videos`: GET `/videos` - records path `items`; query `id` from template `{{ config.ids }}`,
  omitted when absent; `part`=`snippet,statistics`; computed output fields `published_at`, `title`.
- `playlists`: GET `/playlists` - records path `items`; query `id` from template `{{ config.ids }}`,
  omitted when absent; `part`=`snippet,statistics`; computed output fields `published_at`, `title`.
- `playlist_items`: GET `/playlistItems` - records path `items`; query `maxResults`=`50`;
  `part`=`snippet,contentDetails`; cursor pagination; cursor parameter `pageToken`; next token from
  `nextPageToken`; computed output fields `published_at`, `title`, `video_id`; fan-out; ids from
  config field `playlist_ids`; id sent as query parameter `playlistId`; stamps `playlist_id`.
- `comment_threads`: GET `/commentThreads` - records path `items`; query `maxResults`=`100`;
  `part`=`snippet`; cursor pagination; cursor parameter `pageToken`; next token from
  `nextPageToken`; computed output fields `published_at`, `text`; fan-out; ids from config field
  `video_ids`; id sent as query parameter `videoId`; stamps `video_id`.
- `search`: GET `/search` - records path `items`; query `channelId` from template `{{
  config.channel_ids }}`, omitted when absent; `maxResults`=`50`; `part`=`snippet`; `q` from
  template `{{ config.search_query }}`, omitted when absent; `type`=`video`; cursor pagination;
  cursor parameter `pageToken`; next token from `nextPageToken`; computed output fields `id`,
  `published_at`, `title`.
- `video_categories`: GET `/videoCategories` - records path `items`; query `part`=`snippet`;
  `regionCode` from template `{{ config.region_code }}`, default `US`; computed output fields
  `title`.
- `i18n_regions`: GET `/i18nRegions` - records path `items`; query `part`=`snippet`; computed output
  fields `name`.
- `i18n_languages`: GET `/i18nLanguages` - records path `items`; query `part`=`snippet`; computed
  output fields `name`.

## Write actions & risks

This connector is read-only. Read behavior: external YouTube Data API read of public channel, video,
playlist, playlist item, comment, search result, and reference data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 9 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=6, out_of_scope=2, requires_elevated_scope=32.
