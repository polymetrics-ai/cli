# Overview

Reads and writes the documented Luma public API for events, calendars, guests, contacts, tags,
coupons, ticket types, memberships, webhooks, and organization resources.

Readable streams: `events`, `event_guests`, `event_hosts`, `event`, `calendar`, `calendar_events`,
`guest`, `guests`, `self_user`, `contact_tags`, `event_tags`, `calendar_admins`, `entity_lookup`,
`event_lookup`, `calendar_contacts`, `event_coupons`, `calendar_coupons`, `event_ticket_types`,
`event_ticket_type`, `membership_tiers`, `webhooks`, `webhook`, `organization_admins`,
`organization_calendars`, `organization_events`.

Write actions: `create_event`, `update_event`, `update_guest_status`, `send_event_invites`,
`add_event_guests`, `add_event_host`, `update_event_host`, `remove_event_host`,
`create_event_coupon`, `update_event_coupon`, `create_calendar_coupon`, `update_calendar_coupon`,
`import_calendar_contacts`, `create_contact_tag`, `update_contact_tag`, `delete_contact_tag`,
`apply_contact_tag`, `unapply_contact_tag`, `create_event_tag`, `update_event_tag`,
`delete_event_tag`, `apply_event_tag`, `unapply_event_tag`, `add_calendar_event`,
`approve_calendar_event`, `reject_calendar_event`, `create_image_upload_url`, `create_ticket_type`,
`update_ticket_type`, `delete_ticket_type`, `add_membership_member`,
`update_membership_member_status`, `create_webhook`, `update_webhook`, `delete_webhook`,
`request_event_cancellation`, `cancel_event`, `update_calendar`, `create_organization_calendar`,
`transfer_event_calendar`.

Service API documentation: https://docs.luma.com/reference/getting-started-with-your-api.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Luma API key, sent as the x-luma-api-key header. Never
  logged.
- `base_url` (optional, string); default `https://public-api.luma.com`; format `uri`; Luma API base
  URL override for tests or proxies.
- `event_api_id` (optional, string).
- `event_id` (optional, string); Luma event ID used by event detail, guest, coupon, ticket type, and
  lookup streams.
- `event_ticket_type_id` (optional, string); Luma ticket type ID for the ticket type detail stream.
- `guest_id` (optional, string); Luma guest ID or accepted guest identifier for the guest detail
  stream.
- `mode` (optional, string).
- `slug` (optional, string); Luma slug for entity lookup.
- `webhook_id` (optional, string); Luma webhook ID for the webhook detail stream.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://public-api.luma.com`.

Authentication behavior:

- API key authentication in `x-luma-api-key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v1/users/get-self`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `pagination_cursor`; next token from
`next_cursor`; stop flag `has_more`.

Pagination by stream: cursor: `events`, `event_guests`, `calendar_events`, `guests`,
`calendar_contacts`, `event_coupons`, `calendar_coupons`, `membership_tiers`, `webhooks`,
`organization_calendars`, `organization_events`; none: `event_hosts`, `event`, `calendar`, `guest`,
`self_user`, `contact_tags`, `event_tags`, `calendar_admins`, `entity_lookup`, `event_lookup`,
`event_ticket_types`, `event_ticket_type`, `webhook`, `organization_admins`.

- `events`: GET `/v1/calendars/events/list` - records path `entries`; cursor pagination; cursor
  parameter `pagination_cursor`; next token from `next_cursor`; stop flag `has_more`; computed
  output fields `api_id`, `calendar_api_id`.
- `event_guests`: GET `/v1/events/guests/list` - records path `entries`; query `event_id`=`{{
  config.event_api_id }}`; cursor pagination; cursor parameter `pagination_cursor`; next token from
  `next_cursor`; stop flag `has_more`; computed output fields `api_id`, `email`, `event_api_id`,
  `name`, `user_api_id`.
- `event_hosts`: GET `/v1/events/get` - records path `hosts`; query `event_id`=`{{
  config.event_api_id }}`; computed output fields `access_level`, `api_id`.
- `event`: GET `/v1/events/get` - records path `.`; query `event_id`=`{{ config.event_id }}`; emits
  passthrough records.
- `calendar`: GET `/v1/calendars/get` - records path `.`; emits passthrough records.
- `calendar_events`: GET `/v1/calendars/events/list` - records path `entries`; cursor pagination;
  cursor parameter `pagination_cursor`; next token from `next_cursor`; stop flag `has_more`; emits
  passthrough records.
- `guest`: GET `/v1/events/guests/get` - records path `.`; query `event_id`=`{{ config.event_id }}`;
  `id`=`{{ config.guest_id }}`; emits passthrough records.
- `guests`: GET `/v1/events/guests/list` - records path `entries`; query `event_id`=`{{
  config.event_id }}`; cursor pagination; cursor parameter `pagination_cursor`; next token from
  `next_cursor`; stop flag `has_more`; emits passthrough records.
- `self_user`: GET `/v1/users/get-self` - records path `.`; emits passthrough records.
- `contact_tags`: GET `/v1/calendars/contact-tags/list` - records path `entries`; emits passthrough
  records.
- `event_tags`: GET `/v1/calendars/event-tags/list` - records path `entries`; emits passthrough
  records.
- `calendar_admins`: GET `/v1/calendars/admins/list` - records path `entries`; emits passthrough
  records.
- `entity_lookup`: GET `/v1/entities/lookup` - records path `entity`; query `slug`=`{{ config.slug
  }}`; emits passthrough records.
- `event_lookup`: GET `/v1/calendars/events/lookup` - records path `event`; query `event_id`=`{{
  config.event_id }}`; emits passthrough records.
- `calendar_contacts`: GET `/v1/calendars/contacts/list` - records path `entries`; cursor
  pagination; cursor parameter `pagination_cursor`; next token from `next_cursor`; stop flag
  `has_more`; emits passthrough records.
- `event_coupons`: GET `/v1/events/coupons/list` - records path `entries`; query `event_id`=`{{
  config.event_id }}`; cursor pagination; cursor parameter `pagination_cursor`; next token from
  `next_cursor`; stop flag `has_more`; emits passthrough records.
- `calendar_coupons`: GET `/v1/calendars/coupons/list` - records path `entries`; cursor pagination;
  cursor parameter `pagination_cursor`; next token from `next_cursor`; stop flag `has_more`; emits
  passthrough records.
- `event_ticket_types`: GET `/v1/events/ticket-types/list` - records path `entries`; query
  `event_id`=`{{ config.event_id }}`; emits passthrough records.
- `event_ticket_type`: GET `/v1/events/ticket-types/get` - records path `.`; query
  `event_ticket_type_id`=`{{ config.event_ticket_type_id }}`; emits passthrough records.
- `membership_tiers`: GET `/v1/memberships/tiers/list` - records path `entries`; cursor pagination;
  cursor parameter `pagination_cursor`; next token from `next_cursor`; stop flag `has_more`; emits
  passthrough records.
- `webhooks`: GET `/v1/webhooks/list` - records path `entries`; cursor pagination; cursor parameter
  `pagination_cursor`; next token from `next_cursor`; stop flag `has_more`; emits passthrough
  records.
- `webhook`: GET `/v2/webhooks/get` - records path `.`; query `id`=`{{ config.webhook_id }}`; emits
  passthrough records.
- `organization_admins`: GET `/v1/organizations/admins/list` - records path `entries`; emits
  passthrough records.
- `organization_calendars`: GET `/v1/organizations/calendars/list` - records path `entries`; cursor
  pagination; cursor parameter `pagination_cursor`; next token from `next_cursor`; stop flag
  `has_more`; emits passthrough records.
- `organization_events`: GET `/v1/organizations/events/list` - records path `entries`; cursor
  pagination; cursor parameter `pagination_cursor`; next token from `next_cursor`; stop flag
  `has_more`; emits passthrough records.

## Write actions & risks

Overall write risk: live Luma API mutations can create, update, invite, tag, cancel, transfer, or
delete event/calendar/member/webhook data and may send guest invitations.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_event`: POST `/v1/events/create` - kind `create`; body type `json`; required record fields
  `name`, `start_at`, `timezone`; accepted fields `can_register_for_multiple_tickets`, `coordinate`,
  `cover_url`, `description_md`, `end_at`, `feedback_email`, `geo_address_json`,
  `location_visibility`, `max_capacity`, `meeting_url`, `name`, `name_requirement`,
  `phone_number_requirement`, `registration_open`, `registration_questions`, `reminders_disabled`,
  `show_guest_list`, `slug`, and 5 more; risk: create event through the Luma API.
- `update_event`: POST `/v1/events/update` - kind `update`; body type `json`; required record fields
  `event_id`; accepted fields `can_register_for_multiple_tickets`, `coordinate`, `cover_url`,
  `description_md`, `end_at`, `event_id`, `feedback_email`, `geo_address_json`,
  `location_visibility`, `max_capacity`, `meeting_url`, `name`, `name_requirement`,
  `phone_number_requirement`, `registration_open`, `registration_questions`, `reminders_disabled`,
  `show_guest_list`, and 7 more; risk: update event through the Luma API.
- `update_guest_status`: POST `/v1/events/guests/update-status` - kind `update`; body type `json`;
  required record fields `event_id`, `guest_id`, `status`; accepted fields `event_id`, `guest_id`,
  `message`, `send_email`, `should_refund`, `status`; risk: update guest status through the Luma
  API.
- `send_event_invites`: POST `/v1/events/guests/send-invites` - kind `update`; body type `json`;
  required record fields `event_id`, `guests`; accepted fields `event_id`, `guests`, `message`;
  confirmation `destructive`; risk: send event invites through the Luma API.
- `add_event_guests`: POST `/v1/events/guests/add` - kind `create`; body type `json`; required
  record fields `event_id`, `guests`; accepted fields `approval_status`, `event_id`, `guests`,
  `send_email`, `ticket`, `tickets`; risk: add event guests through the Luma API.
- `add_event_host`: POST `/v1/events/hosts/add` - kind `create`; body type `json`; required record
  fields `event_id`, `email`; accepted fields `access_level`, `email`, `event_id`, `is_visible`,
  `name`; risk: add event host through the Luma API.
- `update_event_host`: POST `/v1/events/hosts/update` - kind `update`; body type `json`; required
  record fields `event_id`, `email`; accepted fields `access_level`, `email`, `event_id`,
  `is_visible`; risk: update event host through the Luma API.
- `remove_event_host`: POST `/v1/events/hosts/remove` - kind `delete`; body type `json`; required
  record fields `event_id`, `email`; accepted fields `email`, `event_id`; confirmation
  `destructive`; risk: remove event host through the Luma API.
- `create_event_coupon`: POST `/v1/events/coupons/create` - kind `create`; body type `json`;
  required record fields `code`, `discount`, `event_id`; accepted fields `code`, `discount`,
  `event_id`, `event_ticket_type_id`, `remaining_count`, `valid_end_at`, `valid_start_at`; risk:
  create event coupon through the Luma API.
- `update_event_coupon`: POST `/v1/events/coupons/update` - kind `update`; body type `json`;
  required record fields `event_id`, `code`; accepted fields `code`, `event_id`, `remaining_count`,
  `valid_end_at`, `valid_start_at`; risk: update event coupon through the Luma API.
- `create_calendar_coupon`: POST `/v1/calendars/coupons/create` - kind `create`; body type `json`;
  required record fields `code`, `discount`; accepted fields `code`, `discount`, `remaining_count`,
  `valid_end_at`, `valid_start_at`; risk: create calendar coupon through the Luma API.
- `update_calendar_coupon`: POST `/v1/calendars/coupons/update` - kind `update`; body type `json`;
  required record fields `code`; accepted fields `code`, `remaining_count`, `valid_end_at`,
  `valid_start_at`; risk: update calendar coupon through the Luma API.
- `import_calendar_contacts`: POST `/v1/calendars/contacts/import` - kind `create`; body type
  `json`; required record fields `contacts`; accepted fields `contacts`, `tags`; risk: import
  calendar contacts through the Luma API.
- `create_contact_tag`: POST `/v1/calendars/contact-tags/create` - kind `create`; body type `json`;
  required record fields `name`; accepted fields `color`, `name`; risk: create contact tag through
  the Luma API.
- `update_contact_tag`: POST `/v1/calendars/contact-tags/update` - kind `update`; body type `json`;
  required record fields `tag_id`; accepted fields `color`, `name`, `tag_id`; risk: update contact
  tag through the Luma API.
- `delete_contact_tag`: POST `/v1/calendars/contact-tags/delete` - kind `delete`; body type `json`;
  required record fields `tag_id`; accepted fields `tag_id`; confirmation `destructive`; risk:
  delete contact tag through the Luma API.
- `apply_contact_tag`: POST `/v1/calendars/contact-tags/apply` - kind `update`; body type `json`;
  required record fields `tag`; accepted fields `emails`, `tag`, `user_ids`; risk: apply contact tag
  through the Luma API.
- `unapply_contact_tag`: POST `/v1/calendars/contact-tags/unapply` - kind `update`; body type
  `json`; required record fields `tag`; accepted fields `emails`, `tag`, `user_ids`; risk: unapply
  contact tag through the Luma API.
- `create_event_tag`: POST `/v1/calendars/event-tags/create` - kind `create`; body type `json`;
  required record fields `name`; accepted fields `color`, `name`; risk: create event tag through the
  Luma API.
- `update_event_tag`: POST `/v1/calendars/event-tags/update` - kind `update`; body type `json`;
  required record fields `tag_id`; accepted fields `color`, `name`, `tag_id`; risk: update event tag
  through the Luma API.
- `delete_event_tag`: POST `/v1/calendars/event-tags/delete` - kind `delete`; body type `json`;
  required record fields `tag_id`; accepted fields `tag_id`; confirmation `destructive`; risk:
  delete event tag through the Luma API.
- `apply_event_tag`: POST `/v1/calendars/event-tags/apply` - kind `update`; body type `json`;
  required record fields `tag`, `event_ids`; accepted fields `event_ids`, `tag`; risk: apply event
  tag through the Luma API.
- `unapply_event_tag`: POST `/v1/calendars/event-tags/unapply` - kind `update`; body type `json`;
  required record fields `tag`, `event_ids`; accepted fields `event_ids`, `tag`; risk: unapply event
  tag through the Luma API.
- `add_calendar_event`: POST `/v1/calendars/events/add` - kind `create`; body type `json`; accepted
  fields `coordinate`, `duration_interval`, `event_api_id`, `event_id`, `geo_address_json`,
  `geo_latitude`, `geo_longitude`, `host`, `name`, `platform`, `start_at`, `submission_mode`,
  `timezone`, `url`; risk: add calendar event through the Luma API.
- `approve_calendar_event`: POST `/v1/calendars/events/approve` - kind `update`; body type `json`;
  required record fields `calendar_event_id`; accepted fields `calendar_event_id`; risk: approve
  calendar event through the Luma API.
- `reject_calendar_event`: POST `/v1/calendars/events/reject` - kind `update`; body type `json`;
  required record fields `calendar_event_id`; accepted fields `calendar_event_id`, `message`;
  confirmation `destructive`; risk: reject calendar event through the Luma API.
- `create_image_upload_url`: POST `/v1/images/create-upload-url` - kind `create`; body type `json`;
  accepted fields `content_type`; risk: create image upload url through the Luma API.
- `create_ticket_type`: POST `/v1/events/ticket-types/create` - kind `create`; body type `json`;
  required record fields `event_id`, `name`, `type`; accepted fields `cents`, `currency`,
  `description`, `event_id`, `is_flexible`, `is_hidden`, `max_capacity`, `min_cents`, `name`,
  `require_approval`, `type`, `valid_end_at`, `valid_start_at`; risk: create ticket type through the
  Luma API.
- `update_ticket_type`: POST `/v1/events/ticket-types/update` - kind `update`; body type `json`;
  required record fields `event_ticket_type_id`; accepted fields `cents`, `currency`, `description`,
  `event_ticket_type_id`, `is_flexible`, `is_hidden`, `max_capacity`, `min_cents`, `name`,
  `require_approval`, `type`, `valid_end_at`, `valid_start_at`; risk: update ticket type through the
  Luma API.
- `delete_ticket_type`: POST `/v1/events/ticket-types/delete` - kind `delete`; body type `json`;
  required record fields `event_ticket_type_id`; accepted fields `event_ticket_type_id`;
  confirmation `destructive`; risk: delete ticket type through the Luma API.
- `add_membership_member`: POST `/v1/memberships/members/add` - kind `create`; body type `json`;
  required record fields `email`, `membership_tier_id`; accepted fields `email`,
  `membership_tier_id`, `registration_answers`, `skip_payment`; risk: add membership member through
  the Luma API.
- `update_membership_member_status`: POST `/v1/memberships/members/update-status` - kind `update`;
  body type `json`; required record fields `user_id`, `status`; accepted fields `status`, `user_id`;
  risk: update membership member status through the Luma API.
- `create_webhook`: POST `/v2/webhooks/create` - kind `create`; body type `json`; required record
  fields `url`, `event_types`; accepted fields `event_types`, `url`; risk: create webhook through
  the Luma API.
- `update_webhook`: POST `/v2/webhooks/update` - kind `update`; body type `json`; required record
  fields `id`; accepted fields `event_types`, `id`, `status`; risk: update webhook through the Luma
  API.
- `delete_webhook`: POST `/v1/webhooks/delete` - kind `delete`; body type `json`; required record
  fields `id`; accepted fields `id`; confirmation `destructive`; risk: delete webhook through the
  Luma API.
- `request_event_cancellation`: POST `/v1/events/cancel/request` - kind `update`; body type `json`;
  required record fields `event_id`; accepted fields `event_id`; confirmation `destructive`; risk:
  request event cancellation through the Luma API.
- `cancel_event`: POST `/v1/events/cancel` - kind `delete`; body type `json`; required record fields
  `event_id`, `cancellation_token`; accepted fields `cancellation_token`, `event_id`,
  `should_refund`; confirmation `destructive`; risk: cancel event through the Luma API.
- `update_calendar`: POST `/v1/calendars/update` - kind `update`; body type `json`; required record
  fields `calendar_id`; accepted fields `avatar_url`, `calendar_id`, `description`, `name`, `slug`,
  `tint_color`; risk: update calendar through the Luma API.
- `create_organization_calendar`: POST `/v2/organizations/calendars/create` - kind `create`; body
  type `json`; required record fields `name`; accepted fields `avatar_url`, `description`, `name`,
  `slug`, `tint_color`; risk: create organization calendar through the Luma API.
- `transfer_event_calendar`: POST `/v1/organizations/events/transfer-calendar` - kind `update`; body
  type `json`; required record fields `event_id`, `calendar_id`; accepted fields `calendar_id`,
  `event_id`; confirmation `destructive`; risk: transfer event calendar through the Luma API.

## Known limits

- Published rate limit metadata: requests_per_minute=200.
- Batch defaults: read_page_size=50, write_batch_size=1.
- API coverage includes 25 stream-backed endpoint group(s), 40 write-backed endpoint group(s).
