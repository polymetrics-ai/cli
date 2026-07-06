# Overview

Reads and writes Circa events, contacts, companies, teams, custom fields, and event/company
sub-resources through the Circa REST API.

Readable streams: `events`, `contacts`, `companies`, `teams`, `fields`, `event_contacts`,
`event_staff`, `event_expenses`, `company_contacts`.

Write actions: `create_contact`, `update_contact`, `delete_contact`, `create_event`, `update_event`,
`delete_event`, `create_company`, `update_company`, `delete_company`, `add_event_contact`,
`update_event_contact`, `remove_event_contact`.

Service API documentation: https://docs.circa.co/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Circa API key, sent as a Bearer token. Never logged.
- `base_url` (optional, string); default `https://app.circa.co/api/v1`; format `uri`; Circa API base
  URL override for tests or proxies.
- `mode` (optional, string).
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; only
  events/contacts/companies updated at or after this time are read on a fresh sync.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://app.circa.co/api/v1`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/teams` with query `page`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; no page-size parameter; starts at
1; page size 25.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `events`: GET `/events` - records path `data`; page-number pagination; page parameter `page`; no
  page-size parameter; starts at 1; page size 25; incremental cursor `updated_at`; sent as
  `updated_at[min]`; formatted as `rfc3339`; initial lower bound from `start_date`.
- `contacts`: GET `/contacts` - records path `data`; page-number pagination; page parameter `page`;
  no page-size parameter; starts at 1; page size 25; incremental cursor `updated_at`; sent as
  `updated_at[min]`; formatted as `rfc3339`; initial lower bound from `start_date`.
- `companies`: GET `/companies` - records path `data`; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 1; page size 25; incremental cursor `updated_at`; sent
  as `updated_at[min]`; formatted as `rfc3339`; initial lower bound from `start_date`.
- `teams`: GET `/teams` - records path `data`; page-number pagination; page parameter `page`; no
  page-size parameter; starts at 1; page size 25.
- `fields`: GET `/fields` - records path `data`; page-number pagination; page parameter `page`; no
  page-size parameter; starts at 1; page size 25.
- `event_contacts`: GET `/events/{{ fanout.id }}/contacts` - records path `data`; page-number
  pagination; page parameter `page`; no page-size parameter; starts at 1; page size 25; fan-out; ids
  from request `/events`; id-list records path `data`; id field `id`; id inserted into the request
  path; stamps `event_id`.
- `event_staff`: GET `/events/{{ fanout.id }}/staff` - records path `data`; page-number pagination;
  page parameter `page`; no page-size parameter; starts at 1; page size 25; fan-out; ids from
  request `/events`; id-list records path `data`; id field `id`; id inserted into the request path;
  stamps `event_id`.
- `event_expenses`: GET `/events/{{ fanout.id }}/expenses` - records path `data`; page-number
  pagination; page parameter `page`; no page-size parameter; starts at 1; page size 25; fan-out; ids
  from request `/events`; id-list records path `data`; id field `id`; id inserted into the request
  path; stamps `event_id`.
- `company_contacts`: GET `/companies/{{ fanout.id }}/contacts` - records path `data`; page-number
  pagination; page parameter `page`; no page-size parameter; starts at 1; page size 25; fan-out; ids
  from request `/companies`; id-list records path `data`; id field `id`; id inserted into the
  request path; stamps `company_id`.

## Write actions & risks

Overall write risk: external mutation of Circa contacts, events, companies, and event-contact
registrations; create/update/delete affect live CRM/event data an operator relies on.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_contact`: POST `/contacts` - kind `create`; body type `json`; required record fields
  `first_name`, `last_name`; accepted fields `address`, `city`, `company`, `country`, `description`,
  `email`, `email_opt_in`, `first_name`, `hot_lead`, `last_name`, `linkedin`, `mobile_phone`,
  `office_phone`, `owner`, `postal_index`, `state`, `title`, `twitter`, and 1 more; risk: external
  mutation; creates a new CRM contact record.
- `update_contact`: PATCH `/contacts/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `address`, `city`, `company`, `country`,
  `email`, `email_opt_in`, `first_name`, `hot_lead`, `id`, `last_name`, `mobile_phone`,
  `office_phone`, `owner`, `postal_index`, `state`, `title`, `website`; risk: external mutation;
  updates an existing CRM contact record.
- `delete_contact`: DELETE `/contacts/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: irreversible external deletion of a CRM contact; approval required.
- `create_event`: POST `/events` - kind `create`; body type `json`; required record fields `name`;
  accepted fields `end_date`, `leader`, `location`, `name`, `roles`, `start_date`, `status`,
  `team_id`, `types`, `website`; risk: external mutation; creates a new event record.
- `update_event`: PATCH `/events/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `end_date`, `id`, `name`, `start_date`,
  `status`, `team_id`, `website`; risk: external mutation; updates an existing event record.
- `delete_event`: DELETE `/events/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; risk: irreversible external deletion of an event; approval required.
- `create_company`: POST `/companies` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `address`, `city`, `country`, `name`, `state`, `website`, `zip_code`;
  risk: external mutation; creates a new company record.
- `update_company`: PATCH `/companies/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `address`, `city`, `country`, `id`,
  `name`, `state`, `website`, `zip_code`; risk: external mutation; updates an existing company
  record.
- `delete_company`: DELETE `/companies/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: irreversible external deletion of a company; approval required.
- `add_event_contact`: POST `/events/{{ record.event_id }}/contacts` - kind `create`; body type
  `json`; path fields `event_id`; required record fields `event_id`, `contact_id`; accepted fields
  `contact_id`, `event_id`; risk: external mutation; registers an existing contact onto an event.
- `update_event_contact`: PATCH `/events/{{ record.event_id }}/contacts/{{ record.contact_id }}` -
  kind `update`; body type `json`; path fields `event_id`, `contact_id`; required record fields
  `event_id`, `contact_id`; accepted fields `attendance_status`, `contact_id`, `event_id`,
  `registration_status`; risk: external mutation; updates an event-contact's attendance/registration
  status.
- `remove_event_contact`: DELETE `/events/{{ record.event_id }}/contacts/{{ record.contact_id }}` -
  kind `delete`; body type `none`; path fields `event_id`, `contact_id`; required record fields
  `event_id`, `contact_id`; accepted fields `contact_id`, `event_id`; missing records treated as
  success for status `404`; risk: irreversible external removal of a contact's event registration;
  approval required.

## Known limits

- Batch defaults: read_page_size=25.
- API coverage includes 9 stream-backed endpoint group(s), 12 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=6, out_of_scope=3.
