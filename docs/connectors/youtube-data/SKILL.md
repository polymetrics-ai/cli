---
name: pm-youtube-data
description: YouTube Data connector knowledge and safe action guide.
---

# pm-youtube-data

## Purpose

Reads channels, videos, playlists, playlist items, comment threads, search results, video categories, and i18n region/language reference data through the YouTube Data API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- channel_ids
- ids
- mode
- playlist_ids
- region_code
- search_query
- video_ids
- api_key (secret)

## ETL Streams

- channels:
  - primary key: id
  - fields: id(), title(), view_count()
- videos:
  - primary key: id
  - fields: id(), published_at(), title()
- playlists:
  - primary key: id
  - fields: id(), published_at(), title()
- playlist_items:
  - primary key: id
  - cursor: published_at
  - fields: id(), playlist_id(), published_at(), title(), video_id()
- comment_threads:
  - primary key: id
  - cursor: published_at
  - fields: id(), published_at(), text(), video_id()
- search:
  - primary key: id
  - cursor: published_at
  - fields: id(), published_at(), title()
- video_categories:
  - primary key: id
  - fields: id(), title()
- i18n_regions:
  - primary key: id
  - fields: id(), name()
- i18n_languages:
  - primary key: id
  - fields: id(), name()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external YouTube Data API read of public channel, video, playlist, playlist item, comment, search result, and reference data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect youtube-data
```

### Inspect as structured JSON

```bash
pm connectors inspect youtube-data --json
```

## Agent Rules

- Run pm connectors inspect youtube-data before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
