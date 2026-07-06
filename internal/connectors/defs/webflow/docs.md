# Overview

Reads Webflow sites, collections, collection items, pages, forms, form submissions, assets, asset
folders, webhooks, redirects, custom domains, components, orders, products, and ecommerce settings,
and writes CMS collection-item lifecycle actions, form-submission hidden-field updates, asset
metadata, webhook subscriptions, and ecommerce order/inventory mutations, using the Webflow Data API
v2.

Readable streams: `collections`, `pages`, `forms`, `sites`, `assets`, `webhooks`, `redirects`,
`form_submissions`, `orders`, `products`, `custom_domains`, `components`, `asset_folders`,
`ecommerce_settings`, `collection_items`.

Write actions: `create_collection_item`, `update_collection_item`, `delete_collection_item`,
`publish_collection_item`, `update_form_submission`, `update_asset`, `delete_asset`,
`create_webhook`, `delete_webhook`, `update_order`, `fulfill_order`, `refund_order`,
`update_inventory`.

Service API documentation: https://developers.webflow.com/data/reference.

## Auth setup

Connection fields:

- `accept_version` (optional, string); Optional Webflow Data API version to send as the
  Accept-Version header (e.g. 1.0.0 or 2.0.0). Omitted entirely when unset.
- `api_key` (required, secret, string); Webflow API token (Site or OAuth access token). Used only
  for Bearer auth (Authorization: Bearer <api_key>); never logged.
- `base_url` (optional, string); default `https://api.webflow.com`; format `uri`; Webflow API base
  URL override for tests or proxies.
- `mode` (optional, string).
- `site_id` (required, string); The Webflow site ID to read collections, pages, and forms from.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.webflow.com`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v2/sites/{{ config.site_id }}/collections`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `collections`, `pages`, `forms`, `sites`, `webhooks`, `custom_domains`,
`asset_folders`, `ecommerce_settings`; offset_limit: `assets`, `redirects`, `form_submissions`,
`orders`, `products`, `components`, `collection_items`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `collections`: GET `/v2/sites/{{ config.site_id }}/collections` - records path `collections`;
  emits passthrough records.
- `pages`: GET `/v2/sites/{{ config.site_id }}/pages` - records path `pages`; emits passthrough
  records.
- `forms`: GET `/v2/sites/{{ config.site_id }}/forms` - records path `forms`; emits passthrough
  records.
- `sites`: GET `/v2/sites` - records path `sites`.
- `assets`: GET `/v2/sites/{{ config.site_id }}/assets` - records path `assets`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100.
- `webhooks`: GET `/v2/sites/{{ config.site_id }}/webhooks` - records path `webhooks`.
- `redirects`: GET `/v2/sites/{{ config.site_id }}/redirects` - records path `redirects`;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100.
- `form_submissions`: GET `/v2/sites/{{ config.site_id }}/form_submissions` - records path
  `formSubmissions`; offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
  page size 100.
- `orders`: GET `/v2/sites/{{ config.site_id }}/orders` - records path `orders`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100.
- `products`: GET `/v2/sites/{{ config.site_id }}/products` - records path `items`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100.
- `custom_domains`: GET `/v2/sites/{{ config.site_id }}/custom_domains` - records path
  `customDomains`.
- `components`: GET `/v2/sites/{{ config.site_id }}/components` - records path `components`;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100.
- `asset_folders`: GET `/v2/sites/{{ config.site_id }}/asset_folders` - records path `assetFolders`.
- `ecommerce_settings`: GET `/v2/sites/{{ config.site_id }}/ecommerce/settings` - records path `.`.
- `collection_items`: GET `/v2/collections/{{ fanout.id }}/items` - records path `items`;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100;
  incremental cursor `lastUpdated`; sent as `lastUpdated[gte]`; formatted as `rfc3339`; fan-out; ids
  from request `/v2/sites/{{ config.site_id }}/collections`; id-list records path `collections`; id
  field `id`; id inserted into the request path; stamps `collectionId`.

## Write actions & risks

Overall write risk: external mutations against a live Webflow site: CMS collection-item
create/update/delete/publish, form-submission hidden-field updates, asset rename/delete, webhook
subscribe/unsubscribe, and ecommerce order fulfillment/refund/inventory changes;
publish/fulfill/refund actions have customer- or site-visitor-visible side effects and refund is
irreversible.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_collection_item`: POST `/v2/collections/{{ record.collection_id }}/items` - kind `create`;
  body type `json`; body fields `fieldData`, `isArchived`, `isDraft`, `cmsLocaleId`; required record
  fields `collection_id`, `fieldData`; accepted fields `cmsLocaleId`, `collection_id`, `fieldData`,
  `isArchived`, `isDraft`; risk: creates a new staged (unpublished-to-live) CMS item in the given
  collection; consumes the site's CMS item quota; the item is not visible on the live site until a
  subsequent publish_collection_item call.
- `update_collection_item`: PATCH `/v2/collections/{{ record.collection_id }}/items/{{ record.id }}`
  - kind `update`; body type `json`; path fields `collection_id`, `id`; body fields `fieldData`,
  `isArchived`, `isDraft`, `cmsLocaleId`; required record fields `collection_id`, `id`; accepted
  fields `cmsLocaleId`, `collection_id`, `fieldData`, `id`, `isArchived`, `isDraft`; risk:
  overwrites staged field data on an existing CMS item (primary locale only, unless cmsLocaleId is
  set); changes are not visible on the live site until a subsequent publish_collection_item call.
- `delete_collection_item`: DELETE `/v2/collections/{{ record.collection_id }}/items/{{ record.id
  }}` - kind `delete`; body type `none`; path fields `collection_id`, `id`; required record fields
  `collection_id`, `id`; accepted fields `collection_id`, `id`; missing records treated as success
  for status `404`; risk: permanently removes a staged CMS item from the collection; if the item was
  previously published live, it remains live until a subsequent unpublish, since delete only affects
  the staged copy per Webflow's staged/live item model.
- `publish_collection_item`: POST `/v2/collections/{{ record.collection_id }}/items/publish` - kind
  `update`; body type `json`; path fields `collection_id`; body fields `itemIds`; required record
  fields `collection_id`, `itemIds`; accepted fields `collection_id`, `itemIds`; risk: publishes the
  current staged content of the named item ids live immediately, making them visible on the site's
  live domain(s); does not itself trigger a full site publish/build.
- `update_form_submission`: PATCH `/v2/sites/{{ config.site_id }}/form_submissions/{{ record.id }}`
  - kind `update`; body type `json`; path fields `id`; body fields `formSubmissionData`; required
  record fields `id`, `formSubmissionData`; accepted fields `formSubmissionData`, `id`; risk:
  overwrites the values of hidden fields already defined on the form's schema for one submission;
  cannot add new fields or edit visible/user-submitted answers.
- `update_asset`: PATCH `/v2/assets/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; body fields `displayName`, `altText`; required record fields `id`; accepted fields
  `altText`, `displayName`, `id`; risk: renames or re-captions an existing site asset (display name
  / alt text only); does not replace the underlying file.
- `delete_asset`: DELETE `/v2/assets/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; risk: permanently deletes an uploaded asset from the site's asset library; any live
  page still referencing the asset's URL will show a broken image/file link.
- `create_webhook`: POST `/v2/sites/{{ config.site_id }}/webhooks` - kind `create`; body type
  `json`; body fields `triggerType`, `url`, `filter`; required record fields `triggerType`, `url`;
  accepted fields `filter`, `triggerType`, `url`; risk: registers a new outbound webhook
  subscription for the site (up to 75 per triggerType); Webflow will begin POSTing event payloads to
  the given url immediately for matching events.
- `delete_webhook`: DELETE `/v2/webhooks/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: permanently removes a webhook subscription; Webflow stops sending event
  payloads for it immediately.
- `update_order`: PATCH `/v2/sites/{{ config.site_id }}/orders/{{ record.orderId }}` - kind
  `update`; body type `json`; path fields `orderId`; body fields `comment`, `shippingProvider`,
  `shippingTracking`, `shippingTrackingURL`; required record fields `orderId`; accepted fields
  `comment`, `orderId`, `shippingProvider`, `shippingTracking`, `shippingTrackingURL`; risk: updates
  internal order record-keeping fields (comment/shipping provider/tracking number/tracking URL) on a
  live ecommerce order; does not change order status or trigger customer notifications by itself.
- `fulfill_order`: POST `/v2/sites/{{ config.site_id }}/orders/{{ record.orderId }}/fulfill` - kind
  `update`; body type `json`; path fields `orderId`; body fields `sendOrderFulfilledEmail`; required
  record fields `orderId`; accepted fields `orderId`, `sendOrderFulfilledEmail`; risk: marks a real
  customer order as fulfilled and, unless sendOrderFulfilledEmail is explicitly set false, sends the
  customer a fulfillment notification email.
- `refund_order`: POST `/v2/sites/{{ config.site_id }}/orders/{{ record.orderId }}/refund` - kind
  `update`; body type `none`; path fields `orderId`; required record fields `orderId`; accepted
  fields `orderId`; risk: DESTRUCTIVE FINANCIAL ACTION: reverses the underlying Stripe charge and
  refunds the customer's payment in full, setting the order's status to refunded; cannot be undone
  through the API.
- `update_inventory`: PATCH `/v2/collections/{{ record.sku_collection_id }}/items/{{ record.sku_id
  }}/inventory` - kind `update`; body type `json`; path fields `sku_collection_id`, `sku_id`; body
  fields `inventoryType`, `quantity`, `updateQuantity`; required record fields `sku_collection_id`,
  `sku_id`, `inventoryType`; accepted fields `inventoryType`, `quantity`, `sku_collection_id`,
  `sku_id`, `updateQuantity`; risk: changes live storefront stock levels for a SKU either absolutely
  (quantity) or incrementally (updateQuantity); an incorrect value can oversell or wrongly zero-out
  a product's live availability.

## Known limits

- API coverage includes 15 stream-backed endpoint group(s), 13 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=3, destructive_admin=5, duplicate_of=18, non_data_endpoint=2, out_of_scope=28,
  requires_elevated_scope=30.
