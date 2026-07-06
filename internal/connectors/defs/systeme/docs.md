# Overview

Reads Systeme.io contacts, tags, contact fields, funnels, and funnel steps, and writes
contact/tag/contact-field/funnel lifecycle mutations and contact-tag assignment, through the
Systeme.io public API.

Readable streams: `contacts`, `tags`, `contact_fields`, `funnels`, `funnel_steps`, `webhooks`.

Write actions: `create_contact`, `update_contact`, `delete_contact`, `create_tag`, `update_tag`,
`delete_tag`, `add_contact_tag`, `remove_contact_tag`, `create_contact_field`,
`update_contact_field`, `delete_contact_field`, `create_funnel`, `create_funnel_step`,
`create_webhook`.

Service API documentation: https://developer.systeme.io/reference/api.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Systeme.io API key, sent as the X-API-Key header. Never
  logged.
- `base_url` (optional, string); default `https://api.systeme.io/api`; format `uri`; Systeme.io API
  base URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.systeme.io/api`.

Authentication behavior:

- API key authentication in `X-API-Key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/contacts`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 100.

- `contacts`: GET `/contacts` - records path `items`; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 100; computed output fields `created_at`.
- `tags`: GET `/tags` - records path `items`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100.
- `contact_fields`: GET `/contact_fields` - records path `items`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `funnels`: GET `/funnels` - records path `items`; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 100.
- `funnel_steps`: GET `/funnels/{{ fanout.id }}/steps` - records path `items`; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100; fan-out;
  ids from request `/funnels`; id-list records path `items`; id field `id`; id inserted into the
  request path; stamps `funnel_id`.
- `webhooks`: GET `/webhooks` - records path `items`; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 100.

## Write actions & risks

Overall write risk: external Systeme.io API mutation (contact/tag/contact-field/funnel lifecycle,
contact-tag assignment).

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_contact`: POST `/contacts` - kind `create`; body type `json`; required record fields
  `email`; accepted fields `email`, `fields`, `locale`; risk: creates a new contact; low-risk
  external mutation, no approval required.
- `update_contact`: PATCH `/contacts/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `email`, `fields`, `id`, `locale`; risk:
  updates an existing contact's email/locale/custom-field values; external mutation, no approval
  required.
- `delete_contact`: DELETE `/contacts/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: irreversibly deletes a contact; approval required.
- `create_tag`: POST `/tags` - kind `create`; body type `json`; required record fields `name`;
  accepted fields `name`; risk: creates a new tag; low-risk external mutation, no approval required.
- `update_tag`: PUT `/tags/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`, `name`; accepted fields `id`, `name`; risk: renames an existing tag;
  external mutation, no approval required.
- `delete_tag`: DELETE `/tags/{{ record.id }}` - kind `delete`; body type `none`; path fields `id`;
  required record fields `id`; accepted fields `id`; missing records treated as success for status
  `404`; risk: irreversibly deletes a tag, removing it from every contact it is assigned to;
  approval required.
- `add_contact_tag`: POST `/contacts/{{ record.contact_id }}/tags` - kind `create`; body type
  `json`; path fields `contact_id`; required record fields `contact_id`, `tag_id`; accepted fields
  `contact_id`, `tag_id`; risk: assigns a tag to a contact; assigning certain tags can trigger
  Systeme.io automations (enrollment in a course/campaign); external mutation, no approval required.
- `remove_contact_tag`: DELETE `/contacts/{{ record.contact_id }}/tags/{{ record.tag_id }}` - kind
  `delete`; body type `none`; path fields `contact_id`, `tag_id`; required record fields
  `contact_id`, `tag_id`; accepted fields `contact_id`, `tag_id`; missing records treated as success
  for status `404`; risk: removes a tag from a contact; removing certain tags can trigger Systeme.io
  automations; external mutation, no approval required.
- `create_contact_field`: POST `/contact_fields` - kind `create`; body type `json`; required record
  fields `slug`, `type`; accepted fields `slug`, `type`; risk: creates a new custom contact field
  definition; low-risk external mutation, no approval required.
- `update_contact_field`: PATCH `/contact_fields/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`; accepted fields `id`, `slug`, `type`; risk: updates
  an existing custom contact field definition; external mutation, no approval required.
- `delete_contact_field`: DELETE `/contact_fields/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; risk: irreversibly deletes a custom contact field definition
  and its stored values on every contact; approval required.
- `create_funnel`: POST `/funnels` - kind `create`; body type `json`; required record fields `name`;
  accepted fields `name`; risk: creates a new sales funnel; low-risk external mutation, no approval
  required.
- `create_funnel_step`: POST `/funnels/{{ record.funnel_id }}/steps` - kind `create`; body type
  `json`; path fields `funnel_id`; required record fields `funnel_id`, `name`; accepted fields
  `funnel_id`, `name`; risk: creates a new step within an existing funnel; low-risk external
  mutation, no approval required.
- `create_webhook`: POST `/webhooks` - kind `create`; body type `json`; required record fields
  `url`, `event`; accepted fields `event`, `url`; risk: creates a new outgoing webhook subscription;
  low-risk external mutation, no approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 6 stream-backed endpoint group(s), 14 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=4, requires_elevated_scope=7.
