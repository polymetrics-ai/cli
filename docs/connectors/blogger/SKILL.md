---
name: pm-blogger
description: Blogger connector knowledge and safe action guide.
---

# pm-blogger

## Purpose

Reads Blogger (Google Blogger API v3) blogs, posts, pages, comments, and page-view counts using an OAuth 2.0 refresh-token grant. Read-only.

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
- blog_id
- page_size
- token_url
- client_id (secret)
- client_refresh_token (secret)
- client_secret (secret)

## ETL Streams

- blogs:
  - primary key: id
  - cursor: updated
  - fields: description(), id(), kind(), name(), pages_total(), posts_total(), published(), updated(), url()
- posts:
  - primary key: id
  - cursor: updated
  - fields: author_display_name(), author_id(), blog_id(), content(), id(), kind(), published(), replies_total(), status(), title(), updated(), url()
- pages:
  - primary key: id
  - cursor: updated
  - fields: author_display_name(), author_id(), blog_id(), content(), id(), kind(), published(), status(), title(), updated(), url()
- comments:
  - primary key: id
  - cursor: updated
  - fields: author_display_name(), author_id(), blog_id(), content(), id(), kind(), post_id(), published(), status(), updated()
- pageviews:
  - primary key: blog_id, time_range
  - fields: blog_id(), count(), time_range()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Blogger API read of blog/post/page/comment metadata and page-view counts
- approval: none; read-only, no reverse-ETL write surface
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect blogger
```

### Inspect as structured JSON

```bash
pm connectors inspect blogger --json
```

## Agent Rules

- Run pm connectors inspect blogger before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
