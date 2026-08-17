---
name: pm-ninjaone-rmm
description: NinjaOne RMM connector knowledge and safe action guide.
---

# pm-ninjaone-rmm

## Purpose

Reads NinjaOne RMM organizations, devices, locations, activities, and policies through the NinjaOne v2 REST API.

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
- page_size
- api_key (secret)

## ETL Streams

- organizations:
  - primary key: id
  - fields: description(), id(), name(), node_approval_mode()
- devices:
  - primary key: id
  - fields: approval_status(), dns_name(), id(), location_id(), node_class(), offline(), organization_id(), system_name()
- locations:
  - primary key: id
  - fields: address(), description(), id(), name(), organization_id()
- activities:
  - primary key: id
  - cursor: activityTime
  - fields: activityTime(), activity_type(), device_id(), id(), message(), status()
- policies:
  - primary key: id
  - fields: description(), id(), name(), node_class()

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
