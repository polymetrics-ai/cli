# Overview

Reads and writes Brevo (formerly Sendinblue) contacts, email campaigns, contact lists, segments,
senders, sender domains, CRM companies/deals, and webhooks through the Brevo REST API.

Readable streams: `contacts`, `email_campaigns`, `contacts_lists`, `senders`, `senders_domains`,
`contacts_segments`, `companies`, `crm_deals`, `webhooks`.

Write actions: `create_contact`, `update_contact`, `delete_contact`, `create_contacts_list`,
`create_sender`, `update_sender`, `delete_sender`, `create_company`, `update_company`,
`create_deal`, `update_deal`, `create_webhook`, `update_webhook`, `delete_webhook`.

Service API documentation: https://developers.brevo.com/reference.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Brevo API key. Sent as the api-key header; never logged.
- `base_url` (optional, string); default `https://api.brevo.com/v3`; format `uri`; Brevo API base
  URL override for tests or proxies.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-1000) for paginated endpoints.
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; only contacts/campaigns
  modified at or after this time are read (sent as modifiedSince).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.brevo.com/v3`, `max_pages=0`, `page_size=100`.

Authentication behavior:

- API key authentication in `api-key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/account`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `senders`, `senders_domains`, `webhooks`; offset_limit: `contacts`,
`email_campaigns`, `contacts_lists`, `contacts_segments`, `crm_deals`; page_number: `companies`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `contacts`: GET `/contacts` - records path `contacts`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; incremental cursor `modifiedAt`; sent as
  `modifiedSince`; formatted as `rfc3339`; initial lower bound from `start_date`.
- `email_campaigns`: GET `/emailCampaigns` - records path `campaigns`; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 100; incremental cursor
  `modifiedAt`; sent as `modifiedSince`; formatted as `rfc3339`; initial lower bound from
  `start_date`.
- `contacts_lists`: GET `/contacts/lists` - records path `lists`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100.
- `senders`: GET `/senders` - records path `senders`.
- `senders_domains`: GET `/senders/domains` - records path `domains`.
- `contacts_segments`: GET `/contacts/segments` - records path `segments`; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 100.
- `companies`: GET `/companies` - records path `items`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 100; incremental cursor `last_updated_at`;
  sent as `modifiedSince`; formatted as `rfc3339`; initial lower bound from `start_date`; computed
  output fields `last_updated_at`.
- `crm_deals`: GET `/crm/deals` - records path `items`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; incremental cursor `last_updated_date`; sent as
  `modifiedSince`; formatted as `rfc3339`; initial lower bound from `start_date`; computed output
  fields `last_updated_date`.
- `webhooks`: GET `/webhooks` - records path `webhooks`.

## Write actions & risks

Overall write risk: external mutation of contacts, contact lists, senders, CRM companies/deals, and
webhooks; webhook writes register live event delivery to a caller-chosen URL.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_contact`: POST `/contacts` - kind `create`; body type `json`; accepted fields
  `attributes`, `email`, `emailBlacklisted`, `ext_id`, `listIds`, `smsBlacklisted`, `updateEnabled`;
  risk: creates a new marketing contact; low-risk external mutation, no approval required.
- `update_contact`: PUT `/contacts/{{ record.identifier }}` - kind `update`; body type `json`; path
  fields `identifier`; required record fields `identifier`; accepted fields `attributes`,
  `emailBlacklisted`, `ext_id`, `identifier`, `listIds`, `smsBlacklisted`, `unlinkListIds`; risk:
  mutates an existing contact's attributes, list membership, or blacklist status; changing
  emailBlacklisted/smsBlacklisted affects real send eligibility.
- `delete_contact`: DELETE `/contacts/{{ record.identifier }}` - kind `delete`; body type `none`;
  path fields `identifier`; required record fields `identifier`; accepted fields `identifier`;
  missing records treated as success for status `404`; risk: permanently removes a contact and its
  engagement history; irreversible.
- `create_contacts_list`: POST `/contacts/lists` - kind `create`; body type `json`; required record
  fields `name`, `folderId`; accepted fields `folderId`, `name`; risk: creates a new contact list
  under an existing folder; low-risk external mutation, no approval required.
- `create_sender`: POST `/senders` - kind `create`; body type `json`; required record fields `name`,
  `email`; accepted fields `email`, `ips`, `name`; risk: registers a new verified-sending identity;
  Brevo emails a verification link to the address before it can send.
- `update_sender`: PUT `/senders/{{ record.senderId }}` - kind `update`; body type `json`; path
  fields `senderId`; required record fields `senderId`; accepted fields `email`, `ips`, `name`,
  `senderId`; risk: mutates an existing sender's from-name, email, or dedicated-IP pool; affects all
  campaigns using this sender going forward.
- `delete_sender`: DELETE `/senders/{{ record.senderId }}` - kind `delete`; body type `none`; path
  fields `senderId`; required record fields `senderId`; accepted fields `senderId`; missing records
  treated as success for status `404`; risk: permanently removes a sending identity; any scheduled
  campaign still referencing it will fail to send.
- `create_company`: POST `/companies` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `attributes`, `countryCode`, `linkedContactsIds`, `linkedDealsIds`,
  `name`; risk: creates a new CRM company record; low-risk external mutation, no approval required.
- `update_company`: PATCH `/companies/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `attributes`, `countryCode`, `id`,
  `linkedContactsIds`, `linkedDealsIds`, `name`; risk: mutates an existing CRM company's name,
  attributes, or linked contact/deal set.
- `create_deal`: POST `/crm/deals` - kind `create`; body type `json`; required record fields `name`;
  accepted fields `attributes`, `linkedCompaniesIds`, `linkedContactsIds`, `name`; risk: creates a
  new CRM deal record; low-risk external mutation, no approval required.
- `update_deal`: PATCH `/crm/deals/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `attributes`, `id`, `linkedCompaniesIds`,
  `linkedContactsIds`; risk: mutates an existing CRM deal's stage, amount, or linked contact/company
  set.
- `create_webhook`: POST `/webhooks` - kind `create`; body type `json`; required record fields
  `url`, `events`; accepted fields `description`, `events`, `type`, `url`; risk: registers live
  event delivery (opens/clicks/bounces/unsubscribes) to an external endpoint of the caller's
  choosing; review the target before enabling, per metadata.json risk.write.
- `update_webhook`: PUT `/webhooks/{{ record.webhookId }}` - kind `update`; body type `json`; path
  fields `webhookId`; required record fields `webhookId`; accepted fields `description`, `events`,
  `url`, `webhookId`; risk: re-points an already-registered webhook's delivery URL or event set;
  redirects live event delivery immediately.
- `delete_webhook`: DELETE `/webhooks/{{ record.webhookId }}` - kind `delete`; body type `none`;
  path fields `webhookId`; required record fields `webhookId`; accepted fields `webhookId`; missing
  records treated as success for status `404`; risk: permanently removes a webhook subscription;
  irreversible.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 9 stream-backed endpoint group(s), 14 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=4, destructive_admin=8, duplicate_of=21, non_data_endpoint=7, out_of_scope=183,
  requires_elevated_scope=32.
