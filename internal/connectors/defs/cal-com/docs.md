# Overview

Reads Cal.com bookings, event types, schedules, webhooks, and profile, and manages bookings/event
types/schedules/webhooks through the Cal.com v2 REST API.

Readable streams: `bookings`, `schedules`, `event_types`, `webhooks`, `my_profile`.

Write actions: `create_booking`, `cancel_booking`, `confirm_booking`, `decline_booking`,
`reschedule_booking`, `create_event_type`, `update_event_type`, `delete_event_type`,
`create_schedule`, `update_schedule`, `delete_schedule`, `create_webhook`, `delete_webhook`.

Service API documentation: https://cal.com/docs/api-reference/introduction.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Cal.com API key. Sent as a Bearer token (Authorization:
  Bearer <api_key>). Never logged.
- `api_version` (optional, string); default `2024-08-13`; Value sent as the cal-api-version header
  on every request.
- `base_url` (optional, string); default `https://api.cal.com`; format `uri`; Cal.com API base URL
  override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `api_version=2024-08-13`, `base_url=https://api.cal.com`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v2/me`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `skip`; limit parameter `take`; page
size 100.

Pagination by stream: none: `my_profile`; offset_limit: `bookings`, `schedules`, `event_types`,
`webhooks`.

- `bookings`: GET `/v2/bookings` - records path `data`; offset/limit pagination; offset parameter
  `skip`; limit parameter `take`; page size 100.
- `schedules`: GET `/v2/schedules` - records path `data`; offset/limit pagination; offset parameter
  `skip`; limit parameter `take`; page size 100.
- `event_types`: GET `/v2/event-types` - records path `data`; offset/limit pagination; offset
  parameter `skip`; limit parameter `take`; page size 100.
- `webhooks`: GET `/v2/webhooks` - records path `data`; offset/limit pagination; offset parameter
  `skip`; limit parameter `take`; page size 100.
- `my_profile`: GET `/v2/me` - records path `data`.

## Write actions & risks

Overall write risk: external mutation of live scheduling data:
creates/cancels/confirms/declines/reschedules real bookings (notifying attendees),
creates/updates/deletes event types and availability schedules (changes public booking
availability), and creates/deletes webhook subscriptions.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_booking`: POST `/v2/bookings` - kind `create`; body type `json`; required record fields
  `start`, `eventTypeId`, `attendee`; accepted fields `attendee`, `bookingFieldsResponses`,
  `eventTypeId`, `guests`, `start`; risk: external mutation; books a real meeting slot on the target
  event type and notifies attendees; approval required.
- `cancel_booking`: POST `/v2/bookings/{{ record.uid }}/cancel` - kind `update`; body type `json`;
  path fields `uid`; required record fields `uid`; accepted fields `cancelSubsequentBookings`,
  `cancellationReason`, `uid`; risk: external mutation; cancels a real booking and notifies
  attendees; approval required.
- `confirm_booking`: POST `/v2/bookings/{{ record.uid }}/confirm` - kind `update`; body type `none`;
  path fields `uid`; required record fields `uid`; accepted fields `uid`; risk: external mutation;
  confirms a booking pending host approval, notifying the attendee; approval required.
- `decline_booking`: POST `/v2/bookings/{{ record.uid }}/decline` - kind `update`; body type `json`;
  path fields `uid`; required record fields `uid`; accepted fields `reason`, `uid`; risk: external
  mutation; declines a booking pending host approval, notifying the attendee; approval required.
- `reschedule_booking`: POST `/v2/bookings/{{ record.uid }}/reschedule` - kind `update`; body type
  `json`; path fields `uid`; required record fields `uid`, `start`; accepted fields `rescheduledBy`,
  `reschedulingReason`, `start`, `uid`; risk: external mutation; moves a real booking to a new time
  and notifies attendees; approval required.
- `create_event_type`: POST `/v2/event-types` - kind `create`; body type `json`; required record
  fields `title`, `slug`, `lengthInMinutes`; accepted fields `description`, `hidden`,
  `lengthInMinutes`, `scheduleId`, `slug`, `title`; risk: external mutation; creates a new
  publicly-bookable event type; approval required.
- `update_event_type`: PATCH `/v2/event-types/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`; accepted fields `description`, `hidden`, `id`,
  `lengthInMinutes`, `scheduleId`, `slug`, `title`; risk: external mutation; changes the public
  scheduling configuration of an existing event type; approval required.
- `delete_event_type`: DELETE `/v2/event-types/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: destructive;
  permanently deletes an event type, breaking any existing public booking links; approval required.
- `create_schedule`: POST `/v2/schedules` - kind `create`; body type `json`; required record fields
  `name`, `timeZone`, `isDefault`; accepted fields `availability`, `isDefault`, `name`, `overrides`,
  `timeZone`; risk: external mutation; creates a new availability schedule, which can be attached to
  event types and change public availability; approval required.
- `update_schedule`: PATCH `/v2/schedules/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `availability`, `id`, `isDefault`,
  `name`, `overrides`, `timeZone`; risk: external mutation; changes a real availability schedule's
  hours/timezone, directly affecting public bookable slots; approval required.
- `delete_schedule`: DELETE `/v2/schedules/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: destructive; permanently
  deletes an availability schedule; approval required.
- `create_webhook`: POST `/v2/webhooks` - kind `create`; body type `json`; required record fields
  `subscriberUrl`, `triggers`, `active`; accepted fields `active`, `payloadTemplate`,
  `subscriberUrl`, `triggers`; risk: external mutation; registers a new webhook endpoint that will
  receive live booking event payloads; approval required.
- `delete_webhook`: DELETE `/v2/webhooks/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: destructive; permanently
  deletes a webhook subscription; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s), 13 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=2, destructive_admin=5, duplicate_of=12, non_data_endpoint=12, out_of_scope=18,
  requires_elevated_scope=54.
