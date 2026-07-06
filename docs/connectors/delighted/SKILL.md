---
name: pm-delighted
description: Delighted connector knowledge and safe action guide.
---

# pm-delighted

## Purpose

Reads Delighted survey responses, people, bounces, unsubscribes, and aggregate metrics through the Delighted REST API; can create/update and delete people.

## Icon

- asset: icons/delighted.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://delighted.com/docs/api

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- start_date
- api_key (secret)

## ETL Streams

- survey_responses:
  - primary key: id
  - cursor: updated_at
  - fields: comment(), created_at(), id(), notes(), permalink(), person(), person_properties(), score(), survey_type(), tags(), updated_at()
- people:
  - primary key: id
  - fields: created_at(), email(), id(), last_responded_at(), last_sent_at(), name(), next_survey_scheduled_at(), phone_number()
- bounces:
  - primary key: person_id
  - fields: bounced_at(), email(), name(), person_id()
- unsubscribes:
  - primary key: person_id
  - fields: email(), name(), person_id(), unsubscribed_at()
- metrics:
  - fields: detractor_count(), detractor_percent(), nps(), passive_count(), passive_percent(), promoter_count(), promoter_percent(), response_count()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_person:
  - endpoint: POST /people.json
  - risk: creates or updates a Delighted person and may trigger survey workflow depending on account settings
- delete_person:
  - endpoint: DELETE /people/{{ record.person_id }}.json
  - required fields: person_id
  - risk: deletes a Delighted person record

## Security

- read risk: external Delighted API read of survey responses, people, and aggregate NPS metrics
- write risk: creates/updates Delighted people and deletes existing people
- approval: reverse-ETL writes require plan preview and approval token
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect delighted
```

### Inspect as structured JSON

```bash
pm connectors inspect delighted --json
```

## Agent Rules

- Run pm connectors inspect delighted before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
