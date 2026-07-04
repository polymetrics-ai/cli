# Overview

Judge.me Reviews reads the current Judge.me OpenAPI surface for reviews, widget payloads, review counts, reviewers, webhooks, shop information, and settings. It also keeps the legacy `products` and `widgets` streams from `internal/connectors/judge-me-reviews`, because the legacy Go connector emits those records even though those two list paths are no longer present in the current OpenAPI document.

The OpenAPI document at `https://judge.me/api/docs.yaml` declares `https://api.judge.me/api/v1` as the API server. `base_url` remains configurable for tests and proxies.

## Auth setup

Provide a Judge.me API token through the `api_key` secret. The bundle sends it as the documented `X-Api-Token` header. Provide `shop_domain` as config; it is sent as the `shop_domain` query parameter by the bundle auth layer so reads and writes share the same shop scoping.

## Streams notes

The legacy `reviews`, `products`, and `widgets` streams keep the original record projection: primary key `id`, cursor field `created_at`, and the same flattened reviewer fields on review records. `reviews` also exposes documented optional filters for `product_id`, `rating`, and `reviewer_id`.

All documented widget, count, detail, reviewer, webhook, shop, and settings GET endpoints are represented as streams. Singleton HTML/count/settings endpoints use `records.path: "."` or the documented envelope key and stamp a deterministic `id` when the API payload has no natural primary key.

## Write actions & risks

Write actions cover every dialect-expressible documented mutation: creating reviews, moderating reviews, upserting reviewers, submitting reviewer data requests, creating/updating/deleting webhooks, bulk-creating webhooks, updating or uninstalling the shop, creating checkout comments, and creating public or private replies.

The highest-risk action is `uninstall_shop`, which calls `DELETE /shops` and is marked destructive. Webhook deletes are also destructive and idempotent for 404 replay handling. All writes require the normal reverse ETL plan, preview, approval, execute flow.

## Known limits

- The OpenAPI document no longer lists the legacy `/products` and `/widgets` list endpoints. They remain in this bundle solely because the read-only legacy Go connector emits those record sets.
- Widget endpoints that accept `page`/`per_page` return one HTML/count payload per request rather than a normal array envelope. The bundle exposes `widget_page` and `widget_per_page` config values but does not auto-paginate those singleton payloads.
- `reviewer` read lookups can pass `external_id` and `email` query parameters. The `update_reviewer` write action covers the documented endpoint by Judge.me id only; write actions do not currently have a query-template field for the alternate `id=-1` lookup mode.
- The legacy streams still use fixed `pagination.page_size: 100` and `max_pages: 1000` for the engine stop threshold. `config.page_size` drives the outgoing `per_page` request value, but `max_pages` is not runtime-configurable in the declarative pagination dialect.
