---
name: pm-pexels-api
description: Pexels API connector knowledge and safe action guide.
---

# pm-pexels-api

## Purpose

Reads Pexels photo/video search and curated/popular results plus featured and personal collections and their media through the Pexels REST API.

## Icon

- id: pexels
- asset: icons/pexels.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://www.pexels.com/api/documentation/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- collection_media_sort
- collection_media_type
- color
- locale
- orientation
- query
- size
- api_key (secret) (required)

## ETL Streams

- photos:
  - primary key: id
  - fields: alt(string), id(integer), photographer(string), photographer_url(string), src(object), url(string)
- curated_photos:
  - primary key: id
  - fields: alt(string), id(integer), photographer(string), photographer_url(string), src(object), url(string)
- videos:
  - primary key: id
  - fields: duration(integer), id(integer), image(string), url(string), user(object)
- popular_videos:
  - primary key: id
  - fields: duration(integer), id(integer), image(string), url(string), user(object)
- featured_collections:
  - primary key: id
  - fields: description(string), id(string), media_count(integer), photos_count(integer), private(boolean), title(string), videos_count(integer)
- my_collections:
  - primary key: id
  - fields: description(string), id(string), media_count(integer), photos_count(integer), private(boolean), title(string), videos_count(integer)
- collection_media:
  - primary key: id
  - fields: alt(string), collection_id(string), duration(integer), height(integer), id(integer), image(string), photographer(string), photographer_url(string), src(object), type(string), url(string), user(object), video_files(array), video_pictures(array), width(integer)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Pexels API read of photo/video search, curated/popular results, and collection metadata/media; all publicly-licensed stock media, no PII
- approval: none; read-only, no writes (the Pexels API has no create/update/delete endpoint anywhere in its documented surface, per its own docs: "Collections cannot be created or modified using the Pexels API")
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect pexels-api
```

### Inspect as structured JSON

```bash
pm connectors inspect pexels-api --json
```

## Agent Rules

- Run pm connectors inspect pexels-api before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
