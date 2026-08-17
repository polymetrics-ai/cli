---
name: pm-zapier-supported-storage
description: Zapier Supported Storage connector knowledge and safe action guide.
---

# pm-zapier-supported-storage

## Purpose

Reads and writes Zapier Storage key/value records.

## Icon

- asset: icons/zapiersupportedstorage.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://help.zapier.com/hc/en-us/articles/8496293271053-Save-and-retrieve-data-from-Zaps

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- secret (secret)

## ETL Streams

- records:
  - primary key: id
  - cursor: updated_at
  - fields: id(), key(), updated_at(), value()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- set_record:
  - endpoint: PATCH /api/records
  - risk: creates or overwrites a single key/value pair in the caller's Zapier Storage bucket (optionally only when the existing value matches only_if_value); external mutation, no approval required
- increment_record:
  - endpoint: PATCH /api/records
  - risk: atomically increments a numeric-valued key by amount (creating it at amount if absent); external mutation, no approval required
- delete_record:
  - endpoint: DELETE /api/records?key={{ record.key }}
  - required fields: key
  - risk: irreversibly deletes a single key from the caller's Zapier Storage bucket
- delete_all_records:
  - endpoint: DELETE /api/records
  - risk: irreversibly deletes EVERY key in the caller's Zapier Storage bucket (whole-bucket wipe); destructive, requires explicit confirmation

## Security

- read risk: external Zapier Storage API read of stored key/value records
- write risk: external mutation of a shared per-Zap/per-app key/value store: set/increment a single key, delete a single key, or wipe the entire bucket (delete_all_records, destructive)
- approval: required for write actions; delete_all_records requires explicit destructive confirmation
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect zapier-supported-storage
```

### Inspect as structured JSON

```bash
pm connectors inspect zapier-supported-storage --json
```

## Agent Rules

- Run pm connectors inspect zapier-supported-storage before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
