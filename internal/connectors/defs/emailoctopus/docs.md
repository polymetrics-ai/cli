# Overview

Reads and writes EmailOctopus lists, campaigns, campaign summary reports, list contacts, list tags,
and list custom fields through the EmailOctopus v1.6 REST API.

Readable streams: `lists`, `campaigns`, `list_contacts`, `list_tags`, `campaign_summary_reports`.

Write actions: `create_list`, `update_list`, `delete_list`, `create_list_contact`,
`update_list_contact`, `delete_list_contact`, `create_list_tag`, `update_list_tag`,
`delete_list_tag`, `create_list_field`, `update_list_field`, `delete_list_field`,
`start_automation`.

Service API documentation: https://emailoctopus.com/api-documentation.

## Auth setup

Connection fields:

- `api_key` (optional, secret, string); EmailOctopus API key, sent as the api_key query parameter on
  every request. Never logged.
- `base_url` (optional, string); default `https://emailoctopus.com/api/1.6`; format `uri`;
  EmailOctopus API base URL. Defaults to the production endpoint; override for test proxies.
- `list_id` (optional, string); EmailOctopus list ID to read contacts from; required only for the
  list_contacts stream, substituted into /lists/<list_id>/contacts.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://emailoctopus.com/api/1.6`.

Authentication behavior:

- API key authentication in query parameter `api_key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/lists`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `paging.next`; next
URLs stay on the configured API host.

Pagination by stream: next_url: `lists`, `campaigns`, `list_contacts`, `list_tags`; none:
`campaign_summary_reports`.

- `lists`: GET `/lists` - records path `data`; query `limit`=`100`; follows a next-page URL from the
  response body; URL path `paging.next`; next URLs stay on the configured API host; computed output
  fields `pending_count`, `subscribed_count`, `unsubscribed_count`.
- `campaigns`: GET `/campaigns` - records path `data`; query `limit`=`100`; follows a next-page URL
  from the response body; URL path `paging.next`; next URLs stay on the configured API host;
  computed output fields `from_email_address`, `from_name`.
- `list_contacts`: GET `/lists/{{ config.list_id }}/contacts` - records path `data`; query
  `limit`=`100`; follows a next-page URL from the response body; URL path `paging.next`; next URLs
  stay on the configured API host.
- `list_tags`: GET `/lists/{{ config.list_id }}/tags` - records path `data`; query `limit`=`100`;
  follows a next-page URL from the response body; URL path `paging.next`; next URLs stay on the
  configured API host.
- `campaign_summary_reports`: GET `/campaigns/{{ fanout.id }}/reports/summary` - records at response
  root; computed output fields `bounced_hard`, `bounced_soft`, `clicked_total`, `clicked_unique`,
  `opened_total`, `opened_unique`; fan-out; ids from request `/campaigns`; id-list records path
  `data`; id field `id`; id inserted into the request path; stamps `campaign_id`.

## Write actions & risks

Overall write risk: external EmailOctopus API mutations covering list/contact/tag/custom-field
lifecycle management, plus start_automation, which enrolls a contact into a live automation sequence
and triggers its configured email sends.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_list`: POST `/lists` - kind `create`; body type `json`; required record fields `name`;
  accepted fields `name`; risk: creates a new contact list; low-risk external mutation, no approval
  required.
- `update_list`: PUT `/lists/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  body fields `name`; required record fields `id`, `name`; accepted fields `id`, `name`; risk:
  renames an existing list; the id used by campaigns/API integrations to reference it is unchanged.
- `delete_list`: DELETE `/lists/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; risk: permanently removes a list and all of its contacts/tags/custom fields.
- `create_list_contact`: POST `/lists/{{ record.list_id }}/contacts` - kind `create`; body type
  `json`; path fields `list_id`; required record fields `list_id`, `email_address`; accepted fields
  `email_address`, `fields`, `list_id`, `status`, `tags`; risk: adds a new contact to a list,
  immediately eligible to receive future campaigns targeting it (unless status is PENDING on a
  double opt-in list).
- `update_list_contact`: PUT `/lists/{{ record.list_id }}/contacts/{{ record.member_id }}` - kind
  `update`; body type `json`; path fields `list_id`, `member_id`; required record fields `list_id`,
  `member_id`; accepted fields `email_address`, `fields`, `list_id`, `member_id`, `status`, `tags`;
  risk: mutates an existing contact's email/fields/tags/status; a status change to UNSUBSCRIBED or
  SUBSCRIBED changes future campaign eligibility for this recipient.
- `delete_list_contact`: DELETE `/lists/{{ record.list_id }}/contacts/{{ record.member_id }}` - kind
  `delete`; body type `none`; path fields `list_id`, `member_id`; required record fields `list_id`,
  `member_id`; accepted fields `list_id`, `member_id`; missing records treated as success for status
  `404`; risk: permanently removes a contact from a list and its subscription/consent history.
- `create_list_tag`: POST `/lists/{{ record.list_id }}/tags` - kind `create`; body type `json`; path
  fields `list_id`; body fields `tag`; required record fields `list_id`, `tag`; accepted fields
  `list_id`, `tag`; risk: creates a new tag on a list, up to that list's tag-count limit; low-risk
  external mutation, no approval required.
- `update_list_tag`: PUT `/lists/{{ record.list_id }}/tags/{{ record.tag }}` - kind `update`; body
  type `json`; path fields `list_id`, `tag`; body fields `new_tag`; required record fields
  `list_id`, `tag`, `new_tag`; accepted fields `list_id`, `new_tag`, `tag`; risk: renames an
  existing tag on a list; any external automation/segment referencing the old tag name stops
  matching contacts by that name.
- `delete_list_tag`: DELETE `/lists/{{ record.list_id }}/tags/{{ record.tag }}` - kind `delete`;
  body type `none`; path fields `list_id`, `tag`; required record fields `list_id`, `tag`; accepted
  fields `list_id`, `tag`; missing records treated as success for status `404`; risk: permanently
  removes a tag from a list and from every contact currently carrying it.
- `create_list_field`: POST `/lists/{{ record.list_id }}/fields` - kind `create`; body type `json`;
  path fields `list_id`; required record fields `list_id`, `label`, `tag`, `type`; accepted fields
  `fallback`, `label`, `list_id`, `tag`, `type`; risk: creates a new custom field on a list; the
  field's type (NUMBER/TEXT/DATE) cannot be changed after creation.
- `update_list_field`: PUT `/lists/{{ record.list_id }}/fields/{{ record.tag }}` - kind `update`;
  body type `json`; path fields `list_id`, `tag`; body fields `label`, `new_tag`, `fallback`;
  required record fields `list_id`, `tag`, `label`, `new_tag`; accepted fields `fallback`, `label`,
  `list_id`, `new_tag`, `tag`; risk: renames a custom field's label/tag or changes its fallback
  default; any email template referencing the old field tag stops resolving a value.
- `delete_list_field`: DELETE `/lists/{{ record.list_id }}/fields/{{ record.tag }}` - kind `delete`;
  body type `none`; path fields `list_id`, `tag`; required record fields `list_id`, `tag`; accepted
  fields `list_id`, `tag`; missing records treated as success for status `404`; risk: permanently
  removes a custom field and its stored values from every contact on the list.
- `start_automation`: POST `/automations/{{ record.automation_id }}/queue` - kind `create`; body
  type `json`; path fields `automation_id`; body fields `list_member_id`; required record fields
  `automation_id`, `list_member_id`; accepted fields `automation_id`, `list_member_id`; risk:
  enrolls a contact into a live automation sequence, triggering its configured emails/delays; the
  automation must already have the 'Started via API' trigger enabled in the EmailOctopus dashboard.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s), 13 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=6, out_of_scope=10.
