---
name: pm-rss
description: RSS connector knowledge and safe action guide.
---

# pm-rss

## Purpose

Reads RSS channel metadata and feed items from any RSS 2.0 feed URL. Read-only and credential-free.

## Icon

- asset: icons/rss.svg
- source: upstream_registry
- review_status: upstream_seeded

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- No secret authentication is required for this connector.

## Configuration

- feed_url

## ETL Streams

- items:
  - primary key: id
  - cursor: published_at
  - fields: description(), id(), link(), published_at(), title()
- channel:
  - primary key: id
  - fields: description(), id(), link(), title(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external RSS feed read (XML over HTTP/HTTPS)
- approval: none; read-only, credential-free feed reader
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect rss
```

### Inspect as structured JSON

```bash
pm connectors inspect rss --json
```

## Agent Rules

- Run pm connectors inspect rss before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
