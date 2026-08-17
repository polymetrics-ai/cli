---
name: pm-shutterstock
description: Shutterstock connector knowledge and safe action guide.
---

# pm-shutterstock

## Purpose

Reads Shutterstock media, collection, license, editorial, catalog, contributor, and subscription metadata; writes collection/lightbox metadata through safe collection endpoints.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- audio_collection_ids
- audio_ids
- base_url
- category
- contributor_collection_ids
- contributor_id
- contributor_ids
- editorial_ids
- editorial_image_ids
- editorial_image_livefeed_ids
- editorial_livefeed_ids
- editorial_video_ids
- image_collection_ids
- image_ids
- orientation
- query
- sfx_ids
- sort
- video_collection_ids
- video_ids
- visual_asset_ids
- access_token (secret)

## ETL Streams

- images:
  - primary key: id
  - cursor: updated_at
  - fields: description(), id(), media_type(), updated_at()
- videos:
  - primary key: id
  - cursor: updated_at
  - fields: description(), id(), media_type(), updated_at()
- audio:
  - primary key: id
  - cursor: updated_at
  - fields: description(), id(), media_type(), updated_at()
- list_similar_images:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_image_recommendations:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_updated_images:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_image_list:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_image:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_image_license_list:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_image_collection_list:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_image_collection:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_image_collection_items:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- find_similar_videos:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_updated_videos:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_video_list:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_video:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_video_license_list:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_video_collection_list:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_video_collection:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_video_collection_items:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_track_list:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_track:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_track_license_list:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_track_collection_list:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_track_collection:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_track_collection_items:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- search_sfx:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_sfx:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_sfx_list:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_sfx_license_list:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- search_editorial_images:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_updated_editorial_images:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_editorial_image:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_editorial_image_list:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_editorial_image_license_list:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_editorial_image_livefeeds:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_editorial_image_livefeed:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_editorial_image_livefeed_items:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_editorial:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_editorial_livefeeds:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_editorial_livefeed:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_editorial_livefeed_items:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- search_editorial:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_updated_editorial:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- search_editorial_videos:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_editorial_video:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_editorial_video_list:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_editorial_video_license_list:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_similar_cv_images:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_similar_cv_videos:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- search_catalog:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_collections:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_contributors:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_contributor:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_contributor_collections:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_contributor_collection:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_contributor_collection_items:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()
- get_user_subscriptions:
  - primary key: id
  - fields: asset_type(), collection_id(), contributor_id(), created_at(), created_time(), description(), id(), items(), license_id(), media_type(), name(), subscription_id(), updated_at(), updated_time()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_image_collection:
  - endpoint: POST /v2/images/collections
  - risk: external Shutterstock POST /v2/images/collections; approval required
- rename_image_collection:
  - endpoint: POST /v2/images/collections/{{ record.image_collection_id }}
  - required fields: image_collection_id
  - risk: external Shutterstock POST /v2/images/collections/{{ record.image_collection_id }}; approval required
- delete_image_collection:
  - endpoint: DELETE /v2/images/collections/{{ record.image_collection_id }}
  - required fields: image_collection_id
  - risk: destructive external Shutterstock DELETE /v2/images/collections/{{ record.image_collection_id }}; approval required
- add_image_collection_items:
  - endpoint: POST /v2/images/collections/{{ record.image_collection_id }}/items
  - required fields: image_collection_id
  - risk: external Shutterstock POST /v2/images/collections/{{ record.image_collection_id }}/items; approval required
- create_video_collection:
  - endpoint: POST /v2/videos/collections
  - risk: external Shutterstock POST /v2/videos/collections; approval required
- rename_video_collection:
  - endpoint: POST /v2/videos/collections/{{ record.video_collection_id }}
  - required fields: video_collection_id
  - risk: external Shutterstock POST /v2/videos/collections/{{ record.video_collection_id }}; approval required
- delete_video_collection:
  - endpoint: DELETE /v2/videos/collections/{{ record.video_collection_id }}
  - required fields: video_collection_id
  - risk: destructive external Shutterstock DELETE /v2/videos/collections/{{ record.video_collection_id }}; approval required
- add_video_collection_items:
  - endpoint: POST /v2/videos/collections/{{ record.video_collection_id }}/items
  - required fields: video_collection_id
  - risk: external Shutterstock POST /v2/videos/collections/{{ record.video_collection_id }}/items; approval required
- create_audio_collection:
  - endpoint: POST /v2/audio/collections
  - risk: external Shutterstock POST /v2/audio/collections; approval required
- rename_audio_collection:
  - endpoint: POST /v2/audio/collections/{{ record.audio_collection_id }}
  - required fields: audio_collection_id
  - risk: external Shutterstock POST /v2/audio/collections/{{ record.audio_collection_id }}; approval required
- delete_audio_collection:
  - endpoint: DELETE /v2/audio/collections/{{ record.audio_collection_id }}
  - required fields: audio_collection_id
  - risk: destructive external Shutterstock DELETE /v2/audio/collections/{{ record.audio_collection_id }}; approval required
- add_audio_collection_items:
  - endpoint: POST /v2/audio/collections/{{ record.audio_collection_id }}/items
  - required fields: audio_collection_id
  - risk: external Shutterstock POST /v2/audio/collections/{{ record.audio_collection_id }}/items; approval required
- create_catalog_collection:
  - endpoint: POST /v2/catalog/collections
  - risk: external Shutterstock POST /v2/catalog/collections; approval required
- update_catalog_collection:
  - endpoint: PATCH /v2/catalog/collections/{{ record.catalog_collection_id }}
  - required fields: catalog_collection_id
  - risk: external Shutterstock PATCH /v2/catalog/collections/{{ record.catalog_collection_id }}; approval required
- delete_catalog_collection:
  - endpoint: DELETE /v2/catalog/collections/{{ record.catalog_collection_id }}
  - required fields: catalog_collection_id
  - risk: destructive external Shutterstock DELETE /v2/catalog/collections/{{ record.catalog_collection_id }}; approval required
- add_catalog_collection_items:
  - endpoint: POST /v2/catalog/collections/{{ record.catalog_collection_id }}/items
  - required fields: catalog_collection_id
  - risk: external Shutterstock POST /v2/catalog/collections/{{ record.catalog_collection_id }}/items; approval required

## Security

- read risk: external Shutterstock API read of media, collections, licenses, editorial, catalog, contributor, and subscription metadata
- write risk: external Shutterstock collection/lightbox create, rename, delete, and item-add mutations; licensing/download writes are intentionally excluded
- approval: required before collection write actions; licensing and download endpoints are not exposed as writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect shutterstock
```

### Inspect as structured JSON

```bash
pm connectors inspect shutterstock --json
```

## Agent Rules

- Run pm connectors inspect shutterstock before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
