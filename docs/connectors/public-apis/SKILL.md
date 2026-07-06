---
name: pm-public-apis
description: Public APIs connector knowledge and safe action guide.
---

# pm-public-apis

## Purpose

Reads public API directory entries and categories from the api.publicapis.org directory API. Read-only and credential-free.

## Icon

- asset: icons/public-apis.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://github.com/public-apis/public-apis

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- No secret authentication is required for this connector.

## Configuration

- base_url
- mode

## ETL Streams

- entries:
  - primary key: id
  - fields: api(), auth(), category(), cors(), description(), https(), id(), link()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external public-apis.org directory read of API listing metadata
- approval: none; read-only, credential-free public directory API
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect public-apis
```

### Inspect as structured JSON

```bash
pm connectors inspect public-apis --json
```

## Agent Rules

- Run pm connectors inspect public-apis before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
