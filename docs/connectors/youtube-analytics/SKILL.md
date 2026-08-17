---
name: pm-youtube-analytics
description: YouTube Analytics connector knowledge and safe action guide.
---

# pm-youtube-analytics

## Purpose

Reads YouTube Reporting API jobs, report types, and generated reports via the Google OAuth 2.0 refresh-token grant.

## Icon

- asset: icons/youtube-analytics.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.google.com/youtube/analytics

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- content_owner_id
- job_id
- max_pages
- mode
- page_size
- scopes
- token_url
- client_id (secret)
- client_secret (secret)
- refresh_token (secret)

## ETL Streams

- jobs:
  - primary key: id
  - fields: create_time(), expire_time(), id(), name(), report_type_id(), system_managed()
- report_types:
  - primary key: id
  - fields: deprecate_time(), id(), name(), system_managed()
- reports:
  - primary key: id
  - fields: create_time(), download_url(), end_time(), id(), job_expire_time(), job_id(), start_time()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external YouTube Reporting API read of reporting-job/report-type/report metadata
- approval: none; read-only, no reverse-ETL write surface
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect youtube-analytics
```

### Inspect as structured JSON

```bash
pm connectors inspect youtube-analytics --json
```

## Agent Rules

- Run pm connectors inspect youtube-analytics before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
