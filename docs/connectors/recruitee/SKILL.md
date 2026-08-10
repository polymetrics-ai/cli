---
name: pm-recruitee
description: Recruitee connector knowledge and safe action guide.
---

# pm-recruitee

## Purpose

Reads Recruitee offers, candidates, departments, sources, and tags through the Recruitee REST API.

## Icon

- id: recruitee
- asset: icons/recruitee.svg
- source: upstream_registry
- review_status: upstream_seeded

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- company_id (required)
- api_key (secret) (required)

## ETL Streams

- offers:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), id(integer), status(string), title(string), updated_at(string)
- candidates:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), email(string), id(integer), name(string), updated_at(string)
- departments:
  - primary key: id
  - fields: id(integer), name(string)
- sources:
  - primary key: id
  - fields: id(integer), name(string)
- tags:
  - primary key: id
  - fields: id(integer), name(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Recruitee API read of ATS offer and candidate data
- approval: none; read-only ATS API
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect recruitee
```

### Inspect as structured JSON

```bash
pm connectors inspect recruitee --json
```

## Agent Rules

- Run pm connectors inspect recruitee before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
