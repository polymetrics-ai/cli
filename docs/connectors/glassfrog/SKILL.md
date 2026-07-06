---
name: pm-glassfrog
description: GlassFrog connector knowledge and safe action guide.
---

# pm-glassfrog

## Purpose

Reads GlassFrog circles, roles, people, projects, and assignments through the GlassFrog API v3 (read-only full-refresh source).

## Icon

- asset: icons/glassfrog.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://documenter.getpostman.com/view/1014385/glassfrog-api-v3/2SJViY

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- api_key (secret)

## ETL Streams

- assignments:
  - primary key: id
  - fields: election(), exclude_from_meetings(), focus(), id(), person_id(), role_id()
- circles:
  - primary key: id
  - fields: id(), name(), organization_id(), short_name(), strategy(), supported_role_id()
- people:
  - primary key: id
  - fields: email(), external_id(), id(), name(), tag_names()
- projects:
  - primary key: id
  - fields: archived_at(), created_at(), description(), effort(), id(), link(), private_to_circle(), roi(), status(), value(), waiting_on_what(), waiting_on_who()
- roles:
  - primary key: id
  - fields: elected_until(), id(), is_core(), name(), name_with_circle_for_core_roles(), organization_id(), purpose()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external GlassFrog API read of circle, role, person, project, and assignment data
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect glassfrog
```

### Inspect as structured JSON

```bash
pm connectors inspect glassfrog --json
```

## Agent Rules

- Run pm connectors inspect glassfrog before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
