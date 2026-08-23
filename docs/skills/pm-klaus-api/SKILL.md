---
name: pm-klaus-api
description: Klaus API connector knowledge and safe action guide.
---

# pm-klaus-api

## Purpose

Reads Klaus (Zendesk QA) users and rating categories through the Klaus public REST API. The reviews stream is not yet migrated (ENGINE_GAP, see docs.md).

## Icon

- id: klaus-api
- asset: icons/klaus-api.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://help.klausapp.com/en/articles/2911907-klaus-api

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- account (required)
- base_url
- mode
- workspace
- api_key (secret) (required)

## ETL Streams

- users:
  - primary key: id
  - fields: email(string), id(string), name(string)
- categories:
  - primary key: id
  - fields: archived(boolean), critical(boolean), description(string), groupId(string), groupName(string), groupPosition(integer), id(string), maxRating(integer), name(string), position(integer), rootCauses(array), scorecards(array), weight(number)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Klaus API read of user and quality-review configuration data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect klaus-api
```

### Inspect as structured JSON

```bash
pm connectors inspect klaus-api --json
```

## Agent Rules

- Run pm connectors inspect klaus-api before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
