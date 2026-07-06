---
name: pm-harness
description: Harness connector knowledge and safe action guide.
---

# pm-harness

## Purpose

Reads Harness NextGen organizations, projects, services, connectors, and pipelines through the Harness platform REST API.

## Icon

- asset: icons/harness.svg
- source: upstream_registry
- review_status: upstream_seeded

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- account_id
- base_url
- mode
- page_size
- api_key (secret)

## ETL Streams

- organizations:
  - primary key: identifier
  - fields: account_identifier(), description(), identifier(), name()
- projects:
  - primary key: identifier
  - fields: account_identifier(), color(), description(), identifier(), modules(), name(), org_identifier()
- services:
  - primary key: identifier
  - fields: account_identifier(), deleted(), description(), identifier(), name(), org_identifier(), project_identifier()
- connectors:
  - primary key: identifier
  - fields: description(), identifier(), name(), org_identifier(), project_identifier(), type()
- pipelines:
  - primary key: identifier
  - fields: description(), identifier(), name(), org_identifier(), project_identifier(), stage_count()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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
