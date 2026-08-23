---
name: pm-agilecrm
description: AgileCRM connector knowledge and safe action guide.
---

# pm-agilecrm

## Purpose

Reads AgileCRM contacts, deals, tasks, milestone pipelines, campaigns, and support tickets, and writes contact/deal/task create, update, and delete actions, through the AgileCRM REST API.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- domain (required)
- email (required)
- mode
- api_key (secret) (required)

## ETL Streams

- contacts:
  - primary key: id
  - cursor: created_time
  - fields: created_time(integer), id(integer), lead_score(integer), owner_id(string), properties(array), star_value(integer), tags(array), type(string), updated_time(integer)
- deals:
  - primary key: id
  - cursor: created_time
  - fields: close_date(integer), created_time(integer), expected_value(number), id(integer), milestone(string), name(string), owner_id(string), pipeline_id(integer), probability(integer)
- tasks:
  - primary key: id
  - cursor: created_time
  - fields: created_time(integer), due(integer), id(integer), is_complete(boolean), owner_id(string), priority_type(string), status(string), subject(string), type(string)
- milestone:
  - primary key: id
  - fields: id(integer), milestones(string), name(string), pipeline_default(boolean)
- campaigns:
  - primary key: id
  - cursor: created_time
  - fields: created_time(integer), creatorName(string), domainUserId(integer), id(integer), name(string), rules(string), updated_time(integer)
- tickets_filters:
  - primary key: id
  - cursor: updated_time
  - fields: conditions(array), id(integer), is_default_filter(boolean), name(string), owner_id(integer), updated_time(integer)
- tickets:
  - primary key: id
  - cursor: last_updated_time
  - fields: contactID(integer), created_time(integer), filter_id(string), id(integer), is_favorite(boolean), is_spam(boolean), last_updated_time(integer), priority(string), requester_email(string), requester_name(string), source(string), status(string), subject(string), type(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_contact:
  - endpoint: POST /contacts
  - required fields: properties
  - risk: external mutation; creates a live AgileCRM contact/company; approval required
- update_contact:
  - endpoint: PUT /contacts/edit-properties
  - required fields: id, properties
  - risk: external mutation; overwrites live AgileCRM contact property fields; approval required
- delete_contact:
  - endpoint: DELETE /contacts/{{ record.id }}
  - required fields: id
  - risk: external mutation; irreversibly deletes a live AgileCRM contact; approval required
- create_deal:
  - endpoint: POST /opportunity
  - required fields: name
  - risk: external mutation; creates a live AgileCRM deal; approval required
- update_deal:
  - endpoint: PUT /opportunity/partial-update
  - required fields: id
  - risk: external mutation; overwrites live AgileCRM deal fields; approval required
- delete_deal:
  - endpoint: DELETE /opportunity/{{ record.id }}
  - required fields: id
  - risk: external mutation; irreversibly deletes a live AgileCRM deal; approval required
- create_task:
  - endpoint: POST /tasks
  - required fields: subject, type
  - risk: external mutation; creates a live AgileCRM task; approval required
- update_task:
  - endpoint: PUT /tasks/partial-update
  - required fields: id
  - risk: external mutation; overwrites live AgileCRM task fields; approval required
- delete_task:
  - endpoint: DELETE /tasks/{{ record.id }}
  - required fields: id
  - risk: external mutation; irreversibly deletes a live AgileCRM task; approval required

## Security

- read risk: external AgileCRM API read of contacts, deals, tasks, pipeline, campaign, and ticket data
- write risk: external mutation of live AgileCRM contacts, deals, and tasks including irreversible deletes; approval required for every write action
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect agilecrm
```

### Inspect as structured JSON

```bash
pm connectors inspect agilecrm --json
```

## Agent Rules

- Run pm connectors inspect agilecrm before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
