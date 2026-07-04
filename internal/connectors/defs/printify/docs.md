# Overview

Printify reads and writes the documented Printify public API surface from `https://api.printify.com`, covering v1 shop/catalog/product/order/upload/webhook endpoints and v2 catalog shipping endpoints. The legacy streams keep their narrow mapped fields; expanded streams use documented top-level response fields.

## Auth setup

Provide a Printify personal access token via the `api_token` secret. Requests use `Authorization: Bearer <api_token>`. `base_url` now defaults to the API root `https://api.printify.com` because this bundle covers both `/v1` and `/v2` paths.

## Streams notes

Streams: shops, products, orders, blueprints, print_providers, blueprint_detail, blueprint_print_providers, blueprint_variants, shipping_profiles, print_provider_detail, product_detail, product_gpsr, order_detail, uploads, upload_detail, webhooks, v2_shipping_methods, v2_shipping_standard, v2_shipping_priority, v2_shipping_express, v2_shipping_economy.

Shop-scoped streams use `shop_id`. Catalog subresource streams use `blueprint_id` and/or `print_provider_id`. Detail streams use the matching `product_id`, `order_id`, or `image_id` config field. `products`, `orders`, and `uploads` use Printify `next_page_url` pagination with `limit=100`; their fixtures are single-page because static fixtures cannot know the replay server's absolute next URL.

`products` and `orders` preserve legacy's mapped fields and cursor metadata. The nested legacy `raw` copy is still a known engine gap: the declarative dialect cannot currently put the entire raw source object under one output field.

## Write actions & risks

Write actions: disconnect_shop, create_product, update_product, delete_product, publish_product, mark_product_publishing_succeeded, mark_product_publishing_failed, unpublish_product, submit_order, submit_express_order, send_order_to_production, calculate_order_shipping, cancel_order, upload_image, archive_uploaded_image, create_webhook, update_webhook, delete_webhook, simulate_webhook.

These actions can create and mutate live Printify products, orders, uploads, shop connections, and webhooks. Destructive actions are marked with `confirm: destructive`; all writes still require the normal reverse-ETL plan, preview, approval, and execute flow.

## Known limits

- Legacy's nested `raw` field remains blocked by the engine dialect. `projection: passthrough` would flatten raw fields instead of nesting them under `raw`, and `computed_fields` has no whole-record reference.
- `base_url` override semantics changed from the legacy `/v1` base to the API root so one bundle can cover documented `/v1` and `/v2` endpoints.
- Product and order write schemas intentionally validate only the stable documented top-level requirements; Printify performs deeper validation for nested product artwork, line item, and address payloads.
- OAuth app authorization/token endpoints are excluded as non-data credential lifecycle endpoints.
