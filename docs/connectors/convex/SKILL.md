---
name: pm-convex
description: Convex connector knowledge and safe action guide.
---

# pm-convex

## Purpose

Reads Convex tables and documents through the deployment HTTP API.

## Icon

- asset: icons/convex.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.convex.dev/http-api/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- deployment_url
- mode
- table
- access_key (secret)

## ETL Streams

- tables:
  - primary key: name
  - fields: name()
- documents:
  - primary key: id
  - fields: _id(), id()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Convex deployment API read of table metadata and documents
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect convex
```

### Inspect as structured JSON

```bash
pm connectors inspect convex --json
```

## Agent Rules

- Run pm connectors inspect convex before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
