---
name: pm-savvycal
description: SavvyCal connector knowledge and safe action guide.
---

# pm-savvycal

## Purpose

Reads SavvyCal events, scheduling links, contacts, time zones, webhooks, and workflows, and writes scheduling-link and webhook lifecycle mutations, through the SavvyCal API.

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

- base_url
- page_size
- api_key (secret) (required)

## ETL Streams

- events:
  - primary key: id
  - fields: id(string), name(string)
- links:
  - primary key: id
  - fields: id(string), name(string)
- contacts:
  - primary key: id
  - fields: id(string), name(string)
- time_zones:
  - primary key: name
  - fields: abbreviation(string), display_name(string), dst(boolean), name(string), utc_offset(integer)
- webhooks:
  - primary key: id
  - fields: events(array), id(string), inserted_at(string), updated_at(string), url(string)
- workflows:
  - primary key: id
  - fields: id(string), inserted_at(string), name(string), scope_slug(string), updated_at(string)
- workflow_rules:
  - primary key: id
  - fields: id(string), position(integer), type(string), workflow_id(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Reverse ETL Actions

- create_personal_link:
  - endpoint: POST /v1/links
  - required fields: name
  - risk: creates a new scheduling link in the authenticated user's personal scope; external mutation, approval required
- create_scope_link:
  - endpoint: POST /v1/scopes/{{ record.scope_slug }}/links
  - required fields: scope_slug, name
  - risk: creates a new scheduling link under a specific team or individual scope; external mutation, approval required
- update_link:
  - endpoint: PATCH /v1/links/{{ record.id }}
  - required fields: id
  - risk: external mutation updating an existing scheduling link's name/slug/duration; approval required
- delete_link:
  - endpoint: DELETE /v1/links/{{ record.id }}
  - required fields: id
  - risk: destructive/irreversible: permanently deletes a scheduling link; approval required
- duplicate_link:
  - endpoint: POST /v1/links/{{ record.id }}/duplicate
  - required fields: id
  - risk: creates a copy of an existing scheduling link; low-risk external mutation, no approval required
- toggle_link:
  - endpoint: POST /v1/links/{{ record.id }}/toggle
  - required fields: id
  - risk: flips a scheduling link between active and disabled state, changing its public bookability; approval required
- cancel_event:
  - endpoint: POST /v1/events/{{ record.id }}/cancel
  - required fields: id
  - risk: destructive/irreversible: cancels a scheduled event, notifying attendees; approval required
- create_webhook:
  - endpoint: POST /v1/webhooks
  - required fields: url
  - risk: creates a new webhook subscription that will POST event notifications to an external URL; approval required
- delete_webhook:
  - endpoint: DELETE /v1/webhooks/{{ record.id }}
  - required fields: id
  - risk: destructive/irreversible: permanently deletes a webhook subscription; approval required

## Security

- read risk: external SavvyCal API read of event, scheduling link, contact, time zone, webhook, and workflow data
- write risk: external SavvyCal mutations: scheduling link create/update/delete/duplicate/toggle, event cancellation, webhook subscription create/delete
- approval: required for all write actions except duplicate_link (low-risk copy operation); delete_link/delete_webhook/cancel_event are destructive and irreversible
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect savvycal
```

### Inspect as structured JSON

```bash
pm connectors inspect savvycal --json
```

## Agent Rules

- Run pm connectors inspect savvycal before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
