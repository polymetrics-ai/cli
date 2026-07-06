---
name: pm-trustpilot
description: Trustpilot connector knowledge and safe action guide.
---

# pm-trustpilot

## Purpose

Reads Trustpilot business-unit reviews, invitations, and business-unit profile metadata.

## Icon

- asset: icons/trustpilot.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.trustpilot.com/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- business_unit_id
- api_key (secret)

## ETL Streams

- reviews:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), id(), stars(), title()
- invitations:
  - primary key: id
  - fields: created_at(), id(), status()
- business_units:
  - primary key: id
  - fields: display_name(), id()
- categories:
  - primary key: category_id
  - fields: category_id(), display_name(), is_primary(), name(), relevance(), source()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Trustpilot API read of business-unit reviews, invitations, and profile metadata
- approval: none; read-only, no reverse-ETL writes implemented by legacy
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect trustpilot
```

### Inspect as structured JSON

```bash
pm connectors inspect trustpilot --json
```

## Agent Rules

- Run pm connectors inspect trustpilot before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
