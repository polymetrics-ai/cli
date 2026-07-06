# Overview

Reads Calendly scheduled events (and their invitees), event types, organization memberships, groups,
routing forms and submissions, webhook subscriptions, availability schedules, activity log entries,
and the current user, and manages bookings/webhooks/memberships/invitations/event types through the
Calendly v2 REST API.

Readable streams: `scheduled_events`, `event_types`, `organization_memberships`, `groups`, `users`,
`routing_forms`, `routing_form_submissions`, `webhook_subscriptions`, `user_availability_schedules`,
`group_relationships`, `activity_log_entries`, `invitees`.

Write actions: `cancel_scheduled_event`, `create_invitee`, `create_webhook_subscription`,
`delete_webhook_subscription`, `remove_organization_membership`, `invite_user_to_organization`,
`create_one_off_event_type`, `create_share`.

Service API documentation: https://developer.calendly.com/api-docs.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Calendly personal access token or OAuth token. Used only for
  Bearer auth (Authorization: Bearer <api_key>); never logged.
- `base_url` (optional, string); default `https://api.calendly.com`; format `uri`; Calendly API base
  URL override for tests or proxies.
- `organization_uri` (required, string); format `uri`; The authenticated user's Calendly
  organization URI (e.g. https://api.calendly.com/organizations/AAAAAAAAAAAAAAAA), used to scope
  every organization-level list request (organization=<organization_uri>). See docs.md's Known
  limits.
- `page_size` (optional, string); default `100`; Records per page (1-100), sent as the count query
  param.
- `routing_form_uri` (optional, string); format `uri`; The Calendly routing form URI whose
  submissions the routing_form_submissions stream reads (Calendly's `routing_form` query filter is a
  required parameter naming one specific form, not the whole organization). Resolvable from the
  routing_forms stream's own uri field. See docs.md's Known limits.
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; only scheduled_events
  starting at or after this time are read (Calendly's min_start_time filter).
- `user_uri` (optional, string); format `uri`; The Calendly user URI whose availability schedules
  the user_availability_schedules stream reads (Calendly's `user` query filter is a required
  parameter naming one specific user, not the whole organization). Same per-account-invariant-value
  pattern as organization_uri: Calendly's API gives no 'list every user' endpoint and the engine
  dialect cannot chain a prior request's response into this one, so the operator configures it once
  - resolvable by calling GET https://api.calendly.com/users/me and reading resource.uri (the
  authenticated user's own URI is the common case). See docs.md's Known limits.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.calendly.com`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/users/me`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `pagination.next_page`;
next URLs stay on the configured API host.

Pagination by stream: next_url: `scheduled_events`, `event_types`, `organization_memberships`,
`groups`, `routing_forms`, `routing_form_submissions`, `webhook_subscriptions`,
`group_relationships`, `activity_log_entries`, `invitees`; none: `users`,
`user_availability_schedules`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `scheduled_events`: GET `/scheduled_events` - records path `collection`; query `count`=`{{
  config.page_size }}`; `organization`=`{{ config.organization_uri }}`; follows a next-page URL from
  the response body; URL path `pagination.next_page`; next URLs stay on the configured API host;
  incremental cursor `start_time`; sent as `min_start_time`; formatted as `rfc3339`; initial lower
  bound from `start_date`; computed output fields `id`.
- `event_types`: GET `/event_types` - records path `collection`; query `count`=`{{ config.page_size
  }}`; `organization`=`{{ config.organization_uri }}`; follows a next-page URL from the response
  body; URL path `pagination.next_page`; next URLs stay on the configured API host; computed output
  fields `id`.
- `organization_memberships`: GET `/organization_memberships` - records path `collection`; query
  `count`=`{{ config.page_size }}`; `organization`=`{{ config.organization_uri }}`; follows a
  next-page URL from the response body; URL path `pagination.next_page`; next URLs stay on the
  configured API host; computed output fields `id`, `user`, `user_email`, `user_name`.
- `groups`: GET `/groups` - records path `collection`; query `count`=`{{ config.page_size }}`;
  `organization`=`{{ config.organization_uri }}`; follows a next-page URL from the response body;
  URL path `pagination.next_page`; next URLs stay on the configured API host; computed output fields
  `id`.
- `users`: GET `/users/me` - single-object response; records path `resource`; computed output fields
  `id`.
- `routing_forms`: GET `/routing_forms` - records path `collection`; query `count`=`{{
  config.page_size }}`; `organization`=`{{ config.organization_uri }}`; follows a next-page URL from
  the response body; URL path `pagination.next_page`; next URLs stay on the configured API host;
  incremental cursor `updated_at`; formatted as `rfc3339`; computed output fields `id`.
- `routing_form_submissions`: GET `/routing_form_submissions` - records path `collection`; query
  `count`=`{{ config.page_size }}`; `routing_form`=`{{ config.routing_form_uri }}`; follows a
  next-page URL from the response body; URL path `pagination.next_page`; next URLs stay on the
  configured API host; incremental cursor `updated_at`; formatted as `rfc3339`; computed output
  fields `id`.
- `webhook_subscriptions`: GET `/webhook_subscriptions` - records path `collection`; query
  `count`=`{{ config.page_size }}`; `organization`=`{{ config.organization_uri }}`;
  `scope`=`organization`; follows a next-page URL from the response body; URL path
  `pagination.next_page`; next URLs stay on the configured API host; incremental cursor
  `updated_at`; formatted as `rfc3339`; computed output fields `id`.
- `user_availability_schedules`: GET `/user_availability_schedules` - records path `collection`;
  query `user`=`{{ config.user_uri }}`; computed output fields `id`.
- `group_relationships`: GET `/group_relationships` - records path `collection`; query `count`=`{{
  config.page_size }}`; `organization`=`{{ config.organization_uri }}`; follows a next-page URL from
  the response body; URL path `pagination.next_page`; next URLs stay on the configured API host;
  incremental cursor `updated_at`; formatted as `rfc3339`; computed output fields `id`.
- `activity_log_entries`: GET `/activity_log_entries` - records path `collection`; query `count`=`{{
  config.page_size }}`; `organization`=`{{ config.organization_uri }}`; follows a next-page URL from
  the response body; URL path `pagination.next_page`; next URLs stay on the configured API host;
  incremental cursor `occurred_at`; sent as `min_occurred_at`; formatted as `rfc3339`; computed
  output fields `id`.
- `invitees`: GET `/scheduled_events/{{ fanout.id }}/invitees` - records path `collection`; query
  `count`=`{{ config.page_size }}`; follows a next-page URL from the response body; URL path
  `pagination.next_page`; next URLs stay on the configured API host; incremental cursor
  `updated_at`; formatted as `rfc3339`; computed output fields `id`; fan-out; ids from request
  `/scheduled_events?organization={{ config.organization_uri | urlencode }}&count={{
  config.page_size }}`; id-list records path `collection`; id field `uri`; id inserted into the
  request path; stamps `scheduled_event_id`.

## Write actions & risks

Overall write risk: external mutation of live scheduling data: cancels real scheduled events and
books new ones (notifying invitees), creates/deletes webhook subscriptions, removes organization
memberships, sends organization invitation emails, and creates one-off event types/shareable booking
links.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `cancel_scheduled_event`: POST `/scheduled_events/{{ record.uuid }}/cancellation` - kind `update`;
  body type `json`; path fields `uuid`; required record fields `uuid`; accepted fields `reason`,
  `uuid`; risk: external mutation; cancels a real scheduled event and notifies invitees; approval
  required.
- `create_invitee`: POST `/invitees` - kind `create`; body type `json`; required record fields
  `event_type`, `start_time`, `invitee`; accepted fields `event_type`, `invitee`, `start_time`;
  risk: external mutation; books a real meeting slot on the target event type and notifies the
  invitee; approval required.
- `create_webhook_subscription`: POST `/webhook_subscriptions` - kind `create`; body type `json`;
  required record fields `url`, `events`, `organization`, `scope`; accepted fields `events`,
  `organization`, `scope`, `url`, `user`; risk: external mutation; registers a new webhook endpoint
  that will receive live invitee/routing-form event payloads; approval required.
- `delete_webhook_subscription`: DELETE `/webhook_subscriptions/{{ record.uuid }}` - kind `delete`;
  body type `none`; path fields `uuid`; required record fields `uuid`; accepted fields `uuid`; risk:
  destructive; permanently deletes a webhook subscription; approval required.
- `remove_organization_membership`: DELETE `/organization_memberships/{{ record.uuid }}` - kind
  `delete`; body type `none`; path fields `uuid`; required record fields `uuid`; accepted fields
  `uuid`; risk: destructive; permanently removes a user's membership from the organization, revoking
  their access; approval required.
- `invite_user_to_organization`: POST `/organizations/{{ record.organization_uuid }}/invitations` -
  kind `create`; body type `json`; path fields `organization_uuid`; body fields `email`; required
  record fields `organization_uuid`, `email`; accepted fields `email`, `organization_uuid`; risk:
  external mutation; sends a real organization-invitation email to the given address; approval
  required.
- `create_one_off_event_type`: POST `/one_off_event_types` - kind `create`; body type `json`;
  required record fields `name`, `host`, `duration`, `date_setting`; accepted fields `date_setting`,
  `duration`, `host`, `location`, `name`; risk: external mutation; publishes a new one-off
  publicly-bookable event type; approval required.
- `create_share`: POST `/shares` - kind `create`; body type `json`; required record fields
  `event_type`; accepted fields `event_type`, `max_spots`; risk: external mutation; creates a new
  shareable booking link with its own spot limit for an event type; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 12 stream-backed endpoint group(s), 8 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=2, duplicate_of=11, non_data_endpoint=1, out_of_scope=1.
