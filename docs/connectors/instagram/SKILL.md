---
name: pm-instagram
description: Instagram connector knowledge and safe action guide.
---

# pm-instagram

## Purpose

Reads Instagram Business/Creator account profile, media, and stories through the Facebook Graph API.

## Icon

- asset: icons/instagram.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.facebook.com/docs/instagram-platform/changelog

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- ig_user_id
- page_size
- access_token (secret)

## ETL Streams

- users:
  - primary key: id
  - fields: biography(), followers_count(), follows_count(), id(), media_count(), name(), profile_picture_url(), username(), website()
- media:
  - primary key: id
  - cursor: timestamp
  - fields: caption(), comments_count(), id(), like_count(), media_product_type(), media_type(), media_url(), permalink(), thumbnail_url(), timestamp(), username()
- stories:
  - primary key: id
  - fields: caption(), id(), media_product_type(), media_type(), media_url(), permalink(), thumbnail_url(), timestamp(), username()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Facebook Graph API read of Instagram Business/Creator account data
- approval: none; read-only source
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect instagram
```

### Inspect as structured JSON

```bash
pm connectors inspect instagram --json
```

## Agent Rules

- Run pm connectors inspect instagram before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
