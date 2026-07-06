---
name: pm-facebook-pages
description: Facebook Pages connector knowledge and safe action guide.
---

# pm-facebook-pages

## Purpose

Reads Facebook Page metadata and posts from the Graph API. Read-only.

## Icon

- asset: icons/facebook.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.facebook.com/docs/pages/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- page_id
- page_size
- access_token (secret)

## ETL Streams

- page:
  - primary key: id
  - fields: category(), fan_count(), id(), link(), name()
- posts:
  - primary key: id
  - cursor: updated_time
  - fields: created_time(), id(), message(), permalink_url(), updated_time()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Facebook Graph API read of page metadata and posts
- approval: none; read-only, no writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect facebook-pages
```

### Inspect as structured JSON

```bash
pm connectors inspect facebook-pages --json
```

## Agent Rules

- Run pm connectors inspect facebook-pages before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
