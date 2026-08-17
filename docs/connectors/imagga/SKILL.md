---
name: pm-imagga
description: Imagga connector knowledge and safe action guide.
---

# pm-imagga

## Purpose

Reads Imagga account API usage and per-image tags/categories via the Imagga REST API. Read-only. The colors and faces_detections detection streams are not yet ported — see docs.md Known limits.

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
- image_urls
- api_key (secret)
- api_secret (secret)

## ETL Streams

- usage:
  - primary key: period
  - fields: daily_processed(), monthly_limit(), monthly_processed(), period(), requests()
- tags:
  - primary key: image_url, tag
  - fields: confidence(), image_url(), tag()
- categories:
  - primary key: image_url, category
  - fields: category(), confidence(), image_url()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Imagga API read of account usage data and per-image tags/categories
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect imagga
```

### Inspect as structured JSON

```bash
pm connectors inspect imagga --json
```

## Agent Rules

- Run pm connectors inspect imagga before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
