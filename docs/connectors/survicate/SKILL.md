---
name: pm-survicate
description: Survicate connector knowledge and safe action guide.
---

# pm-survicate

## Purpose

Reads Survicate surveys, survey questions, responses, and respondent attributes, and manages GDPR personal-data requests, through the Survicate Data Export API v2. Read-only.

## Icon

- asset: icons/survicate.svg
- source: official
- review_status: official_verified
- review_url: https://developers.survicate.com/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- api_key (secret)

## ETL Streams

- surveys:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), id(), name(), updated_at()
- survey_questions:
  - primary key: survey_id, id
  - fields: answer_choices(), columns(), fields(), id(), introduction(), question(), survey_id(), type()
- responses:
  - primary key: uuid
  - fields: attributes(), collected_at(), device_type(), language(), operating_system(), questions(), respondent_uuid(), survey_id(), url(), uuid()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Survicate API read of survey, response, and respondent data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect survicate
```

### Inspect as structured JSON

```bash
pm connectors inspect survicate --json
```

## Agent Rules

- Run pm connectors inspect survicate before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
