# Overview

Reads Picqer products, customers, orders, picklists, warehouses, suppliers, purchase orders,
returns, and warehouse-operations reference data, and writes order/purchase-order/return lifecycle
and catalog mutations through the Picqer REST API.

Readable streams: `products`, `customers`, `orders`, `picklists`, `warehouses`, `suppliers`, `tags`,
`purchaseorders`, `receipts`, `returns`, `return_statuses`, `return_reasons`, `backorders`,
`comments`, `stockhistory`, `users`, `product_fields`, `customer_fields`, `order_fields`,
`pricelists`, `shippingproviders`, `vatgroups`, `locations`, `location_types`, `picking_containers`,
`picklist_batches`, `shipments`, `packagings`, `packingstations`, `webshoporders`, `hooks`.

Write actions: `create_customer`, `update_customer`, `delete_customer`, `create_supplier`,
`update_supplier`, `create_tag`, `update_tag`, `delete_tag`, `update_product`, `pause_order`,
`resume_order`, `reopen_order`, `cancel_order`, `create_purchaseorder`,
`mark_purchaseorder_as_purchased`, `close_purchaseorder`, `cancel_purchaseorder`, `create_receipt`,
`complete_receipt`, `create_return`, `update_return`, `delete_return`, `process_backorders`,
`create_location`, `update_location`, `delete_location`, `create_location_type`,
`update_location_type`, `create_picking_container`, `update_picking_container`,
`create_picklist_batch`, `create_shipment`, `create_packaging`, `update_packaging`, `create_hook`,
`delete_hook`.

Service API documentation: https://picqer.com/en/api.

## Auth setup

Connection fields:

- `api_key` (optional, secret, string); Picqer API key, sent as the HTTP Basic auth username
  (password left blank). Never logged.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `organization_name` (required, string); Picqer subdomain; the connector always derives
  https://<organization_name>.picqer.com/api/v1 from it.
- `page_size` (optional, string); default `100`; Records per page (1-100).
- `password` (optional, secret, string); Optional HTTP Basic auth password. Picqer's own API
  convention leaves this blank.
- `username` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`, `password`.

Default configuration values: `max_pages=0`, `page_size=100`.

Authentication behavior:

- HTTP Basic authentication using `secrets.api_key`, `secrets.password` when `{{ secrets.api_key
  }}`.
- HTTP Basic authentication using `config.username`, `secrets.password`.

Requests use base URL `https://{{ config.organization_name }}.picqer.com/api/v1` after applying
configuration defaults.

Connection checks call GET `/products` with query `offset`=`0`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; page size 100.

- `products`: GET `/products` - records path `.`; offset/limit pagination; offset parameter
  `offset`; page size 100; computed output fields `id`; emits passthrough records.
- `customers`: GET `/customers` - records path `.`; offset/limit pagination; offset parameter
  `offset`; page size 100; computed output fields `id`; emits passthrough records.
- `orders`: GET `/orders` - records path `.`; offset/limit pagination; offset parameter `offset`;
  page size 100; computed output fields `id`; emits passthrough records.
- `picklists`: GET `/picklists` - records path `.`; offset/limit pagination; offset parameter
  `offset`; page size 100; computed output fields `id`; emits passthrough records.
- `warehouses`: GET `/warehouses` - records path `.`; offset/limit pagination; offset parameter
  `offset`; page size 100; computed output fields `id`; emits passthrough records.
- `suppliers`: GET `/suppliers` - records path `.`; offset/limit pagination; offset parameter
  `offset`; page size 100; computed output fields `id`; emits passthrough records.
- `tags`: GET `/tags` - records path `.`; offset/limit pagination; offset parameter `offset`; page
  size 100; computed output fields `id`; emits passthrough records.
- `purchaseorders`: GET `/purchaseorders` - records path `.`; offset/limit pagination; offset
  parameter `offset`; page size 100; computed output fields `id`; emits passthrough records.
- `receipts`: GET `/receipts` - records path `.`; offset/limit pagination; offset parameter
  `offset`; page size 100; computed output fields `id`; emits passthrough records.
- `returns`: GET `/returns` - records path `.`; offset/limit pagination; offset parameter `offset`;
  page size 100; computed output fields `id`; emits passthrough records.
- `return_statuses`: GET `/return_statuses` - records path `.`; offset/limit pagination; offset
  parameter `offset`; page size 100; computed output fields `id`; emits passthrough records.
- `return_reasons`: GET `/return_reasons` - records path `.`; offset/limit pagination; offset
  parameter `offset`; page size 100; computed output fields `id`; emits passthrough records.
- `backorders`: GET `/backorders` - records path `.`; offset/limit pagination; offset parameter
  `offset`; page size 100; computed output fields `id`; emits passthrough records.
- `comments`: GET `/comments` - records path `.`; offset/limit pagination; offset parameter
  `offset`; page size 100; computed output fields `id`; emits passthrough records.
- `stockhistory`: GET `/stockhistory` - records path `.`; offset/limit pagination; offset parameter
  `offset`; page size 100; computed output fields `id`; emits passthrough records.
- `users`: GET `/users` - records path `.`; offset/limit pagination; offset parameter `offset`; page
  size 100; computed output fields `id`; emits passthrough records.
- `product_fields`: GET `/productfields` - records path `.`; offset/limit pagination; offset
  parameter `offset`; page size 100; computed output fields `id`; emits passthrough records.
- `customer_fields`: GET `/customerfields` - records path `.`; offset/limit pagination; offset
  parameter `offset`; page size 100; computed output fields `id`; emits passthrough records.
- `order_fields`: GET `/orderfields` - records path `.`; offset/limit pagination; offset parameter
  `offset`; page size 100; computed output fields `id`; emits passthrough records.
- `pricelists`: GET `/pricelists` - records path `.`; offset/limit pagination; offset parameter
  `offset`; page size 100; computed output fields `id`; emits passthrough records.
- `shippingproviders`: GET `/shippingproviders` - records path `.`; offset/limit pagination; offset
  parameter `offset`; page size 100; computed output fields `id`; emits passthrough records.
- `vatgroups`: GET `/vatgroups` - records path `.`; offset/limit pagination; offset parameter
  `offset`; page size 100; computed output fields `id`; emits passthrough records.
- `locations`: GET `/locations` - records path `.`; offset/limit pagination; offset parameter
  `offset`; page size 100; computed output fields `id`; emits passthrough records.
- `location_types`: GET `/location_types` - records path `.`; offset/limit pagination; offset
  parameter `offset`; page size 100; computed output fields `id`; emits passthrough records.
- `picking_containers`: GET `/picking-containers` - records path `.`; offset/limit pagination;
  offset parameter `offset`; page size 100; computed output fields `id`; emits passthrough records.
- `picklist_batches`: GET `/picklists/batches` - records path `.`; offset/limit pagination; offset
  parameter `offset`; page size 100; computed output fields `id`; emits passthrough records.
- `shipments`: GET `/shipments` - records path `.`; offset/limit pagination; offset parameter
  `offset`; page size 100; computed output fields `id`; emits passthrough records.
- `packagings`: GET `/packagings` - records path `.`; offset/limit pagination; offset parameter
  `offset`; page size 100; computed output fields `id`; emits passthrough records.
- `packingstations`: GET `/packingstations` - records path `.`; offset/limit pagination; offset
  parameter `offset`; page size 100; computed output fields `id`; emits passthrough records.
- `webshoporders`: GET `/webshoporders` - records path `.`; offset/limit pagination; offset
  parameter `offset`; page size 100; computed output fields `id`; emits passthrough records.
- `hooks`: GET `/hooks` - records path `.`; offset/limit pagination; offset parameter `offset`; page
  size 100; computed output fields `id`; emits passthrough records.

## Write actions & risks

Overall write risk: creates/updates/deletes customers, suppliers, tags, locations, purchase orders,
returns, and warehouse-operations records, and mutates order fulfillment lifecycle
(pause/resume/reopen/cancel).

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_customer`: POST `/customers` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `contactname`, `emailaddress`, `language`, `name`, `telephone`,
  `vatnumber`; risk: creates a new WMS customer record; low-risk external mutation, no approval
  required.
- `update_customer`: PUT `/customers/{{ record.idcustomer }}` - kind `update`; body type `json`;
  path fields `idcustomer`; required record fields `idcustomer`; accepted fields `contactname`,
  `emailaddress`, `idcustomer`, `name`, `telephone`; risk: updates an existing customer's contact
  details; external mutation, approval required.
- `delete_customer`: DELETE `/customers/{{ record.idcustomer }}` - kind `delete`; body type `none`;
  path fields `idcustomer`; required record fields `idcustomer`; accepted fields `idcustomer`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: permanently
  deletes a customer record; destructive external mutation, approval required.
- `create_supplier`: POST `/suppliers` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `address`, `city`, `country`, `name`, `zipcode`; risk: creates a new
  supplier record; low-risk external mutation, no approval required.
- `update_supplier`: PUT `/suppliers/{{ record.idsupplier }}` - kind `update`; body type `json`;
  path fields `idsupplier`; required record fields `idsupplier`; accepted fields `address`, `city`,
  `country`, `idsupplier`, `name`; risk: updates an existing supplier's contact details; external
  mutation, approval required.
- `create_tag`: POST `/tags` - kind `create`; body type `json`; required record fields `title`,
  `color`, `inherit`; accepted fields `color`, `inherit`, `title`; risk: creates a new tag; low-risk
  external mutation, no approval required.
- `update_tag`: PUT `/tags/{{ record.idtag }}` - kind `update`; body type `json`; path fields
  `idtag`; required record fields `idtag`; accepted fields `color`, `idtag`, `inherit`, `title`;
  risk: updates an existing tag's title/color/inherit setting; external mutation, approval required.
- `delete_tag`: DELETE `/tags/{{ record.idtag }}` - kind `delete`; body type `none`; path fields
  `idtag`; required record fields `idtag`; accepted fields `idtag`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: permanently deletes a tag and removes
  it from every linked order/customer/product; destructive external mutation, approval required.
- `update_product`: PUT `/products/{{ record.idproduct }}` - kind `update`; body type `json`; path
  fields `idproduct`; required record fields `idproduct`; accepted fields `active`, `idproduct`,
  `name`, `price`, `productcode`; risk: updates an existing product's catalog fields
  (name/price/active status); external mutation, approval required.
- `pause_order`: POST `/orders/{{ record.idorder }}/pause` - kind `update`; body type `none`; path
  fields `idorder`; required record fields `idorder`; accepted fields `idorder`; risk: pauses
  picking/fulfillment of an order; external mutation, approval required.
- `resume_order`: POST `/orders/{{ record.idorder }}/resume` - kind `update`; body type `none`; path
  fields `idorder`; required record fields `idorder`; accepted fields `idorder`; risk: resumes
  picking/fulfillment of a paused order; external mutation, approval required.
- `reopen_order`: POST `/orders/{{ record.idorder }}/reopen` - kind `update`; body type `none`; path
  fields `idorder`; required record fields `idorder`; accepted fields `idorder`; risk: reopens a
  completed/closed order for further processing; external mutation, approval required.
- `cancel_order`: DELETE `/orders/{{ record.idorder }}` - kind `delete`; body type `none`; path
  fields `idorder`; required record fields `idorder`; accepted fields `idorder`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: cancels an order (Picqer's
  DELETE-shaped cancel, reversible via undo-cancellation, but stops all further fulfillment
  immediately); destructive external mutation, approval required.
- `create_purchaseorder`: POST `/purchaseorders` - kind `create`; body type `json`; required record
  fields `idsupplier`, `idwarehouse`; accepted fields `idsupplier`, `idwarehouse`, `products`,
  `remarks`; risk: creates a new purchase order (concept status) with an optional initial product
  line list; low-risk external mutation, no approval required.
- `mark_purchaseorder_as_purchased`: POST `/purchaseorders/{{ record.idpurchaseorder
  }}/mark-as-purchased` - kind `update`; body type `none`; path fields `idpurchaseorder`; required
  record fields `idpurchaseorder`; accepted fields `idpurchaseorder`; risk: transitions a concept
  purchase order to purchased status, signalling it has been placed with the supplier; external
  mutation, approval required.
- `close_purchaseorder`: POST `/purchaseorders/{{ record.idpurchaseorder }}/close` - kind `update`;
  body type `none`; path fields `idpurchaseorder`; required record fields `idpurchaseorder`;
  accepted fields `idpurchaseorder`; risk: closes a purchase order, marking it as finished even if
  not all products were received; external mutation, approval required.
- `cancel_purchaseorder`: POST `/purchaseorders/{{ record.idpurchaseorder }}/cancel` - kind
  `update`; body type `none`; path fields `idpurchaseorder`; required record fields
  `idpurchaseorder`; accepted fields `idpurchaseorder`; confirmation `destructive`; risk: cancels a
  purchase order; destructive external mutation, approval required.
- `create_receipt`: POST `/receipts` - kind `create`; body type `json`; required record fields
  `idpurchaseorder`; accepted fields `idpurchaseorder`, `version`; risk: starts a new
  goods-receiving session against a purchase order (Picqer's v2 receipts API also accepts idsupplier
  in place of idpurchaseorder for supplier-only receiving; this action only models the
  idpurchaseorder-required shape, see docs.md Known limits); low-risk external mutation, no approval
  required.
- `complete_receipt`: PUT `/receipts/{{ record.idreceipt }}` - kind `update`; body type `json`; path
  fields `idreceipt`; body fields `status`; required record fields `idreceipt`, `status`; accepted
  fields `idreceipt`, `status`; risk: marks a goods-receiving session as complete (Picqer's
  documented PUT /receipts/{idreceipt} {"status": "completed"} shape), finalizing received stock
  quantities in the background; external mutation, approval required.
- `create_return`: POST `/returns` - kind `create`; body type `json`; required record fields `name`;
  accepted fields `address`, `city`, `country`, `emailaddress`, `idreturn_status`, `idtemplate`,
  `name`, `reference`, `zipcode`; risk: creates a new customer return record; low-risk external
  mutation, no approval required.
- `update_return`: PUT `/returns/{{ record.idreturn }}` - kind `update`; body type `json`; path
  fields `idreturn`; required record fields `idreturn`; accepted fields `address`, `idreturn`,
  `idreturn_status`, `name`; risk: updates an existing return's status/contact details; external
  mutation, approval required.
- `delete_return`: DELETE `/returns/{{ record.idreturn }}` - kind `delete`; body type `none`; path
  fields `idreturn`; required record fields `idreturn`; accepted fields `idreturn`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: permanently deletes a
  return record; destructive external mutation, approval required.
- `process_backorders`: POST `/backorders/process` - kind `custom`; body type `json`; accepted
  fields `idbackorders`; risk: processes one or more backorders for fulfillment now that stock is
  available; external mutation, approval required.
- `create_location`: POST `/locations` - kind `create`; body type `json`; required record fields
  `name`, `idwarehouse`; accepted fields `idwarehouse`, `name`, `parent_idlocation`; risk: creates a
  new warehouse storage location; low-risk external mutation, no approval required.
- `update_location`: PUT `/locations/{{ record.idlocation }}` - kind `update`; body type `json`;
  path fields `idlocation`; required record fields `idlocation`; accepted fields `idlocation`,
  `name`, `remarks`; risk: updates an existing warehouse location's name/remarks; external mutation,
  approval required.
- `delete_location`: DELETE `/locations/{{ record.idlocation }}` - kind `delete`; body type `none`;
  path fields `idlocation`; required record fields `idlocation`; accepted fields `idlocation`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: permanently
  deletes a warehouse storage location; destructive external mutation, approval required.
- `create_location_type`: POST `/location_types` - kind `create`; body type `json`; required record
  fields `name`, `color`; accepted fields `color`, `name`; risk: creates a new location type;
  low-risk external mutation, no approval required.
- `update_location_type`: PUT `/location_types/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`; accepted fields `color`, `id`, `name`; risk:
  updates an existing location type's name/color; external mutation, approval required.
- `create_picking_container`: POST `/picking-containers` - kind `create`; body type `json`; required
  record fields `name`; accepted fields `name`; risk: creates a new picking container; low-risk
  external mutation, no approval required.
- `update_picking_container`: PUT `/picking-containers/{{ record.idpicking_container }}` - kind
  `update`; body type `json`; path fields `idpicking_container`; required record fields
  `idpicking_container`; accepted fields `idpicking_container`, `name`; risk: updates an existing
  picking container's name; external mutation, approval required.
- `create_picklist_batch`: POST `/picklists/batches` - kind `create`; body type `json`; required
  record fields `idwarehouse`; accepted fields `idwarehouse`, `title`, `type`; risk: creates a new
  picklist batch for warehouse picking; low-risk external mutation, no approval required.
- `create_shipment`: POST `/picklists/{{ record.idpicklist }}/shipments` - kind `create`; body type
  `json`; path fields `idpicklist`; required record fields `idpicklist`,
  `idshippingprovider_profile`; accepted fields `idpackaging`, `idpicklist`,
  `idshippingprovider_profile`, `weight`; risk: creates a shipment for a picklist, booking it with
  the configured shipping provider and generating a shipping label; external mutation, approval
  required.
- `create_packaging`: POST `/packagings` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `barcode`, `height`, `length`, `name`, `width`; risk: creates a new
  packaging type; low-risk external mutation, no approval required.
- `update_packaging`: PUT `/packagings/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `active`, `id`, `name`; risk: updates an
  existing packaging type's dimensions/active status; external mutation, approval required.
- `create_hook`: POST `/hooks` - kind `create`; body type `json`; required record fields `name`,
  `event`, `address`; accepted fields `address`, `event`, `name`, `secret`; risk: registers a new
  webhook subscription that will receive event notifications; low-risk external mutation, no
  approval required.
- `delete_hook`: DELETE `/hooks/{{ record.idhook }}` - kind `delete`; body type `none`; path fields
  `idhook`; required record fields `idhook`; accepted fields `idhook`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: deactivates a webhook subscription
  (Picqer's DELETE call deactivates rather than hard-deletes the record); destructive external
  mutation, approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 31 stream-backed endpoint group(s), 36 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=2, deprecated=2, destructive_admin=5, duplicate_of=27, non_data_endpoint=11,
  out_of_scope=57.
