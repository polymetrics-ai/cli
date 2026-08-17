---
name: pm-plausible
description: Plausible connector knowledge and safe action guide.
---

# pm-plausible

## Purpose

Reads Plausible Analytics sites and stats reports through the Stats API.

## Icon

- asset: icons/plausible.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://plausible.io/docs/stats-api

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- compare
- date
- filters
- metrics
- mode
- period
- property
- site_id
- api_token (secret)

## ETL Streams

- sites:
  - primary key: site_id
  - fields: domain(), site_id()
- aggregate:
  - primary key: site_id
  - fields: bounce_rate(), events(), pageviews(), site_id(), visit_duration(), visitors(), visits()
- timeseries:
  - primary key: date
  - fields: bounce_rate(), date(), events(), pageviews(), site_id(), visit_duration(), visitors(), visits()
- breakdown:
  - primary key: property_value
  - fields: bounce_rate(), events(), pageviews(), property_value(), site_id(), visit_duration(), visitors(), visits()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Plausible Analytics API read of site analytics data
- approval: none; read-only analytics sync
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect plausible
```

### Inspect as structured JSON

```bash
pm connectors inspect plausible --json
```

## Agent Rules

- Run pm connectors inspect plausible before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
