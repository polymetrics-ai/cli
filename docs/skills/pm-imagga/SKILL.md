---
name: pm-imagga
description: Imagga connector knowledge and safe action guide.
---

# pm-imagga

## Purpose

Reads Imagga account API usage and per-image tags/categories via the Imagga REST API. Read-only. The colors and faces_detections detection streams are not yet ported — see docs.md Known limits.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- image_urls
- api_key (secret) (required)
- api_secret (secret) (required)

## ETL Streams

- usage:
  - primary key: period
  - fields: daily_processed(integer), monthly_limit(integer), monthly_processed(integer), period(string), requests(integer)
- tags:
  - primary key: image_url, tag
  - fields: confidence(number), image_url(string), tag(string)
- categories:
  - primary key: image_url, category
  - fields: category(string), confidence(number), image_url(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

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
