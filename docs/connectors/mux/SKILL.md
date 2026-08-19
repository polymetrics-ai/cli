---
name: pm-mux
description: Mux connector knowledge and safe action guide.
---

# pm-mux

## Purpose

Reads Mux Video assets, live streams, direct uploads, and system signing keys through the Mux REST API using HTTP Basic authentication.

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
- mode
- username (required)
- password (secret) (required)

## ETL Streams

- assets:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), duration(number), encoding_tier(string), id(string), master_access(string), max_resolution_tier(string), mp4_support(string), status(string), test(boolean)
- live_streams:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), id(string), latency_mode(string), max_continuous_duration(number), reconnect_window(number), status(string), stream_key(string), test(boolean)
- uploads:
  - primary key: id
  - fields: asset_id(string), cors_origin(string), id(string), status(string), test(boolean), timeout(number), url(string)
- signing_keys:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), id(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Mux API read of video asset, live stream, upload, and signing key data
- approval: none; read-only, no reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect mux
```

### Inspect as structured JSON

```bash
pm connectors inspect mux --json
```

## Agent Rules

- Run pm connectors inspect mux before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
