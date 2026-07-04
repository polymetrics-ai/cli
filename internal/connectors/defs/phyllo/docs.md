# Overview

Phyllo is a wave2 fan-out declarative-HTTP migration, expanded in Pass B to the full documented
Phyllo v1 REST surface. It reads Phyllo users, accounts, profiles, social contents/comments/content
groups, audience demographics, social-platform and e-commerce income (transactions/payouts/
balances), work platforms, and webhooks, and writes user/webhook lifecycle and account-config
mutations, through `https://api.getphyllo.com/v1/...`. The source of truth for Pass B's endpoint
inventory is the official Phyllo Postman collection (`getphyllo/phyllo-postman`,
"Phyllo APIs.postman_collection.json") since Phyllo's own docs.getphyllo.com reference is a
JS-rendered Stoplight SPA with no static export reachable outside a browser (the same "JS-backed"
constraint the pre-Pass-B bundle's own comment noted). This bundle is engine-vs-legacy
parity-tested against `internal/connectors/phyllo` (the hand-written connector it migrates) for its
original 4 streams; the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Phyllo `client_id` and `client_secret` secret pair; they are sent as HTTP Basic auth
(`Authorization: Basic base64(client_id:client_secret)`) and are never logged, matching legacy's
`connsdk.Basic(id, secretValue)` (`phyllo.go:127`). `base_url` defaults to the production host
`https://api.getphyllo.com` and may be set explicitly to
`https://api.sandbox.getphyllo.com`/`https://api.staging.getphyllo.com` to target a non-production
environment, or overridden entirely for tests/proxies.

## Streams notes

All 13 streams share the same base shape: `GET` against a Phyllo v1 list endpoint, records at the
top-level `data` key, `"projection": "passthrough"` (every raw field survives; `schemas/*.json`
properties document the field set, they are not an allow-list — matching legacy's verbatim-emit
`Read` for the original 4 streams, and the identical passthrough shape every other Phyllo list
endpoint uses). Pagination is offset+limit (`pagination.type: offset_limit`, `limit_param: limit`,
`offset_param: offset`, `page_size: 50`), stopping on a short page. None of Phyllo's list endpoints
document a server-side incremental filter parameter, so no stream declares an `incremental` block
(full-refresh reads only, matching every stream's real capability).

**Original 4 (legacy-parity) streams**: `users`, `accounts`, `profiles`, `social_contents` — primary
key `["id"]`, unchanged from the wave2 bundle. Legacy's `Read` passes `connsdk.Harvest` a callback
that emits every record verbatim (`func(rec connsdk.Record) error { return emit(connectors.Record(rec)) }`,
`phyllo.go:86`); `commonFields()` (`phyllo.go:104-106`) only documents legacy's `Catalog` metadata,
never gating what `Read` emits.

**9 new Pass B streams**: `work_platforms` (`/v1/work-platforms`, the platform catalog — YouTube,
Instagram, TikTok, etc. — that every account/profile is scoped to), `audience` (`/v1/audience`,
demographic breakdown for a connected account, primary key `["account_id"]` since it is one
demographic snapshot per account rather than an id-per-item collection), `social_content_groups`
(`/v1/social/content-groups`, playlist/series-style groupings of content items), `social_comments`
(`/v1/social/comments`, comments on social content items), `social_income_transactions` /
`social_income_payouts` (`/v1/social/income/{transactions,payouts}`, creator-platform earnings),
`commerce_income_transactions` / `commerce_income_payouts` / `commerce_income_balances`
(`/v1/commerce/income/{transactions,payouts,balances}`, e-commerce-platform earnings), and
`webhooks` (`/v1/webhooks`, registered event subscriptions). `accounts`/`profiles` accept optional
`user_id`/`work_platform_id` config filters (`phyllo_user_id`/`phyllo_work_platform_id`,
`omit_when_absent`); every account-scoped stream (`profiles`, `social_contents`, `audience`,
`social_content_groups`, `social_comments`, and all 5 income streams) accepts an optional
`phyllo_account_id` filter the same way — Phyllo's docs mark `account_id` as effectively required
for a meaningful result on several of these, but the engine has no per-param required-query
enforcement mechanism (conventions.md §3's `stream.Query` dialect is opt-in tolerance only, never a
hard requirement layer), so this bundle exposes it as an optional filter rather than inventing
client-side validation the dialect doesn't support; omitting it returns Phyllo's own
account-unscoped default result set.

## Write actions & risks

Six write actions, all newly added in Pass B (legacy `phyllo.Write` always returned
`connectors.ErrUnsupportedOperation`; this bundle now sets `capabilities.write: true`):

- **`create_user`** (`POST /v1/users`) — creates a new Phyllo end-user record (`name` +
  caller-assigned `external_id`) that every subsequent Connect/account/profile flow anchors to.
  Low risk, no approval required.
- **`create_webhook`** (`POST /v1/webhooks`) — registers a new webhook subscription (`url` +
  `events[]`). Low risk, no approval required.
- **`update_account`** (`PATCH /v1/accounts/{id}`) — changes an account's identity/engagement/
  income monitoring configuration; `body_fields: ["data"]` restricts the JSON body to Phyllo's own
  `{"data": {...}}` envelope shape (the API's PATCH body is a single nested `data` object, not a
  flat field set). Approval required.
- **`update_webhook`** (`PUT /v1/webhooks/{id}`) — replaces a webhook's `url`/`events[]`. Approval
  required.
- **`disconnect_account`** (`POST /v1/accounts/{id}/disconnect`) — permanently revokes Phyllo's
  connection to the linked creator platform account; `body_type: none`, no body. Destructive,
  approval required (`confirm: destructive`).
- **`delete_webhook`** (`DELETE /v1/webhooks/{id}`) — removes a webhook subscription;
  `delete.missing_ok_status: [404]` (idempotent delete). Destructive, approval required
  (`confirm: destructive`).

## Known limits

- **`environment`-based base-URL derivation is dropped; `base_url` must be set explicitly for
  non-production hosts.** Legacy derives the effective base URL from a separate `environment`
  config value (`"api.sandbox"` -> `https://api.sandbox.getphyllo.com`, `"api.staging"` ->
  `https://api.staging.getphyllo.com`, anything else -> the production default) only when
  `base_url` itself is unset (`phyllo.go:129-142`). The engine's `spec.json` `"default"`
  materialization mechanism (conventions.md §3) fills in a FIXED literal for a genuinely-absent
  key; it cannot express "the default value is a function of ANOTHER config key's value" (the same
  documented limitation as sentry's hostname-derived URL and chargebee's site-derived URL,
  conventions.md §3's `spec.json "default"` section). This bundle therefore requires the caller to
  set `base_url` directly to the desired sandbox/staging/production host — a documented
  config-surface narrowing, not a silent behavior change: every legacy-accepted final base URL
  (production default, sandbox, staging, or a raw override) remains reachable, just via one
  config key (`base_url`) instead of two (`base_url` OR `environment`). `environment` is not
  declared in this bundle's `spec.json` (F6, REVIEW.md: a declared-but-unwireable config key is
  worse than an absent one).
- **Async refresh/fetch-historic/search endpoints are out of scope.** Every `/refresh`,
  `/fetch-historic`, and `/search` endpoint across profiles/social-contents/social-content-groups/
  social-income/commerce-income triggers an out-of-band job whose result is delivered later via
  webhook, or is a request-body-driven bulk-id lookup rather than a paginated list read — neither
  shape is a synchronous "read this stream" or "mutate this one record" operation this dialect
  models; see `api_surface.json`'s `out_of_scope` entries for the per-endpoint reasoning.
- **`sdk-tokens` and `webhooks/send` are non-data endpoints**, not covered as streams or writes
  (session-token issuance and mock-webhook-delivery testing respectively; see `api_surface.json`).
- Every single-object detail GET (`/v1/{resource}/{id}`) is `duplicate_of` its sibling list
  stream's per-item record shape and is not separately modeled — see `api_surface.json`.
