# Overview

Reads and writes BigMailer brands, account users, and brand-scoped contacts, lists, custom fields,
message types, segments, senders, templates, suppression lists, and campaigns through the BigMailer
REST API.

Readable streams: `brands`, `users`, `contacts`, `lists`, `fields`, `connections`, `message_types`,
`segments`, `senders`, `templates`, `suppression_lists`, `bulk_campaigns`, `rss_campaigns`,
`transactional_campaigns`, `test_campaigns`.

Write actions: `create_brand`, `update_brand`, `create_contact`, `update_contact`, `upsert_contact`,
`delete_contact`, `create_list`, `update_list`, `delete_list`, `create_field`, `update_field`,
`delete_field`, `create_message_type`, `update_message_type`, `delete_message_type`,
`create_segment`, `update_segment`, `delete_segment`, `create_sender`, `update_sender`,
`delete_sender`, `create_template`, `update_template`, `delete_template`, `create_user`,
`update_user`, `delete_user`.

Service API documentation: https://www.bigmailer.io/api.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); BigMailer API key, sent as the X-API-Key header. Never
  logged.
- `base_url` (optional, string); default `https://api.bigmailer.io/v1`; format `uri`; BigMailer API
  base URL override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.bigmailer.io/v1`.

Authentication behavior:

- API key authentication in `X-API-Key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/brands` with query `limit`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `cursor`; next token from `cursor`; stop
flag `has_more`.

- `brands`: GET `/brands` - records path `data`; query `limit`=`100`; cursor pagination; cursor
  parameter `cursor`; next token from `cursor`; stop flag `has_more`.
- `users`: GET `/users` - records path `data`; query `limit`=`100`; cursor pagination; cursor
  parameter `cursor`; next token from `cursor`; stop flag `has_more`.
- `contacts`: GET `/brands/{{ fanout.id }}/contacts` - records path `data`; query `limit`=`100`;
  cursor pagination; cursor parameter `cursor`; next token from `cursor`; stop flag `has_more`;
  fan-out; ids from request `/brands`; id-list records path `data`; id field `id`; id inserted into
  the request path; stamps `brand_id`.
- `lists`: GET `/brands/{{ fanout.id }}/lists` - records path `data`; query `limit`=`100`; cursor
  pagination; cursor parameter `cursor`; next token from `cursor`; stop flag `has_more`; fan-out;
  ids from request `/brands`; id-list records path `data`; id field `id`; id inserted into the
  request path; stamps `brand_id`.
- `fields`: GET `/brands/{{ fanout.id }}/fields` - records path `data`; query `limit`=`100`; cursor
  pagination; cursor parameter `cursor`; next token from `cursor`; stop flag `has_more`; fan-out;
  ids from request `/brands`; id-list records path `data`; id field `id`; id inserted into the
  request path; stamps `brand_id`.
- `connections`: GET `/connections` - records path `data`; query `limit`=`100`; cursor pagination;
  cursor parameter `cursor`; next token from `cursor`; stop flag `has_more`.
- `message_types`: GET `/brands/{{ fanout.id }}/message-types` - records path `data`; query
  `limit`=`100`; cursor pagination; cursor parameter `cursor`; next token from `cursor`; stop flag
  `has_more`; fan-out; ids from request `/brands`; id-list records path `data`; id field `id`; id
  inserted into the request path; stamps `brand_id`.
- `segments`: GET `/brands/{{ fanout.id }}/segments` - records path `data`; query `limit`=`100`;
  cursor pagination; cursor parameter `cursor`; next token from `cursor`; stop flag `has_more`;
  fan-out; ids from request `/brands`; id-list records path `data`; id field `id`; id inserted into
  the request path; stamps `brand_id`.
- `senders`: GET `/brands/{{ fanout.id }}/senders` - records path `data`; query `limit`=`100`;
  cursor pagination; cursor parameter `cursor`; next token from `cursor`; stop flag `has_more`;
  fan-out; ids from request `/brands`; id-list records path `data`; id field `id`; id inserted into
  the request path; stamps `brand_id`.
- `templates`: GET `/brands/{{ fanout.id }}/templates` - records path `data`; query `limit`=`100`;
  cursor pagination; cursor parameter `cursor`; next token from `cursor`; stop flag `has_more`;
  fan-out; ids from request `/brands`; id-list records path `data`; id field `id`; id inserted into
  the request path; stamps `brand_id`.
- `suppression_lists`: GET `/brands/{{ fanout.id }}/suppression-lists` - records path `data`; query
  `limit`=`100`; cursor pagination; cursor parameter `cursor`; next token from `cursor`; stop flag
  `has_more`; fan-out; ids from request `/brands`; id-list records path `data`; id field `id`; id
  inserted into the request path; stamps `brand_id`.
- `bulk_campaigns`: GET `/brands/{{ fanout.id }}/bulk-campaigns` - records path `data`; query
  `limit`=`100`; cursor pagination; cursor parameter `cursor`; next token from `cursor`; stop flag
  `has_more`; fan-out; ids from request `/brands`; id-list records path `data`; id field `id`; id
  inserted into the request path; stamps `brand_id`.
- `rss_campaigns`: GET `/brands/{{ fanout.id }}/rss-campaigns` - records path `data`; query
  `limit`=`100`; cursor pagination; cursor parameter `cursor`; next token from `cursor`; stop flag
  `has_more`; fan-out; ids from request `/brands`; id-list records path `data`; id field `id`; id
  inserted into the request path; stamps `brand_id`.
- `transactional_campaigns`: GET `/brands/{{ fanout.id }}/transactional-campaigns` - records path
  `data`; query `limit`=`100`; cursor pagination; cursor parameter `cursor`; next token from
  `cursor`; stop flag `has_more`; fan-out; ids from request `/brands`; id-list records path `data`;
  id field `id`; id inserted into the request path; stamps `brand_id`.
- `test_campaigns`: GET `/brands/{{ fanout.id }}/test-campaigns` - records path `data`; query
  `limit`=`100`; cursor pagination; cursor parameter `cursor`; next token from `cursor`; stop flag
  `has_more`; fan-out; ids from request `/brands`; id-list records path `data`; id field `id`; id
  inserted into the request path; stamps `brand_id`.

## Write actions & risks

Overall write risk: external mutation of BigMailer brands, contacts, lists, custom fields, message
types, segments, senders, templates, and account users; can send real emails indirectly (e.g. via a
sender/template referenced by a later campaign) but issues no send action itself.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_brand`: POST `/brands` - kind `create`; body type `json`; required record fields `name`,
  `from_name`, `from_email`, `connection_id`; accepted fields `bounce_danger_percent`,
  `connection_id`, `contact_limit`, `from_email`, `from_name`, `logo`, `max_soft_bounces`, `name`,
  `unsubscribe_text`, `url`; risk: external mutation; creates a new BigMailer brand (sending
  identity); approval required.
- `update_brand`: POST `/brands/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `bounce_danger_percent`, `connection_id`,
  `contact_limit`, `from_email`, `from_name`, `id`, `logo`, `max_soft_bounces`, `name`,
  `unsubscribe_text`, `url`; risk: external mutation; approval required.
- `create_contact`: POST `/brands/{{ record.brand_id }}/contacts` - kind `create`; body type `json`;
  path fields `brand_id`; required record fields `brand_id`, `email`; accepted fields `brand_id`,
  `email`, `field_values`, `list_ids`, `unsubscribe_all`, `unsubscribe_ids`; risk: external
  mutation; creates a contact in a BigMailer brand; approval required.
- `update_contact`: POST `/brands/{{ record.brand_id }}/contacts/{{ record.id }}` - kind `update`;
  body type `json`; path fields `brand_id`, `id`; required record fields `brand_id`, `id`; accepted
  fields `brand_id`, `email`, `field_values`, `id`, `list_ids`, `unsubscribe_all`,
  `unsubscribe_ids`; risk: external mutation; approval required.
- `upsert_contact`: POST `/brands/{{ record.brand_id }}/contacts/upsert` - kind `upsert`; body type
  `json`; path fields `brand_id`; required record fields `brand_id`, `email`; accepted fields
  `brand_id`, `email`, `field_values`, `list_ids`, `unsubscribe_all`, `unsubscribe_ids`; risk:
  external mutation; creates the contact if the email is new, otherwise updates the existing
  contact; approval required.
- `delete_contact`: DELETE `/brands/{{ record.brand_id }}/contacts/{{ record.id }}` - kind `delete`;
  body type `none`; path fields `brand_id`, `id`; required record fields `brand_id`, `id`; accepted
  fields `brand_id`, `id`; missing records treated as success for status `404`; risk: permanently
  removes a contact from a brand; irreversible; approval required.
- `create_list`: POST `/brands/{{ record.brand_id }}/lists` - kind `create`; body type `json`; path
  fields `brand_id`; required record fields `brand_id`, `name`; accepted fields `brand_id`, `name`;
  risk: external mutation; creates a contact list in a BigMailer brand; approval required.
- `update_list`: POST `/brands/{{ record.brand_id }}/lists/{{ record.id }}` - kind `update`; body
  type `json`; path fields `brand_id`, `id`; required record fields `brand_id`, `id`; accepted
  fields `brand_id`, `id`, `name`; risk: external mutation; approval required.
- `delete_list`: DELETE `/brands/{{ record.brand_id }}/lists/{{ record.id }}` - kind `delete`; body
  type `none`; path fields `brand_id`, `id`; required record fields `brand_id`, `id`; accepted
  fields `brand_id`, `id`; missing records treated as success for status `404`; risk: permanently
  removes a list from a brand (contacts in the list are NOT deleted); irreversible; approval
  required.
- `create_field`: POST `/brands/{{ record.brand_id }}/fields` - kind `create`; body type `json`;
  path fields `brand_id`; required record fields `brand_id`, `name`, `type`; accepted fields
  `brand_id`, `merge_tag_name`, `name`, `sample_value`, `type`; risk: external mutation; creates a
  custom contact field in a BigMailer brand; approval required.
- `update_field`: POST `/brands/{{ record.brand_id }}/fields/{{ record.id }}` - kind `update`; body
  type `json`; path fields `brand_id`, `id`; required record fields `brand_id`, `id`; accepted
  fields `brand_id`, `id`, `merge_tag_name`, `name`, `sample_value`; risk: external mutation;
  approval required.
- `delete_field`: DELETE `/brands/{{ record.brand_id }}/fields/{{ record.id }}` - kind `delete`;
  body type `none`; path fields `brand_id`, `id`; required record fields `brand_id`, `id`; accepted
  fields `brand_id`, `id`; missing records treated as success for status `404`; risk: permanently
  removes a custom contact field from a brand; irreversible; approval required.
- `create_message_type`: POST `/brands/{{ record.brand_id }}/message-types` - kind `create`; body
  type `json`; path fields `brand_id`; required record fields `brand_id`, `name`; accepted fields
  `brand_id`, `name`; risk: external mutation; creates a message type (unsubscribe category) in a
  BigMailer brand; approval required.
- `update_message_type`: POST `/brands/{{ record.brand_id }}/message-types/{{ record.id }}` - kind
  `update`; body type `json`; path fields `brand_id`, `id`; required record fields `brand_id`, `id`;
  accepted fields `brand_id`, `id`, `name`; risk: external mutation; approval required.
- `delete_message_type`: DELETE `/brands/{{ record.brand_id }}/message-types/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `brand_id`, `id`; required record fields `brand_id`, `id`;
  accepted fields `brand_id`, `id`; missing records treated as success for status `404`; risk:
  permanently removes a message type from a brand; irreversible; approval required.
- `create_segment`: POST `/brands/{{ record.brand_id }}/segments` - kind `create`; body type `json`;
  path fields `brand_id`; required record fields `brand_id`, `name`, `operator`, `conditions`;
  accepted fields `brand_id`, `conditions`, `name`, `operator`; risk: external mutation; creates a
  contact segment in a BigMailer brand; approval required.
- `update_segment`: POST `/brands/{{ record.brand_id }}/segments/{{ record.id }}` - kind `update`;
  body type `json`; path fields `brand_id`, `id`; required record fields `brand_id`, `id`; accepted
  fields `brand_id`, `conditions`, `id`, `name`, `operator`; risk: external mutation; approval
  required.
- `delete_segment`: DELETE `/brands/{{ record.brand_id }}/segments/{{ record.id }}` - kind `delete`;
  body type `none`; path fields `brand_id`, `id`; required record fields `brand_id`, `id`; accepted
  fields `brand_id`, `id`; missing records treated as success for status `404`; risk: permanently
  removes a segment from a brand; irreversible; approval required.
- `create_sender`: POST `/brands/{{ record.brand_id }}/senders` - kind `create`; body type `json`;
  path fields `brand_id`; required record fields `brand_id`, `identity`; accepted fields `brand_id`,
  `identity`, `identity_type`, `share_type`; risk: external mutation; adds a sender domain/email
  identity to a BigMailer brand; approval required.
- `update_sender`: POST `/brands/{{ record.brand_id }}/senders/{{ record.id }}` - kind `update`;
  body type `json`; path fields `brand_id`, `id`; required record fields `brand_id`, `id`; accepted
  fields `brand_id`, `id`, `share_type`; risk: external mutation; approval required.
- `delete_sender`: DELETE `/brands/{{ record.brand_id }}/senders/{{ record.id }}` - kind `delete`;
  body type `none`; path fields `brand_id`, `id`; required record fields `brand_id`, `id`; accepted
  fields `brand_id`, `id`; missing records treated as success for status `404`; risk: permanently
  removes a sender identity from a brand; irreversible; approval required.
- `create_template`: POST `/brands/{{ record.brand_id }}/templates` - kind `create`; body type
  `json`; path fields `brand_id`; required record fields `brand_id`, `name`, `type`, `html`;
  accepted fields `brand_id`, `html`, `name`, `shared_with_account`, `type`; risk: external
  mutation; creates a campaign template in a BigMailer brand; approval required.
- `update_template`: POST `/brands/{{ record.brand_id }}/templates/{{ record.id }}` - kind `update`;
  body type `json`; path fields `brand_id`, `id`; required record fields `brand_id`, `id`; accepted
  fields `brand_id`, `html`, `id`, `name`, `shared_with_account`, `type`; risk: external mutation;
  approval required.
- `delete_template`: DELETE `/brands/{{ record.brand_id }}/templates/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `brand_id`, `id`; required record fields `brand_id`, `id`;
  accepted fields `brand_id`, `id`; missing records treated as success for status `404`; risk:
  permanently removes a template from a brand; irreversible; approval required.
- `create_user`: POST `/users` - kind `create`; body type `json`; required record fields `email`,
  `role`; accepted fields `allowed_brands`, `email`, `invitation_message`, `role`; risk: external
  mutation; invites a new user into the BigMailer account; approval required.
- `update_user`: POST `/users/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `allowed_brands`, `email`, `id`, `role`; risk:
  external mutation; approval required.
- `delete_user`: DELETE `/users/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; risk: permanently removes a user from the BigMailer account; irreversible; approval
  required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 15 stream-backed endpoint group(s), 27 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=10, duplicate_of=15, non_data_endpoint=2, out_of_scope=5,
  requires_elevated_scope=6.
