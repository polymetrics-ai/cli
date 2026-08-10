---
name: pm-calendly
description: Calendly connector knowledge and safe action guide.
---

# pm-calendly

## Purpose

Reads Calendly scheduled events (and their invitees), event types, organization memberships, groups, routing forms and submissions, webhook subscriptions, availability schedules, activity log entries, and the current user, and manages bookings/webhooks/memberships/invitations/event types through the Calendly v2 REST API.

## Icon

- id: simple-icons-calendly
- asset: icons/simple-icons/calendly.svg
- title: Calendly
- simple_icon_slug: calendly
- simple_icon_hex: 006BFF
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Calendly
- match: exact-name-or-slug
- matched_by: calendly

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- organization_uri (required)
- page_size
- routing_form_uri
- start_date
- user_uri
- api_key (secret) (required)

## ETL Streams

- scheduled_events:
  - primary key: id
  - cursor: start_time
  - fields: cancellation(object), created_at(string), end_time(string), event_guests(array), event_memberships(array), event_type(string), id(string), invitees_counter(object), location(object), name(string), start_time(string), status(string), updated_at(string), uri(string)
- event_types:
  - primary key: id
  - cursor: updated_at
  - fields: active(boolean), booking_method(string), color(string), created_at(string), deleted_at(string), description_html(string), description_plain(string), duration(integer), id(string), kind(string), name(string), pooling_type(string), scheduling_url(string), secret(boolean), slug(string), type(string), updated_at(string), uri(string)
- organization_memberships:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), id(string), organization(string), role(string), updated_at(string), uri(string), user(string), user_email(string), user_name(string)
- groups:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), id(string), name(string), organization(string), updated_at(string), uri(string)
- users:
  - primary key: id
  - fields: avatar_url(string), created_at(string), current_organization(string), email(string), id(string), name(string), scheduling_url(string), slug(string), timezone(string), updated_at(string), uri(string)
- routing_forms:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), id(string), name(string), organization(string), questions(array), updated_at(string), uri(string)
- routing_form_submissions:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), id(string), questions_and_answers(array), routing_form(string), submitter(string), submitter_type(string), tracking(object), updated_at(string), uri(string)
- webhook_subscriptions:
  - primary key: id
  - cursor: updated_at
  - fields: callback_url(string), created_at(string), creator(string), events(array), id(string), organization(string), retry_started_at(string), scope(string), state(string), updated_at(string), uri(string), user(string)
- user_availability_schedules:
  - primary key: id
  - fields: default(boolean), id(string), name(string), rules(array), timezone(string), uri(string)
- group_relationships:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), group(string), id(string), organization(string), owner(string), updated_at(string), uri(string)
- activity_log_entries:
  - primary key: id
  - cursor: occurred_at
  - fields: action(string), actor(object), details(object), id(string), namespace(string), occurred_at(string), organization(string), uri(string)
- invitees:
  - primary key: id
  - cursor: updated_at
  - fields: cancel_url(string), cancellation(object), created_at(string), email(string), event(string), first_name(string), id(string), last_name(string), name(string), new_invitee(string), old_invitee(string), payment(object), questions_and_answers(array), reschedule_url(string), rescheduled(boolean), routing_form_submission(string), scheduled_event_id(string), status(string), timezone(string), tracking(object), updated_at(string), uri(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- cancel_scheduled_event:
  - endpoint: POST /scheduled_events/{{ record.uuid }}/cancellation
  - required fields: uuid
  - risk: external mutation; cancels a real scheduled event and notifies invitees; approval required
- create_invitee:
  - endpoint: POST /invitees
  - required fields: event_type, start_time, invitee
  - risk: external mutation; books a real meeting slot on the target event type and notifies the invitee; approval required
- create_webhook_subscription:
  - endpoint: POST /webhook_subscriptions
  - required fields: url, events, organization, scope
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
  - required fields: organization_uuid, email
  - risk: external mutation; sends a real organization-invitation email to the given address; approval required
- create_one_off_event_type:
  - endpoint: POST /one_off_event_types
  - required fields: name, host, duration, date_setting
  - risk: external mutation; publishes a new one-off publicly-bookable event type; approval required
- create_share:
  - endpoint: POST /shares
  - required fields: event_type
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
