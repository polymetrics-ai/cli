---
name: pm-oncehub
description: OnceHub connector knowledge and safe action guide.
---

# pm-oncehub

## Purpose

Reads OnceHub bookings, contacts, booking pages, users, and event types through the OnceHub REST API.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- max_pages
- mode
- page_size
- start_date
- api_key (secret)

## ETL Streams

- bookings:
  - primary key: id
  - cursor: last_updated_time
  - fields: booking_page(string), contact(string), creation_time(string), customer_timezone(string), duration_minutes(number), event_type(string), id(string), in_trash(boolean), last_updated_time(string), location_description(string), object(string), owner(string), starting_time(string), status(string), subject(string), tracking_id(string)
- contacts:
  - primary key: id
  - cursor: last_updated_time
  - fields: creation_time(string), email(string), first_name(string), id(string), last_updated_time(string), mobile_phone(string), object(string), owner(string), timezone(string)
- booking_pages:
  - primary key: id
  - fields: active(boolean), id(string), label(string), name(string), object(string), timezone(string), url(string)
- users:
  - primary key: id
  - fields: email(string), first_name(string), id(string), last_name(string), object(string), role_name(string), status(string)
- event_types:
  - primary key: id
  - fields: id(string), label(string), name(string), object(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external OnceHub API read of scheduling, contact, and user data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run OnceHub's declared streams and reverse-ETL actions.
- Usage: pm oncehub <command> [flags]
- Read streams
- Reverse ETL writes
- Other Commands
  - api delete contacts id - Documented DELETE /contacts/{id} (not implemented) [intent=direct_write availability=not_implemented operation=oncehub.delete.contacts-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete users id - Documented DELETE /users/{id} (not implemented) [intent=direct_write availability=not_implemented operation=oncehub.delete.users-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete v2 bookings id - Documented DELETE /v2/bookings/{id} (not implemented) [intent=direct_write availability=not_implemented operation=oncehub.delete.v2-bookings-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete webhooks id - Documented DELETE /webhooks/{id} (not implemented) [intent=direct_write availability=not_implemented operation=oncehub.delete.webhooks-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api get booking-calendars - Documented GET /booking-calendars (not implemented) [intent=direct_read availability=not_implemented operation=oncehub.get.booking-calendars]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get booking-calendars id - Documented GET /booking-calendars/{id} (not implemented) [intent=direct_read availability=not_implemented operation=oncehub.get.booking-calendars-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get booking-calendars id time-slots - Documented GET /booking-calendars/{id}/time-slots (not implemented) [intent=direct_read availability=not_implemented operation=oncehub.get.booking-calendars-id-time-slots]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get bookings - Documented GET /bookings (not implemented) [intent=direct_read availability=not_implemented operation=oncehub.get.bookings]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get bookings id - Documented GET /bookings/{id} (not implemented) [intent=direct_read availability=not_implemented operation=oncehub.get.bookings-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get contacts - Documented GET /contacts (not implemented) [intent=direct_read availability=not_implemented operation=oncehub.get.contacts]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get contacts id - Documented GET /contacts/{id} (not implemented) [intent=direct_read availability=not_implemented operation=oncehub.get.contacts-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get notifications sms - Documented GET /notifications/sms (not implemented) [intent=direct_read availability=not_implemented operation=oncehub.get.notifications-sms]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get teams - Documented GET /teams (not implemented) [intent=direct_read availability=not_implemented operation=oncehub.get.teams]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get teams id - Documented GET /teams/{id} (not implemented) [intent=direct_read availability=not_implemented operation=oncehub.get.teams-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get test - Documented GET /test (not implemented) [intent=direct_read availability=not_implemented operation=oncehub.get.test]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get users - Documented GET /users (not implemented) [intent=direct_read availability=not_implemented operation=oncehub.get.users]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get users id - Documented GET /users/{id} (not implemented) [intent=direct_read availability=not_implemented operation=oncehub.get.users-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get users id scheduling-availability - Documented GET /users/{id}/scheduling-availability (not implemented) [intent=direct_read availability=not_implemented operation=oncehub.get.users-id-scheduling-availability]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 availability - Documented GET /v2/availability (not implemented) [intent=direct_read availability=not_implemented operation=oncehub.get.v2-availability]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 webhooks - Documented GET /v2/webhooks (not implemented) [intent=direct_read availability=not_implemented operation=oncehub.get.v2-webhooks]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get webhooks - Documented GET /webhooks (not implemented) [intent=direct_read availability=not_implemented operation=oncehub.get.webhooks]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get webhooks id - Documented GET /webhooks/{id} (not implemented) [intent=direct_read availability=not_implemented operation=oncehub.get.webhooks-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api patch contacts id - Documented PATCH /contacts/{id} (not implemented) [intent=direct_write availability=not_implemented operation=oncehub.patch.contacts-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api patch users id - Documented PATCH /users/{id} (not implemented) [intent=direct_write availability=not_implemented operation=oncehub.patch.users-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api patch users id scheduling-availability - Documented PATCH /users/{id}/scheduling-availability (not implemented) [intent=direct_write availability=not_implemented operation=oncehub.patch.users-id-scheduling-availability]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post booking-calendars id one-time-links - Documented POST /booking-calendars/{id}/one-time-links (not implemented) [intent=direct_write availability=not_implemented operation=oncehub.post.booking-calendars-id-one-time-links]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post booking-calendars id schedule - Documented POST /booking-calendars/{id}/schedule (not implemented) [intent=direct_write availability=not_implemented operation=oncehub.post.booking-calendars-id-schedule]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post bookings id cancel - Documented POST /bookings/{id}/cancel (not implemented) [intent=direct_write availability=not_implemented operation=oncehub.post.bookings-id-cancel]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post bookings id no-show - Documented POST /bookings/{id}/no-show (not implemented) [intent=direct_write availability=not_implemented operation=oncehub.post.bookings-id-no-show]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post bookings id reassign - Documented POST /bookings/{id}/reassign (not implemented) [intent=direct_write availability=not_implemented operation=oncehub.post.bookings-id-reassign]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post bookings id request-reschedule - Documented POST /bookings/{id}/request-reschedule (not implemented) [intent=direct_write availability=not_implemented operation=oncehub.post.bookings-id-request-reschedule]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post contacts - Documented POST /contacts (not implemented) [intent=direct_write availability=not_implemented operation=oncehub.post.contacts]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post users - Documented POST /users (not implemented) [intent=direct_write availability=not_implemented operation=oncehub.post.users]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v2 bookings - Documented POST /v2/bookings (not implemented) [intent=direct_write availability=not_implemented operation=oncehub.post.v2-bookings]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post webhooks - Documented POST /webhooks (not implemented) [intent=direct_write availability=not_implemented operation=oncehub.post.webhooks]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api webhook booking-canceled-post - Documented WEBHOOK booking.canceled#POST (not implemented) [intent=docs_only availability=not_implemented operation=oncehub.webhook.booking-canceled-post]; approval: not implemented: the runtime has no inbound webhook receiver executor for top-level webhook operations; risk: medium; notes: named_dependency=engine.webhook_receiver_executor: the runtime has no inbound webhook receiver executor for top-level webhook operations
  - api webhook booking-canceled-reschedule-requested-post - Documented WEBHOOK booking.canceled_reschedule_requested#POST (not implemented) [intent=docs_only availability=not_implemented operation=oncehub.webhook.booking-canceled-reschedule-requested-post]; approval: not implemented: the runtime has no inbound webhook receiver executor for top-level webhook operations; risk: medium; notes: named_dependency=engine.webhook_receiver_executor: the runtime has no inbound webhook receiver executor for top-level webhook operations
  - api webhook booking-canceled-then-rescheduled-post - Documented WEBHOOK booking.canceled_then_rescheduled#POST (not implemented) [intent=docs_only availability=not_implemented operation=oncehub.webhook.booking-canceled-then-rescheduled-post]; approval: not implemented: the runtime has no inbound webhook receiver executor for top-level webhook operations; risk: medium; notes: named_dependency=engine.webhook_receiver_executor: the runtime has no inbound webhook receiver executor for top-level webhook operations
  - api webhook booking-completed-post - Documented WEBHOOK booking.completed#POST (not implemented) [intent=docs_only availability=not_implemented operation=oncehub.webhook.booking-completed-post]; approval: not implemented: the runtime has no inbound webhook receiver executor for top-level webhook operations; risk: medium; notes: named_dependency=engine.webhook_receiver_executor: the runtime has no inbound webhook receiver executor for top-level webhook operations
  - api webhook booking-no-show-post - Documented WEBHOOK booking.no_show#POST (not implemented) [intent=docs_only availability=not_implemented operation=oncehub.webhook.booking-no-show-post]; approval: not implemented: the runtime has no inbound webhook receiver executor for top-level webhook operations; risk: medium; notes: named_dependency=engine.webhook_receiver_executor: the runtime has no inbound webhook receiver executor for top-level webhook operations
  - api webhook booking-reassigned-post - Documented WEBHOOK booking.reassigned#POST (not implemented) [intent=docs_only availability=not_implemented operation=oncehub.webhook.booking-reassigned-post]; approval: not implemented: the runtime has no inbound webhook receiver executor for top-level webhook operations; risk: medium; notes: named_dependency=engine.webhook_receiver_executor: the runtime has no inbound webhook receiver executor for top-level webhook operations
  - api webhook booking-rescheduled-post - Documented WEBHOOK booking.rescheduled#POST (not implemented) [intent=docs_only availability=not_implemented operation=oncehub.webhook.booking-rescheduled-post]; approval: not implemented: the runtime has no inbound webhook receiver executor for top-level webhook operations; risk: medium; notes: named_dependency=engine.webhook_receiver_executor: the runtime has no inbound webhook receiver executor for top-level webhook operations
  - api webhook booking-scheduled-post - Documented WEBHOOK booking.scheduled#POST (not implemented) [intent=docs_only availability=not_implemented operation=oncehub.webhook.booking-scheduled-post]; approval: not implemented: the runtime has no inbound webhook receiver executor for top-level webhook operations; risk: medium; notes: named_dependency=engine.webhook_receiver_executor: the runtime has no inbound webhook receiver executor for top-level webhook operations
  - api webhook conversation-abandoned-post - Documented WEBHOOK conversation.abandoned#POST (not implemented) [intent=docs_only availability=not_implemented operation=oncehub.webhook.conversation-abandoned-post]; approval: not implemented: the runtime has no inbound webhook receiver executor for top-level webhook operations; risk: medium; notes: named_dependency=engine.webhook_receiver_executor: the runtime has no inbound webhook receiver executor for top-level webhook operations
  - api webhook conversation-closed-post - Documented WEBHOOK conversation.closed#POST (not implemented) [intent=docs_only availability=not_implemented operation=oncehub.webhook.conversation-closed-post]; approval: not implemented: the runtime has no inbound webhook receiver executor for top-level webhook operations; risk: medium; notes: named_dependency=engine.webhook_receiver_executor: the runtime has no inbound webhook receiver executor for top-level webhook operations
  - api webhook conversation-started-post - Documented WEBHOOK conversation.started#POST (not implemented) [intent=docs_only availability=not_implemented operation=oncehub.webhook.conversation-started-post]; approval: not implemented: the runtime has no inbound webhook receiver executor for top-level webhook operations; risk: medium; notes: named_dependency=engine.webhook_receiver_executor: the runtime has no inbound webhook receiver executor for top-level webhook operations
  - booking pages list - Run the booking pages ETL stream [intent=etl availability=implemented stream=booking_pages]; notes: discrepancy=present-in-surface-absent-from-artifact
  - bookings list - Run the bookings ETL stream [intent=etl availability=implemented stream=bookings]; notes: discrepancy=present-in-surface-absent-from-artifact
  - contacts list - Run the contacts ETL stream [intent=etl availability=implemented stream=contacts]; notes: discrepancy=present-in-surface-absent-from-artifact
  - event types list - Run the event types ETL stream [intent=etl availability=implemented stream=event_types]; notes: discrepancy=present-in-surface-absent-from-artifact
  - users list - Run the users ETL stream [intent=etl availability=implemented stream=users]; notes: discrepancy=present-in-surface-absent-from-artifact

## Commands

### Inspect as a manual

```bash
pm connectors inspect oncehub
```

### Inspect as structured JSON

```bash
pm connectors inspect oncehub --json
```

## Agent Rules

- Run pm connectors inspect oncehub before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
