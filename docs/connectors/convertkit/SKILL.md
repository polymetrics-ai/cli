---
name: pm-convertkit
description: ConvertKit connector knowledge and safe action guide.
---

# pm-convertkit

## Purpose

Reads ConvertKit (Kit) subscribers, forms, sequences, tags, broadcasts, custom fields, and purchases, and writes subscriber/tag/form/sequence/broadcast/custom-field/purchase/webhook mutations, through the ConvertKit v3 REST API.

## Icon

- asset: icons/convertkit.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.convertkit.com/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- access_token (secret)
- api_key (secret)
- api_secret (secret)

## ETL Streams

- subscribers:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), email_address(), first_name(), id(), state()
- forms:
  - primary key: id
  - cursor: created_at
  - fields: archived(), created_at(), format(), id(), name(), type()
- sequences:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), hold(), id(), name(), repeat()
- tags:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), id(), name()
- broadcasts:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), description(), id(), public(), published_at(), subject()
- custom_fields:
  - primary key: id
  - fields: id(), key(), label(), name()
- purchases:
  - primary key: id
  - cursor: transaction_time
  - fields: currency(), discount(), email_address(), id(), shipping(), status(), subtotal(), tax(), total(), transaction_id(), transaction_time()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- update_subscriber:
  - endpoint: PUT /subscribers/{{ record.id }}
  - required fields: id
  - risk: mutates an existing subscriber's name/email/custom-field values; external mutation, no approval required
- create_tag:
  - endpoint: POST /tags
  - risk: creates a new tag on the account; low-risk external mutation, no approval required
- tag_subscriber:
  - endpoint: POST /tags/{{ record.tag_id }}/subscribe
  - required fields: tag_id
  - risk: applies a tag to a subscriber (creating the subscriber if the email is new); external mutation, no approval required
- remove_tag_from_subscriber:
  - endpoint: DELETE /subscribers/{{ record.subscriber_id }}/tags/{{ record.tag_id }}
  - required fields: subscriber_id, tag_id
  - risk: removes a tag from a subscriber; external mutation, no approval required
- subscribe_to_form:
  - endpoint: POST /forms/{{ record.form_id }}/subscribe
  - required fields: form_id
  - risk: subscribes an email address to a form (creating the subscriber if the email is new); external mutation, no approval required
- subscribe_to_sequence:
  - endpoint: POST /sequences/{{ record.sequence_id }}/subscribe
  - required fields: sequence_id
  - risk: subscribes an email address to a sequence (creating the subscriber if the email is new); external mutation, no approval required
- create_broadcast:
  - endpoint: POST /broadcasts
  - risk: creates a draft or scheduled email broadcast; a scheduled broadcast (send_at/published_at set) will send to the account's live subscriber list, external mutation, approval required
- update_broadcast:
  - endpoint: PUT /broadcasts/{{ record.id }}
  - required fields: id
  - risk: mutates a draft or scheduled broadcast's content/send time; external mutation, approval required
- delete_broadcast:
  - endpoint: DELETE /broadcasts/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a draft or scheduled broadcast record; irreversible, approval required
- create_custom_field:
  - endpoint: POST /custom_fields
  - risk: creates a new custom subscriber field on the account (up to 140 total); low-risk external mutation, no approval required
- update_custom_field:
  - endpoint: PUT /custom_fields/{{ record.id }}
  - required fields: id
  - risk: renames a custom field's label (the underlying key is unchanged per Kit's own docs); external mutation, no approval required
- create_purchase:
  - endpoint: POST /purchases
  - risk: records a new purchase-tracking transaction for a subscriber; external mutation, no approval required
- create_webhook:
  - endpoint: POST /automations/hooks
  - risk: creates a webhook that POSTs subscriber-event payloads to an external URL the caller controls; external mutation, approval required
- delete_webhook:
  - endpoint: DELETE /automations/hooks/{{ record.rule_id }}
  - required fields: rule_id
  - risk: permanently deletes a webhook automation; irreversible, approval required

## Security

- read risk: external ConvertKit API read of subscriber and campaign data
- write risk: external mutation: creates/updates subscribers, tags, forms/sequences subscriptions, broadcasts (including scheduling live sends), custom fields, purchase records, and webhooks; deletes are limited to broadcasts/webhooks/tag-removal (no subscriber/custom-field/global-unsubscribe deletes)
- approval: required for all write actions; read-only otherwise
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect convertkit
```

### Inspect as structured JSON

```bash
pm connectors inspect convertkit --json
```

## Agent Rules

- Run pm connectors inspect convertkit before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
