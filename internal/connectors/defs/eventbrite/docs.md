# Overview

Reads Eventbrite organizations, events, attendees, orders, and ticket classes through the Eventbrite
v3 REST API. Read-only source.

Readable streams: `organizations`, `events`, `attendees`, `orders`, `ticket_classes`.

This connector is read-only; no write actions are declared.

Service API documentation: https://www.eventbrite.com/platform/api.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://www.eventbriteapi.com/v3`; format `uri`;
  Eventbrite API base URL override for tests or proxies.
- `event_id` (optional, string); Eventbrite event id the 'attendees', 'orders', and 'ticket_classes'
  streams are scoped to (required for those streams; substituted into the event-scoped path).
- `organization_id` (optional, string); Eventbrite organization id the 'events' stream is scoped to
  (required for that stream; substituted into the organization-scoped path).
- `private_token` (required, secret, string); Eventbrite private OAuth token, sent as a Bearer
  credential. Never logged.
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; only objects changed at
  or after this time are read (events, attendees, orders, ticket_classes).

Secret fields are redacted in logs and write previews: `private_token`.

Default configuration values: `base_url=https://www.eventbriteapi.com/v3`.

Authentication behavior:

- Bearer token authentication using `secrets.private_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/users/me/organizations/`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `continuation`; next token from
`pagination.continuation`; stop flag `pagination.has_more_items`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `organizations`: GET `/users/me/organizations/` - records path `organizations`; query
  `expand`=`venue,ticket_classes`; cursor pagination; cursor parameter `continuation`; next token
  from `pagination.continuation`; stop flag `pagination.has_more_items`.
- `events`: GET `/organizations/{{ config.organization_id }}/events/` - records path `events`; query
  `changed_since` from template `{{ incremental.lower_bound }}`, omitted when absent;
  `expand`=`venue,ticket_classes`; cursor pagination; cursor parameter `continuation`; next token
  from `pagination.continuation`; stop flag `pagination.has_more_items`; incremental cursor
  `changed`; sent as `changed_since`; formatted as `rfc3339`; initial lower bound from `start_date`;
  computed output fields `description`, `end`, `name`, `start`.
- `attendees`: GET `/events/{{ config.event_id }}/attendees/` - records path `attendees`; query
  `changed_since` from template `{{ incremental.lower_bound }}`, omitted when absent;
  `expand`=`venue,ticket_classes`; cursor pagination; cursor parameter `continuation`; next token
  from `pagination.continuation`; stop flag `pagination.has_more_items`; incremental cursor
  `changed`; sent as `changed_since`; formatted as `rfc3339`; initial lower bound from `start_date`;
  computed output fields `email`, `name`.
- `orders`: GET `/events/{{ config.event_id }}/orders/` - records path `orders`; query
  `changed_since` from template `{{ incremental.lower_bound }}`, omitted when absent;
  `expand`=`venue,ticket_classes`; cursor pagination; cursor parameter `continuation`; next token
  from `pagination.continuation`; stop flag `pagination.has_more_items`; incremental cursor
  `changed`; sent as `changed_since`; formatted as `rfc3339`; initial lower bound from `start_date`.
- `ticket_classes`: GET `/events/{{ config.event_id }}/ticket_classes/` - records path
  `ticket_classes`; query `changed_since` from template `{{ config.start_date }}`, omitted when
  absent; `expand`=`venue,ticket_classes`; cursor pagination; cursor parameter `continuation`; next
  token from `pagination.continuation`; stop flag `pagination.has_more_items`; computed output
  fields `cost`, `description`, `fee`, `name`.

## Write actions & risks

This connector is read-only. Read behavior: external Eventbrite API read of organization, event,
attendee, and order data.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=1, out_of_scope=4.
