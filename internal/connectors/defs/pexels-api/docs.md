# Overview

Reads Pexels photo/video search and curated/popular results plus featured and personal collections
and their media through the Pexels REST API.

Readable streams: `photos`, `curated_photos`, `videos`, `popular_videos`, `featured_collections`,
`my_collections`, `collection_media`.

This connector is read-only; no write actions are declared.

Service API documentation: https://www.pexels.com/api/documentation/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Pexels API key, sent as the raw Authorization header value
  (no Bearer prefix). Never logged.
- `base_url` (optional, string); default `https://api.pexels.com`; format `uri`; Pexels API base URL
  override for tests or proxies.
- `collection_media_sort` (optional, string); Optional sort order for the 'collection_media' stream:
  'asc' or 'desc'.
- `collection_media_type` (optional, string); Optional filter for the 'collection_media' stream:
  'photos' or 'videos' (omitted = both).
- `color` (optional, string); Optional photo color filter.
- `locale` (optional, string); Optional locale for search results.
- `orientation` (optional, string); Optional photo/video orientation filter (landscape, portrait,
  square).
- `query` (optional, string); default `people`.
- `size` (optional, string); Optional minimum photo/video size filter (large, medium, small).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.pexels.com`, `query=people`.

Authentication behavior:

- API key authentication in `Authorization` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v1/search`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `next_page`; next URLs
stay on the configured API host.

- `photos`: GET `/v1/search` - records path `photos`; query `color` from template `{{ config.color
  }}`, omitted when absent; `locale` from template `{{ config.locale }}`, omitted when absent;
  `orientation` from template `{{ config.orientation }}`, omitted when absent; `page`=`1`;
  `per_page`=`40`; `query` from template `{{ config.query }}`, default `people`; `size` from
  template `{{ config.size }}`, omitted when absent; follows a next-page URL from the response body;
  URL path `next_page`; next URLs stay on the configured API host.
- `curated_photos`: GET `/v1/curated` - records path `photos`; query `color` from template `{{
  config.color }}`, omitted when absent; `locale` from template `{{ config.locale }}`, omitted when
  absent; `orientation` from template `{{ config.orientation }}`, omitted when absent; `page`=`1`;
  `per_page`=`40`; `size` from template `{{ config.size }}`, omitted when absent; follows a
  next-page URL from the response body; URL path `next_page`; next URLs stay on the configured API
  host.
- `videos`: GET `/v1/videos/search` - records path `videos`; query `color` from template `{{
  config.color }}`, omitted when absent; `locale` from template `{{ config.locale }}`, omitted when
  absent; `orientation` from template `{{ config.orientation }}`, omitted when absent; `page`=`1`;
  `per_page`=`40`; `query` from template `{{ config.query }}`, default `people`; `size` from
  template `{{ config.size }}`, omitted when absent; follows a next-page URL from the response body;
  URL path `next_page`; next URLs stay on the configured API host.
- `popular_videos`: GET `/v1/videos/popular` - records path `videos`; query `color` from template
  `{{ config.color }}`, omitted when absent; `locale` from template `{{ config.locale }}`, omitted
  when absent; `orientation` from template `{{ config.orientation }}`, omitted when absent;
  `page`=`1`; `per_page`=`40`; `size` from template `{{ config.size }}`, omitted when absent;
  follows a next-page URL from the response body; URL path `next_page`; next URLs stay on the
  configured API host.
- `featured_collections`: GET `/v1/collections/featured` - records path `collections`; query
  `page`=`1`; `per_page`=`40`; follows a next-page URL from the response body; URL path `next_page`;
  next URLs stay on the configured API host.
- `my_collections`: GET `/v1/collections` - records path `collections`; query `page`=`1`;
  `per_page`=`40`; follows a next-page URL from the response body; URL path `next_page`; next URLs
  stay on the configured API host.
- `collection_media`: GET `/v1/collections/{{ fanout.id }}` - records path `media`; query
  `page`=`1`; `per_page`=`40`; `sort` from template `{{ config.collection_media_sort }}`, omitted
  when absent; `type` from template `{{ config.collection_media_type }}`, omitted when absent;
  follows a next-page URL from the response body; URL path `next_page`; next URLs stay on the
  configured API host; fan-out; ids from request `/v1/collections`; id-list records path
  `collections`; id field `id`; id inserted into the request path; stamps `collection_id`; emits
  passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external Pexels API read of photo/video search,
curated/popular results, and collection metadata/media; all publicly-licensed stock media, no PII.

## Known limits

- Batch defaults: read_page_size=40.
- API coverage includes 7 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  requires_elevated_scope=2.
