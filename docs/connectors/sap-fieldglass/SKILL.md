---
name: pm-sap-fieldglass
description: SAP Fieldglass connector knowledge and safe action guide.
---

# pm-sap-fieldglass

## Purpose

Reads SAP Fieldglass workers, job postings, and time sheets through the SAP Fieldglass API. Read-only.

## Icon

- asset: icons/sapfieldglass.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://api.sap.com/package/SAPFieldglass/rest

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- access_token (secret)

## ETL Streams

- workers:
  - primary key: id
  - fields: id(), name(), status(), stream()
- job_postings:
  - primary key: id
  - fields: id(), name(), status(), stream()
- time_sheets:
  - primary key: id
  - fields: id(), name(), status(), stream()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external SAP Fieldglass API read of worker, job posting, and time sheet data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect sap-fieldglass
```

### Inspect as structured JSON

```bash
pm connectors inspect sap-fieldglass --json
```

## Agent Rules

- Run pm connectors inspect sap-fieldglass before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
