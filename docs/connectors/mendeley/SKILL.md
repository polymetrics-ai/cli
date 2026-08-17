---
name: pm-mendeley
description: Mendeley connector knowledge and safe action guide.
---

# pm-mendeley

## Purpose

Reads documents, folders, groups, and annotations from the Mendeley reference manager REST API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- name_for_institution
- query_for_catalog
- start_date
- client_id (secret)
- client_refresh_token (secret)
- client_secret (secret)

## ETL Streams

- documents:
  - primary key: id
  - cursor: last_modified
  - fields: abstract(), created(), group_id(), id(), last_modified(), profile_id(), source(), title(), type(), year()
- folders:
  - primary key: id
  - cursor: modified
  - fields: created(), group_id(), id(), modified(), name(), parent_id()
- groups:
  - primary key: id
  - fields: access_level(), created(), description(), id(), name(), owning_profile_id(), role(), webpage()
- annotations:
  - primary key: id
  - cursor: last_modified
  - fields: created(), document_id(), filehash(), id(), last_modified(), privacy_level(), profile_id(), text(), type()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Mendeley API reads performed by the legacy connector via a Tier-2 hook
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect mendeley
```

### Inspect as structured JSON

```bash
pm connectors inspect mendeley --json
```

## Agent Rules

- Run pm connectors inspect mendeley before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
