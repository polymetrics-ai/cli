---
name: pm-agilecrm
description: AgileCRM connector knowledge and safe action guide.
---

# pm-agilecrm

## Purpose

Reads AgileCRM contacts, deals, tasks, milestone pipelines, campaigns, and support tickets, and writes contact/deal/task create, update, and delete actions, through the AgileCRM REST API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- domain
- email
- mode
- api_key (secret)

## ETL Streams

- contacts:
  - primary key: id
  - cursor: created_time
  - fields: created_time(), id(), lead_score(), owner_id(), properties(), star_value(), tags(), type(), updated_time()
- deals:
  - primary key: id
  - cursor: created_time
  - fields: close_date(), created_time(), expected_value(), id(), milestone(), name(), owner_id(), pipeline_id(), probability()
- tasks:
  - primary key: id
  - cursor: created_time
  - fields: created_time(), due(), id(), is_complete(), owner_id(), priority_type(), status(), subject(), type()
- milestone:
  - primary key: id
  - fields: id(), milestones(), name(), pipeline_default()
- campaigns:
  - primary key: id
  - cursor: created_time
  - fields: created_time(), creatorName(), domainUserId(), id(), name(), rules(), updated_time()
- tickets_filters:
  - primary key: id
  - cursor: updated_time
  - fields: conditions(), id(), is_default_filter(), name(), owner_id(), updated_time()
- tickets:
  - primary key: id
  - cursor: last_updated_time
  - fields: contactID(), created_time(), filter_id(), id(), is_favorite(), is_spam(), last_updated_time(), priority(), requester_email(), requester_name(), source(), status(), subject(), type()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_contact:
  - endpoint: POST /contacts
  - risk: external mutation; creates a live AgileCRM contact/company; approval required
- update_contact:
  - endpoint: PUT /contacts/edit-properties
  - risk: external mutation; overwrites live AgileCRM contact property fields; approval required
- delete_contact:
  - endpoint: DELETE /contacts/{{ record.id }}
  - required fields: id
  - risk: external mutation; irreversibly deletes a live AgileCRM contact; approval required
- create_deal:
  - endpoint: POST /opportunity
  - risk: external mutation; creates a live AgileCRM deal; approval required
- update_deal:
  - endpoint: PUT /opportunity/partial-update
  - risk: external mutation; overwrites live AgileCRM deal fields; approval required
- delete_deal:
  - endpoint: DELETE /opportunity/{{ record.id }}
  - required fields: id
  - risk: external mutation; irreversibly deletes a live AgileCRM deal; approval required
- create_task:
  - endpoint: POST /tasks
  - risk: external mutation; creates a live AgileCRM task; approval required
- update_task:
  - endpoint: PUT /tasks/partial-update
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
