---
name: pm-adjust
description: Adjust connector knowledge and safe action guide.
---

# pm-adjust

## Purpose

Reads Adjust report-service report rows for configured dimensions and metrics. Read-only.

## Icon

- asset: icons/adjust.svg
- source: official
- review_status: official_verified
- review_url: https://dev.adjust.com/en/api/rs-api/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- additional_metrics
- base_url
- dimensions
- end_date
- metrics
- mode
- start_date
- api_token (secret)

## ETL Streams

- reports:
  - fields: app(), clicks(), cost(), country(), date(), installs()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Adjust report-service read of configured dimensions/metrics
- approval: none; read-only reporting API
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect adjust
```

### Inspect as structured JSON

```bash
pm connectors inspect adjust --json
```

## Agent Rules

- Run pm connectors inspect adjust before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
