# Overview

Reads Svix applications, endpoints, event types, messages, message delivery attempts, background
tasks, connectors, and operational webhook endpoints, and writes
application/endpoint/event-type/connector/operational-webhook-endpoint lifecycle mutations and
outgoing messages, through the Svix REST API.

Readable streams: `applications`, `endpoints`, `event_types`, `messages`, `background_tasks`,
`connectors`, `operational_webhook_endpoints`.

Write actions: `create_application`, `update_application`, `delete_application`, `create_endpoint`,
`update_endpoint`, `delete_endpoint`, `create_event_type`, `update_event_type`, `delete_event_type`,
`send_message`, `create_connector`, `update_connector`, `delete_connector`,
`create_operational_webhook_endpoint`, `update_operational_webhook_endpoint`,
`delete_operational_webhook_endpoint`.

Service API documentation: https://docs.svix.com/api-reference.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Svix API key, sent as a Bearer token (Authorization: Bearer
  <api_key>). Never logged.
- `base_url` (optional, string); default `https://api.svix.com/api/v1`; format `uri`; Svix API base
  URL. Svix is multi-region (us/eu/ca/au/in); override with the region-specific host (e.g.
  https://api.eu.svix.com/api/v1) if the account is not on the default/US region, or for
  tests/proxies.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.svix.com/api/v1`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/app`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `iterator`; next token from `iterator`.

- `applications`: GET `/app` - records path `data`; query `limit`=`50`; cursor pagination; cursor
  parameter `iterator`; next token from `iterator`; computed output fields `created_at`.
- `endpoints`: GET `/app/{{ fanout.id }}/endpoint` - records path `data`; query `limit`=`50`; cursor
  pagination; cursor parameter `iterator`; next token from `iterator`; computed output fields
  `created_at`, `updated_at`; fan-out; ids from request `/app`; id-list records path `data`; id
  field `id`; id inserted into the request path; stamps `app_id`.
- `event_types`: GET `/event-type` - records path `data`; query `include_archived`=`true`;
  `limit`=`50`; `with_content`=`true`; cursor pagination; cursor parameter `iterator`; next token
  from `iterator`; computed output fields `created_at`, `updated_at`.
- `messages`: GET `/app/{{ fanout.id }}/msg` - records path `data`; query `limit`=`50`; cursor
  pagination; cursor parameter `iterator`; next token from `iterator`; fan-out; ids from request
  `/app`; id-list records path `data`; id field `id`; id inserted into the request path; stamps
  `app_id`.
- `background_tasks`: GET `/background-task` - records path `data`; query `limit`=`50`; cursor
  pagination; cursor parameter `iterator`; next token from `iterator`; computed output fields
  `updated_at`.
- `connectors`: GET `/connector` - records path `data`; query `limit`=`50`; cursor pagination;
  cursor parameter `iterator`; next token from `iterator`; computed output fields `created_at`,
  `updated_at`.
- `operational_webhook_endpoints`: GET `/operational-webhook/endpoint` - records path `data`; query
  `limit`=`50`; cursor pagination; cursor parameter `iterator`; next token from `iterator`; computed
  output fields `created_at`, `updated_at`.

## Write actions & risks

Overall write risk: external Svix API mutation
(application/endpoint/event-type/connector/operational-webhook-endpoint lifecycle, outgoing message
creation).

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_application`: POST `/app` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `metadata`, `name`, `throttleRate`, `uid`; risk: creates a new Svix
  application (a webhook-sending namespace); low-risk external mutation, no approval required.
- `update_application`: PUT `/app/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `name`; accepted fields `id`, `metadata`, `name`,
  `throttleRate`, `uid`; risk: replaces an existing application's metadata/name/throttle rate;
  external mutation, no approval required.
- `delete_application`: DELETE `/app/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; risk: irreversibly deletes an application and all its endpoints, messages, and
  delivery history; approval required.
- `create_endpoint`: POST `/app/{{ record.app_id }}/endpoint` - kind `create`; body type `json`;
  path fields `app_id`; required record fields `app_id`, `url`; accepted fields `app_id`,
  `channels`, `description`, `disabled`, `filterTypes`, `metadata`, `uid`, `url`; risk: creates a
  new webhook delivery endpoint on an application; the endpoint immediately starts receiving future
  events; low-risk external mutation, no approval required.
- `update_endpoint`: PUT `/app/{{ record.app_id }}/endpoint/{{ record.id }}` - kind `update`; body
  type `json`; path fields `app_id`, `id`; required record fields `app_id`, `id`, `url`; accepted
  fields `app_id`, `channels`, `description`, `disabled`, `filterTypes`, `id`, `uid`, `url`; risk:
  replaces an existing endpoint's delivery URL/filters/disabled state; changing url redirects all
  future webhook deliveries for that endpoint; external mutation, no approval required.
- `delete_endpoint`: DELETE `/app/{{ record.app_id }}/endpoint/{{ record.id }}` - kind `delete`;
  body type `none`; path fields `app_id`, `id`; required record fields `app_id`, `id`; accepted
  fields `app_id`, `id`; missing records treated as success for status `404`; risk: irreversibly
  deletes a webhook delivery endpoint and stops all future deliveries to it; approval required.
- `create_event_type`: POST `/event-type` - kind `create`; body type `json`; required record fields
  `name`, `description`; accepted fields `archived`, `deprecated`, `description`, `groupName`,
  `name`, `schemas`; risk: creates a new event type definition; low-risk external mutation, no
  approval required.
- `update_event_type`: PUT `/event-type/{{ record.name }}` - kind `update`; body type `json`; path
  fields `name`; required record fields `name`, `description`; accepted fields `archived`,
  `deprecated`, `description`, `groupName`, `name`, `schemas`; risk: replaces an existing event
  type's description/schema/archived state; external mutation, no approval required.
- `delete_event_type`: DELETE `/event-type/{{ record.name }}` - kind `delete`; body type `none`;
  path fields `name`; required record fields `name`; accepted fields `name`; missing records treated
  as success for status `404`; risk: archives (soft-deletes) an event type definition; approval
  required.
- `send_message`: POST `/app/{{ record.app_id }}/msg` - kind `create`; body type `json`; path fields
  `app_id`; required record fields `app_id`, `eventType`, `payload`; accepted fields `app_id`,
  `channels`, `eventId`, `eventType`, `payload`, `tags`; risk: sends a real outgoing webhook message
  that Svix immediately attempts to deliver to every matching endpoint on the application; approval
  required.
- `create_connector`: POST `/connector` - kind `create`; body type `json`; required record fields
  `name`, `transformation`; accepted fields `allowedEventTypes`, `description`, `instructions`,
  `kind`, `logo`, `name`, `productType`, `transformation`; risk: creates a new outgoing-webhook
  payload-transformation connector template; low-risk external mutation, no approval required.
- `update_connector`: PUT `/connector/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `name`, `transformation`; accepted fields `description`,
  `id`, `name`, `transformation`; risk: replaces an existing connector's transformation
  JS/description; changes the payload shape delivered to every endpoint using this connector;
  external mutation, no approval required.
- `delete_connector`: DELETE `/connector/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: irreversibly deletes a connector transformation template; approval
  required.
- `create_operational_webhook_endpoint`: POST `/operational-webhook/endpoint` - kind `create`; body
  type `json`; required record fields `url`; accepted fields `description`, `disabled`,
  `filterTypes`, `uid`, `url`; risk: creates a new operational webhook endpoint (Svix account-level
  events, e.g. message.attempt.exhausted); low-risk external mutation, no approval required.
- `update_operational_webhook_endpoint`: PUT `/operational-webhook/endpoint/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`, `url`; accepted fields
  `description`, `disabled`, `filterTypes`, `id`, `url`; risk: replaces an existing operational
  webhook endpoint's delivery URL/filters/disabled state; external mutation, no approval required.
- `delete_operational_webhook_endpoint`: DELETE `/operational-webhook/endpoint/{{ record.id }}` -
  kind `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; missing records treated as success for status `404`; risk: irreversibly deletes an
  operational webhook endpoint and stops all future account-level event deliveries to it; approval
  required.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 7 stream-backed endpoint group(s), 16 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=2, destructive_admin=4, duplicate_of=15, non_data_endpoint=8, out_of_scope=56,
  requires_elevated_scope=20.
