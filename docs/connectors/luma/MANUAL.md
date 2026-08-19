# pm connectors inspect luma

```text
NAME
  pm connectors inspect luma - Luma connector manual

SYNOPSIS
  pm connectors inspect luma
  pm connectors inspect luma --json
  pm credentials add <name> --connector luma [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes the documented Luma public API for events, calendars, guests, contacts, tags, coupons, ticket types, memberships, webhooks, and organization resources.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  event_api_id
  event_id
  event_ticket_type_id
  guest_id
  mode
  slug
  webhook_id
  api_key (secret) (required)

ETL STREAMS
  events:
    primary key: api_id
    fields: api_id(string), calendar_api_id(string), cover_url(string), created_at(string), description(string), end_at(string), name(string), start_at(string), timezone(string), url(string), visibility(string)
  event_guests:
    primary key: api_id
    fields: api_id(string), approval_status(string), checked_in_at(string), email(string), event_api_id(string), name(string), registered_at(string), user_api_id(string), user_name(string)
  event_hosts:
    primary key: api_id
    fields: access_level(string), api_id(string), avatar_url(string), email(string), name(string)
  event:
    primary key: id
    fields: access(string), calendar_id(string), coordinate(object), cover_url(string), created_at(string), description(string), description_md(string), duration_interval(string), end_at(string), feedback_email(object), geo_address_json(object), guest_counts(object), hosts(array), id(string), location_type(string), location_visibility(string), meeting_url(string), name(string), platform(string), registration_open(boolean), registration_questions(array), start_at(string), timezone(string), url(string), user_id(string), visibility(string), waitlist_status(string)
  calendar:
    primary key: id
    fields: avatar_url(string), coordinate(object), cover_image_url(string), description(string), id(string), instagram_handle(string), is_personal(boolean), location(object), name(string), slug(string), social_image_url(string), twitter_handle(string), url(string), website(string), youtube_handle(string)
  calendar_events:
    primary key: id
    fields: calendar_id(string), coordinate(object), created_at(string), duration_interval(string), end_at(string), geo_address_json(object), host(string), id(string), name(string), platform(string), start_at(string), tags(array), timezone(string), url(string)
  guest:
    primary key: id
    fields: approval_status(string), check_in_qr_code(string), eth_address(string), event_ticket_orders(array), event_tickets(array), id(string), invited_at(string), joined_at(string), phone_number(string), registered_at(string), registration_answers(array), solana_address(string), user_email(string), user_first_name(string), user_id(string), user_last_name(string), user_name(string), utm_source(string)
  guests:
    primary key: id
    fields: approval_status(string), check_in_qr_code(string), eth_address(string), event_tickets(array), id(string), invited_at(string), joined_at(string), phone_number(string), registered_at(string), registration_answers(array), solana_address(string), user_email(string), user_first_name(string), user_id(string), user_last_name(string), user_name(string), utm_source(string)
  self_user:
    primary key: id
    fields: avatar_url(string), email(string), first_name(string), id(string), last_name(string), name(string)
  contact_tags:
    primary key: id
    fields: color(string), id(string), name(string)
  event_tags:
    primary key: id
    fields: color(string), id(string), name(string)
  calendar_admins:
    primary key: id
    fields: avatar_url(string), email(string), first_name(string), id(string), last_name(string), name(string)
  entity_lookup:
    fields: calendar(object), event(object), type(string)
  event_lookup:
    primary key: id
    fields: id(string), status(string)
  calendar_contacts:
    primary key: id
    fields: avatar_url(string), created_at(string), email(string), event_approved_count(number), event_checked_in_count(number), first_name(string), id(string), last_name(string), membership(object), name(string), revenue_usd_cents(number), tags(array), user_id(string)
  event_coupons:
    primary key: id
    fields: cents_off(number), code(string), currency(string), event_ticket_type_id(string), id(string), percent_off(number), remaining_count(integer), valid_end_at(string), valid_start_at(string)
  calendar_coupons:
    primary key: id
    fields: cents_off(number), code(string), currency(string), event_ticket_type_id(string), id(string), percent_off(number), remaining_count(integer), valid_end_at(string), valid_start_at(string)
  event_ticket_types:
    primary key: id
    fields: cents(number), currency(string), description(string), id(string), is_flexible(boolean), is_hidden(boolean), max_capacity(number), min_cents(number), name(string), require_approval(boolean), type(string), valid_end_at(string), valid_start_at(string)
  event_ticket_type:
    primary key: id
    fields: cents(number), currency(string), description(string), id(string), is_flexible(boolean), is_hidden(boolean), max_capacity(number), min_cents(number), name(string), require_approval(boolean), type(string), valid_end_at(string), valid_start_at(string)
  membership_tiers:
    primary key: id
    fields: access_info(object), description(string), id(string), name(string), tint_color(string)
  webhooks:
    primary key: id
    fields: created_at(string), event_types(array), id(string), secret(string), status(string), url(string)
  webhook:
    primary key: id
    fields: created_at(string), event_types(array), id(string), secret(string), status(string), url(string)
  organization_admins:
    primary key: id
    fields: api_id(string), avatar_url(string), email(string), first_name(string), id(string), last_name(string), name(string)
  organization_calendars:
    primary key: id
    fields: avatar_url(string), coordinate(object), cover_image_url(string), description(string), id(string), instagram_handle(string), is_personal(boolean), location(object), name(string), slug(string), social_image_url(string), twitter_handle(string), url(string), website(string), youtube_handle(string)
  organization_events:
    primary key: id
    fields: api_id(string), calendar_api_id(string), calendar_id(string), coordinate(object), cover_url(string), created_at(string), description(string), description_md(string), duration_interval(string), end_at(string), feedback_email(object), geo_address_json(object), geo_latitude(string), geo_longitude(string), id(string), location_type(string), location_visibility(string), managing_calendars(array), meeting_url(string), name(string), platform(string), registration_open(boolean), registration_questions(array), start_at(string), timezone(string), url(string), user_api_id(string), user_id(string), visibility(string), waitlist_status(string), zoom_meeting_url(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

REVERSE ETL ACTIONS
  create_event:
    endpoint: POST /v1/events/create
    required fields: name, start_at, timezone
    risk: create event through the Luma API.
  update_event:
    endpoint: POST /v1/events/update
    required fields: event_id
    risk: update event through the Luma API.
  update_guest_status:
    endpoint: POST /v1/events/guests/update-status
    required fields: event_id, guest_id, status
    risk: update guest status through the Luma API.
  send_event_invites:
    endpoint: POST /v1/events/guests/send-invites
    required fields: event_id, guests
    risk: send event invites through the Luma API.
  add_event_guests:
    endpoint: POST /v1/events/guests/add
    required fields: event_id, guests
    risk: add event guests through the Luma API.
  add_event_host:
    endpoint: POST /v1/events/hosts/add
    required fields: event_id, email
    risk: add event host through the Luma API.
  update_event_host:
    endpoint: POST /v1/events/hosts/update
    required fields: event_id, email
    risk: update event host through the Luma API.
  remove_event_host:
    endpoint: POST /v1/events/hosts/remove
    required fields: event_id, email
    risk: remove event host through the Luma API.
  create_event_coupon:
    endpoint: POST /v1/events/coupons/create
    required fields: code, discount, event_id
    risk: create event coupon through the Luma API.
  update_event_coupon:
    endpoint: POST /v1/events/coupons/update
    required fields: event_id, code
    risk: update event coupon through the Luma API.
  create_calendar_coupon:
    endpoint: POST /v1/calendars/coupons/create
    required fields: code, discount
    risk: create calendar coupon through the Luma API.
  update_calendar_coupon:
    endpoint: POST /v1/calendars/coupons/update
    required fields: code
    risk: update calendar coupon through the Luma API.
  import_calendar_contacts:
    endpoint: POST /v1/calendars/contacts/import
    required fields: contacts
    risk: import calendar contacts through the Luma API.
  create_contact_tag:
    endpoint: POST /v1/calendars/contact-tags/create
    required fields: name
    risk: create contact tag through the Luma API.
  update_contact_tag:
    endpoint: POST /v1/calendars/contact-tags/update
    required fields: tag_id
    risk: update contact tag through the Luma API.
  delete_contact_tag:
    endpoint: POST /v1/calendars/contact-tags/delete
    required fields: tag_id
    risk: delete contact tag through the Luma API.
  apply_contact_tag:
    endpoint: POST /v1/calendars/contact-tags/apply
    required fields: tag
    risk: apply contact tag through the Luma API.
  unapply_contact_tag:
    endpoint: POST /v1/calendars/contact-tags/unapply
    required fields: tag
    risk: unapply contact tag through the Luma API.
  create_event_tag:
    endpoint: POST /v1/calendars/event-tags/create
    required fields: name
    risk: create event tag through the Luma API.
  update_event_tag:
    endpoint: POST /v1/calendars/event-tags/update
    required fields: tag_id
    risk: update event tag through the Luma API.
  delete_event_tag:
    endpoint: POST /v1/calendars/event-tags/delete
    required fields: tag_id
    risk: delete event tag through the Luma API.
  apply_event_tag:
    endpoint: POST /v1/calendars/event-tags/apply
    required fields: tag, event_ids
    risk: apply event tag through the Luma API.
  unapply_event_tag:
    endpoint: POST /v1/calendars/event-tags/unapply
    required fields: tag, event_ids
    risk: unapply event tag through the Luma API.
  add_calendar_event:
    endpoint: POST /v1/calendars/events/add
    risk: add calendar event through the Luma API.
  approve_calendar_event:
    endpoint: POST /v1/calendars/events/approve
    required fields: calendar_event_id
    risk: approve calendar event through the Luma API.
  reject_calendar_event:
    endpoint: POST /v1/calendars/events/reject
    required fields: calendar_event_id
    risk: reject calendar event through the Luma API.
  create_image_upload_url:
    endpoint: POST /v1/images/create-upload-url
    risk: create image upload url through the Luma API.
  create_ticket_type:
    endpoint: POST /v1/events/ticket-types/create
    required fields: event_id, name, type
    risk: create ticket type through the Luma API.
  update_ticket_type:
    endpoint: POST /v1/events/ticket-types/update
    required fields: event_ticket_type_id
    risk: update ticket type through the Luma API.
  delete_ticket_type:
    endpoint: POST /v1/events/ticket-types/delete
    required fields: event_ticket_type_id
    risk: delete ticket type through the Luma API.
  add_membership_member:
    endpoint: POST /v1/memberships/members/add
    required fields: email, membership_tier_id
    risk: add membership member through the Luma API.
  update_membership_member_status:
    endpoint: POST /v1/memberships/members/update-status
    required fields: user_id, status
    risk: update membership member status through the Luma API.
  create_webhook:
    endpoint: POST /v2/webhooks/create
    required fields: url, event_types
    risk: create webhook through the Luma API.
  update_webhook:
    endpoint: POST /v2/webhooks/update
    required fields: id
    risk: update webhook through the Luma API.
  delete_webhook:
    endpoint: POST /v1/webhooks/delete
    required fields: id
    risk: delete webhook through the Luma API.
  request_event_cancellation:
    endpoint: POST /v1/events/cancel/request
    required fields: event_id
    risk: request event cancellation through the Luma API.
  cancel_event:
    endpoint: POST /v1/events/cancel
    required fields: event_id, cancellation_token
    risk: cancel event through the Luma API.
  update_calendar:
    endpoint: POST /v1/calendars/update
    required fields: calendar_id
    risk: update calendar through the Luma API.
  create_organization_calendar:
    endpoint: POST /v2/organizations/calendars/create
    required fields: name
    risk: create organization calendar through the Luma API.
  transfer_event_calendar:
    endpoint: POST /v1/organizations/events/transfer-calendar
    required fields: event_id, calendar_id
    risk: transfer event calendar through the Luma API.

SECURITY
  read risk: external Luma public API read of calendar, event, guest, contact, tag, coupon, ticket, membership, webhook, and organization data
  write risk: live Luma API mutations can create, update, invite, tag, cancel, transfer, or delete event/calendar/member/webhook data and may send guest invitations
  approval: reverse ETL writes require plan, preview, approval token, and destructive confirmation for delete/cancel/invite/transfer operations
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect luma

  # Inspect as structured JSON
  pm connectors inspect luma --json

AGENT WORKFLOW
  - Run pm connectors inspect luma before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
