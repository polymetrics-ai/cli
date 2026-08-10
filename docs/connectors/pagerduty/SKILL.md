---
name: pm-pagerduty
description: PagerDuty connector knowledge and safe action guide.
---

# pm-pagerduty

## Purpose

Reads PagerDuty incidents, users, services, and teams through the REST API.

## Icon

- id: pagerduty
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
- api_key (secret) (required)

## ETL Streams

- incidents:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), id(string), incident_number(integer), status(string), title(string)
- users:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), email(string), id(string), name(string), role(string)
- services:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), description(string), id(string), name(string), status(string)
- teams:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), description(string), id(string), name(string)

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
