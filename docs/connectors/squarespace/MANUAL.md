# pm connectors inspect squarespace

```text
NAME
  pm connectors inspect squarespace - Squarespace connector manual

SYNOPSIS
  pm connectors inspect squarespace
  pm connectors inspect squarespace --json
  pm credentials add <name> --connector squarespace [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Squarespace orders, products, inventory, profiles, transactions, store pages, webhook subscriptions, and contacts, and writes webhook subscription mutations through the Squarespace Commerce API.

ICON
  id: simple-icons-squarespace
  asset: icons/simple-icons/squarespace.svg
  title: Squarespace
  simple_icon_slug: squarespace
  simple_icon_hex: 000000
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Squarespace
  match: exact-name-or-slug
  matched_by: squarespace

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_key (secret)

ETL STREAMS
  orders:
    primary key: id
    cursor: modifiedOn
    fields: createdOn(string), id(string), modifiedOn(string), orderNumber(string)
  products:
    primary key: id
    cursor: modifiedOn
    fields: createdOn(string), id(string), modifiedOn(string), name(string)
  inventory:
    primary key: sku
    fields: modifiedOn(string), quantity(integer), sku(string)
  profiles:
    primary key: id
    fields: createdOn(string), id(string), modifiedOn(string), name(string)
  transactions:
    primary key: id
    fields: createdOn(string), customerEmail(string), discounts(array), id(string), modifiedOn(string), payments(array), salesLineItems(array), salesOrderId(string), shippingLineItems(array), total(object), totalNetPayment(object), totalNetSales(object), totalNetShipping(object), totalSales(object), totalTaxes(object), voided(boolean)
  store_pages:
    primary key: id
    fields: id(string), isEnabled(boolean), title(string), urlSlug(string)
  webhook_subscriptions:
    primary key: id
    fields: clientId(string), createdOn(string), endpointUrl(string), id(string), topics(array), updatedOn(string), websiteId(string)
  contacts:
    primary key: id
    fields: createdOn(string), defaultShippingAddress(object), firstName(string), id(string), lastName(string), locale(string), primaryEmail(object)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_webhook_subscription:
    endpoint: POST /webhook_subscriptions
    required fields: endpointUrl
    risk: registers a new HTTPS endpoint to receive live order/contact/address event notifications; low-risk external mutation, no approval required
  delete_webhook_subscription:
    endpoint: DELETE /webhook_subscriptions/{{ record.id }}
    required fields: id
    risk: permanently removes a webhook subscription, stopping future event notifications to that endpoint; external mutation, approval required

SECURITY
  read risk: external Squarespace API read of commerce orders, products, inventory, profiles, transactions, store pages, webhook subscriptions, and contacts
  write risk: external Squarespace API mutation (webhook subscription create/delete)
  approval: reverse ETL plan approval required before destructive writes (delete_webhook_subscription); create_webhook_subscription is low-risk
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Squarespace's declared streams and reverse-ETL actions.
  Usage: pm squarespace <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    api delete v1 commerce discounts discountid - Documented DELETE /v1/commerce/discounts/{discountId} (not implemented) [intent=direct_write availability=not_implemented operation=squarespace.delete.v1-commerce-discounts-discountid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v1 contacts contactid - Documented DELETE /v1/contacts/{contactId} (not implemented) [intent=direct_write availability=not_implemented operation=squarespace.delete.v1-contacts-contactid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v1 contacts contactid address-book addressbookentryid - Documented DELETE /v1/contacts/{contactId}/address-book/{addressBookEntryId} (not implemented) [intent=direct_write availability=not_implemented operation=squarespace.delete.v1-contacts-contactid-address-book-addressbookentryid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v2 commerce products productid - Documented DELETE /v2/commerce/products/{productId} (not implemented) [intent=direct_write availability=not_implemented operation=squarespace.delete.v2-commerce-products-productid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v2 commerce products productid images imageid - Documented DELETE /v2/commerce/products/{productId}/images/{imageId} (not implemented) [intent=direct_write availability=not_implemented operation=squarespace.delete.v2-commerce-products-productid-images-imageid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v2 commerce products productid variants variantid - Documented DELETE /v2/commerce/products/{productId}/variants/{variantId} (not implemented) [intent=direct_write availability=not_implemented operation=squarespace.delete.v2-commerce-products-productid-variants-variantid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api get 1-0 authorization member - Documented GET /1.0/authorization/member (not implemented) [intent=direct_read availability=not_implemented operation=squarespace.get.1-0-authorization-member]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get 1-0 authorization website - Documented GET /1.0/authorization/website (not implemented) [intent=direct_read availability=not_implemented operation=squarespace.get.1-0-authorization-website]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get 1-0 commerce inventory variantidcsvs - Documented GET /1.0/commerce/inventory/{variantIdCsvs} (not implemented) [intent=direct_read availability=not_implemented operation=squarespace.get.1-0-commerce-inventory-variantidcsvs]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get 1-0 commerce orders id - Documented GET /1.0/commerce/orders/{id} (not implemented) [intent=direct_read availability=not_implemented operation=squarespace.get.1-0-commerce-orders-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get 1-0 commerce transactions documentids - Documented GET /1.0/commerce/transactions/{documentIds} (not implemented) [intent=direct_read availability=not_implemented operation=squarespace.get.1-0-commerce-transactions-documentids]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get 1-0 profiles profileidcsvs - Documented GET /1.0/profiles/{profileIdCsvs} (not implemented) [intent=direct_read availability=not_implemented operation=squarespace.get.1-0-profiles-profileidcsvs]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get 1-0 webhook-subscriptions subscriptionid - Documented GET /1.0/webhook_subscriptions/{subscriptionId} (not implemented) [intent=direct_read availability=not_implemented operation=squarespace.get.1-0-webhook-subscriptions-subscriptionid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 commerce discounts - Documented GET /v1/commerce/discounts (not implemented) [intent=direct_read availability=not_implemented operation=squarespace.get.v1-commerce-discounts]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 commerce discounts discountid - Documented GET /v1/commerce/discounts/{discountId} (not implemented) [intent=direct_read availability=not_implemented operation=squarespace.get.v1-commerce-discounts-discountid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 contacts contactid - Documented GET /v1/contacts/{contactId} (not implemented) [intent=direct_read availability=not_implemented operation=squarespace.get.v1-contacts-contactid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 contacts contactid address-book - Documented GET /v1/contacts/{contactId}/address-book (not implemented) [intent=direct_read availability=not_implemented operation=squarespace.get.v1-contacts-contactid-address-book]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 contacts contactid address-book addressbookentryid - Documented GET /v1/contacts/{contactId}/address-book/{addressBookEntryId} (not implemented) [intent=direct_read availability=not_implemented operation=squarespace.get.v1-contacts-contactid-address-book-addressbookentryid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 fulfillments fulfillment-options - Documented GET /v1/fulfillments/fulfillment-options (not implemented) [intent=direct_read availability=not_implemented operation=squarespace.get.v1-fulfillments-fulfillment-options]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 commerce products productid images imageid status - Documented GET /v2/commerce/products/{productId}/images/{imageId}/status (not implemented) [intent=direct_read availability=not_implemented operation=squarespace.get.v2-commerce-products-productid-images-imageid-status]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 commerce products productidcsvs - Documented GET /v2/commerce/products/{productIdCsvs} (not implemented) [intent=direct_read availability=not_implemented operation=squarespace.get.v2-commerce-products-productidcsvs]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api patch v1 contacts contactid - Documented PATCH /v1/contacts/{contactId} (not implemented) [intent=direct_write availability=not_implemented operation=squarespace.patch.v1-contacts-contactid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post 1-0 commerce inventory adjustments - Documented POST /1.0/commerce/inventory/adjustments (not implemented) [intent=direct_write availability=not_implemented operation=squarespace.post.1-0-commerce-inventory-adjustments]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post 1-0 commerce orders - Documented POST /1.0/commerce/orders (not implemented) [intent=direct_write availability=not_implemented operation=squarespace.post.1-0-commerce-orders]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post 1-0 commerce orders id fulfillments - Documented POST /1.0/commerce/orders/{id}/fulfillments (not implemented) [intent=direct_write availability=not_implemented operation=squarespace.post.1-0-commerce-orders-id-fulfillments]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post 1-0 webhook-subscriptions subscriptionid - Documented POST /1.0/webhook_subscriptions/{subscriptionId} (not implemented) [intent=direct_write availability=not_implemented operation=squarespace.post.1-0-webhook-subscriptions-subscriptionid]; approval: not implemented: the REST write executor lacks the provider-specific top-level body envelope required by this operation; risk: high; notes: named_dependency=engine.rest_write_body_envelope: the REST write executor lacks the provider-specific top-level body envelope required by this operation
    api post 1-0 webhook-subscriptions subscriptionid actions rotatesecret - Documented POST /1.0/webhook_subscriptions/{subscriptionId}/actions/rotateSecret (not implemented) [intent=direct_write availability=not_implemented operation=squarespace.post.1-0-webhook-subscriptions-subscriptionid-actions-rotatesecret]; approval: not implemented: the provider webhook-URL mutation has no dedicated security-reviewed execution contract; risk: high; notes: named_dependency=review.webhook_url_mutation: the provider webhook-URL mutation has no dedicated security-reviewed execution contract
    api post 1-0 webhook-subscriptions subscriptionid actions sendtestnotification - Documented POST /1.0/webhook_subscriptions/{subscriptionId}/actions/sendTestNotification (not implemented) [intent=direct_write availability=not_implemented operation=squarespace.post.1-0-webhook-subscriptions-subscriptionid-actions-sendtestnotification]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 analytics transaction-summaries - Documented POST /v1/analytics/transaction-summaries (not implemented) [intent=direct_write availability=not_implemented operation=squarespace.post.v1-analytics-transaction-summaries]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 commerce discounts - Documented POST /v1/commerce/discounts (not implemented) [intent=direct_write availability=not_implemented operation=squarespace.post.v1-commerce-discounts]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 contacts - Documented POST /v1/contacts (not implemented) [intent=direct_write availability=not_implemented operation=squarespace.post.v1-contacts]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 contacts contactid address-book - Documented POST /v1/contacts/{contactId}/address-book (not implemented) [intent=direct_write availability=not_implemented operation=squarespace.post.v1-contacts-contactid-address-book]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 contacts query - Documented POST /v1/contacts/query (not implemented) [intent=direct_write availability=not_implemented operation=squarespace.post.v1-contacts-query]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 commerce products - Documented POST /v2/commerce/products (not implemented) [intent=direct_write availability=not_implemented operation=squarespace.post.v2-commerce-products]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 commerce products productid - Documented POST /v2/commerce/products/{productId} (not implemented) [intent=direct_write availability=not_implemented operation=squarespace.post.v2-commerce-products-productid]; approval: not implemented: the REST write executor lacks the provider-specific top-level body envelope required by this operation; risk: high; notes: named_dependency=engine.rest_write_body_envelope: the REST write executor lacks the provider-specific top-level body envelope required by this operation
    api post v2 commerce products productid images - Documented POST /v2/commerce/products/{productId}/images (not implemented) [intent=direct_write availability=not_implemented operation=squarespace.post.v2-commerce-products-productid-images]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 commerce products productid images imageid - Documented POST /v2/commerce/products/{productId}/images/{imageId} (not implemented) [intent=direct_write availability=not_implemented operation=squarespace.post.v2-commerce-products-productid-images-imageid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 commerce products productid images imageid order - Documented POST /v2/commerce/products/{productId}/images/{imageId}/order (not implemented) [intent=direct_write availability=not_implemented operation=squarespace.post.v2-commerce-products-productid-images-imageid-order]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 commerce products productid variants - Documented POST /v2/commerce/products/{productId}/variants (not implemented) [intent=direct_write availability=not_implemented operation=squarespace.post.v2-commerce-products-productid-variants]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 commerce products productid variants variantid - Documented POST /v2/commerce/products/{productId}/variants/{variantId} (not implemented) [intent=direct_write availability=not_implemented operation=squarespace.post.v2-commerce-products-productid-variants-variantid]; approval: not implemented: the REST write executor lacks the provider-specific top-level body envelope required by this operation; risk: high; notes: named_dependency=engine.rest_write_body_envelope: the REST write executor lacks the provider-specific top-level body envelope required by this operation
    api post v2 commerce products productid variants variantid image - Documented POST /v2/commerce/products/{productId}/variants/{variantId}/image (not implemented) [intent=direct_write availability=not_implemented operation=squarespace.post.v2-commerce-products-productid-variants-variantid-image]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put v1 commerce discounts discountid - Documented PUT /v1/commerce/discounts/{discountId} (not implemented) [intent=direct_write availability=not_implemented operation=squarespace.put.v1-commerce-discounts-discountid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put v1 contacts contactid address-book addressbookentryid - Documented PUT /v1/contacts/{contactId}/address-book/{addressBookEntryId} (not implemented) [intent=direct_write availability=not_implemented operation=squarespace.put.v1-contacts-contactid-address-book-addressbookentryid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    contacts list - Run the contacts ETL stream [intent=etl availability=implemented stream=contacts]
    create webhook subscription apply - Plan and execute the create webhook subscription reverse-ETL action [intent=reverse_etl availability=implemented write=create_webhook_subscription]; approval: requires plan, preview, approval, and execute; risk: registers a new HTTPS endpoint to receive live order/contact/address event notifications; low-risk external mutation, no approval required; flags: --endpointUrl (required)
    delete webhook subscription apply - Plan and execute the delete webhook subscription reverse-ETL action [intent=reverse_etl availability=implemented write=delete_webhook_subscription]; approval: requires plan, preview, approval, and execute; risk: permanently removes a webhook subscription, stopping future event notifications to that endpoint; external mutation, approval required; flags: --id (required)
    inventory list - Run the inventory ETL stream [intent=etl availability=implemented stream=inventory]
    orders list - Run the orders ETL stream [intent=etl availability=implemented stream=orders]
    products list - Run the products ETL stream [intent=etl availability=implemented stream=products]
    profiles list - Run the profiles ETL stream [intent=etl availability=implemented stream=profiles]
    store pages list - Run the store pages ETL stream [intent=etl availability=implemented stream=store_pages]
    transactions list - Run the transactions ETL stream [intent=etl availability=implemented stream=transactions]
    webhook subscriptions list - Run the webhook subscriptions ETL stream [intent=etl availability=implemented stream=webhook_subscriptions]

EXAMPLES
  # Inspect as a manual
  pm connectors inspect squarespace

  # Inspect as structured JSON
  pm connectors inspect squarespace --json

AGENT WORKFLOW
  - Run pm connectors inspect squarespace before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
