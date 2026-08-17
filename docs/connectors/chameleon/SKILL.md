---
name: pm-chameleon
description: Chameleon connector knowledge and safe action guide.
---

# pm-chameleon

## Purpose

Reads Chameleon surveys, tours, launchers, tooltips, and segments through the Chameleon v3 REST API.

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

- base_url
- mode
- api_key (secret)

## ETL Streams

- surveys:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), id(), is_live(), state(), title(), type(), updated_at()
- tours:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), id(), is_live(), state(), title(), type(), updated_at()
- launchers:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), id(), is_live(), state(), title(), type(), updated_at()
- tooltips:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), id(), is_live(), state(), title(), type(), updated_at()
- segments:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), description(), id(), name(), updated_at()
- embeds:
  - primary key: id
  - cursor: updated_at
  - fields: archived_at(), created_at(), dashboard_url(), description(), id(), name(), position(), published_at(), segment_ids(), style(), tag_ids(), updated_at()
- event_names:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), dashboard_url(), description(), id(), kind(), last_seen_at(), name(), published_at(), source(), uid(), updated_at()
- tags:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), description(), disabled_at(), id(), last_seen_at(), models_count(), name(), uid(), updated_at()
- deliveries:
  - primary key: id
  - cursor: updated_at
  - fields: at(), at_href(), created_at(), from(), group_kind(), id(), idempotency_key(), interaction_id(), model_id(), model_kind(), once(), options(), profile_id(), until(), updated_at(), use_segmentation()
- webhooks:
  - primary key: id
  - fields: id(), last_item_at(), last_item_error(), last_item_state(), name(), uid()
- companies:
  - primary key: id
  - fields: clv(), created_at(), domain(), id(), plan(), uid()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- publish_survey:
  - endpoint: PATCH /edit/surveys/{{ record.id }}
  - required fields: id
  - risk: external mutation publishing/unpublishing a live in-product Microsurvey to end-users; approval required
- publish_tour:
  - endpoint: PATCH /edit/tours/{{ record.id }}
  - required fields: id
  - risk: external mutation publishing/unpublishing a live in-product Tour to end-users; approval required
- publish_launcher:
  - endpoint: PATCH /edit/launchers/{{ record.id }}
  - required fields: id
  - risk: external mutation publishing/unpublishing a live in-product Launcher to end-users; approval required
- publish_tooltip:
  - endpoint: PATCH /edit/tooltips/{{ record.id }}
  - required fields: id
  - risk: external mutation publishing/unpublishing a live in-product Tooltip to end-users; approval required
- publish_embed:
  - endpoint: PATCH /edit/embeds/{{ record.id }}
  - required fields: id
  - risk: external mutation publishing/unpublishing a live in-product Embeddable to end-users; approval required
- create_delivery:
  - endpoint: POST /edit/deliveries
  - risk: external mutation directly triggering a Tour or Microsurvey experience for one specific end-user; approval required
- delete_delivery:
  - endpoint: DELETE /edit/deliveries/{{ record.id }}
  - required fields: id
  - risk: cancels a not-yet-triggered Delivery; irreversible once the target has already been shown, approval required
- create_webhook:
  - endpoint: POST /edit/webhooks
  - risk: external mutation creating a new outbound webhook subscription that will POST Chameleon event data to a third-party URL; approval required
- delete_webhook:
  - endpoint: DELETE /edit/webhooks/{{ record.id }}
  - required fields: id
  - risk: irreversible removal of an outbound webhook subscription; approval required

## Security

- read risk: external Chameleon API read of in-product experience, segment, tag, event, delivery, webhook, and company data
- write risk: external mutations publishing/unpublishing in-product experiences, triggering/cancelling user-targeted Deliveries, and creating/deleting outbound Webhook subscriptions; every write action requires approval
- approval: read: none; write: required for every action
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect chameleon
```

### Inspect as structured JSON

```bash
pm connectors inspect chameleon --json
```

## Agent Rules

- Run pm connectors inspect chameleon before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
