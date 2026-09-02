---
name: pm-metabase
description: Metabase connector knowledge and safe action guide.
---

# pm-metabase

## Purpose

Reads Metabase cards, dashboards, collections, databases, and users through the Metabase REST API using session-token authentication. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

## Icon

- id: metabase
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
- instance_api_url (required)
- mode
- username (required)
- password (secret)
- session_token (secret)

## ETL Streams

- cards:
  - primary key: id
  - fields: archived(boolean), collection_id(integer), created_at(string), creator_id(integer), database_id(integer), description(string), display(string), id(integer), name(string), query_type(string), updated_at(string)
- dashboards:
  - primary key: id
  - fields: archived(boolean), collection_id(integer), created_at(string), creator_id(integer), description(string), id(integer), name(string), updated_at(string)
- collections:
  - primary key: id
  - fields: archived(boolean), description(string), id(string), location(string), name(string), personal_owner_id(integer), slug(string)
- databases:
  - primary key: id
  - fields: created_at(string), engine(string), id(integer), is_on_demand(boolean), is_sample(boolean), name(string), timezone(string), updated_at(string)
- users:
  - primary key: id
  - fields: common_name(string), date_joined(string), email(string), first_name(string), id(integer), is_active(boolean), is_superuser(boolean), last_login(string), last_name(string), updated_at(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

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
