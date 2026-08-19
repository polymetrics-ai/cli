---
name: pm-smaily
description: Smaily connector knowledge and safe action guide.
---

# pm-smaily

## Purpose

Reads Smaily campaigns, segments, contacts, templates, automations, and organization users; creates/updates subscribers and segments, unsubscribes recipients, sends messages, and triggers automation workflows.

## Icon

- id: smaily
- asset: icons/smaily.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://smaily.com/help/api/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- api_username (required)
- base_url (required)
- segment_id
- api_password (secret) (required)

## ETL Streams

- campaigns:
  - primary key: id
  - fields: created_at(string), id(integer), name(string)
- segments:
  - primary key: id
  - fields: created_at(string), id(integer), name(string)
- subscribers:
  - primary key: id
  - fields: created_at(string), id(integer), name(string)
- templates:
  - primary key: id
  - fields: created_at(string), id(integer), name(string)
- automations:
  - primary key: id
  - fields: created_at(string), id(integer), name(string)
- segment_rules:
  - primary key: id
  - fields: filter_data(array), filter_type(string), id(integer), name(string), subscribers_count(integer)
- segment_subscribers:
  - primary key: email
  - fields: created_at(string), email(string), is_unsubscribed(integer), last_click_at(string), last_open_at(string), modified_at(string), subscribed_at(string), total_clicks(string), total_opens(string)
- ab_tests:
  - primary key: id
  - fields: created_at(string), id(integer), name(string)
- organization_users:
  - primary key: id
  - fields: email(string), id(integer), label(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Reverse ETL Actions

- create_or_update_subscriber:
  - endpoint: POST api/contact.php
  - required fields: email
  - risk: external mutation; creates or updates a subscriber (matched by email) on the connected Smaily account; does not trigger automation workflows; approval required
- create_or_update_segment:
  - endpoint: POST api/list.php
  - required fields: name, filter_type, filter_data
  - risk: external mutation; creates a new segment or, when id is set, overwrites an existing segment's filter definition on the connected Smaily account; approval required
- unsubscribe_recipient:
  - endpoint: POST api/unsubscribe.php
  - required fields: email, campaign_id
  - risk: external mutation; unsubscribes a recipient from a specific campaign (reflected in that campaign's statistics); approval required
- send_message:
  - endpoint: POST api/message/send.php
  - required fields: autoresponder_id, to
  - risk: external mutation; sends a real, individually-templated outbound email to real recipients using an automation workflow's template (without triggering the workflow itself); approval required
- trigger_automation_workflow:
  - endpoint: POST api/autoresponder.php
  - required fields: autoresponder, addresses
  - risk: external mutation; opts in subscribers and triggers a 'form submitted' automation workflow for them, updating subscriber data before any scheduled messages send; approval required
- launch_ab_test:
  - endpoint: POST api/split.php
  - required fields: splits, list, size, win_at
  - risk: external mutation; creates and, unless save_as_draft is set, immediately launches a real A/B test campaign send to a percentage of a real subscriber list, with the winning variant auto-sent to the remainder at win_at; approval required

## Security

- read risk: read-only campaign/segment/contact/template/automation/organization-user data from a connected Smaily account
- write risk: creates/updates subscribers and segments, unsubscribes a recipient from a campaign, sends an individually-templated outbound email, and triggers an automation workflow for real subscribers
- approval: required for all 5 write actions; read is unapproved
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect smaily
```

### Inspect as structured JSON

```bash
pm connectors inspect smaily --json
```

## Agent Rules

- Run pm connectors inspect smaily before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
