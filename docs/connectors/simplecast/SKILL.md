---
name: pm-simplecast
description: Simplecast connector knowledge and safe action guide.
---

# pm-simplecast

## Purpose

Reads Simplecast podcasts and episodes through the Simplecast REST API.

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
- access_token (secret) (required)

## ETL Streams

- podcasts:
  - primary key: id
  - cursor: updated_at
  - fields: id(string), status(string), title(string), updated_at(string)
- episodes:
  - primary key: id
  - cursor: updated_at
  - fields: id(string), status(string), title(string), updated_at(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Simplecast API read of podcast and episode data
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect simplecast
```

### Inspect as structured JSON

```bash
pm connectors inspect simplecast --json
```

## Agent Rules

- Run pm connectors inspect simplecast before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
