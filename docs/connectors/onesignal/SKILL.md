---
name: pm-onesignal
description: OneSignal connector knowledge and safe action guide.
---

# pm-onesignal

## Purpose

Reads OneSignal account-level applications through the OneSignal REST API. Device/notification/outcome streams remain quarantined (ENGINE_GAP: no per-stream auth override).

## Icon

- asset: icons/onesignal.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://documentation.onesignal.com/reference

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- user_auth_key (secret)

## ETL Streams

- apps:
  - primary key: id
  - fields: created_at(), id(), messageable_players(), name(), organization_id(), players(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external OneSignal API read of account-level application metadata
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect onesignal
```

### Inspect as structured JSON

```bash
pm connectors inspect onesignal --json
```

## Agent Rules

- Run pm connectors inspect onesignal before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
