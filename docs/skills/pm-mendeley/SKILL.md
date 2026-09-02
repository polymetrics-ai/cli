---
name: pm-mendeley
description: Mendeley connector knowledge and safe action guide.
---

# pm-mendeley

## Purpose

Reads documents, folders, groups, and annotations from the Mendeley reference manager REST API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

## Icon

- id: simple-icons-mendeley
- asset: icons/simple-icons/mendeley.svg
- title: Mendeley
- simple_icon_slug: mendeley
- simple_icon_hex: 9D1620
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Mendeley
- match: exact-name-or-slug
- matched_by: mendeley

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- name_for_institution (required)
- query_for_catalog (required)
- start_date (required)
- client_id (secret) (required)
- client_refresh_token (secret) (required)
- client_secret (secret) (required)

## ETL Streams

- documents:
  - primary key: id
  - cursor: last_modified
  - fields: abstract(string), created(string), group_id(string), id(string), last_modified(string), profile_id(string), source(string), title(string), type(string), year(integer)
- folders:
  - primary key: id
  - cursor: modified
  - fields: created(string), group_id(string), id(string), modified(string), name(string), parent_id(string)
- groups:
  - primary key: id
  - fields: access_level(string), created(string), description(string), id(string), name(string), owning_profile_id(string), role(string), webpage(string)
- annotations:
  - primary key: id
  - cursor: last_modified
  - fields: created(string), document_id(string), filehash(string), id(string), last_modified(string), privacy_level(string), profile_id(string), text(string), type(string)

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
