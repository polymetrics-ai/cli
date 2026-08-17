---
name: pm-calendly
description: Calendly connector knowledge and safe action guide.
---

# pm-calendly

## Purpose

Reads Calendly scheduled events (and their invitees), event types, organization memberships, groups, routing forms and submissions, webhook subscriptions, availability schedules, activity log entries, and the current user, and manages bookings/webhooks/memberships/invitations/event types through the Calendly v2 REST API.

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
- organization_uri
- page_size
- routing_form_uri
- start_date
- user_uri
- api_key (secret)

## ETL Streams

- scheduled_events:
  - primary key: id
  - cursor: start_time
  - fields: cancellation(), created_at(), end_time(), event_guests(), event_memberships(), event_type(), id(), invitees_counter(), location(), name(), start_time(), status(), updated_at(), uri()
- event_types:
  - primary key: id
  - cursor: updated_at
  - fields: active(), booking_method(), color(), created_at(), deleted_at(), description_html(), description_plain(), duration(), id(), kind(), name(), pooling_type(), scheduling_url(), secret(), slug(), type(), updated_at(), uri()
- organization_memberships:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), id(), organization(), role(), updated_at(), uri(), user(), user_email(), user_name()
- groups:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), id(), name(), organization(), updated_at(), uri()
- users:
  - primary key: id
  - fields: avatar_url(), created_at(), current_organization(), email(), id(), name(), scheduling_url(), slug(), timezone(), updated_at(), uri()
- routing_forms:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), id(), name(), organization(), questions(), updated_at(), uri()
- routing_form_submissions:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), id(), questions_and_answers(), routing_form(), submitter(), submitter_type(), tracking(), updated_at(), uri()
- webhook_subscriptions:
  - primary key: id
  - cursor: updated_at
  - fields: callback_url(), created_at(), creator(), events(), id(), organization(), retry_started_at(), scope(), state(), updated_at(), uri(), user()
- user_availability_schedules:
  - primary key: id
  - fields: default(), id(), name(), rules(), timezone(), uri()
- group_relationships:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), group(), id(), organization(), owner(), updated_at(), uri()
- activity_log_entries:
  - primary key: id
  - cursor: occurred_at
  - fields: action(), actor(), details(), id(), namespace(), occurred_at(), organization(), uri()
- invitees:
  - primary key: id
  - cursor: updated_at
  - fields: cancel_url(), cancellation(), created_at(), email(), event(), first_name(), id(), last_name(), name(), new_invitee(), old_invitee(), payment(), questions_and_answers(), reschedule_url(), rescheduled(), routing_form_submission(), scheduled_event_id(), status(), timezone(), tracking(), updated_at(), uri()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- cancel_scheduled_event:
  - endpoint: POST /scheduled_events/{{ record.uuid }}/cancellation
  - required fields: uuid
  - risk: external mutation; cancels a real scheduled event and notifies invitees; approval required
- create_invitee:
  - endpoint: POST /invitees
  - risk: external mutation; books a real meeting slot on the target event type and notifies the invitee; approval required
- create_webhook_subscription:
  - endpoint: POST /webhook_subscriptions
  - risk: external mutation; registers a new webhook endpoint that will receive live invitee/routing-form event payloads; approval required
- delete_webhook_subscription:
  - endpoint: DELETE /webhook_subscriptions/{{ record.uuid }}
  - required fields: uuid
  - risk: destructive; permanently deletes a webhook subscription; approval required
- remove_organization_membership:
  - endpoint: DELETE /organization_memberships/{{ record.uuid }}
  - required fields: uuid
  - risk: destructive; permanently removes a user's membership from the organization, revoking their access; approval required
- invite_user_to_organization:
  - endpoint: POST /organizations/{{ record.organization_uuid }}/invitations
  - required fields: organization_uuid
  - optional fields: email
  - risk: external mutation; sends a real organization-invitation email to the given address; approval required
- create_one_off_event_type:
  - endpoint: POST /one_off_event_types
  - risk: external mutation; publishes a new one-off publicly-bookable event type; approval required
- create_share:
  - endpoint: POST /shares
  - risk: external mutation; creates a new shareable booking link with its own spot limit for an event type; approval required

## Security

- read risk: external Calendly API read of scheduling, organization, routing-form, webhook, and activity-log data
- write risk: external mutation of live scheduling data: cancels real scheduled events and books new ones (notifying invitees), creates/deletes webhook subscriptions, removes organization memberships, sends organization invitation emails, and creates one-off event types/shareable booking links
- approval: required for every write action
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect calendly
```

### Inspect as structured JSON

```bash
pm connectors inspect calendly --json
```

## Agent Rules

- Run pm connectors inspect calendly before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
