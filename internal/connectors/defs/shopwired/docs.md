# Overview

ShopWired is a declarative HTTP connector for the documented ShopWired REST API v1. Pass B expands beyond the original products, orders, customers, and categories streams to cover the ReadMe API reference list, detail, and search GET endpoints as streams, and JSON POST/PUT/PATCH/DELETE endpoints as write actions.

The legacy Go connector emitted a fixed six-field projection for its four original streams. This bundle keeps schema projection for `products`, `orders`, `customers`, and `categories`, including the legacy `id`/`name`/`updated_at` fallbacks, while expanded API streams use passthrough so raw documented response fields are preserved.

## Auth setup

The current ShopWired docs use HTTP Basic authentication where the API key is the username and the API secret is the password. Provide `api_key` and `api_secret` for that flow. OAuth bearer tokens are also supported with `access_token`. For compatibility with the legacy connector, if only `api_key` is configured the engine sends it as the `X-API-Key` header.

`base_url` defaults to `https://api.ecommerceapi.uk/v1`, matching the current ShopWired reference. Override it if an account still uses the legacy `https://api.shopwired.co.uk` host.

## Streams notes

Streams are generated from the ReadMe OpenAPI snippets linked by `https://shopwired.readme.io/llms.txt`. List endpoints use documented `count`/`offset` pagination, with a static `count` of 100. The legacy streams project `id`, `name`, `sku`, `email`, `status`, and `updated_at`; detail and expanded list streams preserve passthrough records. Detail endpoints use `pagination.type: none` and require the documented path parameter through config, such as `id`, `product_id`, `sku`, `comment_id`, or `country_id` depending on the stream path.

Count endpoints such as `/products/count` are listed in `api_surface.json` as `non_data_endpoint` because they return aggregate metadata rather than records. Optional documented query parameters are exposed as optional config-backed query values and omitted when absent.

## Write actions & risks

Write actions are fixed to documented ShopWired endpoints and validate required path/body fields from the OpenAPI snippets where available. They are not generic HTTP writes. Deletions, refunds, cancellations, stock returns, dispatch requests, wishlist modifications, and all-app-data deletion are marked destructive when their documented operation can remove data or trigger irreversible store behavior.

All writes mutate external ShopWired store state and must go through plan, preview, approval, and execute.

## Known limits

- The current docs use Basic/OAuth authentication and `https://api.ecommerceapi.uk/v1`; the legacy Go connector used `X-API-Key` against `https://api.shopwired.co.uk`. The bundle supports both auth shapes, but defaults to the current documented host.
- Runtime `page_size`/`max_pages` overrides from legacy are not modeled because engine pagination fields are static. The bundle uses documented `count`/`offset` pagination with `count=100`.
- Expanded response schemas are intentionally permissive passthrough schemas. The ReadMe OpenAPI snippets define many resource-specific schemas, but preserving raw fields across the broad Pass B surface is safer than projecting a partial field list.
