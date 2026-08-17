---
name: pm-open-exchange-rates
description: Open Exchange Rates connector knowledge and safe action guide.
---

# pm-open-exchange-rates

## Purpose

Reads Open Exchange Rates account usage/plan status through the Open Exchange Rates JSON API (read-only). Live/historical/currencies rate-map streams remain quarantined (ENGINE_GAP).

## Icon

- asset: icons/open-exchange-rates.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.openexchangerates.org/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- app_id (secret)

## ETL Streams

- usage:
  - primary key: app_id
  - fields: app_id(), daily_average(), days_elapsed(), days_remaining(), plan(), requests(), requests_quota(), requests_remaining(), status()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Open Exchange Rates API read of account usage/plan status
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect open-exchange-rates
```

### Inspect as structured JSON

```bash
pm connectors inspect open-exchange-rates --json
```

## Agent Rules

- Run pm connectors inspect open-exchange-rates before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
