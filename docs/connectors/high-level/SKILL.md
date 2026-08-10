---
name: pm-high-level
description: High Level connector knowledge and safe action guide.
---

# pm-high-level

## Purpose

Reads HighLevel (Go HighLevel / LeadConnector) contacts, opportunities, pipelines, custom fields, and form submissions for a location through the HighLevel REST API.

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

- api_version
- base_url
- location_id (required)
- api_key (secret) (required)

## ETL Streams

- pipelines:
  - primary key: id
  - fields: dateAdded(string), dateUpdated(string), id(string), locationId(string), name(string), stages(array)
- contacts:
  - primary key: id
  - cursor: dateUpdated
  - fields: contactName(string), dateAdded(string), dateUpdated(string), email(string), firstName(string), id(string), lastName(string), locationId(string), phone(string), source(string), type(string)
- opportunities:
  - primary key: id
  - cursor: dateUpdated
  - fields: assignedTo(string), contactId(string), dateAdded(string), dateUpdated(string), id(string), monetaryValue(number), name(string), pipelineId(string), pipelineStageId(string), source(string), status(string)
- custom_fields:
  - primary key: id
  - fields: dataType(string), fieldKey(string), id(string), model(string), name(string), position(integer)
- form_submissions:
  - primary key: id
  - cursor: createdAt
  - fields: contactId(string), createdAt(string), email(string), formId(string), id(string), locationId(string), name(string)

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
