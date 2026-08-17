---
name: pm-shortcut
description: Shortcut connector knowledge and safe action guide.
---

# pm-shortcut

## Purpose

Reads Shortcut stories, epics, projects, and iterations through the Shortcut REST API.

## Icon

- asset: icons/shortcut.svg
- source: official
- review_status: official_verified
- review_url: https://developer.shortcut.com/api/rest/v3

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- page_size
- api_token (secret)

## ETL Streams

- stories:
  - primary key: id
  - cursor: updated_at
  - fields: id(), name(), state(), updated_at()
- epics:
  - primary key: id
  - cursor: updated_at
  - fields: id(), name(), state(), updated_at()
- projects:
  - primary key: id
  - cursor: updated_at
  - fields: id(), name(), state(), updated_at()
- iterations:
  - primary key: id
  - cursor: updated_at
  - fields: id(), name(), state(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Shortcut API read of story, epic, project, and iteration data
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect shortcut
```

### Inspect as structured JSON

```bash
pm connectors inspect shortcut --json
```

## Agent Rules

- Run pm connectors inspect shortcut before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
