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
- api_key (secret)

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

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Pexels API read of photo/video search, curated/popular results, and collection metadata/media; all publicly-licensed stock media, no PII
- approval: none; read-only, no writes (the Pexels API has no create/update/delete endpoint anywhere in its documented surface, per its own docs: "Collections cannot be created or modified using the Pexels API")
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run Pexels API's declared streams and reverse-ETL actions.
- Usage: pm pexels-api <command> [flags]
- Read streams
- Other Commands
  - api get v1 photos id - Documented GET /v1/photos/{id} (not implemented) [intent=direct_read availability=not_implemented operation=pexels-api.get.v1-photos-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 videos videos id - Documented GET /v1/videos/videos/{id} (not implemented) [intent=direct_read availability=not_implemented operation=pexels-api.get.v1-videos-videos-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - collection media list - Run the collection media ETL stream [intent=etl availability=implemented stream=collection_media]
  - curated photos list - Run the curated photos ETL stream [intent=etl availability=implemented stream=curated_photos]
  - featured collections list - Run the featured collections ETL stream [intent=etl availability=implemented stream=featured_collections]
  - my collections list - Run the my collections ETL stream [intent=etl availability=implemented stream=my_collections]
  - photos list - Run the photos ETL stream [intent=etl availability=implemented stream=photos]
  - popular videos list - Run the popular videos ETL stream [intent=etl availability=implemented stream=popular_videos]
  - videos list - Run the videos ETL stream [intent=etl availability=implemented stream=videos]

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
