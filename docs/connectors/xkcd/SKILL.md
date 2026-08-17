---
name: pm-xkcd
description: XKCD connector knowledge and safe action guide.
---

# pm-xkcd

## Purpose

Reads public XKCD comic metadata from the JSON API. Read-only.

## Icon

- asset: icons/xkcd.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://xkcd.com/json.html

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- No secret authentication is required for this connector.

## Configuration

- base_url
- comic_number

## ETL Streams

- latest:
  - primary key: num
  - fields: alt(), day(), img(), link(), month(), news(), num(), safe_title(), title(), transcript(), year()
- comic:
  - primary key: num
  - fields: alt(), day(), img(), link(), month(), news(), num(), safe_title(), title(), transcript(), year()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: public XKCD comic metadata read, no credentials involved
- approval: none; read-only public API
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect xkcd
```

### Inspect as structured JSON

```bash
pm connectors inspect xkcd --json
```

## Agent Rules

- Run pm connectors inspect xkcd before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
