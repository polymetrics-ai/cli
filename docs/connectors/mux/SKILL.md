---
name: pm-mux
description: Mux connector knowledge and safe action guide.
---

# pm-mux

## Purpose

Reads Mux Video assets, live streams, direct uploads, and system signing keys through the Mux REST API using HTTP Basic authentication.

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
- mode
- username
- password (secret)

## ETL Streams

- assets:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), duration(), encoding_tier(), id(), master_access(), max_resolution_tier(), mp4_support(), status(), test()
- live_streams:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), id(), latency_mode(), max_continuous_duration(), reconnect_window(), status(), stream_key(), test()
- uploads:
  - primary key: id
  - fields: asset_id(), cors_origin(), id(), status(), test(), timeout(), url()
- signing_keys:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), id()

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
