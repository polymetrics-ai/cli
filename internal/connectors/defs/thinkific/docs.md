# Overview

Thinkific is a wave2 fan-out declarative-HTTP migration of `internal/connectors/thinkific` (the
hand-written legacy connector this bundle migrates; the legacy package stays registered and
unchanged until wave6's registry flip), since Pass B expanded to the full documented Thinkific
Admin API v1 surface (`https://developers.thinkific.com/openapi/thinkific-admin-api-v1.yaml`):
courses, users, enrollments, collections (categories), products, orders, coupons, promotions,
groups, instructors, course reviews, custom profile field definitions, site scripts, and product
publish requests, plus the write actions the dialect can express for each. `capabilities.write` is
now `true`.

## Auth setup

Provide a Thinkific API key via the `api_key` secret and the site subdomain via the `subdomain`
config value; both are sent as static request headers (`X-Auth-API-Key`/`X-Auth-Subdomain`,
`streams.json` `base.headers`), matching legacy's `connsdk.Requester.DefaultHeaders` construction
(`thinkific.go:110`: `{"X-Auth-API-Key": key, "X-Auth-Subdomain": subdomain}`) and the published
OpenAPI spec's `ApiKey`/`ApiKeySubdomain` security scheme pair. `base.auth` declares
`[{"mode": "none"}]` since Thinkific's auth is header-only, not a bearer/basic/api-key-* mode the
dialect's `auth` block otherwise expresses. `api_key` is never logged (`x-secret: true`). `base_url`
defaults to `https://api.thinkific.com` and may be overridden for tests/proxies. The published spec
also documents an `OAuthAccessToken` alternative (bearer-token OAuth app installs); this bundle
models only the static API-key/subdomain header pair, matching legacy's own auth shape — OAuth app
installation is a distinct, more complex credential-provisioning flow out of scope for this pass.

## Streams notes

13 streams, all `GET`, all under `/api/public/v1`, all records at the `items` envelope key, primary
key `["id"]` (Thinkific's real wire type — a JSON integer, not a string — for every stream's `id`
field, matching each stream's fixture and never widened to a string union):

- `courses` — the original legacy-parity stream. Every emitted field (`id`, `name`, `slug`,
  `created_at`) matches the raw API's own field names exactly — no `computed_fields` rename
  needed, plain schema projection reproduces legacy's inline record construction
  (`thinkific.go:86`) field-for-field.
- `users`, `enrollments`, `collections`, `products`, `orders`, `coupons`, `promotions`, `groups`,
  `instructors`, `course_reviews`, `custom_profile_field_definitions`, `site_scripts`,
  `product_publish_requests` — new Pass B streams, each a plain `GET .../items`-enveloped list
  matching the published OpenAPI response schema field-for-field (e.g. `enrollments`' `user_id`/
  `course_id`/`percentage_completed`/`completed`/`expired` fields, `orders`' `amount_cents`/
  `amount_dollars`/`status`, `site_scripts`' `page_scopes`/`location`/`load_method`/`category`).
  `collections` (the OpenAPI spec's own name for what Thinkific's UI calls "Categories") declares
  `created_at`/`default`/`product_ids` per its documented required-field set.
  `product_publish_requests`' schema declares `id` as its primary key even though the published
  OpenAPI response component (`GetProductPublishResponse`) mistakenly `$ref`s `PromotionResponse`
  rather than a dedicated response schema — the sibling by-id/approve/deny endpoints
  (`/product_publish_requests/{id}[/approve|/deny]`) all document `id` as an integer path
  parameter, confirming it is the real identifier despite the response-schema documentation bug;
  this is a documented interpretation, not a guess without supporting evidence.

Pagination is shared by every stream via `base.pagination` (`pagination.type: page_number`,
`page_param: page`, `size_param: limit`, `page_size: 100`) — matches legacy's hand-rolled loop
(`thinkific.go:72-93`) and the published spec's own `page`/`limit` query parameters (documented
identically on every list endpoint), stopping when a page returns fewer records than the configured
size.

## Write actions & risks

28 write actions, all under `/api/public/v1`, `body_type: "json"` unless noted:

- **Users**: `create_user` (POST `/users`), `update_user` (PUT `/users/{id}`), `delete_user`
  (DELETE `/users/{id}`, destructive/confirm-gated — permanently removes the account and its course
  access).
- **Enrollments**: `create_enrollment` (POST `/enrollments`, grants course access),
  `update_enrollment` (PUT `/enrollments/{id}`, can extend or revoke the access window via
  `activated_at`/`expiry_date`).
- **Coupons**: `create_coupon`, `update_coupon`, `delete_coupon` (standard create/update/delete
  against `/coupons[/{id}]`; deleting an in-use code breaks checkout for anyone holding it).
- **Collections (Categories)**: `create_collection`, `update_collection`, `delete_collection`
  against `/collections[/{id}]`, plus `add_products_to_collection` (POST
  `/collection_memberships/{collection_id}`, `body_fields: ["product_ids"]`) and
  `remove_products_from_collection` (DELETE `/collection_memberships/{collection_id}`, same
  `body_fields` restriction — a DELETE-with-body action, `delete.missing_ok_status: [404]`).
- **Groups**: `create_group`, `delete_group` against `/groups[/{id}]` (no PUT/update endpoint is
  published for groups — name is set only at creation), and `add_user_to_groups` (POST
  `/group_users`, a `kind: custom` action since it neither creates nor updates a single addressable
  resource but rather attaches an existing user to one or more existing groups by name).
- **Instructors**: `create_instructor`, `update_instructor`, `delete_instructor` against
  `/instructors[/{id}]` — deleting an instructor removes their attribution from every course
  crediting them.
- **Promotions**: `create_promotion`, `update_promotion`, `delete_promotion` against
  `/promotions[/{id}]` — directly changes checkout pricing for the targeted products; `update`/
  `delete` are flagged accordingly in `risk`.
- **Site Scripts**: `create_site_script`, `update_site_script` (both destructive/confirm-gated —
  they inject arbitrary third-party HTML/JavaScript site-wide into every scoped page) and
  `delete_site_script` against `/site_scripts[/{id}]`.
- **Course Reviews**: `create_course_review` (POST `/course_reviews`) — an approved review is
  publicly visible on the course landing page immediately.
- **Product Publish Requests**: `approve_product_publish_request` (POST
  `/product_publish_requests/{id}/approve`) and `deny_product_publish_request` (POST
  `/product_publish_requests/{id}/deny`), both `kind: custom` action-trigger writes carrying the
  documented `ProductPublishRequestBody` (`user_id` required, `response_text`/
  `notify_requester` optional) — approving makes a course publicly purchasable.

`metadata.json` now declares `capabilities.write: true`; `risk.approval` names exactly which
actions require approval (`site_scripts` create/update, `delete_user`) versus which are low-risk
additive/idempotent mutations.

## Known limits

- Detail-by-id endpoints (`/courses/{id}`, `/users/{id}`, `/enrollments/{id}`, `/collections/{id}`,
  `/products/{id}`, `/orders/{id}`, `/coupons/{id}`, `/course_reviews/{id}`, `/instructors/{id}`,
  `/groups/{id}`, `/promotions/{id}`, `/site_scripts/{id}`, `/product_publish_requests/{id}`) are
  not modeled as separate streams — each returns the identical record shape its corresponding list
  stream already emits; see `api_surface.json`'s `duplicate_of` exclusions.
- **Bundles are not modeled at all.** Thinkific's own API exposes only `/bundles/{id}` and its
  sub-resources (`/bundles/{id}/courses`, `/bundles/{id}/enrollments`); there is no `GET /bundles`
  list endpoint anywhere in the published spec, so there is no discovery path to obtain a bundle id
  without an id already in hand from an external source. See `api_surface.json`'s
  `requires_elevated_scope` exclusions.
- **Chapters and Contents are not modeled.** Both are detail-only (`/chapters/{id}`,
  `/chapters/{id}/contents`, `/contents/{id}`) with no top-level list endpoint; the only path to
  discover a chapter id is a course's own `chapter_ids` array, which would require sub-resource
  fan-out over every course — out of scope for this pass (`requires_elevated_scope`).
- **Group-analyst assignment sub-resources are not modeled** (`/groups/{group_id}/analysts`,
  `/group_analysts/{user_id}/groups`, and their mutations) — a narrower group-membership-management
  variant with no top-level discovery endpoint of its own; `add_user_to_groups` already covers the
  common "attach a user to groups" case.
- **External Orders (manual/off-platform sale recording) are not modeled** — `/external_orders`
  and its `/transactions/purchase`/`/transactions/refund` sub-endpoints are a distinct bookkeeping
  workflow for recording sales made outside Thinkific checkout, with no corresponding read stream
  to round-trip against; out of scope for this pass.
- **`page_size` is not runtime-configurable.** Legacy exposes a config-driven `page_size` override
  (`thinkific.go:124-130`, `pageSize(cfg, 100)`, any positive integer, defaulting to 100 when unset
  or invalid). The engine's `page_number` paginator constructor reads `PaginationSpec.PageSize` as
  a static bundle-level integer from `streams.json`, not a config-templated field, so there is no
  mechanism to make it runtime-configurable from `config.page_size` without inventing Go. This
  bundle hardcodes `page_size: 100`, legacy's own default, matching every input that does not
  explicitly override the page size (the common case); an operator who previously set a
  smaller/larger `page_size` config value loses that override here. `page_size` is not declared in
  `spec.json` at all (F6, REVIEW.md: a declared-but-unwireable key is worse than an absent one).
  This applies identically to every stream in this bundle (shared `base.pagination`).
- No incremental cursor is modeled on any stream. Legacy's catalog declares `created_at` as a
  `CursorFields` hint but the `Read` loop never actually filters by it (no incremental request
  param is ever sent) — this bundle matches that exact behavior on `courses`:
  `schemas/courses.json` declares `x-cursor-field: created_at` for catalog-hint parity, but no
  `incremental` block is declared, so every sync is full-refresh, exactly like legacy. New Pass B
  streams do not declare cursor metadata, since neither legacy nor the published OpenAPI spec
  documents a server-side updated-since filter parameter on those list endpoints.
