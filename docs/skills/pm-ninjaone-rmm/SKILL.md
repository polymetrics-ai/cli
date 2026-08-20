---
name: pm-ninjaone-rmm
description: NinjaOne RMM connector knowledge and safe action guide.
---

# pm-ninjaone-rmm

## Purpose

Reads NinjaOne RMM organizations, devices, locations, activities, and policies through the NinjaOne v2 REST API.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- page_size
- api_key (secret) (required)

## ETL Streams

- organizations:
  - primary key: id
  - fields: description(string), id(integer), name(string), node_approval_mode(string)
- devices:
  - primary key: id
  - fields: approval_status(string), dns_name(string), id(integer), location_id(integer), node_class(string), offline(boolean), organization_id(integer), system_name(string)
- locations:
  - primary key: id
  - fields: address(string), description(string), id(integer), name(string), organization_id(integer)
- activities:
  - primary key: id
  - cursor: activityTime
  - fields: activityTime(number), activity_type(string), device_id(integer), id(integer), message(string), status(string)
- policies:
  - primary key: id
  - fields: description(string), id(integer), name(string), node_class(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external NinjaOne RMM API read of managed device and organization data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect ninjaone-rmm
```

### Inspect as structured JSON

```bash
pm connectors inspect ninjaone-rmm --json
```

## Agent Rules

- Run pm connectors inspect ninjaone-rmm before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
