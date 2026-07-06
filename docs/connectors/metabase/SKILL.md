---
name: pm-metabase
description: Metabase connector knowledge and safe action guide.
---

# pm-metabase

## Purpose

Reads Metabase cards, dashboards, collections, databases, and users through the Metabase REST API using session-token authentication. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

## Icon

- asset: icons/metabase.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://www.metabase.com/docs/latest/api-documentation

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- instance_api_url
- mode
- username
- password (secret)
- session_token (secret)

## ETL Streams

- cards:
  - primary key: id
  - fields: archived(), collection_id(), created_at(), creator_id(), database_id(), description(), display(), id(), name(), query_type(), updated_at()
- dashboards:
  - primary key: id
  - fields: archived(), collection_id(), created_at(), creator_id(), description(), id(), name(), updated_at()
- collections:
  - primary key: id
  - fields: archived(), description(), id(), location(), name(), personal_owner_id(), slug()
- databases:
  - primary key: id
  - fields: created_at(), engine(), id(), is_on_demand(), is_sample(), name(), timezone(), updated_at()
- users:
  - primary key: id
  - fields: common_name(), date_joined(), email(), first_name(), id(), is_active(), is_superuser(), last_login(), last_name(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Metabase API reads performed by the legacy connector via a Tier-2 hook
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect metabase
```

### Inspect as structured JSON

```bash
pm connectors inspect metabase --json
```

## Agent Rules

- Run pm connectors inspect metabase before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
