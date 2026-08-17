---
name: pm-care-quality-commission
description: Care Quality Commission connector knowledge and safe action guide.
---

# pm-care-quality-commission

## Purpose

Reads Care Quality Commission (CQC) registered locations, providers, and inspection areas from the public CQC Syndication API. Read-only (full-refresh).

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
- api_key (secret)

## ETL Streams

- locations:
  - primary key: locationId
  - fields: locationId(), locationName(), postalCode()
- providers:
  - primary key: providerId
  - fields: providerId(), providerName()
- inspection_areas:
  - primary key: inspectionAreaId
  - fields: endDate(), inspectionAreaId(), inspectionAreaName(), inspectionAreaType(), inspectionCategories(), orgInspectionAreaRetirementDate(), status(), supersededBy()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external CQC Syndication API read of publicly published care provider/location data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect care-quality-commission
```

### Inspect as structured JSON

```bash
pm connectors inspect care-quality-commission --json
```

## Agent Rules

- Run pm connectors inspect care-quality-commission before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
