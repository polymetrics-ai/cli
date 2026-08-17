---
name: pm-appsflyer
description: AppsFlyer connector knowledge and safe action guide.
---

# pm-appsflyer

## Purpose

Reads AppsFlyer raw-data CSV export reports (installs, in-app events) through the AppsFlyer Pull API. Read-only.

## Icon

- asset: icons/appsflyer.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://dev.appsflyer.com/hc/reference

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- app_id
- base_url
- end_date
- mode
- start_date
- timezone
- api_token (secret)

## ETL Streams

- installs_report:
  - fields: appsflyer_id(), campaign(), event_time(), media_source()
- in_app_events_report:
  - fields: appsflyer_id(), campaign(), event_time(), media_source()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external AppsFlyer API read of raw installs/in-app-event export reports
- approval: none; read-only, no writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect appsflyer
```

### Inspect as structured JSON

```bash
pm connectors inspect appsflyer --json
```

## Agent Rules

- Run pm connectors inspect appsflyer before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
