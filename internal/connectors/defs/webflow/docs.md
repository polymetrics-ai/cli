# Overview

Webflow reads sites, CMS collections and items, pages, forms and form submissions, assets and
asset folders, webhooks, redirects, custom domains, reusable components, and ecommerce orders/
products/inventory/settings for a single configured Webflow site, and writes CMS collection-item
lifecycle actions, form-submission hidden-field updates, asset metadata edits, webhook
subscription management, and ecommerce order/inventory mutations, using the Webflow Data API v2.
This bundle originated as a migration of `internal/connectors/webflow` (the hand-written,
read-only-only legacy connector); this pass (Pass B full-surface expansion) extends it to the full
practical Webflow Data API v2 surface reachable with a Site token or OAuth access token — see
`api_surface.json` for exactly what is and isn't covered, and why. The legacy package stays
registered and unchanged until wave6's registry flip. The catalog entry (`source-webflow`,
`internal/connectors/catalog_data.json`) declares `"type": "source"`; this bundle now also supports
write actions ahead of any catalog `"type"` change, which is out of scope for this connector-defs
pass.

## Auth setup

Provide a Webflow API token (Site token or OAuth access token) via the `api_key` secret; it is
used only for Bearer auth (`Authorization: Bearer <api_key>`) and is never logged. `site_id` is a
required config value naming the Webflow site every site-scoped stream and write reads/writes
against. An optional `accept_version` config value is sent as the `Accept-Version` header
(Webflow's API-versioning mechanism); when unset, the header is omitted entirely (not sent empty).

Several endpoints Webflow documents publicly require credentials this bundle does not model and
are excluded on that basis (see `api_surface.json`): custom-code registration needs a bearer token
minted by an OAuth **Data Client App**'s code-grant flow specifically (not a Site token or a plain
OAuth access token); Workspace Audit Logs, Site Activity Logs, 301-redirect/robots.txt/well-known-
file site configuration, and Workspace Management (site create/update/delete/plan) all require an
**Enterprise-plan** workspace.

## Streams notes

- `sites` — `GET /v2/sites`, records at `sites`, no pagination (Webflow does not paginate the
  sites-list response). Lists every site the token can access, not just the configured `site_id`.
- `collections` — `GET /v2/sites/{site_id}/collections`, records at `collections`, no pagination.
- `pages` — `GET /v2/sites/{site_id}/pages`, records at `pages`, no pagination.
- `forms` — `GET /v2/sites/{site_id}/forms`, records at `forms`, no pagination.
- `components` — `GET /v2/sites/{site_id}/components`, `offset_limit` pagination (`limit`/`offset`,
  page size 100), records at `components`.
- `assets` — `GET /v2/sites/{site_id}/assets`, `offset_limit` pagination, records at `assets`.
- `asset_folders` — `GET /v2/sites/{site_id}/asset_folders`, records at `assetFolders`, no
  pagination.
- `webhooks` — `GET /v2/sites/{site_id}/webhooks`, records at `webhooks`, no pagination.
- `redirects` — `GET /v2/sites/{site_id}/redirects`, `offset_limit` pagination, records at
  `redirects`.
- `custom_domains` — `GET /v2/sites/{site_id}/custom_domains`, records at `customDomains`, no
  pagination.
- `form_submissions` — `GET /v2/sites/{site_id}/form_submissions`, `offset_limit` pagination,
  records at `formSubmissions`. Site-wide (not per-form); the real API also supports narrowing via
  an `elementId` query param, not wired here since this stream's scope is the full-site feed.
- `orders` — `GET /v2/sites/{site_id}/orders`, `offset_limit` pagination, records at `orders`,
  primary key `orderId` (Webflow's own field name for the order identifier, distinct from every
  other stream's `id`).
- `products` — `GET /v2/sites/{site_id}/products`, `offset_limit` pagination, records at `items`
  (each element bundles a `product` and its `skus`, per Webflow's own response shape).
- `ecommerce_settings` — `GET /v2/sites/{site_id}/ecommerce/settings`, a single-object detail
  response (`records.path: "."`), primary key `siteId`; one record per sync, no pagination.
- `collection_items` — `GET /v2/collections/{collection_id}/items`, `offset_limit` pagination,
  records at `items`, incremental via `lastUpdated`/`lastUpdated[gte]`. Fans out over every
  collection returned by this site's own `collections` list (`fan_out.ids_from.request`), stamping
  `collectionId` onto every emitted item so records from different collections are distinguishable
  after the fan-out flattens them into one stream.

None of the non-fan-out streams declares an incremental cursor — Webflow's list endpoints for
sites/collections/pages/forms/components/assets/asset_folders/webhooks/redirects/custom_domains/
form_submissions/orders/products/ecommerce_settings expose no documented server-side "modified
since" filter, so every sync of those streams is a full read (§8 incremental truth table: no
server-side filter, no incremental block).

## Write actions & risks

- `create_collection_item` / `update_collection_item` / `delete_collection_item` /
  `publish_collection_item` — full staged-CMS-item lifecycle: create or update a **staged** item
  (not yet visible on the live site), delete a staged item, or publish specific item ids live
  immediately. Creating/updating never itself makes content live; `publish_collection_item` is the
  action with real live-site visibility impact.
- `update_form_submission` — updates only **hidden fields** already defined on the form's own
  schema; cannot alter visitor-submitted answers or add new fields.
- `update_asset` / `delete_asset` — rename/re-caption or permanently delete an uploaded asset.
  Deleting an asset still referenced by a live page leaves a broken image/file link; Webflow does
  not cascade-fix references. Uploading a NEW asset (`POST .../assets`) is not modeled — see
  `api_surface.json`'s `binary_payload` exclusion (it is a presigned-URL-plus-direct-S3-POST flow,
  not a single JSON/form request).
- `create_webhook` / `delete_webhook` — subscribe/unsubscribe a URL to a named Webflow event
  trigger type. Webflow begins/stops POSTing event payloads to the given URL immediately.
- `update_order` — updates only record-keeping fields (`comment`/`shippingProvider`/
  `shippingTracking`/`shippingTrackingURL`); never changes order status.
- `fulfill_order` — marks a real order fulfilled and, unless `sendOrderFulfilledEmail` is
  explicitly `false`, emails the customer.
- `refund_order` — **irreversible financial action**: reverses the underlying Stripe charge and
  sets the order to `refunded`. No request body; the entire effect is keyed off the `orderId` path
  parameter alone.
- `update_inventory` — sets a SKU's live stock level absolutely (`quantity`) or incrementally
  (`updateQuantity`); an incorrect value can oversell or wrongly zero out a live product.

All twelve actions require operator approval per `metadata.json`'s `risk.approval`; `refund_order`
and `publish_collection_item` carry the highest blast radius (irreversible payment reversal, and
immediate live-site content visibility, respectively) and should be gated more strictly in practice
than the others.

## Known limits

- Page/component DOM content (the actual rich-text/element tree of a page or component) and its
  corresponding write endpoints are not modeled — `ENGINE_GAP`: Webflow's DOM node representation is
  an arbitrarily-nested, type-polymorphic tree (text nodes vs. component-instance-property nodes
  have different shapes at every depth) with no flat draft-07 schema representation this dialect can
  express without either losing structure or degrading to an unstructured passthrough blob. See
  `api_surface.json`'s `out_of_scope` entries for `/v2/pages/{page_id}/dom` and the components
  equivalent.
- Custom code registration/application (site- and page-level) requires a bearer token minted via an
  OAuth **Data Client App**'s code-grant flow specifically, per Webflow's own endpoint
  documentation — not the Site-token/plain-OAuth-access-token auth this bundle's `spec.json` models.
  Excluded as `requires_elevated_scope`.
- Custom fonts and site-collaboration Comments (Designer in-app annotation threads) are out of
  scope: neither has an identified reverse-ETL/business-data use case, and font upload/replace is a
  multipart binary flow in any case.
- Webflow Analyze (traffic/top-pages/top-dimensions/top-events/time-on-page reports), Workspace
  Audit Logs, Site Activity Logs, Enterprise 301-redirect/robots.txt/well-known-file site
  configuration, and Workspace Management (site create/update/delete/plan) all require plan tiers
  or app types (Enterprise workspace, Analyze add-on) not confirmed available for this connector's
  modeled auth; excluded as `requires_elevated_scope`.
- Per-SKU inventory reading (`GET .../inventory`) is not modeled as its own stream: it has no
  independent list form (one call per SKU id), and the SKU ids it needs live nested inside each
  `products` stream record rather than as top-level list-response entries, which this dialect's
  `fan_out.ids_from.request` (top-level `records_path` only) cannot extract from. `update_inventory`
  (the write) is still fully covered — see `api_surface.json`.
- Product/SKU catalog authoring (`create_product`, `update_product`, `create_sku`, `update_sku`) is
  out of scope this pass — the write surface here focuses on ecommerce *operations*
  (fulfillment/refund/inventory) rather than catalog design, which typically happens in the Webflow
  Designer/E-commerce panel rather than via reverse ETL.
- CMS collection/field schema mutations (create/delete collection, create/update/delete field) and
  whole-site publish are excluded as `out_of_scope`/`destructive_admin` — they are structural
  schema-design or whole-site-deploy operations, a different risk class from the per-record CMS
  item writes modeled here.
- `collection_items`' fan-out reads every collection on the configured site on every sync (one
  preliminary paginated `GET /v2/sites/{site_id}/collections` call), then repeats the full paginated
  item read once per collection id — a site with many collections issues proportionally more
  requests per sync, matching the dialect's documented `fan_out` behavior (§3), not a bug.
  `collectionId` is stamped onto every emitted item so downstream consumers can group by source
  collection after the fan-out flattens results into one stream.
