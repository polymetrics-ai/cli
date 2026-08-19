---
name: pm-guru
description: Guru connector knowledge and safe action guide.
---

# pm-guru

## Purpose

Reads Guru collections, groups, members, and teams through the Guru REST API using HTTP Basic authentication (email + API token).

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- max_pages
- mode
- page_size
- username (required)
- password (secret) (required)

## ETL Streams

- collections:
  - primary key: id
  - fields: collectionType(string), color(string), dateCreated(string), description(string), id(string), name(string), publicCardsEnabled(boolean), slug(string)
- groups:
  - primary key: id
  - fields: dateCreated(string), groupType(string), id(string), memberCount(integer), modifiable(boolean), name(string)
- members:
  - primary key: id
  - fields: dateCreated(string), email(string), firstName(string), id(string), lastName(string), status(string)
- teams:
  - primary key: id
  - fields: dateCreated(string), id(string), name(string), status(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Guru API read of collections, groups, members, and teams
- approval: none; read-only source connector
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect guru
```

### Inspect as structured JSON

```bash
pm connectors inspect guru --json
```

## Agent Rules

- Run pm connectors inspect guru before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
