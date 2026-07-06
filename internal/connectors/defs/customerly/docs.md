# Overview

Reads Customerly users, leads, and accounts, and writes user/lead/tag/message/attribute/company
mutations through the Customerly v1 REST API.

Readable streams: `users`, `leads`, `accounts`.

Write actions: `delete_user`, `delete_lead`, `unsubscribe_user`, `add_tag`, `delete_tag`,
`send_message`, `add_user_attributes`, `add_company_attributes`, `add_user_to_company`.

Service API documentation: https://docs.customerly.io.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Customerly API key, sent as a Bearer token. Never logged.
- `base_url` (optional, string); default `https://api.customerly.io/v1`; format `uri`; Customerly
  API base URL override for tests or proxies.
- `mode` (optional, string).
- `page_size` (optional, string); default `50`; Records per page (1-100).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.customerly.io/v1`, `page_size=50`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/users/list` with query `page`=`0`; `per_page`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 0; page size 50.

Pagination by stream: none: `accounts`; page_number: `users`, `leads`.

- `users`: GET `/users/list` - records path `data.users`; query `sort`=`last_update`;
  `sort_direction`=`desc`; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 0; page size 50.
- `leads`: GET `/leads/list` - records path `data.leads`; query `sort`=`last_update`;
  `sort_direction`=`desc`; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 0; page size 50.
- `accounts`: GET `/accounts/` - records path `data`.

## Write actions & risks

Overall write risk: external mutation of live Customerly
users/leads/tags/messages/attributes/companies, including irreversible user and lead deletion;
approval required.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `delete_user`: DELETE `/users?user_id={{ record.user_id }}` - kind `delete`; body type `none`;
  path fields `user_id`; required record fields `user_id`; accepted fields `user_id`; missing
  records treated as success for status `404`; risk: external mutation; irreversibly deletes a live
  Customerly user and every conversation/survey/campaign record tied to them; approval required.
- `delete_lead`: DELETE `/leads?email={{ record.email }}` - kind `delete`; body type `none`; path
  fields `email`; required record fields `email`; accepted fields `email`; missing records treated
  as success for status `404`; risk: external mutation; irreversibly deletes a live Customerly lead
  and every associated record; approval required.
- `unsubscribe_user`: POST `/users/unsubscribe/{{ record.user_id }}` - kind `update`; body type
  `none`; path fields `user_id`; required record fields `user_id`; accepted fields `user_id`; risk:
  external mutation; unsubscribes a live user from Customerly messaging; approval required.
- `add_tag`: POST `/tags` - kind `custom`; body type `json`; required record fields `tag`; accepted
  fields `leads`, `tag`, `untag`, `users`; risk: external mutation; adds or removes a tag on one or
  more live users/leads.
- `delete_tag`: DELETE `/tags` - kind `delete`; body type `json`; body fields `tag`; required record
  fields `tag`; accepted fields `tag`; risk: external mutation; permanently removes a tag definition
  from the app; it is un-applied from every contact that carried it; approval required.
- `send_message`: POST `/messages` - kind `create`; body type `json`; required record fields `from`,
  `to`, `content`; accepted fields `attachments`, `content`, `from`, `to`; risk: sends a
  user-visible message from Customerly on the sender's behalf; may notify the recipient.
- `add_user_attributes`: POST `/users/add-attributes/{{ record.user_id }}` - kind `update`; body
  type `json`; path fields `user_id`; required record fields `user_id`, `attributes`; accepted
  fields `attributes`, `user_id`; risk: external mutation; adds/overwrites custom attribute values
  on a live user.
- `add_company_attributes`: POST `/company/add-attributes/{{ record.company_id }}` - kind `update`;
  body type `json`; path fields `company_id`; required record fields `company_id`; accepted fields
  `attributes`, `company_id`, `company_name`; risk: external mutation; adds/overwrites custom
  attribute values (and optionally renames) a live company.
- `add_user_to_company`: POST `/users/add-to-company` - kind `custom`; body type `json`; required
  record fields `company_id`; accepted fields `company_attributes`, `company_id`, `company_name`,
  `email`, `internal_user_id`, `user_id`; risk: external mutation; links a live user to a company,
  creating the company if it does not already exist.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 3 stream-backed endpoint group(s), 9 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, duplicate_of=3, out_of_scope=14.
