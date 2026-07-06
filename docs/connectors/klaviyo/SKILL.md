---
name: pm-klaviyo
description: Klaviyo connector knowledge and safe action guide.
---

# pm-klaviyo

## Purpose

Reads Klaviyo profiles, events, campaigns, lists, metrics, and segments through the Klaviyo REST (JSON:API) API.

## Icon

- asset: icons/klaviyo.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.klaviyo.com/en/docs/api_versioning_and_deprecation_policy

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- revision
- api_key (secret)

## ETL Streams

- profiles:
  - primary key: id
  - cursor: updated
  - fields: created(), email(), external_id(), first_name(), id(), last_name(), organization(), phone_number(), type(), updated()
- events:
  - primary key: id
  - cursor: datetime
  - fields: datetime(), id(), timestamp(), type(), uuid()
- campaigns:
  - primary key: id
  - cursor: updated_at
  - fields: archived(), channel(), created_at(), id(), name(), scheduled_at(), send_time(), status(), type(), updated_at()
- lists:
  - primary key: id
  - cursor: updated
  - fields: created(), id(), name(), type(), updated()
- metrics:
  - primary key: id
  - cursor: updated
  - fields: created(), id(), integration_name(), name(), type(), updated()
- segments:
  - primary key: id
  - cursor: updated
  - fields: created(), id(), is_active(), is_processing(), name(), type(), updated()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Klaviyo API read of customer profile, event, and campaign data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect klaviyo
```

### Inspect as structured JSON

```bash
pm connectors inspect klaviyo --json
```

## Agent Rules

- Run pm connectors inspect klaviyo before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
