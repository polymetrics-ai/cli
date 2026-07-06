---
name: pm-google-directory
description: Google Directory connector knowledge and safe action guide.
---

# pm-google-directory

## Purpose

Reads Google Admin SDK Directory users, groups, organizational units, and ChromeOS devices via bearer-token OAuth. Read-only.

## Icon

- asset: icons/googledirectory.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.google.com/admin-sdk/directory/reference/rest

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- customer_id
- max_pages
- mode
- page_size
- access_token (secret)

## ETL Streams

- users:
  - primary key: id
  - fields: id(), name(), org_unit_path(), primary_email()
- groups:
  - primary key: id
  - fields: description(), email(), id(), name()
- orgunits:
  - primary key: id
  - fields: description(), id(), name(), org_unit_path()
- chromeos_devices:
  - primary key: id
  - fields: id(), org_unit_path(), serial_number(), status()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Google Admin SDK Directory API read of user/group/org-unit/device metadata
- approval: none; read-only, no reverse-ETL write surface
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect google-directory
```

### Inspect as structured JSON

```bash
pm connectors inspect google-directory --json
```

## Agent Rules

- Run pm connectors inspect google-directory before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
