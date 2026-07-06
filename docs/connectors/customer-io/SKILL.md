---
name: pm-customer-io
description: Customer.io connector knowledge and safe action guide.
---

# pm-customer-io

## Purpose

Reads Customer.io campaigns, newsletters, segments, broadcasts, activities, messages, exports, transactional templates, object types, reporting webhooks, sender identities, snippets, subscription channels/topics, workspaces, and collections; writes snippet/webhook/segment mutations and can send transactional email or trigger broadcasts, through the Customer.io App API.

## Icon

- asset: icons/customer-io.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://customer.io/docs/api/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- page_size
- app_api_key (secret)

## ETL Streams

- campaigns:
  - primary key: id
  - cursor: updated
  - fields: active(), created(), id(), name(), state(), type(), updated()
- newsletters:
  - primary key: id
  - cursor: updated
  - fields: created(), id(), name(), subject(), type(), updated()
- segments:
  - primary key: id
  - cursor: updated
  - fields: description(), id(), name(), state(), type(), updated()
- broadcasts:
  - primary key: id
  - cursor: updated
  - fields: active(), created(), id(), name(), state(), type(), updated()
- activities:
  - primary key: id
  - cursor: timestamp
  - fields: customer_id(), customer_identifiers(), data(), delivery_id(), delivery_type(), id(), timestamp(), type()
- messages:
  - primary key: id
  - fields: action_id(), broadcast_id(), campaign_id(), content_id(), created(), customer_id(), failure_message(), id(), newsletter_id(), recipient(), subject(), type()
- exports:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), description(), downloads(), failed(), id(), status(), total(), type(), updated_at(), user_email(), user_id()
- transactional:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), description(), hide_message_body(), id(), link_tracking(), name(), open_tracking(), queue_drafts(), send_to_unsubscribed(), updated_at()
- object_types:
  - primary key: id
  - fields: enabled(), icon(), id(), name(), singular_name(), singular_slug(), slug()
- reporting_webhooks:
  - primary key: id
  - fields: disabled(), endpoint(), events(), full_resolution(), id(), name(), with_content()
- sender_identities:
  - primary key: id
  - fields: address(), auto_generated(), deduplicate_id(), email(), id(), name(), template_type()
- snippets:
  - primary key: name
  - cursor: updated_at
  - fields: name(), updated_at(), value()
- subscription_channels:
  - primary key: id
  - fields: description(), id(), name(), subscribed_by_default(), type()
- subscription_topics:
  - primary key: id
  - fields: description(), id(), identifier(), name(), subscribed_by_default()
- workspaces:
  - primary key: id
  - fields: billable_messages_sent(), id(), messages_sent(), name(), object_types(), objects(), people()
- collections:
  - primary key: id
  - cursor: updated_at
  - fields: bytes(), created_at(), id(), name(), rows(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_snippet:
  - endpoint: POST /snippets
  - risk: external mutation; creates a reusable content snippet referenced by live messages/newsletters
- update_snippet:
  - endpoint: PUT /snippets
  - risk: external mutation; overwrites the content of a live snippet, changing every message/newsletter that references it
- delete_snippet:
  - endpoint: DELETE /snippets/{{ record.name }}
  - required fields: name
  - risk: external mutation; permanently removes a snippet; irreversible, breaks any message/newsletter still referencing it; approval required
- create_reporting_webhook:
  - endpoint: POST /reporting_webhooks
  - risk: external mutation; registers a new reporting webhook that will deliver live workspace event data to the given endpoint URL
- update_reporting_webhook:
  - endpoint: PUT /reporting_webhooks/{{ record.id }}
  - required fields: id
  - risk: external mutation; changes a live reporting webhook's target endpoint/event selection or enables/disables delivery
- delete_reporting_webhook:
  - endpoint: DELETE /reporting_webhooks/{{ record.id }}
  - required fields: id
  - risk: external mutation; permanently removes a reporting webhook; event delivery to its target URL stops immediately; approval required
- create_manual_segment:
  - endpoint: POST /segments
  - risk: external mutation; creates a new manual segment in the live workspace
- delete_manual_segment:
  - endpoint: DELETE /segments/{{ record.id }}
  - required fields: id
  - risk: external mutation; permanently removes a manual segment; irreversible, any campaign/newsletter targeting it loses that audience slice immediately; approval required
- send_email:
  - endpoint: POST /send/email
  - risk: sends a live transactional email to the given recipient on the workspace's behalf; irreversible once delivered
- trigger_broadcast:
  - endpoint: POST /campaigns/{{ record.broadcast_id }}/triggers
  - required fields: broadcast_id
  - risk: triggers a live API-triggered broadcast to its default audience; sends real messages to recipients, irreversible once delivered

## Security

- read risk: external Customer.io App API read of campaign/newsletter/segment/broadcast/activity/message/export/transactional/webhook/subscription/workspace metadata
- write risk: external mutation of live Customer.io workspace config (snippets/webhooks/segments) and live message sends (transactional email, broadcast triggers); irreversible once delivered; approval required
- approval: read: none; write: required
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect customer-io
```

### Inspect as structured JSON

```bash
pm connectors inspect customer-io --json
```

## Agent Rules

- Run pm connectors inspect customer-io before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
