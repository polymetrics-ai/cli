# Overview

RevenueCat is a declarative HTTP connector for the RevenueCat REST API v2. It reads the documented v2 list and detail endpoints for projects, apps, products, offerings, customers, entitlements, packages, purchases, subscriptions, invoices, paywalls, webhook integrations, virtual currencies, audit logs, charts, and metrics. It also exposes dialect-expressible v2 JSON mutations as reverse-ETL write actions.

The legacy Go connector emitted raw RevenueCat records with an added `stream` marker. All streams use passthrough projection so those raw fields are preserved while schemas pin stable identifiers and commonly documented fields.

## Auth setup

Provide a RevenueCat v2 secret API key via the `api_key` secret. It is sent with bearer authentication. `base_url` defaults to `https://api.revenuecat.com/v2` and may be overridden for fixture replay or a trusted proxy.

`project_id` is optional at connection setup because the `projects` stream is global, but every project-scoped stream and write requires it when used. Detail streams also require the matching id config such as `app_id`, `customer_id`, `product_id`, `subscription_id`, or `virtual_currency_code`.

## Streams notes

The original parity streams remain `projects`, `apps`, `products`, `offerings`, and `customers`. Pass B adds the documented v2 list/detail surface including app details and keys, customer child collections, entitlement/product/package relations, paywalls, webhook integrations, purchases, subscriptions, virtual currencies, audit logs, chart data/options, and metrics.

RevenueCat v2 documents cursor-style `starting_after` pagination and `next_page` links. The legacy connector used page-number requests with `page` and `limit`, so this bundle keeps that behavior for parity and fixture stability. Optional passthrough filters `starting_after`, `created_after`, and `updated_after` are still sent when configured.

Search streams `purchases` and `subscriptions` accept `store_purchase_identifier` and `store_subscription_identifier` respectively. Chart streams require `chart_name`. Single-object helper streams stamp the config id into the emitted record when the API response does not contain a natural id field.

## Write actions & risks

Write actions cover RevenueCat v2 JSON mutations for projects, apps, customers, entitlements, offerings, packages, products, paywalls, webhook integrations, purchases, subscriptions, and virtual currencies. They are fixed endpoint-specific actions, not a generic HTTP write surface. The media asset upload endpoint is excluded because it requires binary/multipart handling outside this dialect.

Destructive actions such as deletes, refunds, subscription cancellation, entitlement revocation, and customer data transfer are marked with destructive confirmation. All writes mutate external RevenueCat state and must go through plan, preview, approval, and execute.

## Known limits

- RevenueCat API v1 is a separate API/key family and is not mixed into this v2 connector. The metadata URL now points at the v2 docs used for this surface.
- RevenueCat's documented `next_page` cursor URL pagination is not perfectly expressible in the current engine because `next_page` is a URL/path while cursor `stop_path` handles boolean true/false stops. The bundle retains legacy page-number pagination with `page`/`limit`.
- `page_size` and `max_pages` remain static bundle values. The legacy connector accepted runtime overrides, but the engine pagination fields are not templated.
- Write schemas validate endpoint identifiers and common required body fields while allowing additional documented request-body fields to pass through to the fixed RevenueCat endpoint.
