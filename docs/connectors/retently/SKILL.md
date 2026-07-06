---
name: pm-retently
description: Retently connector knowledge and safe action guide.
---

# pm-retently

## Purpose

Reads Retently customers, survey responses, surveys, and campaigns through the REST API.

## Icon

- asset: icons/retently.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://www.retently.com/api/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- campaign_id
- created_after
- email
- updated_after
- api_key (secret)

## ETL Streams

- customers:
  - primary key: id
  - cursor: updated_at
  - fields: company(), email(), full_name(), id(), stream(), updated_at()
- responses:
  - primary key: id
  - cursor: created_at
  - fields: comment(), created_at(), customer_id(), id(), score(), stream()
- surveys:
  - primary key: id
  - cursor: updated_at
  - fields: id(), name(), status(), stream(), type(), updated_at()
- campaigns:
  - primary key: id
  - cursor: updated_at
  - fields: id(), name(), status(), stream(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Retently API read of customer and NPS/CSAT survey response data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect retently
```

### Inspect as structured JSON

```bash
pm connectors inspect retently --json
```

## Agent Rules

- Run pm connectors inspect retently before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
