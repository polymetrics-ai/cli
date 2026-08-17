---
name: pm-twilio-taskrouter
description: Twilio TaskRouter connector knowledge and safe action guide.
---

# pm-twilio-taskrouter

## Purpose

Reads Twilio TaskRouter workers, tasks, activities, task queues, and workflows for a workspace.

## Icon

- asset: icons/twilio.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://www.twilio.com/docs/taskrouter/api

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- workspace_sid
- account_sid (secret)
- auth_token (secret)

## ETL Streams

- workers:
  - primary key: sid
  - fields: activity_name(), available(), friendly_name(), sid()
- tasks:
  - primary key: sid
  - fields: assignment_status(), sid(), workflow_sid()
- activities:
  - primary key: sid
  - fields: friendly_name(), sid()
- task_queues:
  - primary key: sid
  - fields: friendly_name(), sid()
- workflows:
  - primary key: sid
  - fields: friendly_name(), sid()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Twilio TaskRouter API read of workspace workers, tasks, activities, task queues, and workflows
- approval: none; read-only, no reverse-ETL writes implemented by legacy
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect twilio-taskrouter
```

### Inspect as structured JSON

```bash
pm connectors inspect twilio-taskrouter --json
```

## Agent Rules

- Run pm connectors inspect twilio-taskrouter before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
