# Overview

Reads and writes EasyPost shipping resources including shipments, trackers, addresses, parcels,
batches, events, claims, pickups, refunds, scan forms, end shippers, users, and webhooks through the
EasyPost REST API.

Readable streams: `shipments`, `trackers`, `addresses`, `parcels`, `insurances`, `batches`,
`carrier_accounts`, `carrier_metadata`, `carrier_types`, `end_shippers`, `events`, `claims`,
`pickups`, `refunds`, `scan_forms`, `child_users`, `referral_customers`, `webhooks`.

Write actions: `create_address`, `create_and_verify_address`, `create_parcel`,
`create_customs_item`, `create_customs_info`, `create_shipment`, `buy_shipment`, `rerate_shipment`,
`insure_shipment`, `refund_shipment`, `create_shipment_form`, `create_tracker`, `delete_tracker`,
`create_batch`, `add_shipments_to_batch`, `remove_shipments_from_batch`, `buy_batch`, `label_batch`,
`create_batch_scan_form`, `create_end_shipper`, `update_end_shipper`, `create_insurance`,
`refund_insurance`, `create_order`, `cancel_claim`, `buy_order`, `create_pickup`, `buy_pickup`,
`cancel_pickup`, `create_refund`, `create_scan_form`, `create_report`, `create_luma_promise`,
`create_luma_shipment`, `buy_luma_shipment`, `create_child_user`, `update_user`,
`delete_child_user`, `create_webhook`, `update_webhook`, `delete_webhook`.

Service API documentation: https://docs.easypost.com/docs/shipments.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.easypost.com/v2`; format `uri`; EasyPost API
  base URL override for tests or proxies.
- `mode` (optional, string).
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; only objects created at
  or after this time are read on a fresh sync (a persisted cursor takes precedence on repeat syncs).
- `username` (required, secret, string); EasyPost API key, sent as the HTTP Basic username with an
  empty password. Never logged.

Secret fields are redacted in logs and write previews: `username`.

Default configuration values: `base_url=https://api.easypost.com/v2`.

Authentication behavior:

- HTTP Basic authentication using `secrets.username`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/shipments` with query `page_size`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `before_id`; next cursor from last record
field `id`; stop flag `has_more`.

Pagination by stream: cursor: `shipments`, `trackers`, `addresses`, `parcels`, `insurances`,
`batches`, `end_shippers`, `events`, `claims`, `pickups`, `refunds`, `scan_forms`, `child_users`,
`referral_customers`; none: `carrier_accounts`, `carrier_metadata`, `carrier_types`, `webhooks`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `shipments`: GET `/shipments` - records path `shipments`; query `page_size`=`100`; cursor
  pagination; cursor parameter `before_id`; next cursor from last record field `id`; stop flag
  `has_more`; incremental cursor `created_at`; sent as `start_datetime`; formatted as `rfc3339`;
  initial lower bound from `start_date`.
- `trackers`: GET `/trackers` - records path `trackers`; query `page_size`=`100`; cursor pagination;
  cursor parameter `before_id`; next cursor from last record field `id`; stop flag `has_more`;
  incremental cursor `created_at`; sent as `start_datetime`; formatted as `rfc3339`; initial lower
  bound from `start_date`.
- `addresses`: GET `/addresses` - records path `addresses`; query `page_size`=`100`; cursor
  pagination; cursor parameter `before_id`; next cursor from last record field `id`; stop flag
  `has_more`; incremental cursor `created_at`; sent as `start_datetime`; formatted as `rfc3339`;
  initial lower bound from `start_date`.
- `parcels`: GET `/parcels` - records path `parcels`; query `page_size`=`100`; cursor pagination;
  cursor parameter `before_id`; next cursor from last record field `id`; stop flag `has_more`;
  incremental cursor `created_at`; sent as `start_datetime`; formatted as `rfc3339`; initial lower
  bound from `start_date`.
- `insurances`: GET `/insurances` - records path `insurances`; query `page_size`=`100`; cursor
  pagination; cursor parameter `before_id`; next cursor from last record field `id`; stop flag
  `has_more`; incremental cursor `created_at`; sent as `start_datetime`; formatted as `rfc3339`;
  initial lower bound from `start_date`.
- `batches`: GET `/batches` - records path `batches`; query `page_size`=`100`; cursor pagination;
  cursor parameter `before_id`; next cursor from last record field `id`; stop flag `has_more`;
  incremental cursor `created_at`; sent as `start_datetime`; formatted as `rfc3339`; initial lower
  bound from `start_date`; emits passthrough records.
- `carrier_accounts`: GET `/carrier_accounts` - records path `.`; emits passthrough records.
- `carrier_metadata`: GET `/metadata/carriers` - records path `carriers`; emits passthrough records.
- `carrier_types`: GET `/carrier_types` - records path `.`; emits passthrough records.
- `end_shippers`: GET `/end_shippers` - records path `end_shippers`; query `page_size`=`100`; cursor
  pagination; cursor parameter `before_id`; next cursor from last record field `id`; stop flag
  `has_more`; emits passthrough records.
- `events`: GET `/events` - records path `events`; query `page_size`=`100`; cursor pagination;
  cursor parameter `before_id`; next cursor from last record field `id`; stop flag `has_more`;
  incremental cursor `created_at`; sent as `start_datetime`; formatted as `rfc3339`; initial lower
  bound from `start_date`; emits passthrough records.
- `claims`: GET `/claims` - records path `claims`; query `page_size`=`100`; cursor pagination;
  cursor parameter `before_id`; next cursor from last record field `id`; stop flag `has_more`;
  incremental cursor `created_at`; sent as `start_datetime`; formatted as `rfc3339`; initial lower
  bound from `start_date`; emits passthrough records.
- `pickups`: GET `/pickups` - records path `pickups`; query `page_size`=`100`; cursor pagination;
  cursor parameter `before_id`; next cursor from last record field `id`; stop flag `has_more`;
  incremental cursor `created_at`; sent as `start_datetime`; formatted as `rfc3339`; initial lower
  bound from `start_date`; emits passthrough records.
- `refunds`: GET `/refunds` - records path `refunds`; query `page_size`=`100`; cursor pagination;
  cursor parameter `before_id`; next cursor from last record field `id`; stop flag `has_more`;
  incremental cursor `created_at`; sent as `start_datetime`; formatted as `rfc3339`; initial lower
  bound from `start_date`; emits passthrough records.
- `scan_forms`: GET `/scan_forms` - records path `scan_forms`; query `page_size`=`100`; cursor
  pagination; cursor parameter `before_id`; next cursor from last record field `id`; stop flag
  `has_more`; incremental cursor `created_at`; sent as `start_datetime`; formatted as `rfc3339`;
  initial lower bound from `start_date`; emits passthrough records.
- `child_users`: GET `/users/children` - records path `children`; query `page_size`=`100`; cursor
  pagination; cursor parameter `after_id`; next cursor from last record field `id`; stop flag
  `has_more`; emits passthrough records.
- `referral_customers`: GET `/referral_customers` - records path `referral_customers`; query
  `page_size`=`100`; cursor pagination; cursor parameter `before_id`; next cursor from last record
  field `id`; stop flag `has_more`; emits passthrough records.
- `webhooks`: GET `/webhooks` - records path `webhooks`; emits passthrough records.

## Write actions & risks

Overall write risk: external EasyPost API mutation of live shipping objects, labels, pickups,
insurance/refund workflows, reports, users, and webhooks; buy/refund/insurance/pickup/order/Luma
actions may incur charges or alter operational shipping state.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_address`: POST `/addresses` - kind `create`; body type `json`; required record fields
  `address`; accepted fields `address`; risk: creates a reusable EasyPost Address object; low-risk
  external mutation, approval required.
- `create_and_verify_address`: POST `/addresses/create_and_verify` - kind `create`; body type
  `json`; required record fields `address`; accepted fields `address`; risk: creates and verifies an
  EasyPost Address object; may return verification failures but does not buy postage, approval
  required.
- `create_parcel`: POST `/parcels` - kind `create`; body type `json`; required record fields
  `parcel`; accepted fields `parcel`; risk: creates a Parcel object describing package dimensions
  and weight; low-risk external mutation, approval required.
- `create_customs_item`: POST `/customs_items` - kind `create`; body type `json`; required record
  fields `customs_item`; accepted fields `customs_item`; risk: creates a CustomsItem declaration
  object used by international shipments; approval required.
- `create_customs_info`: POST `/customs_infos` - kind `create`; body type `json`; required record
  fields `customs_info`; accepted fields `customs_info`; risk: creates a CustomsInfo declaration
  object used by international shipments; approval required.
- `create_shipment`: POST `/shipments` - kind `create`; body type `json`; required record fields
  `shipment`; accepted fields `shipment`; risk: creates and rates a Shipment object; does not
  purchase postage by itself, approval required.
- `buy_shipment`: POST `/shipments/{{ record.id }}/buy` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `rate`; accepted fields `id`, `insurance`, `rate`;
  confirmation `destructive`; risk: purchases a live postage label for an existing Shipment and may
  incur carrier/account charges; approval required.
- `rerate_shipment`: POST `/shipments/{{ record.id }}/rerate` - kind `update`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: refreshes rates on an
  existing Shipment; external mutation of rated shipment state, approval required.
- `insure_shipment`: POST `/shipments/{{ record.id }}/insure` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`, `amount`; accepted fields `amount`, `id`;
  confirmation `destructive`; risk: adds shipping insurance to an existing Shipment and may incur a
  charge; approval required.
- `refund_shipment`: POST `/shipments/{{ record.id }}/refund` - kind `update`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; confirmation `destructive`;
  risk: requests a refund for an existing Shipment label; changes shipment refund state, approval
  required.
- `create_shipment_form`: POST `/shipments/{{ record.id }}/forms` - kind `create`; body type `json`;
  path fields `id`; required record fields `id`, `form`; accepted fields `form`, `id`; risk: creates
  a shipment-associated form/document metadata object; approval required.
- `create_tracker`: POST `/trackers` - kind `create`; body type `json`; required record fields
  `tracker`; accepted fields `tracker`; risk: creates a Tracker for a carrier tracking code;
  low-risk external mutation, approval required.
- `delete_tracker`: DELETE `/trackers/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: deletes an EasyPost Tracker object;
  destructive external mutation, approval required.
- `create_batch`: POST `/batches` - kind `create`; body type `json`; required record fields `batch`;
  accepted fields `batch`; risk: creates a Batch grouping shipments; does not buy postage by itself,
  approval required.
- `add_shipments_to_batch`: POST `/batches/{{ record.id }}/add_shipments` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`, `shipments`; accepted fields `id`,
  `shipments`; risk: adds Shipment references to an existing Batch; approval required.
- `remove_shipments_from_batch`: POST `/batches/{{ record.id }}/remove_shipments` - kind `update`;
  body type `json`; path fields `id`; required record fields `id`, `shipments`; accepted fields
  `id`, `shipments`; risk: removes Shipment references from an existing Batch; approval required.
- `buy_batch`: POST `/batches/{{ record.id }}/buy` - kind `update`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; confirmation `destructive`; risk:
  purchases postage for all eligible shipments in a Batch and may incur charges; approval required.
- `label_batch`: POST `/batches/{{ record.id }}/label` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `file_format`, `id`; risk: generates a
  batch label file after purchase; external mutation of batch artifact state, approval required.
- `create_batch_scan_form`: POST `/batches/{{ record.id }}/scan_form` - kind `create`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: creates a
  ScanForm for an existing Batch; approval required.
- `create_end_shipper`: POST `/end_shippers` - kind `create`; body type `json`; required record
  fields `end_shipper`; accepted fields `end_shipper`; risk: creates an EndShipper sender
  identity/address record; approval required.
- `update_end_shipper`: PUT `/end_shippers/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `end_shipper`; accepted fields `end_shipper`, `id`;
  risk: updates an EndShipper sender identity/address record; approval required.
- `create_insurance`: POST `/insurances` - kind `create`; body type `json`; required record fields
  `insurance`; accepted fields `insurance`; confirmation `destructive`; risk: creates standalone
  shipping insurance and may incur a charge; approval required.
- `refund_insurance`: POST `/insurances/{{ record.id }}/refund` - kind `update`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; confirmation `destructive`;
  risk: requests a refund for standalone insurance; approval required.
- `create_order`: POST `/orders` - kind `create`; body type `json`; required record fields `order`;
  accepted fields `order`; risk: creates an Order grouping multiple shipments; does not buy postage
  by itself, approval required.
- `cancel_claim`: POST `/claims/{{ record.id }}/cancel` - kind `update`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; confirmation `destructive`; risk:
  cancels an existing EasyPost insurance claim; external claim workflow mutation, approval required.
- `buy_order`: POST `/orders/{{ record.id }}/buy` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `carrier`, `service`; accepted fields `carrier`, `id`,
  `service`; confirmation `destructive`; risk: purchases postage for an Order and may incur charges;
  approval required.
- `create_pickup`: POST `/pickups` - kind `create`; body type `json`; required record fields
  `pickup`; accepted fields `pickup`; risk: creates a carrier pickup request for a shipment/address
  window; approval required.
- `buy_pickup`: POST `/pickups/{{ record.id }}/buy` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `carrier`, `service`; accepted fields `carrier`, `id`,
  `service`; confirmation `destructive`; risk: buys/schedules a carrier pickup and may incur carrier
  charges; approval required.
- `cancel_pickup`: POST `/pickups/{{ record.id }}/cancel` - kind `update`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; confirmation `destructive`; risk:
  cancels a scheduled Pickup; external operational mutation, approval required.
- `create_refund`: POST `/refunds` - kind `create`; body type `json`; required record fields
  `refund`; accepted fields `refund`; confirmation `destructive`; risk: creates one or more shipment
  refund requests; approval required.
- `create_scan_form`: POST `/scan_forms` - kind `create`; body type `json`; required record fields
  `shipments`; accepted fields `shipments`; risk: creates a ScanForm manifest for shipment IDs;
  approval required.
- `create_report`: POST `/reports/{{ record.type }}` - kind `create`; body type `json`; path fields
  `type`; required record fields `type`, `start_date`, `end_date`; accepted fields `end_date`,
  `start_date`, `type`; risk: starts an asynchronous EasyPost report export for the requested report
  type/date range; approval required.
- `create_luma_promise`: POST `/luma/promise` - kind `custom`; body type `json`; required record
  fields `shipment`; accepted fields `shipment`; risk: requests a Luma delivery promise/rating
  calculation; no label purchase by itself, approval required.
- `create_luma_shipment`: POST `/shipments/luma` - kind `create`; body type `json`; required record
  fields `shipment`; accepted fields `shipment`; confirmation `destructive`; risk: creates and buys
  a Shipment through Luma one-call buy and may incur postage charges; approval required.
- `buy_luma_shipment`: POST `/shipments/{{ record.id }}/luma` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`, `ruleset_name`; accepted fields `deliver_by_date`,
  `id`, `planned_ship_date`, `ruleset_name`; confirmation `destructive`; risk: buys postage for an
  existing Shipment through Luma and may incur charges; approval required.
- `create_child_user`: POST `/users` - kind `create`; body type `json`; required record fields
  `user`; accepted fields `user`; risk: creates a production-only child user/sub-account under the
  authenticated EasyPost account; elevated account-management mutation, approval required.
- `update_user`: PATCH `/users/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`, `user`; accepted fields `id`, `user`; risk: updates EasyPost
  user/sub-account settings such as child account name; elevated account-management mutation,
  approval required.
- `delete_child_user`: DELETE `/users/{{ record.child_id }}` - kind `delete`; body type `none`; path
  fields `child_id`; required record fields `child_id`; accepted fields `child_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: removes a child user from
  the parent account; destructive account-management mutation, approval required.
- `create_webhook`: POST `/webhooks` - kind `create`; body type `json`; required record fields
  `webhook`; accepted fields `webhook`; risk: registers an outbound Webhook URL/custom headers for
  EasyPost events; approval required.
- `update_webhook`: PATCH `/webhooks/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `custom_headers`, `id`; risk: updates Webhook
  delivery metadata such as custom headers; approval required.
- `delete_webhook`: DELETE `/webhooks/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: deletes a Webhook subscription, stopping
  outbound event delivery; destructive external mutation, approval required.

## Known limits

- Batch defaults: read_page_size=100, write_batch_size=1.
- API coverage includes 18 stream-backed endpoint group(s), 41 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=4, destructive_admin=8, duplicate_of=15, non_data_endpoint=7, out_of_scope=5,
  requires_elevated_scope=14.
