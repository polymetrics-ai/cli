---
name: pm-high-level
description: High Level connector knowledge and safe action guide.
---

# pm-high-level

## Purpose

Reads HighLevel (Go HighLevel / LeadConnector) contacts, opportunities, pipelines, custom fields, and form submissions for a location through the HighLevel REST API.

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

- api_version
- base_url
- location_id
- api_key (secret)

## ETL Streams

- pipelines:
  - primary key: id
  - fields: dateAdded(), dateUpdated(), id(), locationId(), name(), stages()
- contacts:
  - primary key: id
  - cursor: dateUpdated
  - fields: contactName(), dateAdded(), dateUpdated(), email(), firstName(), id(), lastName(), locationId(), phone(), source(), type()
- opportunities:
  - primary key: id
  - cursor: dateUpdated
  - fields: assignedTo(), contactId(), dateAdded(), dateUpdated(), id(), monetaryValue(), name(), pipelineId(), pipelineStageId(), source(), status()
- custom_fields:
  - primary key: id
  - fields: dataType(), fieldKey(), id(), model(), name(), position()
- form_submissions:
  - primary key: id
  - cursor: createdAt
  - fields: contactId(), createdAt(), email(), formId(), id(), locationId(), name()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external HighLevel (LeadConnector) API read of contact, opportunity, pipeline, custom field, and form submission data for a configured location
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect high-level
```

### Inspect as structured JSON

```bash
pm connectors inspect high-level --json
```

## Agent Rules

- Run pm connectors inspect high-level before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
