---
name: pm-unleash
description: Unleash connector knowledge and safe action guide.
---

# pm-unleash

## Purpose

Reads Unleash projects, feature toggles, environments, and segments through admin API list endpoints.

## Icon

- asset: icons/unleash.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.getunleash.io/reference/api/unleash

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- project_id
- api_token (secret)

## ETL Streams

- projects:
  - primary key: id
  - fields: id(), name()
- features:
  - primary key: name
  - fields: enabled(), name(), project(), type()
- environments:
  - primary key: id
  - fields: id(), name()
- segments:
  - primary key: id
  - fields: id(), name()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Unleash admin API read of project, feature toggle, environment, and segment data
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect unleash
```

### Inspect as structured JSON

```bash
pm connectors inspect unleash --json
```

## Agent Rules

- Run pm connectors inspect unleash before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
