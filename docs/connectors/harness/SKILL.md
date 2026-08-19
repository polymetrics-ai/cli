---
name: pm-harness
description: Harness connector knowledge and safe action guide.
---

# pm-harness

## Purpose

Reads Harness NextGen organizations, projects, services, connectors, and pipelines through the Harness platform REST API.

## Icon

- id: harness
- asset: icons/harness.svg
- source: upstream_registry
- review_status: upstream_seeded

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- account_id (required)
- base_url
- mode
- page_size
- api_key (secret) (required)

## ETL Streams

- organizations:
  - primary key: identifier
  - fields: account_identifier(string), description(string), identifier(string), name(string)
- projects:
  - primary key: identifier
  - fields: account_identifier(string), color(string), description(string), identifier(string), modules(array), name(string), org_identifier(string)
- services:
  - primary key: identifier
  - fields: account_identifier(string), deleted(boolean), description(string), identifier(string), name(string), org_identifier(string), project_identifier(string)
- connectors:
  - primary key: identifier
  - fields: description(string), identifier(string), name(string), org_identifier(string), project_identifier(string), type(string)
- pipelines:
  - primary key: identifier
  - fields: description(string), identifier(string), name(string), org_identifier(string), project_identifier(string), stage_count(integer)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Harness NextGen platform API read of organization/project/service/connector/pipeline metadata
- approval: none; read-only source connector
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect harness
```

### Inspect as structured JSON

```bash
pm connectors inspect harness --json
```

## Agent Rules

- Run pm connectors inspect harness before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
