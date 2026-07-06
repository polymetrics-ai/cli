---
name: pm-recruitee
description: Recruitee connector knowledge and safe action guide.
---

# pm-recruitee

## Purpose

Reads Recruitee offers, candidates, departments, sources, and tags through the Recruitee REST API.

## Icon

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
- company_id
- api_key (secret)

## ETL Streams

- offers:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), id(), status(), title(), updated_at()
- candidates:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), email(), id(), name(), updated_at()
- departments:
  - primary key: id
  - fields: id(), name()
- sources:
  - primary key: id
  - fields: id(), name()
- tags:
  - primary key: id
  - fields: id(), name()

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
