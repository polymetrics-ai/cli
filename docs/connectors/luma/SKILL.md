---
name: pm-luma
description: Luma connector knowledge and safe action guide.
---

# pm-luma

## Purpose

Reads and writes the documented Luma public API for events, calendars, guests, contacts, tags, coupons, ticket types, memberships, webhooks, and organization resources.

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
- event_api_id
- event_id
- event_ticket_type_id
- guest_id
- mode
- slug
- webhook_id
- api_key (secret)

## ETL Streams

- events:
  - primary key: api_id
  - fields: api_id(string), calendar_api_id(string), cover_url(string), created_at(string), description(string), end_at(string), name(string), start_at(string), timezone(string), url(string), visibility(string)
- event_guests:
  - primary key: api_id
  - fields: api_id(string), approval_status(string), checked_in_at(string), email(string), event_api_id(string), name(string), registered_at(string), user_api_id(string), user_name(string)
- event_hosts:
  - primary key: api_id
  - fields: access_level(string), api_id(string), avatar_url(string), email(string), name(string)
- event:
  - primary key: id
  - fields: access(string), calendar_id(string), coordinate(object), cover_url(string), created_at(string), description(string), description_md(string), duration_interval(string), end_at(string), feedback_email(object), geo_address_json(object), guest_counts(object), hosts(array), id(string), location_type(string), location_visibility(string), meeting_url(string), name(string), platform(string), registration_open(boolean), registration_questions(array), start_at(string), timezone(string), url(string), user_id(string), visibility(string), waitlist_status(string)
- calendar:
  - primary key: id
  - fields: avatar_url(string), coordinate(object), cover_image_url(string), description(string), id(string), instagram_handle(string), is_personal(boolean), location(object), name(string), slug(string), social_image_url(string), twitter_handle(string), url(string), website(string), youtube_handle(string)
- calendar_events:
  - primary key: id
  - fields: calendar_id(string), coordinate(object), created_at(string), duration_interval(string), end_at(string), geo_address_json(object), host(string), id(string), name(string), platform(string), start_at(string), tags(array), timezone(string), url(string)
- guest:
  - primary key: id
  - fields: approval_status(string), check_in_qr_code(string), eth_address(string), event_ticket_orders(array), event_tickets(array), id(string), invited_at(string), joined_at(string), phone_number(string), registered_at(string), registration_answers(array), solana_address(string), user_email(string), user_first_name(string), user_id(string), user_last_name(string), user_name(string), utm_source(string)
- guests:
  - primary key: id
  - fields: approval_status(string), check_in_qr_code(string), eth_address(string), event_tickets(array), id(string), invited_at(string), joined_at(string), phone_number(string), registered_at(string), registration_answers(array), solana_address(string), user_email(string), user_first_name(string), user_id(string), user_last_name(string), user_name(string), utm_source(string)
- self_user:
  - primary key: id
  - fields: avatar_url(string), email(string), first_name(string), id(string), last_name(string), name(string)
- contact_tags:
  - primary key: id
  - fields: color(string), id(string), name(string)
- event_tags:
  - primary key: id
  - fields: color(string), id(string), name(string)
- calendar_admins:
  - primary key: id
  - fields: avatar_url(string), email(string), first_name(string), id(string), last_name(string), name(string)
- entity_lookup:
  - fields: calendar(object), event(object), type(string)
- event_lookup:
  - primary key: id
  - fields: id(string), status(string)
- calendar_contacts:
  - primary key: id
  - fields: avatar_url(string), created_at(string), email(string), event_approved_count(number), event_checked_in_count(number), first_name(string), id(string), last_name(string), membership(object), name(string), revenue_usd_cents(number), tags(array), user_id(string)
- event_coupons:
  - primary key: id
  - fields: cents_off(number), code(string), currency(string), event_ticket_type_id(string), id(string), percent_off(number), remaining_count(integer), valid_end_at(string), valid_start_at(string)
- calendar_coupons:
  - primary key: id
  - fields: cents_off(number), code(string), currency(string), event_ticket_type_id(string), id(string), percent_off(number), remaining_count(integer), valid_end_at(string), valid_start_at(string)
- event_ticket_types:
  - primary key: id
  - fields: cents(number), currency(string), description(string), id(string), is_flexible(boolean), is_hidden(boolean), max_capacity(number), min_cents(number), name(string), require_approval(boolean), type(string), valid_end_at(string), valid_start_at(string)
- event_ticket_type:
  - primary key: id
  - fields: cents(number), currency(string), description(string), id(string), is_flexible(boolean), is_hidden(boolean), max_capacity(number), min_cents(number), name(string), require_approval(boolean), type(string), valid_end_at(string), valid_start_at(string)
- membership_tiers:
  - primary key: id
  - fields: access_info(object), description(string), id(string), name(string), tint_color(string)
- webhooks:
  - primary key: id
  - fields: created_at(string), event_types(array), id(string), secret(string), status(string), url(string)
- webhook:
  - primary key: id
  - fields: created_at(string), event_types(array), id(string), secret(string), status(string), url(string)
- organization_admins:
  - primary key: id
  - fields: api_id(string), avatar_url(string), email(string), first_name(string), id(string), last_name(string), name(string)
- organization_calendars:
  - primary key: id
  - fields: avatar_url(string), coordinate(object), cover_image_url(string), description(string), id(string), instagram_handle(string), is_personal(boolean), location(object), name(string), slug(string), social_image_url(string), twitter_handle(string), url(string), website(string), youtube_handle(string)
- organization_events:
  - primary key: id
  - fields: api_id(string), calendar_api_id(string), calendar_id(string), coordinate(object), cover_url(string), created_at(string), description(string), description_md(string), duration_interval(string), end_at(string), feedback_email(object), geo_address_json(object), geo_latitude(string), geo_longitude(string), id(string), location_type(string), location_visibility(string), managing_calendars(array), meeting_url(string), name(string), platform(string), registration_open(boolean), registration_questions(array), start_at(string), timezone(string), url(string), user_api_id(string), user_id(string), visibility(string), waitlist_status(string), zoom_meeting_url(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_event:
  - endpoint: POST /v1/events/create
  - required fields: name, start_at, timezone
  - risk: create event through the Luma API.
- update_event:
  - endpoint: POST /v1/events/update
  - required fields: event_id
  - risk: update event through the Luma API.
- update_guest_status:
  - endpoint: POST /v1/events/guests/update-status
  - required fields: event_id, guest_id, status
  - risk: update guest status through the Luma API.
- send_event_invites:
  - endpoint: POST /v1/events/guests/send-invites
  - required fields: event_id, guests
  - risk: send event invites through the Luma API.
- add_event_guests:
  - endpoint: POST /v1/events/guests/add
  - required fields: event_id, guests
  - risk: add event guests through the Luma API.
- add_event_host:
  - endpoint: POST /v1/events/hosts/add
  - required fields: event_id, email
  - risk: add event host through the Luma API.
- update_event_host:
  - endpoint: POST /v1/events/hosts/update
  - required fields: event_id, email
  - risk: update event host through the Luma API.
- remove_event_host:
  - endpoint: POST /v1/events/hosts/remove
  - required fields: event_id, email
  - risk: remove event host through the Luma API.
- create_event_coupon:
  - endpoint: POST /v1/events/coupons/create
  - required fields: code, discount, event_id
  - risk: create event coupon through the Luma API.
- update_event_coupon:
  - endpoint: POST /v1/events/coupons/update
  - required fields: event_id, code
  - risk: update event coupon through the Luma API.
- create_calendar_coupon:
  - endpoint: POST /v1/calendars/coupons/create
  - required fields: code, discount
  - risk: create calendar coupon through the Luma API.
- update_calendar_coupon:
  - endpoint: POST /v1/calendars/coupons/update
  - required fields: code
  - risk: update calendar coupon through the Luma API.
- import_calendar_contacts:
  - endpoint: POST /v1/calendars/contacts/import
  - required fields: contacts
  - risk: import calendar contacts through the Luma API.
- create_contact_tag:
  - endpoint: POST /v1/calendars/contact-tags/create
  - required fields: name
  - risk: create contact tag through the Luma API.
- update_contact_tag:
  - endpoint: POST /v1/calendars/contact-tags/update
  - required fields: tag_id
  - risk: update contact tag through the Luma API.
- delete_contact_tag:
  - endpoint: POST /v1/calendars/contact-tags/delete
  - required fields: tag_id
  - risk: delete contact tag through the Luma API.
- apply_contact_tag:
  - endpoint: POST /v1/calendars/contact-tags/apply
  - required fields: tag
  - risk: apply contact tag through the Luma API.
- unapply_contact_tag:
  - endpoint: POST /v1/calendars/contact-tags/unapply
  - required fields: tag
  - risk: unapply contact tag through the Luma API.
- create_event_tag:
  - endpoint: POST /v1/calendars/event-tags/create
  - required fields: name
  - risk: create event tag through the Luma API.
- update_event_tag:
  - endpoint: POST /v1/calendars/event-tags/update
  - required fields: tag_id
  - risk: update event tag through the Luma API.
- delete_event_tag:
  - endpoint: POST /v1/calendars/event-tags/delete
  - required fields: tag_id
  - risk: delete event tag through the Luma API.
- apply_event_tag:
  - endpoint: POST /v1/calendars/event-tags/apply
  - required fields: tag, event_ids
  - risk: apply event tag through the Luma API.
- unapply_event_tag:
  - endpoint: POST /v1/calendars/event-tags/unapply
  - required fields: tag, event_ids
  - risk: unapply event tag through the Luma API.
- add_calendar_event:
  - endpoint: POST /v1/calendars/events/add
  - risk: add calendar event through the Luma API.
- approve_calendar_event:
  - endpoint: POST /v1/calendars/events/approve
  - required fields: calendar_event_id
  - risk: approve calendar event through the Luma API.
- reject_calendar_event:
  - endpoint: POST /v1/calendars/events/reject
  - required fields: calendar_event_id
  - risk: reject calendar event through the Luma API.
- create_image_upload_url:
  - endpoint: POST /v1/images/create-upload-url
  - risk: create image upload url through the Luma API.
- create_ticket_type:
  - endpoint: POST /v1/events/ticket-types/create
  - required fields: event_id, name, type
  - risk: create ticket type through the Luma API.
- update_ticket_type:
  - endpoint: POST /v1/events/ticket-types/update
  - required fields: event_ticket_type_id
  - risk: update ticket type through the Luma API.
- delete_ticket_type:
  - endpoint: POST /v1/events/ticket-types/delete
  - required fields: event_ticket_type_id
  - risk: delete ticket type through the Luma API.
- add_membership_member:
  - endpoint: POST /v1/memberships/members/add
  - required fields: email, membership_tier_id
  - risk: add membership member through the Luma API.
- update_membership_member_status:
  - endpoint: POST /v1/memberships/members/update-status
  - required fields: user_id, status
  - risk: update membership member status through the Luma API.
- create_webhook:
  - endpoint: POST /v2/webhooks/create
  - required fields: url, event_types
  - risk: create webhook through the Luma API.
- update_webhook:
  - endpoint: POST /v2/webhooks/update
  - required fields: id
  - risk: update webhook through the Luma API.
- delete_webhook:
  - endpoint: POST /v1/webhooks/delete
  - required fields: id
  - risk: delete webhook through the Luma API.
- request_event_cancellation:
  - endpoint: POST /v1/events/cancel/request
  - required fields: event_id
  - risk: request event cancellation through the Luma API.
- cancel_event:
  - endpoint: POST /v1/events/cancel
  - required fields: event_id, cancellation_token
  - risk: cancel event through the Luma API.
- update_calendar:
  - endpoint: POST /v1/calendars/update
  - required fields: calendar_id
  - risk: update calendar through the Luma API.
- create_organization_calendar:
  - endpoint: POST /v2/organizations/calendars/create
  - required fields: name
  - risk: create organization calendar through the Luma API.
- transfer_event_calendar:
  - endpoint: POST /v1/organizations/events/transfer-calendar
  - required fields: event_id, calendar_id
  - risk: transfer event calendar through the Luma API.

## Security

- read risk: external Luma public API read of calendar, event, guest, contact, tag, coupon, ticket, membership, webhook, and organization data
- write risk: live Luma API mutations can create, update, invite, tag, cancel, transfer, or delete event/calendar/member/webhook data and may send guest invitations
- approval: reverse ETL writes require plan, preview, approval token, and destructive confirmation for delete/cancel/invite/transfer operations
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run Luma's declared streams and reverse-ETL actions.
- Usage: pm luma <command> [flags]
- Read streams
- Reverse ETL writes
- Other Commands
  - add calendar event apply - Plan and execute the add calendar event reverse-ETL action [intent=reverse_etl availability=implemented write=add_calendar_event]; approval: requires plan, preview, approval, and execute; risk: add calendar event through the Luma API.
  - add event guests apply - Plan and execute the add event guests reverse-ETL action [intent=reverse_etl availability=not_implemented write=add_event_guests]; approval: requires plan, preview, approval, and execute; risk: add event guests through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - add event host apply - Plan and execute the add event host reverse-ETL action [intent=reverse_etl availability=not_implemented write=add_event_host]; approval: requires plan, preview, approval, and execute; risk: add event host through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - add membership member apply - Plan and execute the add membership member reverse-ETL action [intent=reverse_etl availability=not_implemented write=add_membership_member]; approval: requires plan, preview, approval, and execute; risk: add membership member through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - api post v1 calendars contacts block - Documented POST /v1/calendars/contacts/block (not implemented) [intent=direct_write availability=not_implemented operation=luma.post.v1-calendars-contacts-block]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 calendars contacts remove - Documented POST /v1/calendars/contacts/remove (not implemented) [intent=direct_write availability=not_implemented operation=luma.post.v1-calendars-contacts-remove]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 calendars contacts restore - Documented POST /v1/calendars/contacts/restore (not implemented) [intent=direct_write availability=not_implemented operation=luma.post.v1-calendars-contacts-restore]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 events guests update-tickets - Documented POST /v1/events/guests/update-tickets (not implemented) [intent=direct_write availability=not_implemented operation=luma.post.v1-events-guests-update-tickets]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api webhook calendar-event-added-post - Documented WEBHOOK calendar.event.added#POST (not implemented) [intent=docs_only availability=not_implemented operation=luma.webhook.calendar-event-added-post]; approval: not implemented: the runtime has no inbound webhook receiver executor for top-level webhook operations; risk: medium; notes: named_dependency=engine.webhook_receiver_executor: the runtime has no inbound webhook receiver executor for top-level webhook operations
  - api webhook calendar-event-submitted-post - Documented WEBHOOK calendar.event.submitted#POST (not implemented) [intent=docs_only availability=not_implemented operation=luma.webhook.calendar-event-submitted-post]; approval: not implemented: the runtime has no inbound webhook receiver executor for top-level webhook operations; risk: medium; notes: named_dependency=engine.webhook_receiver_executor: the runtime has no inbound webhook receiver executor for top-level webhook operations
  - api webhook calendar-person-subscribed-post - Documented WEBHOOK calendar.person.subscribed#POST (not implemented) [intent=docs_only availability=not_implemented operation=luma.webhook.calendar-person-subscribed-post]; approval: not implemented: the runtime has no inbound webhook receiver executor for top-level webhook operations; risk: medium; notes: named_dependency=engine.webhook_receiver_executor: the runtime has no inbound webhook receiver executor for top-level webhook operations
  - api webhook event-canceled-post - Documented WEBHOOK event.canceled#POST (not implemented) [intent=docs_only availability=not_implemented operation=luma.webhook.event-canceled-post]; approval: not implemented: the runtime has no inbound webhook receiver executor for top-level webhook operations; risk: medium; notes: named_dependency=engine.webhook_receiver_executor: the runtime has no inbound webhook receiver executor for top-level webhook operations
  - api webhook event-created-post - Documented WEBHOOK event.created#POST (not implemented) [intent=docs_only availability=not_implemented operation=luma.webhook.event-created-post]; approval: not implemented: the runtime has no inbound webhook receiver executor for top-level webhook operations; risk: medium; notes: named_dependency=engine.webhook_receiver_executor: the runtime has no inbound webhook receiver executor for top-level webhook operations
  - api webhook event-updated-post - Documented WEBHOOK event.updated#POST (not implemented) [intent=docs_only availability=not_implemented operation=luma.webhook.event-updated-post]; approval: not implemented: the runtime has no inbound webhook receiver executor for top-level webhook operations; risk: medium; notes: named_dependency=engine.webhook_receiver_executor: the runtime has no inbound webhook receiver executor for top-level webhook operations
  - api webhook guest-registered-post - Documented WEBHOOK guest.registered#POST (not implemented) [intent=docs_only availability=not_implemented operation=luma.webhook.guest-registered-post]; approval: not implemented: the runtime has no inbound webhook receiver executor for top-level webhook operations; risk: medium; notes: named_dependency=engine.webhook_receiver_executor: the runtime has no inbound webhook receiver executor for top-level webhook operations
  - api webhook guest-updated-post - Documented WEBHOOK guest.updated#POST (not implemented) [intent=docs_only availability=not_implemented operation=luma.webhook.guest-updated-post]; approval: not implemented: the runtime has no inbound webhook receiver executor for top-level webhook operations; risk: medium; notes: named_dependency=engine.webhook_receiver_executor: the runtime has no inbound webhook receiver executor for top-level webhook operations
  - api webhook ticket-registered-post - Documented WEBHOOK ticket.registered#POST (not implemented) [intent=docs_only availability=not_implemented operation=luma.webhook.ticket-registered-post]; approval: not implemented: the runtime has no inbound webhook receiver executor for top-level webhook operations; risk: medium; notes: named_dependency=engine.webhook_receiver_executor: the runtime has no inbound webhook receiver executor for top-level webhook operations
  - apply contact tag apply - Plan and execute the apply contact tag reverse-ETL action [intent=reverse_etl availability=not_implemented write=apply_contact_tag]; approval: requires plan, preview, approval, and execute; risk: apply contact tag through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - apply event tag apply - Plan and execute the apply event tag reverse-ETL action [intent=reverse_etl availability=not_implemented write=apply_event_tag]; approval: requires plan, preview, approval, and execute; risk: apply event tag through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - approve calendar event apply - Plan and execute the approve calendar event reverse-ETL action [intent=reverse_etl availability=not_implemented write=approve_calendar_event]; approval: requires plan, preview, approval, and execute; risk: approve calendar event through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - calendar admins list - Run the calendar admins ETL stream [intent=etl availability=implemented stream=calendar_admins]
  - calendar contacts list - Run the calendar contacts ETL stream [intent=etl availability=implemented stream=calendar_contacts]
  - calendar coupons list - Run the calendar coupons ETL stream [intent=etl availability=implemented stream=calendar_coupons]
  - calendar events list - Run the calendar events ETL stream [intent=etl availability=implemented stream=calendar_events]
  - calendar list - Run the calendar ETL stream [intent=etl availability=implemented stream=calendar]
  - cancel event apply - Plan and execute the cancel event reverse-ETL action [intent=reverse_etl availability=not_implemented write=cancel_event]; approval: requires plan, preview, approval, and execute; risk: cancel event through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - contact tags list - Run the contact tags ETL stream [intent=etl availability=implemented stream=contact_tags]
  - create calendar coupon apply - Plan and execute the create calendar coupon reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_calendar_coupon]; approval: requires plan, preview, approval, and execute; risk: create calendar coupon through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - create contact tag apply - Plan and execute the create contact tag reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_contact_tag]; approval: requires plan, preview, approval, and execute; risk: create contact tag through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - create event apply - Plan and execute the create event reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_event]; approval: requires plan, preview, approval, and execute; risk: create event through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - create event coupon apply - Plan and execute the create event coupon reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_event_coupon]; approval: requires plan, preview, approval, and execute; risk: create event coupon through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - create event tag apply - Plan and execute the create event tag reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_event_tag]; approval: requires plan, preview, approval, and execute; risk: create event tag through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - create image upload url apply - Plan and execute the create image upload url reverse-ETL action [intent=reverse_etl availability=implemented write=create_image_upload_url]; approval: requires plan, preview, approval, and execute; risk: create image upload url through the Luma API.
  - create organization calendar apply - Plan and execute the create organization calendar reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_organization_calendar]; approval: requires plan, preview, approval, and execute; risk: create organization calendar through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - create ticket type apply - Plan and execute the create ticket type reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_ticket_type]; approval: requires plan, preview, approval, and execute; risk: create ticket type through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - create webhook apply - Plan and execute the create webhook reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_webhook]; approval: requires plan, preview, approval, and execute; risk: create webhook through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete contact tag apply - Plan and execute the delete contact tag reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_contact_tag]; approval: requires plan, preview, approval, and execute; risk: delete contact tag through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete event tag apply - Plan and execute the delete event tag reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_event_tag]; approval: requires plan, preview, approval, and execute; risk: delete event tag through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete ticket type apply - Plan and execute the delete ticket type reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_ticket_type]; approval: requires plan, preview, approval, and execute; risk: delete ticket type through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete webhook apply - Plan and execute the delete webhook reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_webhook]; approval: requires plan, preview, approval, and execute; risk: delete webhook through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - entity lookup list - Run the entity lookup ETL stream [intent=etl availability=implemented stream=entity_lookup]
  - event coupons list - Run the event coupons ETL stream [intent=etl availability=implemented stream=event_coupons]
  - event guests list - Run the event guests ETL stream [intent=etl availability=implemented stream=event_guests]
  - event hosts list - Run the event hosts ETL stream [intent=etl availability=implemented stream=event_hosts]
  - event list - Run the event ETL stream [intent=etl availability=implemented stream=event]
  - event lookup list - Run the event lookup ETL stream [intent=etl availability=implemented stream=event_lookup]
  - event tags list - Run the event tags ETL stream [intent=etl availability=implemented stream=event_tags]
  - event ticket type list - Run the event ticket type ETL stream [intent=etl availability=implemented stream=event_ticket_type]
  - event ticket types list - Run the event ticket types ETL stream [intent=etl availability=implemented stream=event_ticket_types]
  - events list - Run the events ETL stream [intent=etl availability=implemented stream=events]
  - guest list - Run the guest ETL stream [intent=etl availability=implemented stream=guest]
  - guests list - Run the guests ETL stream [intent=etl availability=implemented stream=guests]
  - import calendar contacts apply - Plan and execute the import calendar contacts reverse-ETL action [intent=reverse_etl availability=not_implemented write=import_calendar_contacts]; approval: requires plan, preview, approval, and execute; risk: import calendar contacts through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - membership tiers list - Run the membership tiers ETL stream [intent=etl availability=implemented stream=membership_tiers]
  - organization admins list - Run the organization admins ETL stream [intent=etl availability=implemented stream=organization_admins]
  - organization calendars list - Run the organization calendars ETL stream [intent=etl availability=implemented stream=organization_calendars]
  - organization events list - Run the organization events ETL stream [intent=etl availability=implemented stream=organization_events]
  - reject calendar event apply - Plan and execute the reject calendar event reverse-ETL action [intent=reverse_etl availability=not_implemented write=reject_calendar_event]; approval: requires plan, preview, approval, and execute; risk: reject calendar event through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - remove event host apply - Plan and execute the remove event host reverse-ETL action [intent=reverse_etl availability=not_implemented write=remove_event_host]; approval: requires plan, preview, approval, and execute; risk: remove event host through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - request event cancellation apply - Plan and execute the request event cancellation reverse-ETL action [intent=reverse_etl availability=not_implemented write=request_event_cancellation]; approval: requires plan, preview, approval, and execute; risk: request event cancellation through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - self user list - Run the self user ETL stream [intent=etl availability=implemented stream=self_user]
  - send event invites apply - Plan and execute the send event invites reverse-ETL action [intent=reverse_etl availability=not_implemented write=send_event_invites]; approval: requires plan, preview, approval, and execute; risk: send event invites through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - transfer event calendar apply - Plan and execute the transfer event calendar reverse-ETL action [intent=reverse_etl availability=not_implemented write=transfer_event_calendar]; approval: requires plan, preview, approval, and execute; risk: transfer event calendar through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - unapply contact tag apply - Plan and execute the unapply contact tag reverse-ETL action [intent=reverse_etl availability=not_implemented write=unapply_contact_tag]; approval: requires plan, preview, approval, and execute; risk: unapply contact tag through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - unapply event tag apply - Plan and execute the unapply event tag reverse-ETL action [intent=reverse_etl availability=not_implemented write=unapply_event_tag]; approval: requires plan, preview, approval, and execute; risk: unapply event tag through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - update calendar apply - Plan and execute the update calendar reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_calendar]; approval: requires plan, preview, approval, and execute; risk: update calendar through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - update calendar coupon apply - Plan and execute the update calendar coupon reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_calendar_coupon]; approval: requires plan, preview, approval, and execute; risk: update calendar coupon through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - update contact tag apply - Plan and execute the update contact tag reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_contact_tag]; approval: requires plan, preview, approval, and execute; risk: update contact tag through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - update event apply - Plan and execute the update event reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_event]; approval: requires plan, preview, approval, and execute; risk: update event through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - update event coupon apply - Plan and execute the update event coupon reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_event_coupon]; approval: requires plan, preview, approval, and execute; risk: update event coupon through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - update event host apply - Plan and execute the update event host reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_event_host]; approval: requires plan, preview, approval, and execute; risk: update event host through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - update event tag apply - Plan and execute the update event tag reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_event_tag]; approval: requires plan, preview, approval, and execute; risk: update event tag through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - update guest status apply - Plan and execute the update guest status reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_guest_status]; approval: requires plan, preview, approval, and execute; risk: update guest status through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - update membership member status apply - Plan and execute the update membership member status reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_membership_member_status]; approval: requires plan, preview, approval, and execute; risk: update membership member status through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - update ticket type apply - Plan and execute the update ticket type reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_ticket_type]; approval: requires plan, preview, approval, and execute; risk: update ticket type through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - update webhook apply - Plan and execute the update webhook reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_webhook]; approval: requires plan, preview, approval, and execute; risk: update webhook through the Luma API.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - webhook list - Run the webhook ETL stream [intent=etl availability=implemented stream=webhook]
  - webhooks list - Run the webhooks ETL stream [intent=etl availability=implemented stream=webhooks]

## Commands

### Inspect as a manual

```bash
pm connectors inspect luma
```

### Inspect as structured JSON

```bash
pm connectors inspect luma --json
```

## Agent Rules

- Run pm connectors inspect luma before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
