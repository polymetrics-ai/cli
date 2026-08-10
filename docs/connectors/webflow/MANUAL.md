# pm connectors inspect webflow

```text
NAME
  pm connectors inspect webflow - Webflow connector manual

SYNOPSIS
  pm connectors inspect webflow
  pm connectors inspect webflow --json
  pm credentials add <name> --connector webflow [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Webflow sites, collections, collection items, pages, forms, form submissions, assets, asset folders, webhooks, redirects, custom domains, components, orders, products, and ecommerce settings, and writes CMS collection-item lifecycle actions, form-submission hidden-field updates, asset metadata, webhook subscriptions, and ecommerce order/inventory mutations, using the Webflow Data API v2.

ICON
  id: webflow
  asset: icons/webflow.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.webflow.com/data/reference

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  accept_version
  base_url
  mode
  site_id (required)
  api_key (secret) (required)

ETL STREAMS
  collections:
    primary key: id
    fields: displayName(string), id(string), slug(string)
  pages:
    primary key: id
    fields: id(string), slug(string), title(string)
  forms:
    primary key: id
    fields: createdOn(string), displayName(string), id(string)
  sites:
    primary key: id
    fields: createdOn(string), displayName(string), id(string), lastPublished(string), lastUpdated(string), parentFolderId(string), shortName(string), timeZone(string), workspaceId(string)
  assets:
    primary key: id
    fields: altText(string), contentType(string), createdOn(string), displayName(string), hostedUrl(string), id(string), lastUpdated(string), originalFileName(string), siteId(string), size(integer)
  webhooks:
    primary key: id
    fields: createdOn(string), filter(object), id(string), lastTriggered(string), siteId(string), triggerType(string), url(string), workspaceId(string)
  redirects:
    primary key: id
    fields: fromUrl(string), id(string), toUrl(string)
  form_submissions:
    primary key: id
    fields: dateSubmitted(string), displayName(string), formId(string), formResponse(object), id(string), localeId(string), siteId(string), workspaceId(string)
  orders:
    primary key: orderId
    fields: acceptedOn(string), comment(string), customerInfo(object), customerPaymentDetails(object), fulfilledOn(string), orderComment(string), orderId(string), purchasedItems(array), refundedOn(string), shippingAddress(object), status(string)
  products:
    primary key: product_id
    fields: product(object), product_id(string), skus(array)
  custom_domains:
    primary key: id
    fields: id(string), lastPublished(string), url(string)
  components:
    primary key: id
    fields: description(string), group(string), id(string), name(string), readonly(boolean)
  asset_folders:
    primary key: id
    fields: assets(array), createdOn(string), displayName(string), id(string), parentFolder(string), siteId(string)
  ecommerce_settings:
    primary key: siteId
    fields: createdOn(string), defaultCurrency(string), siteId(string)
  collection_items:
    primary key: id
    cursor: lastUpdated
    fields: cmsLocaleId(string), collectionId(string), createdOn(string), fieldData(object), id(string), isArchived(boolean), isDraft(boolean), lastPublished(string), lastUpdated(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_collection_item:
    endpoint: POST /v2/collections/{{ record.collection_id }}/items
    required fields: collection_id, fieldData
    optional fields: isArchived, isDraft, cmsLocaleId
    risk: creates a new staged (unpublished-to-live) CMS item in the given collection; consumes the site's CMS item quota; the item is not visible on the live site until a subsequent publish_collection_item call
  update_collection_item:
    endpoint: PATCH /v2/collections/{{ record.collection_id }}/items/{{ record.id }}
    required fields: collection_id, id
    optional fields: fieldData, isArchived, isDraft, cmsLocaleId
    risk: overwrites staged field data on an existing CMS item (primary locale only, unless cmsLocaleId is set); changes are not visible on the live site until a subsequent publish_collection_item call
  delete_collection_item:
    endpoint: DELETE /v2/collections/{{ record.collection_id }}/items/{{ record.id }}
    required fields: collection_id, id
    risk: permanently removes a staged CMS item from the collection; if the item was previously published live, it remains live until a subsequent unpublish, since delete only affects the staged copy per Webflow's staged/live item model
  publish_collection_item:
    endpoint: POST /v2/collections/{{ record.collection_id }}/items/publish
    required fields: collection_id, itemIds
    risk: publishes the current staged content of the named item ids live immediately, making them visible on the site's live domain(s); does not itself trigger a full site publish/build
  update_form_submission:
    endpoint: PATCH /v2/sites/{{ config.site_id }}/form_submissions/{{ record.id }}
    required fields: id, formSubmissionData
    risk: overwrites the values of hidden fields already defined on the form's schema for one submission; cannot add new fields or edit visible/user-submitted answers
  update_asset:
    endpoint: PATCH /v2/assets/{{ record.id }}
    required fields: id
    optional fields: displayName, altText
    risk: renames or re-captions an existing site asset (display name / alt text only); does not replace the underlying file
  delete_asset:
    endpoint: DELETE /v2/assets/{{ record.id }}
    required fields: id
    risk: permanently deletes an uploaded asset from the site's asset library; any live page still referencing the asset's URL will show a broken image/file link
  create_webhook:
    endpoint: POST /v2/sites/{{ config.site_id }}/webhooks
    required fields: triggerType, url
    optional fields: filter
    risk: registers a new outbound webhook subscription for the site (up to 75 per triggerType); Webflow will begin POSTing event payloads to the given url immediately for matching events
  delete_webhook:
    endpoint: DELETE /v2/webhooks/{{ record.id }}
    required fields: id
    risk: permanently removes a webhook subscription; Webflow stops sending event payloads for it immediately
  update_order:
    endpoint: PATCH /v2/sites/{{ config.site_id }}/orders/{{ record.orderId }}
    required fields: orderId
    optional fields: comment, shippingProvider, shippingTracking, shippingTrackingURL
    risk: updates internal order record-keeping fields (comment/shipping provider/tracking number/tracking URL) on a live ecommerce order; does not change order status or trigger customer notifications by itself
  fulfill_order:
    endpoint: POST /v2/sites/{{ config.site_id }}/orders/{{ record.orderId }}/fulfill
    required fields: orderId
    optional fields: sendOrderFulfilledEmail
    risk: marks a real customer order as fulfilled and, unless sendOrderFulfilledEmail is explicitly set false, sends the customer a fulfillment notification email
  refund_order:
    endpoint: POST /v2/sites/{{ config.site_id }}/orders/{{ record.orderId }}/refund
    required fields: orderId
    risk: DESTRUCTIVE FINANCIAL ACTION: reverses the underlying Stripe charge and refunds the customer's payment in full, setting the order's status to refunded; cannot be undone through the API
  update_inventory:
    endpoint: PATCH /v2/collections/{{ record.sku_collection_id }}/items/{{ record.sku_id }}/inventory
    required fields: sku_collection_id, sku_id, inventoryType
    optional fields: quantity, updateQuantity
    risk: changes live storefront stock levels for a SKU either absolutely (quantity) or incrementally (updateQuantity); an incorrect value can oversell or wrongly zero-out a product's live availability

SECURITY
  read risk: external Webflow API read of site collections, pages, forms, assets, webhooks, redirects, ecommerce orders/products, and CMS collection items for a configured site
  write risk: external mutations against a live Webflow site: CMS collection-item create/update/delete/publish, form-submission hidden-field updates, asset rename/delete, webhook subscribe/unsubscribe, and ecommerce order fulfillment/refund/inventory changes; publish/fulfill/refund actions have customer- or site-visitor-visible side effects and refund is irreversible
  approval: required for all write actions; refund_order and publish_collection_item are high-risk (irreversible financial action / immediate live-site visibility) and should be gated more strictly than the others
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect webflow

  # Inspect as structured JSON
  pm connectors inspect webflow --json

AGENT WORKFLOW
  - Run pm connectors inspect webflow before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
