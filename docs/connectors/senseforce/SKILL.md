---
name: pm-senseforce
description: Senseforce connector knowledge and safe action guide.
---

# pm-senseforce

## Purpose

Reads records from a configured Senseforce dataset through the Senseforce API.

## Icon

- asset: icons/senseforce.svg
- source: upstream_registry
- review_status: upstream_seeded

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- backend_url
- dataset_id
- access_token (secret)

## ETL Streams

- records:
  - primary key: id
  - fields: Timestamp(), id(), value()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Senseforce API read of a configured dataset's rows
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect senseforce
```

### Inspect as structured JSON

```bash
pm connectors inspect senseforce --json
```

## Agent Rules

- Run pm connectors inspect senseforce before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
