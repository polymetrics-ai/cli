---
name: pm-squarespace
description: Squarespace connector knowledge and safe action guide.
---

# pm-squarespace

## Purpose

Reads Squarespace orders, products, inventory, profiles, transactions, store pages, webhook subscriptions, and contacts, and writes webhook subscription mutations through the Squarespace Commerce API.

## Icon

- id: simple-icons-squarespace
- asset: icons/simple-icons/squarespace.svg
- title: Squarespace
- simple_icon_slug: squarespace
- simple_icon_hex: 000000
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Squarespace
- match: exact-name-or-slug
- matched_by: squarespace

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- api_key (secret) (required)

## ETL Streams

- orders:
  - primary key: id
  - cursor: modifiedOn
  - fields: createdOn(string), id(string), modifiedOn(string), orderNumber(string)
- products:
  - primary key: id
  - cursor: modifiedOn
  - fields: createdOn(string), id(string), modifiedOn(string), name(string)
- inventory:
  - primary key: sku
  - fields: modifiedOn(string), quantity(integer), sku(string)
- profiles:
  - primary key: id
  - fields: createdOn(string), id(string), modifiedOn(string), name(string)
- transactions:
  - primary key: id
  - fields: createdOn(string), customerEmail(string), discounts(array), id(string), modifiedOn(string), payments(array), salesLineItems(array), salesOrderId(string), shippingLineItems(array), total(object), totalNetPayment(object), totalNetSales(object), totalNetShipping(object), totalSales(object), totalTaxes(object), voided(boolean)
- store_pages:
  - primary key: id
  - fields: id(string), isEnabled(boolean), title(string), urlSlug(string)
- webhook_subscriptions:
  - primary key: id
  - fields: clientId(string), createdOn(string), endpointUrl(string), id(string), topics(array), updatedOn(string), websiteId(string)
- contacts:
  - primary key: id
  - fields: createdOn(string), defaultShippingAddress(object), firstName(string), id(string), lastName(string), locale(string), primaryEmail(object)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_webhook_subscription:
  - endpoint: POST /webhook_subscriptions
  - required fields: endpointUrl
  - risk: registers a new HTTPS endpoint to receive live order/contact/address event notifications; low-risk external mutation, no approval required
- delete_webhook_subscription:
  - endpoint: DELETE /webhook_subscriptions/{{ record.id }}
  - required fields: id
  - risk: permanently removes a webhook subscription, stopping future event notifications to that endpoint; external mutation, approval required

## Security

- read risk: external Squarespace API read of commerce orders, products, inventory, profiles, transactions, store pages, webhook subscriptions, and contacts
- write risk: external Squarespace API mutation (webhook subscription create/delete)
- approval: reverse ETL plan approval required before destructive writes (delete_webhook_subscription); create_webhook_subscription is low-risk
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Declared squarespace API commands.
- Usage: pm squarespace <command> [flags]
- Global flags:
  - --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
- Other Commands
  - operations get-1-0-authorization-member - Declared direct read: GET /1.0/authorization/member. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-1-0-authorization-website - Declared direct read: GET /1.0/authorization/website. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-1-0-commerce-inventory - Declared etl: GET /1.0/commerce/inventory. [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.
  - operations get-1-0-commerce-inventory-variant-id-csvs - Declared direct read: GET /1.0/commerce/inventory/{variantIdCsvs}. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-1-0-commerce-inventory-adjustments - Declared direct write: POST /1.0/commerce/inventory/adjustments. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: CreateInventoryAdjustmentRequest's body is 4 alternative named arrays (incrementOperations/decrementOperations/setFiniteOperations/setUnlimitedOperations) of {variantId,quantity} objects, not a flat record; the engine's default JSON write body cannot construct this nested array-of-operations shape without a Tier-2 WriteHook; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-1-0-commerce-orders - Declared etl: GET /1.0/commerce/orders. [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.
  - operations post-1-0-commerce-orders - Declared direct write: POST /1.0/commerce/orders. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Order.create's real body is a full order graph (lineItems, addresses, fulfillmentStatus, discountLines) with no flat-field shape suitable for reverse ETL; creating orders programmatically is also a rare/unsafe integration pattern for a data-sync connector; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-1-0-commerce-orders-id - Declared direct read: GET /1.0/commerce/orders/{id}. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-1-0-commerce-orders-id-fulfillments - Declared direct write: POST /1.0/commerce/orders/{id}/fulfillments. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: OrderFulfillmentRequest's body is a shipments[] array of nested CreateOrderShipmentRequest objects (carrierName/service/shipDate/trackingNumber each), not a flat record; the engine's default JSON write body copies top-level record fields verbatim and cannot construct a nested array-of-objects payload without a Tier-2 WriteHook; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-1-0-commerce-store-pages - Declared etl: GET /1.0/commerce/store_pages. [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.
  - operations get-1-0-commerce-transactions - Declared etl: GET /1.0/commerce/transactions. [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.
  - operations get-1-0-commerce-transactions-document-ids - Declared direct read: GET /1.0/commerce/transactions/{documentIds}. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-1-0-profiles - Declared etl: GET /1.0/profiles. [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.
  - operations get-1-0-profiles-profile-id-csvs - Declared direct read: GET /1.0/profiles/{profileIdCsvs}. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-1-0-webhook-subscriptions - Declared etl: GET /1.0/webhook_subscriptions. [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.
  - operations post-1-0-webhook-subscriptions - Declared direct write: POST /1.0/webhook_subscriptions. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /1.0/webhook_subscriptions.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations delete-1-0-webhook-subscriptions-subscription-id - Declared direct write: DELETE /1.0/webhook_subscriptions/{subscriptionId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Declared provider mutation: DELETE /1.0/webhook_subscriptions/{subscriptionId}.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-1-0-webhook-subscriptions-subscription-id - Declared direct read: GET /1.0/webhook_subscriptions/{subscriptionId}. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-1-0-webhook-subscriptions-subscription-id - Declared direct write: POST /1.0/webhook_subscriptions/{subscriptionId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: ExternalUpdateWebhookSubscriptionRequest wraps every changed field as {present:true,value:...} (Squarespace's own partial-update convention), not a flat record; the engine's default JSON write body copies top-level record fields verbatim and cannot construct this per-field wrapper without a Tier-2 WriteHook; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-1-0-webhook-subscriptions-subscription-id-actions-rotate-secret - Declared direct write: POST /1.0/webhook_subscriptions/{subscriptionId}/actions/rotateSecret. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: immediately invalidates the subscription's existing signing secret for every downstream consumer verifying webhook signatures; an operator-initiated security action, not a syncable data mutation; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-1-0-webhook-subscriptions-subscription-id-actions-send-test-notification - Declared direct write: POST /1.0/webhook_subscriptions/{subscriptionId}/actions/sendTestNotification. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: fires a synthetic test event at the configured endpoint for manual verification; no data is created, read, or mutated; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-v1-analytics-transaction-summaries - Declared direct write: POST /v1/analytics/transaction-summaries. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: returns aggregated order/donation totals per contact for a caller-supplied contact-id batch, not a syncable object list; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-v1-commerce-discounts - Declared direct read: GET /v1/commerce/discounts. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-v1-commerce-discounts - Declared direct write: POST /v1/commerce/discounts. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations delete-v1-commerce-discounts-discount-id - Declared direct write: DELETE /v1/commerce/discounts/{discountId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Documented provider deletion is not promoted without an operation-specific plan, preview, approval, typed confirmation, and execution contract.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-v1-commerce-discounts-discount-id - Declared direct read: GET /v1/commerce/discounts/{discountId}. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations put-v1-commerce-discounts-discount-id - Declared direct write: PUT /v1/commerce/discounts/{discountId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-v1-contacts - Declared etl: GET /v1/contacts. [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.
  - operations post-v1-contacts - Declared direct write: POST /v1/contacts. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: the Contacts API v1 surface lives at a different version prefix (/v1/...) than every Commerce API write action's base_url (which bakes in the Commerce API's /1.0 segment); reaching it requires the same stream.path absolute-URL override the contacts READ stream uses, but write actions have no equivalent execution-contract skip-marker mechanism (write_request_shape's replay capture server always points at b.HTTP.URL, which an absolute-URL action.path bypasses entirely) — so an untestable write action was not added rather than shipping one that cannot be proven correct in this repo's execution-contract harness; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations delete-v1-contacts-contact-id - Declared direct write: DELETE /v1/contacts/{contactId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: same Contacts API v1 version-prefix/execution-contract-testability gap as POST /v1/contacts above; not added as an untestable write action; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-v1-contacts-contact-id - Declared direct read: GET /v1/contacts/{contactId}. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations patch-v1-contacts-contact-id - Declared direct write: PATCH /v1/contacts/{contactId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: requires an application/merge-patch+json content type the engine's write dialect (json/form/none body_type) does not model; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-v1-contacts-contact-id-address-book - Declared direct read: GET /v1/contacts/{contactId}/address-book. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-v1-contacts-contact-id-address-book - Declared direct write: POST /v1/contacts/{contactId}/address-book. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: creates an address-book entry under a specific contact; narrow sub-resource write with no corresponding read stream in this pass; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations delete-v1-contacts-contact-id-address-book-address-book-entry-id - Declared direct write: DELETE /v1/contacts/{contactId}/address-book/{addressBookEntryId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: deletes a specific address-book entry; narrow sub-resource write with no corresponding read stream in this pass; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-v1-contacts-contact-id-address-book-address-book-entry-id - Declared direct read: GET /v1/contacts/{contactId}/address-book/{addressBookEntryId}. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations put-v1-contacts-contact-id-address-book-address-book-entry-id - Declared direct write: PUT /v1/contacts/{contactId}/address-book/{addressBookEntryId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: replaces a specific address-book entry; narrow sub-resource write with no corresponding read stream in this pass; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-v1-contacts-query - Declared direct write: POST /v1/contacts/query. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: filtered/sorted contact search returning the identical Contact record shape already covered by the contacts list stream; a POST-with-body search endpoint has no GET-list declarative equivalent in this dialect; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-v1-fulfillments-fulfillment-options - Declared direct read: GET /v1/fulfillments/fulfillment-options. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-v2-commerce-products - Declared etl: GET /v2/commerce/products. [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.
  - operations post-v2-commerce-products - Declared direct write: POST /v2/commerce/products. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: CreateProductRequest's body includes nested variants[]/variantAttributes[] arrays required for most real product types; not expressible as the engine's default flat JSON write body without a Tier-2 WriteHook; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations delete-v2-commerce-products-product-id - Declared direct write: DELETE /v2/commerce/products/{productId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: permanently deletes a product and all its variants/images from the merchant's live storefront; higher-risk storefront-content deletion, out of scope for this pass's reverse-ETL write set; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-v2-commerce-products-product-id - Declared direct write: POST /v2/commerce/products/{productId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: UpdateProductRequest wraps every changed field as {present:true,value:...} (Squarespace's own partial-update convention), not a flat record; the engine's default JSON write body copies top-level record fields verbatim and cannot construct this per-field wrapper without a Tier-2 WriteHook; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-v2-commerce-products-product-id-images - Declared direct write: POST /v2/commerce/products/{productId}/images. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: multipart image upload; the engine's write dialect has no multipart/binary body construction; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations delete-v2-commerce-products-product-id-images-image-id - Declared direct write: DELETE /v2/commerce/products/{productId}/images/{imageId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: deletes a single product image; narrow sub-resource write with no corresponding read stream; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-v2-commerce-products-product-id-images-image-id - Declared direct write: POST /v2/commerce/products/{productId}/images/{imageId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: updates image metadata (altText) only; narrow sub-resource write with no corresponding read stream; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-v2-commerce-products-product-id-images-image-id-order - Declared direct write: POST /v2/commerce/products/{productId}/images/{imageId}/order. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: reorders product images; narrow sub-resource write with no corresponding read stream; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-v2-commerce-products-product-id-images-image-id-status - Declared direct read: GET /v2/commerce/products/{productId}/images/{imageId}/status. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-v2-commerce-products-product-id-variants - Declared direct write: POST /v2/commerce/products/{productId}/variants. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: CreateProductVariantRequest's body includes nested attributes/pricing/shippingMeasurements objects; narrow sub-resource write with no corresponding read stream; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations delete-v2-commerce-products-product-id-variants-variant-id - Declared direct write: DELETE /v2/commerce/products/{productId}/variants/{variantId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: deletes a single product variant; narrow sub-resource write with no corresponding read stream; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-v2-commerce-products-product-id-variants-variant-id - Declared direct write: POST /v2/commerce/products/{productId}/variants/{variantId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: UpdateProductVariantRequest uses the same present/value wrapper convention as update_product; narrow sub-resource write with no corresponding read stream; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-v2-commerce-products-product-id-variants-variant-id-image - Declared direct write: POST /v2/commerce/products/{productId}/variants/{variantId}/image. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: associates an existing image with a variant; narrow sub-resource write with no corresponding read stream; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-v2-commerce-products-product-id-csvs - Declared direct read: GET /v2/commerce/products/{productIdCsvs}. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor

## Commands

### Inspect as a manual

```bash
pm connectors inspect squarespace
```

### Inspect as structured JSON

```bash
pm connectors inspect squarespace --json
```

## Agent Rules

- Run pm connectors inspect squarespace before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
