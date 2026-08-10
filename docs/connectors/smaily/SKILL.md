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

- api_username
- base_url
- segment_id
- api_password (secret)

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

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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

## Command Surface

- Run Smaily's declared streams and reverse-ETL actions.
- Usage: pm smaily <command> [flags]
- Read streams
- Reverse ETL writes
- Other Commands
  - ab tests list - Run the ab tests ETL stream [intent=etl availability=implemented stream=ab_tests]
  - api get api campaign-php-id - Documented GET api/campaign.php?id=... (not implemented) [intent=direct_read availability=not_implemented operation=smaily.get.api-campaign-php-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api message action log-php - Documented GET api/message/action/log.php (not implemented) [intent=direct_read availability=not_implemented operation=smaily.get.api-message-action-log-php]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api split-php-id - Documented GET api/split.php?id=... (not implemented) [intent=direct_read availability=not_implemented operation=smaily.get.api-split-php-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api subscribers action log-php - Documented GET api/subscribers/action/log.php (not implemented) [intent=direct_read availability=not_implemented operation=smaily.get.api-subscribers-action-log-php]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api post api campaign-php - Documented POST api/campaign.php (not implemented) [intent=direct_write availability=not_implemented operation=smaily.post.api-campaign-php]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api subscribers forget-php - Documented POST api/subscribers/forget.php (not implemented) [intent=direct_write availability=not_implemented operation=smaily.post.api-subscribers-forget-php]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api subscribers opt-in-php - Documented POST api/subscribers/opt-in.php (not implemented) [intent=direct_write availability=not_implemented operation=smaily.post.api-subscribers-opt-in-php]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - automations list - Run the automations ETL stream [intent=etl availability=implemented stream=automations]
  - campaigns list - Run the campaigns ETL stream [intent=etl availability=implemented stream=campaigns]
  - create or update segment apply - Plan and execute the create or update segment reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_or_update_segment]; approval: requires plan, preview, approval, and execute; risk: external mutation; creates a new segment or, when id is set, overwrites an existing segment's filter definition on the connected Smaily account; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - create or update subscriber apply - Plan and execute the create or update subscriber reverse-ETL action [intent=reverse_etl availability=implemented write=create_or_update_subscriber]; approval: requires plan, preview, approval, and execute; risk: external mutation; creates or updates a subscriber (matched by email) on the connected Smaily account; does not trigger automation workflows; approval required; flags: --email (required)
  - launch ab test apply - Plan and execute the launch ab test reverse-ETL action [intent=reverse_etl availability=not_implemented write=launch_ab_test]; approval: requires plan, preview, approval, and execute; risk: external mutation; creates and, unless save_as_draft is set, immediately launches a real A/B test campaign send to a percentage of a real subscriber list, with the winning variant auto-sent to the remainder at win_at; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - organization users list - Run the organization users ETL stream [intent=etl availability=implemented stream=organization_users]
  - segment rules list - Run the segment rules ETL stream [intent=etl availability=implemented stream=segment_rules]
  - segment subscribers list - Run the segment subscribers ETL stream [intent=etl availability=implemented stream=segment_subscribers]
  - segments list - Run the segments ETL stream [intent=etl availability=implemented stream=segments]
  - send message apply - Plan and execute the send message reverse-ETL action [intent=reverse_etl availability=implemented write=send_message]; approval: requires plan, preview, approval, and execute; risk: external mutation; sends a real, individually-templated outbound email to real recipients using an automation workflow's template (without triggering the workflow itself); approval required; flags: --autoresponder_id (required), --to (required)
  - subscribers list - Run the subscribers ETL stream [intent=etl availability=implemented stream=subscribers]
  - templates list - Run the templates ETL stream [intent=etl availability=implemented stream=templates]
  - trigger automation workflow apply - Plan and execute the trigger automation workflow reverse-ETL action [intent=reverse_etl availability=not_implemented write=trigger_automation_workflow]; approval: requires plan, preview, approval, and execute; risk: external mutation; opts in subscribers and triggers a 'form submitted' automation workflow for them, updating subscriber data before any scheduled messages send; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - unsubscribe recipient apply - Plan and execute the unsubscribe recipient reverse-ETL action [intent=reverse_etl availability=implemented write=unsubscribe_recipient]; approval: requires plan, preview, approval, and execute; risk: external mutation; unsubscribes a recipient from a specific campaign (reflected in that campaign's statistics); approval required; flags: --campaign_id (required), --email (required)

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
