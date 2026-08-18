---
name: pm-glassfrog
description: GlassFrog connector knowledge and safe action guide.
---

# pm-glassfrog

## Purpose

Reads GlassFrog circles, roles, people, projects, and assignments through the GlassFrog API v3 (read-only full-refresh source).

## Icon

- id: glassfrog
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
- api_key (secret) (required)

## ETL Streams

- assignments:
  - primary key: id
  - fields: election(string), exclude_from_meetings(boolean), focus(string), id(integer), person_id(integer), role_id(integer)
- circles:
  - primary key: id
  - fields: id(integer), name(string), organization_id(integer), short_name(string), strategy(string), supported_role_id(integer)
- people:
  - primary key: id
  - fields: email(string), external_id(string), id(integer), name(string), tag_names(array)
- projects:
  - primary key: id
  - fields: archived_at(string), created_at(string), description(string), effort(string), id(integer), link(string), private_to_circle(boolean), roi(string), status(string), value(string), waiting_on_what(string), waiting_on_who(string)
- roles:
  - primary key: id
  - fields: elected_until(string), id(integer), is_core(boolean), name(string), name_with_circle_for_core_roles(string), organization_id(integer), purpose(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

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
