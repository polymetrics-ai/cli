---
name: pm-outbrain-amplify
description: Outbrain Amplify connector knowledge and safe action guide.
---

# pm-outbrain-amplify

## Purpose

Reads Outbrain Amplify marketers, campaigns, and performance reports via the Outbrain Amplify REST API. Read-only.

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
- conversion_count
- end_date
- geo_location_breakdown
- marketer_id
- max_pages
- mode
- page_size
- report_granularity
- start_date
- username
- access_token (secret)
- password (secret)

## ETL Streams

- marketers:
  - primary key: id
  - fields: clicks(integer), created_at(string), enabled(boolean), id(string), impressions(integer), name(string), spend(number), status(string)
- campaigns:
  - primary key: id
  - fields: clicks(integer), created_at(string), enabled(boolean), id(string), impressions(integer), name(string), spend(number), status(string)
- performance_reports:
  - primary key: id
  - fields: clicks(integer), created_at(string), enabled(boolean), id(string), impressions(integer), name(string), spend(number), status(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Outbrain Amplify API read of marketer, campaign, and performance report data
- approval: none; read-only marketing API
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect outbrain-amplify
```

### Inspect as structured JSON

```bash
pm connectors inspect outbrain-amplify --json
```

## Agent Rules

- Run pm connectors inspect outbrain-amplify before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
