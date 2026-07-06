# Overview

Reads Watchmode title search results, streaming sources, regions, networks, genres, list-titles,
releases, per-title details/sources/seasons/episodes/cast-crew, and person details. Read-only.

Readable streams: `search`, `sources`, `regions`, `networks`, `genres`, `titles`, `releases`,
`title_details`, `title_sources`, `title_seasons`, `title_episodes`, `title_cast_crew`,
`person_details`.

This connector is read-only; no write actions are declared.

Service API documentation: https://api.watchmode.com/docs.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Watchmode API key. Sent as the apiKey query parameter on
  every request; never logged.
- `base_url` (optional, string); default `https://api.watchmode.com`; format `uri`; Watchmode API
  base URL override for tests or proxies.
- `mode` (optional, string).
- `page_size` (optional, string); Optional limit query parameter forwarded to the titles and
  changes-feed streams (list-titles/changes endpoints' page size). Watchmode defaults these to 250
  server-side when unset.
- `person_ids` (optional, string); Comma-, whitespace-, or semicolon-separated Watchmode person IDs
  to fan out over for the person_details stream (one request per id). Required for that stream only.
- `regions` (optional, string); Optional comma-separated region codes (e.g. US,CA) forwarded to the
  titles, sources, releases, and title_sources streams when set.
- `search_val` (optional, string); default `Terminator`; Search term for the search stream's
  search_value query parameter (matched against title name).
- `start_date` (optional, string); Optional start_date query parameter forwarded to the search
  stream when set.
- `title_ids` (optional, string); Comma-, whitespace-, or semicolon-separated Watchmode title IDs to
  fan out over for the title_details/title_sources/title_seasons/title_episodes/title_cast_crew
  streams (one request per id). Required for those 5 streams only; other streams do not reference
  it.
- `types` (optional, string); Optional comma-separated title types (e.g. movie,tv_series) forwarded
  to the titles and changes new_titles streams when set.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.watchmode.com`, `search_val=Terminator`.

Authentication behavior:

- API key authentication in query parameter `apiKey` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v1/sources/`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `search`, `sources`, `regions`, `networks`, `genres`, `releases`,
`title_details`, `title_sources`, `title_seasons`, `title_episodes`, `title_cast_crew`,
`person_details`; page_number: `titles`.

- `search`: GET `/v1/search/` - records path `title_results`; query `search_field`=`name`;
  `search_value` from template `{{ config.search_val }}`, default `Terminator`; `start_date` from
  template `{{ config.start_date }}`, omitted when absent; emits passthrough records.
- `sources`: GET `/v1/sources/` - records path `.`; query `regions` from template `{{ config.regions
  }}`, omitted when absent; `start_date` from template `{{ config.start_date }}`, omitted when
  absent; emits passthrough records.
- `regions`: GET `/v1/regions/` - records path `.`; emits passthrough records.
- `networks`: GET `/v1/networks/` - records path `.`; emits passthrough records.
- `genres`: GET `/v1/genres/` - records path `.`; emits passthrough records.
- `titles`: GET `/v1/list-titles/` - records path `titles`; query `limit` from template `{{
  config.page_size }}`, omitted when absent; `regions` from template `{{ config.regions }}`, omitted
  when absent; `types` from template `{{ config.types }}`, omitted when absent; page-number
  pagination; page parameter `page`; no page-size parameter; starts at 1; page size 250; emits
  passthrough records.
- `releases`: GET `/v1/releases/` - records path `releases`; query `regions` from template `{{
  config.regions }}`, omitted when absent; emits passthrough records.
- `title_details`: GET `/v1/title/{{ fanout.id }}/details/` - records path `.`; fan-out; ids from
  config field `title_ids`; id inserted into the request path; stamps `watchmode_title_id`; emits
  passthrough records.
- `title_sources`: GET `/v1/title/{{ fanout.id }}/sources/` - records path `.`; fan-out; ids from
  config field `title_ids`; id inserted into the request path; stamps `watchmode_title_id`; emits
  passthrough records.
- `title_seasons`: GET `/v1/title/{{ fanout.id }}/seasons/` - records path `.`; fan-out; ids from
  config field `title_ids`; id inserted into the request path; stamps `watchmode_title_id`; emits
  passthrough records.
- `title_episodes`: GET `/v1/title/{{ fanout.id }}/episodes/` - records path `.`; fan-out; ids from
  config field `title_ids`; id inserted into the request path; stamps `watchmode_title_id`; emits
  passthrough records.
- `title_cast_crew`: GET `/v1/title/{{ fanout.id }}/cast-crew/` - records path `.`; fan-out; ids
  from config field `title_ids`; id inserted into the request path; stamps `watchmode_title_id`;
  emits passthrough records.
- `person_details`: GET `/v1/person/{{ fanout.id }}/` - records path `.`; fan-out; ids from config
  field `person_ids`; id inserted into the request path; stamps `watchmode_person_id`; emits
  passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external Watchmode API read of public
title/streaming-source/person media metadata.

## Known limits

- API coverage includes 13 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=3, out_of_scope=5, requires_elevated_scope=1.
