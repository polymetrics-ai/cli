# Overview

Reads Adobe Commerce (Magento) products, orders, customers, categories, invoices, shipments, credit
memos, customer groups, and store configuration through the Magento REST API, and writes
product/category updates plus order cancellation.

Readable streams: `products`, `orders`, `customers`, `categories`, `invoices`, `shipments`,
`creditmemos`, `customer_groups`, `store_websites`, `store_views`.

Write actions: `update_product`, `create_category`, `update_category`, `cancel_order`.

Service API documentation: https://developer.adobe.com/commerce/webapi/rest/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Adobe Commerce (Magento) Integration Access Token. Sent as a
  Bearer token; never logged.
- `base_url` (required, string); format `uri`; Full Magento REST base URL including the API version
  prefix, e.g. https://magento.mystore.com/rest/V1.
- `mode` (optional, string).
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; only objects updated at
  or after this time are read (filtered on updated_at).

Secret fields are redacted in logs and write previews: `api_key`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/products`.

## Streams notes

Default pagination: page-number pagination; page parameter `searchCriteria[currentPage]`; size
parameter `searchCriteria[pageSize]`; starts at 1; page size 100.

Pagination by stream: none: `store_websites`, `store_views`; page_number: `products`, `orders`,
`customers`, `categories`, `invoices`, `shipments`, `creditmemos`, `customer_groups`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `products`: GET `/products` - records path `items`; query
  `searchCriteria[filter_groups][0][filters][0][condition_type]` from template `{{
  incremental.lower_bound | const:gt }}`, omitted when absent;
  `searchCriteria[filter_groups][0][filters][0][field]` from template `{{ incremental.lower_bound |
  const:updated_at }}`, omitted when absent; `searchCriteria[filter_groups][0][filters][0][value]`
  from template `{{ incremental.lower_bound }}`, omitted when absent; page-number pagination; page
  parameter `searchCriteria[currentPage]`; size parameter `searchCriteria[pageSize]`; starts at 1;
  page size 100; incremental cursor `updated_at`; formatted as `rfc3339`; initial lower bound from
  `start_date`.
- `orders`: GET `/orders` - records path `items`; query
  `searchCriteria[filter_groups][0][filters][0][condition_type]` from template `{{
  incremental.lower_bound | const:gt }}`, omitted when absent;
  `searchCriteria[filter_groups][0][filters][0][field]` from template `{{ incremental.lower_bound |
  const:updated_at }}`, omitted when absent; `searchCriteria[filter_groups][0][filters][0][value]`
  from template `{{ incremental.lower_bound }}`, omitted when absent; page-number pagination; page
  parameter `searchCriteria[currentPage]`; size parameter `searchCriteria[pageSize]`; starts at 1;
  page size 100; incremental cursor `updated_at`; formatted as `rfc3339`; initial lower bound from
  `start_date`.
- `customers`: GET `/customers/search` - records path `items`; query
  `searchCriteria[filter_groups][0][filters][0][condition_type]` from template `{{
  incremental.lower_bound | const:gt }}`, omitted when absent;
  `searchCriteria[filter_groups][0][filters][0][field]` from template `{{ incremental.lower_bound |
  const:updated_at }}`, omitted when absent; `searchCriteria[filter_groups][0][filters][0][value]`
  from template `{{ incremental.lower_bound }}`, omitted when absent; page-number pagination; page
  parameter `searchCriteria[currentPage]`; size parameter `searchCriteria[pageSize]`; starts at 1;
  page size 100; incremental cursor `updated_at`; formatted as `rfc3339`; initial lower bound from
  `start_date`.
- `categories`: GET `/categories/list` - records path `items`; query
  `searchCriteria[filter_groups][0][filters][0][condition_type]` from template `{{
  incremental.lower_bound | const:gt }}`, omitted when absent;
  `searchCriteria[filter_groups][0][filters][0][field]` from template `{{ incremental.lower_bound |
  const:updated_at }}`, omitted when absent; `searchCriteria[filter_groups][0][filters][0][value]`
  from template `{{ incremental.lower_bound }}`, omitted when absent; page-number pagination; page
  parameter `searchCriteria[currentPage]`; size parameter `searchCriteria[pageSize]`; starts at 1;
  page size 100; incremental cursor `updated_at`; formatted as `rfc3339`; initial lower bound from
  `start_date`.
- `invoices`: GET `/invoices` - records path `items`; query
  `searchCriteria[filter_groups][0][filters][0][condition_type]` from template `{{
  incremental.lower_bound | const:gt }}`, omitted when absent;
  `searchCriteria[filter_groups][0][filters][0][field]` from template `{{ incremental.lower_bound |
  const:updated_at }}`, omitted when absent; `searchCriteria[filter_groups][0][filters][0][value]`
  from template `{{ incremental.lower_bound }}`, omitted when absent; page-number pagination; page
  parameter `searchCriteria[currentPage]`; size parameter `searchCriteria[pageSize]`; starts at 1;
  page size 100; incremental cursor `updated_at`; formatted as `rfc3339`; initial lower bound from
  `start_date`.
- `shipments`: GET `/shipments` - records path `items`; query
  `searchCriteria[filter_groups][0][filters][0][condition_type]` from template `{{
  incremental.lower_bound | const:gt }}`, omitted when absent;
  `searchCriteria[filter_groups][0][filters][0][field]` from template `{{ incremental.lower_bound |
  const:created_at }}`, omitted when absent; `searchCriteria[filter_groups][0][filters][0][value]`
  from template `{{ incremental.lower_bound }}`, omitted when absent; page-number pagination; page
  parameter `searchCriteria[currentPage]`; size parameter `searchCriteria[pageSize]`; starts at 1;
  page size 100; incremental cursor `created_at`; formatted as `rfc3339`; initial lower bound from
  `start_date`.
- `creditmemos`: GET `/creditmemos` - records path `items`; query
  `searchCriteria[filter_groups][0][filters][0][condition_type]` from template `{{
  incremental.lower_bound | const:gt }}`, omitted when absent;
  `searchCriteria[filter_groups][0][filters][0][field]` from template `{{ incremental.lower_bound |
  const:created_at }}`, omitted when absent; `searchCriteria[filter_groups][0][filters][0][value]`
  from template `{{ incremental.lower_bound }}`, omitted when absent; page-number pagination; page
  parameter `searchCriteria[currentPage]`; size parameter `searchCriteria[pageSize]`; starts at 1;
  page size 100; incremental cursor `created_at`; formatted as `rfc3339`; initial lower bound from
  `start_date`.
- `customer_groups`: GET `/customerGroups/search` - records path `items`; query
  `searchCriteria[pageSize]`=`100`; page-number pagination; page parameter
  `searchCriteria[currentPage]`; size parameter `searchCriteria[pageSize]`; starts at 1; page size
  100.
- `store_websites`: GET `/store/websites` - records path `.`.
- `store_views`: GET `/store/storeViews` - records path `.`.

## Write actions & risks

Overall write risk: external mutation of live Magento catalog products/categories and cancellation
of live sales orders; approval required for every write action.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `update_product`: PUT `/products/{{ record.sku }}` - kind `update`; body type `json`; path fields
  `sku`; required record fields `sku`; accepted fields `name`, `price`, `sku`, `status`,
  `visibility`, `weight`; risk: external mutation; overwrites live Magento catalog product fields;
  approval required.
- `create_category`: POST `/categories` - kind `create`; body type `json`; required record fields
  `name`, `parent_id`; accepted fields `is_active`, `name`, `parent_id`, `position`; risk: external
  mutation; creates a live Magento catalog category; approval required.
- `update_category`: PUT `/categories/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `id`, `is_active`, `name`, `position`;
  risk: external mutation; overwrites live Magento catalog category fields; approval required.
- `cancel_order`: POST `/orders/{{ record.entity_id }}/cancel` - kind `update`; body type `none`;
  path fields `entity_id`; required record fields `entity_id`; accepted fields `entity_id`; risk:
  external mutation; irreversibly cancels a live Magento sales order; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 10 stream-backed endpoint group(s), 4 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, destructive_admin=3, duplicate_of=10, non_data_endpoint=6, out_of_scope=38,
  requires_elevated_scope=4.
