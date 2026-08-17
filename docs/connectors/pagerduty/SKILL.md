---
name: pm-pagerduty
description: PagerDuty connector knowledge and safe action guide.
---

# pm-pagerduty

## Purpose

Reads PagerDuty incidents, users, services, and teams through the REST API.

## Icon

- asset: icons/pagerduty.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developer.pagerduty.com/api-reference/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- api_key (secret)

## ETL Streams

- incidents:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), id(), incident_number(), status(), title()
- users:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), email(), id(), name(), role()
- services:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), description(), id(), name(), status()
- teams:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), description(), id(), name()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external PagerDuty API read of incident, user, service, and team data
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect pagerduty
```

### Inspect as structured JSON

```bash
pm connectors inspect pagerduty --json
```

## Agent Rules

- Run pm connectors inspect pagerduty before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
