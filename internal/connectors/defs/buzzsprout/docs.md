# Overview

Reads Buzzsprout podcasts and episodes (titles, publish dates, durations, play counts) and
creates/updates episodes through the Buzzsprout REST API.

Readable streams: `episodes`, `podcasts`.

Write actions: `create_episode`, `update_episode`.

Service API documentation: https://github.com/Buzzsprout/buzzsprout-api.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Buzzsprout API key. Sent as 'Authorization: Token
  token=<api_key>'. Never logged.
- `base_url` (optional, string); default `https://www.buzzsprout.com`; format `uri`; Buzzsprout API
  base URL override for tests or proxies.
- `podcast_id` (optional, string); Buzzsprout podcast id. Required for the 'episodes' stream
  (substituted into its path); not used by the account-level 'podcasts' stream.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://www.buzzsprout.com`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `Token token=` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/podcasts.json`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; no page-size parameter; starts at
1; page size 100.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `episodes`: GET `/api/{{ config.podcast_id }}/episodes.json` - records at response root;
  page-number pagination; page parameter `page`; no page-size parameter; starts at 1; page size 100;
  incremental cursor `published_at`; formatted as `rfc3339`.
- `podcasts`: GET `/api/podcasts.json` - records at response root; page-number pagination; page
  parameter `page`; no page-size parameter; starts at 1; page size 100.

## Write actions & risks

Overall write risk: external mutation of episode metadata/audio (create_episode, update_episode) on
the configured podcast; can trigger audio processing and publish/unpublish an episode.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_episode`: POST `/api/{{ config.podcast_id }}/episodes.json` - kind `create`; body type
  `json`; accepted fields `artist`, `artwork_url`, `audio_url`, `description`, `duration`,
  `email_user_after_audio_processed`, `episode_number`, `explicit`, `guid`, `inactive_at`,
  `private`, `published_at`, `season_number`, `summary`, `tags`, `title`; risk: external mutation;
  creates a new episode (and can trigger audio processing/publication) on the configured podcast;
  approval required.
- `update_episode`: PUT `/api/{{ config.podcast_id }}/episodes/{{ record.id }}.json` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `artist`, `artwork_url`, `audio_url`, `description`, `duration`, `episode_number`, `explicit`,
  `guid`, `id`, `inactive_at`, `private`, `published_at`, `season_number`, `summary`, `tags`,
  `title`; risk: external mutation; overwrites episode metadata on the configured podcast; approval
  required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 2 stream-backed endpoint group(s), 2 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=1.
