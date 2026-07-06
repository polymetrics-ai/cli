# Overview

Reads Clazar cloud GTM data (buyers, listings, contracts, opportunities, private offers, reseller
offers, contacts, and metering records) and writes
buyer/opportunity/contract/private-offer/contact/metering mutations, contract activation, and
metering-record submission, through the Clazar REST API using OAuth2 client credentials.

Readable streams: `buyers`, `listings`, `contracts`, `opportunities`, `private_offers`,
`reseller_offers`, `contacts`, `metering`.

Write actions: `update_buyer`, `update_opportunity`, `update_private_offer`, `update_contract`,
`activate_contract`, `create_contact`, `update_contact`, `delete_contact`, `update_metering_record`,
`create_metering_records`.

Service API documentation: https://developers.clazar.io/reference.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.clazar.io`; format `uri`; Clazar API base URL
  override for tests or proxies; also used to derive the OAuth2 token endpoint (base_url +
  /authenticate/).
- `client_id` (required, secret, string); Clazar OAuth2 client id. Used only for the
  client_credentials token exchange; never logged.
- `client_secret` (required, secret, string); Clazar OAuth2 client secret. Used only for the
  client_credentials token exchange; never logged.
- `mode` (optional, string).
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; only objects last
  modified at or after this time are read on a fresh sync.

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`.

Default configuration values: `base_url=https://api.clazar.io`.

Authentication behavior:

- OAuth 2.0 client credentials authentication using `config.base_url`, `secrets.client_id`,
  `secrets.client_secret`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/listings`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `page_size`;
starts at 1; page size 100.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `buyers`: GET `/buyers` - records path `results`; query `response_format`=`common`; page-number
  pagination; page parameter `page`; size parameter `page_size`; starts at 1; page size 100;
  incremental cursor `last_modified_at`; sent as `last_modified_at_after`; formatted as `rfc3339`;
  initial lower bound from `start_date`.
- `listings`: GET `/listings` - records path `results`; query `response_format`=`common`;
  page-number pagination; page parameter `page`; size parameter `page_size`; starts at 1; page size
  100; incremental cursor `last_modified_at`; sent as `last_modified_at_after`; formatted as
  `rfc3339`; initial lower bound from `start_date`.
- `contracts`: GET `/contracts` - records path `results`; query `response_format`=`common`;
  page-number pagination; page parameter `page`; size parameter `page_size`; starts at 1; page size
  100; incremental cursor `last_modified_at`; sent as `last_modified_at_after`; formatted as
  `rfc3339`; initial lower bound from `start_date`.
- `opportunities`: GET `/opportunities` - records path `results`; query `response_format`=`common`;
  page-number pagination; page parameter `page`; size parameter `page_size`; starts at 1; page size
  100; incremental cursor `last_modified_at`; sent as `last_modified_at_after`; formatted as
  `rfc3339`; initial lower bound from `start_date`.
- `private_offers`: GET `/private_offers` - records path `results`; query
  `response_format`=`common`; page-number pagination; page parameter `page`; size parameter
  `page_size`; starts at 1; page size 100; incremental cursor `last_modified_at`; sent as
  `last_modified_at_after`; formatted as `rfc3339`; initial lower bound from `start_date`.
- `reseller_offers`: GET `/reseller_offers` - records path `results`; page-number pagination; page
  parameter `page`; size parameter `page_size`; starts at 1; page size 100.
- `contacts`: GET `/contacts` - records path `results`; page-number pagination; page parameter
  `page`; size parameter `page_size`; starts at 1; page size 100.
- `metering`: GET `/metering` - records path `results`; page-number pagination; page parameter
  `page`; size parameter `page_size`; starts at 1; page size 100.

## Write actions & risks

Overall write risk: external mutation of Clazar
buyer/opportunity/contract/private-offer/contact/metering-record data; activate_contract
irreversibly transitions a contract's state in the underlying cloud marketplace, and
create_metering_records submits usage data that drives marketplace billing - every write ships with
an explicit per-action risk string.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `update_buyer`: PATCH `/buyers/{{ record.id }}/` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `custom_properties`,
  `external_object_associations`, `id`; risk: external mutation of a buyer's custom properties /
  external-system associations; approval required.
- `update_opportunity`: PATCH `/opportunities/{{ record.id }}/` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`; accepted fields `custom_properties`,
  `external_object_associations`, `id`; risk: external mutation of an opportunity's custom
  properties / external-system associations; approval required.
- `update_private_offer`: PATCH `/private_offers/{{ record.id }}/` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `custom_properties`,
  `external_object_associations`, `id`; risk: external mutation of a private offer's custom
  properties / external-system associations; approval required.
- `update_contract`: PATCH `/contracts/{{ record.id }}/` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `custom_properties`,
  `external_object_associations`, `id`; risk: external mutation of a contract's custom properties /
  external-system associations; approval required.
- `activate_contract`: POST `/contracts/{{ record.id }}/activate/` - kind `update`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: irreversibly
  activates a pending Clazar contract in the underlying cloud marketplace; approval required
  (destructive/high-impact state transition).
- `create_contact`: POST `/contacts/` - kind `create`; body type `json`; accepted fields `email`,
  `full_name`, `phone_number`; risk: creates a new Clazar contact record; low-risk (no external
  marketplace side effects).
- `update_contact`: PATCH `/contacts/{{ record.id }}/` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `email`, `full_name`, `id`,
  `phone_number`; risk: updates a Clazar contact record; low-risk (no external marketplace side
  effects).
- `delete_contact`: DELETE `/contacts/{{ record.id }}/` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: permanently deletes a Clazar contact record; approval required
  (destructive, irreversible).
- `update_metering_record`: PATCH `/metering/{{ record.id }}/` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`, `custom_properties`; accepted fields
  `custom_properties`, `id`; risk: updates only the custom_properties of a submitted metering
  record; low-risk.
- `create_metering_records`: POST `/metering/` - kind `create`; body type `json`; body fields
  `request`; required record fields `request`; accepted fields `request`; risk: submits usage-based
  billing metering records that drive cloud marketplace invoicing for the buyer's contract; approval
  required (financial impact, effectively irreversible once billed).

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 8 stream-backed endpoint group(s), 10 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=8, non_data_endpoint=1, out_of_scope=2.
