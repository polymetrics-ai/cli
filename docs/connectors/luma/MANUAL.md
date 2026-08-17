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
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

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
  api_key (secret)

ETL STREAMS
  events:
    primary key: api_id
    fields: api_id(), calendar_api_id(), cover_url(), created_at(), description(), end_at(), name(), start_at(), timezone(), url(), visibility()
  event_guests:
    primary key: api_id
    fields: api_id(), approval_status(), checked_in_at(), email(), event_api_id(), name(), registered_at(), user_api_id(), user_name()
  event_hosts:
    primary key: api_id
    fields: access_level(), api_id(), avatar_url(), email(), name()
  event:
    primary key: id
    fields: access(), calendar_id(), coordinate(), cover_url(), created_at(), description(), description_md(), duration_interval(), end_at(), feedback_email(), geo_address_json(), guest_counts(), hosts(), id(), location_type(), location_visibility(), meeting_url(), name(), platform(), registration_open(), registration_questions(), start_at(), timezone(), url(), user_id(), visibility(), waitlist_status()
  calendar:
    primary key: id
    fields: avatar_url(), coordinate(), cover_image_url(), description(), id(), instagram_handle(), is_personal(), location(), name(), slug(), social_image_url(), twitter_handle(), url(), website(), youtube_handle()
  calendar_events:
    primary key: id
    fields: calendar_id(), coordinate(), created_at(), duration_interval(), end_at(), geo_address_json(), host(), id(), name(), platform(), start_at(), tags(), timezone(), url()
  guest:
    primary key: id
    fields: approval_status(), check_in_qr_code(), eth_address(), event_ticket_orders(), event_tickets(), id(), invited_at(), joined_at(), phone_number(), registered_at(), registration_answers(), solana_address(), user_email(), user_first_name(), user_id(), user_last_name(), user_name(), utm_source()
  guests:
    primary key: id
    fields: approval_status(), check_in_qr_code(), eth_address(), event_tickets(), id(), invited_at(), joined_at(), phone_number(), registered_at(), registration_answers(), solana_address(), user_email(), user_first_name(), user_id(), user_last_name(), user_name(), utm_source()
  self_user:
    primary key: id
    fields: avatar_url(), email(), first_name(), id(), last_name(), name()
  contact_tags:
    primary key: id
    fields: color(), id(), name()
  event_tags:
    primary key: id
    fields: color(), id(), name()
  calendar_admins:
    primary key: id
    fields: avatar_url(), email(), first_name(), id(), last_name(), name()
  entity_lookup:
    fields: calendar(), event(), type()
  event_lookup:
    primary key: id
    fields: id(), status()
  calendar_contacts:
    primary key: id
    fields: avatar_url(), created_at(), email(), event_approved_count(), event_checked_in_count(), first_name(), id(), last_name(), membership(), name(), revenue_usd_cents(), tags(), user_id()
  event_coupons:
    primary key: id
    fields: cents_off(), code(), currency(), event_ticket_type_id(), id(), percent_off(), remaining_count(), valid_end_at(), valid_start_at()
  calendar_coupons:
    primary key: id
    fields: cents_off(), code(), currency(), event_ticket_type_id(), id(), percent_off(), remaining_count(), valid_end_at(), valid_start_at()
  event_ticket_types:
    primary key: id
    fields: cents(), currency(), description(), id(), is_flexible(), is_hidden(), max_capacity(), min_cents(), name(), require_approval(), type(), valid_end_at(), valid_start_at()
  event_ticket_type:
    primary key: id
    fields: cents(), currency(), description(), id(), is_flexible(), is_hidden(), max_capacity(), min_cents(), name(), require_approval(), type(), valid_end_at(), valid_start_at()
  membership_tiers:
    primary key: id
    fields: access_info(), description(), id(), name(), tint_color()
  webhooks:
    primary key: id
    fields: created_at(), event_types(), id(), secret(), status(), url()
  webhook:
    primary key: id
    fields: created_at(), event_types(), id(), secret(), status(), url()
  organization_admins:
    primary key: id
    fields: api_id(), avatar_url(), email(), first_name(), id(), last_name(), name()
  organization_calendars:
    primary key: id
    fields: avatar_url(), coordinate(), cover_image_url(), description(), id(), instagram_handle(), is_personal(), location(), name(), slug(), social_image_url(), twitter_handle(), url(), website(), youtube_handle()
  organization_events:
    primary key: id
    fields: api_id(), calendar_api_id(), calendar_id(), coordinate(), cover_url(), created_at(), description(), description_md(), duration_interval(), end_at(), feedback_email(), geo_address_json(), geo_latitude(), geo_longitude(), id(), location_type(), location_visibility(), managing_calendars(), meeting_url(), name(), platform(), registration_open(), registration_questions(), start_at(), timezone(), url(), user_api_id(), user_id(), visibility(), waitlist_status(), zoom_meeting_url()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_event:
    endpoint: POST /v1/events/create
    risk: create event through the Luma API.
  update_event:
    endpoint: POST /v1/events/update
    risk: update event through the Luma API.
  update_guest_status:
    endpoint: POST /v1/events/guests/update-status
    risk: update guest status through the Luma API.
  send_event_invites:
    endpoint: POST /v1/events/guests/send-invites
    risk: send event invites through the Luma API.
  add_event_guests:
    endpoint: POST /v1/events/guests/add
    risk: add event guests through the Luma API.
  add_event_host:
    endpoint: POST /v1/events/hosts/add
    risk: add event host through the Luma API.
  update_event_host:
    endpoint: POST /v1/events/hosts/update
    risk: update event host through the Luma API.
  remove_event_host:
    endpoint: POST /v1/events/hosts/remove
    risk: remove event host through the Luma API.
  create_event_coupon:
    endpoint: POST /v1/events/coupons/create
    risk: create event coupon through the Luma API.
  update_event_coupon:
    endpoint: POST /v1/events/coupons/update
    risk: update event coupon through the Luma API.
  create_calendar_coupon:
    endpoint: POST /v1/calendars/coupons/create
    risk: create calendar coupon through the Luma API.
  update_calendar_coupon:
    endpoint: POST /v1/calendars/coupons/update
    risk: update calendar coupon through the Luma API.
  import_calendar_contacts:
    endpoint: POST /v1/calendars/contacts/import
    risk: import calendar contacts through the Luma API.
  create_contact_tag:
    endpoint: POST /v1/calendars/contact-tags/create
    risk: create contact tag through the Luma API.
  update_contact_tag:
    endpoint: POST /v1/calendars/contact-tags/update
    risk: update contact tag through the Luma API.
  delete_contact_tag:
    endpoint: POST /v1/calendars/contact-tags/delete
    risk: delete contact tag through the Luma API.
  apply_contact_tag:
    endpoint: POST /v1/calendars/contact-tags/apply
    risk: apply contact tag through the Luma API.
  unapply_contact_tag:
    endpoint: POST /v1/calendars/contact-tags/unapply
    risk: unapply contact tag through the Luma API.
  create_event_tag:
    endpoint: POST /v1/calendars/event-tags/create
    risk: create event tag through the Luma API.
  update_event_tag:
    endpoint: POST /v1/calendars/event-tags/update
    risk: update event tag through the Luma API.
  delete_event_tag:
    endpoint: POST /v1/calendars/event-tags/delete
    risk: delete event tag through the Luma API.
  apply_event_tag:
    endpoint: POST /v1/calendars/event-tags/apply
    risk: apply event tag through the Luma API.
  unapply_event_tag:
    endpoint: POST /v1/calendars/event-tags/unapply
    risk: unapply event tag through the Luma API.
  add_calendar_event:
    endpoint: POST /v1/calendars/events/add
    risk: add calendar event through the Luma API.
  approve_calendar_event:
    endpoint: POST /v1/calendars/events/approve
    risk: approve calendar event through the Luma API.
  reject_calendar_event:
    endpoint: POST /v1/calendars/events/reject
    risk: reject calendar event through the Luma API.
  create_image_upload_url:
    endpoint: POST /v1/images/create-upload-url
    risk: create image upload url through the Luma API.
  create_ticket_type:
    endpoint: POST /v1/events/ticket-types/create
    risk: create ticket type through the Luma API.
  update_ticket_type:
    endpoint: POST /v1/events/ticket-types/update
    risk: update ticket type through the Luma API.
  delete_ticket_type:
    endpoint: POST /v1/events/ticket-types/delete
    risk: delete ticket type through the Luma API.
  add_membership_member:
    endpoint: POST /v1/memberships/members/add
    risk: add membership member through the Luma API.
  update_membership_member_status:
    endpoint: POST /v1/memberships/members/update-status
    risk: update membership member status through the Luma API.
  create_webhook:
    endpoint: POST /v2/webhooks/create
    risk: create webhook through the Luma API.
  update_webhook:
    endpoint: POST /v2/webhooks/update
    risk: update webhook through the Luma API.
  delete_webhook:
    endpoint: POST /v1/webhooks/delete
    risk: delete webhook through the Luma API.
  request_event_cancellation:
    endpoint: POST /v1/events/cancel/request
    risk: request event cancellation through the Luma API.
  cancel_event:
    endpoint: POST /v1/events/cancel
    risk: cancel event through the Luma API.
  update_calendar:
    endpoint: POST /v1/calendars/update
    risk: update calendar through the Luma API.
  create_organization_calendar:
    endpoint: POST /v2/organizations/calendars/create
    risk: create organization calendar through the Luma API.
  transfer_event_calendar:
    endpoint: POST /v1/organizations/events/transfer-calendar
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
