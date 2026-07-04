# Overview

Amazon Ads reads Amazon Advertising profiles and Sponsored Products campaigns, ad groups, product
ads, portfolios, targeting keywords, and negative keywords through the Amazon Ads API (v2 entity
endpoints). This bundle
migrates `internal/connectors/amazon-ads` (the hand-written connector) to a declarative defs bundle
at capability parity; the legacy package stays registered and unchanged until wave6's registry flip.
It is read-only — the Amazon Ads API does expose mutating operations (create/update campaigns, ad
groups, keywords), but legacy exposes none of them as reverse-ETL targets, matching this bundle's
`capabilities.write: false`.

## Auth setup

Amazon Ads authenticates with Login with Amazon (LWA): a long-lived `refresh_token` is exchanged for
a short-lived `access_token` at the LWA token endpoint, sent on every data request as
`Authorization: Bearer <access_token>`. This exchange is a **Tier-2 AuthHook**
(`internal/connectors/hooks/amazon-ads/hooks.go`) — the declarative dialect's
`oauth2_client_credentials` mode is the client-credentials grant, not the refresh-token grant Amazon
Ads uses, and no declarative `auth` mode can perform a `grant_type=refresh_token` exchange. This is
the same class of "token-exchange auth" Tier-2 trigger conventions.md §1 names for GitHub App's
JWT→installation-token flow and amazon-seller-partner's own LWA refresh_token→access_token exchange
— ported here with a different destination header (`Authorization`, not SP-API's
`x-amz-access-token`).

Provide three secrets: `client_id` (LWA application client ID — also sent verbatim, unhashed, as the
`Amazon-Advertising-API-ClientId` header on every request, matching legacy's `requester` construction),
`client_secret` (LWA application client secret), and `refresh_token` (the long-lived LWA refresh
token) — all `x-secret: true`, never logged. The hook performs `POST {token_url}` with
`grant_type=refresh_token` form-encoded, matching legacy's `refreshTokenAuth.accessToken` exactly.

**The profile-scope header is folded into the returned Authenticator, not declared as a stream
header.** Every Amazon Ads v2 entity endpoint except `/v2/profiles` requires an
`Amazon-Advertising-API-Scope: <profile_id>` header (legacy's `streamEndpoint.scoped` table — see
`streams.go`: `profiles` is unscoped since it enumerates the profiles/scopes themselves; every other
stream is scoped). The engine resolves `base.headers` exactly ONCE per `Read`/`Check` call
(`engine/read.go`'s `newRuntime`), and `StreamSpec` has no per-stream header override in this dialect
— so a header that must be present for 4 streams and absent for the 5th cannot be expressed as a
plain declarative `base.headers` entry. Instead, the AuthHook's returned `Authenticator.Apply` (which
runs LAST in the request pipeline, after `DefaultHeaders` are already set — see `connsdk/http.go`'s
`do`) inspects the fully-built request's `URL.Path` to decide whether to attach the Scope header,
mirroring legacy's per-endpoint `scoped` flag exactly. A profile-scoped stream read with no
`profile_id` configured is a **hard error** (`amazon-ads config profile_id is required for
profile-scoped streams`), matching legacy's own `requester(cfg, scoped=true)` gate — never a silently
unscoped request.

Unlike legacy's `refreshTokenAuth`, which caches the token across calls with a 60-second-before-expiry
refresh window, this hook re-exchanges on every `Check`/`Read` call (the engine calls
`AuthHook.Authenticator` once per `Read`/`Check`, and that single authenticator instance is reused
across every page within that one call — see `engine/read.go`'s `newRuntime`/`selectAuth` — so a
multi-page Read still performs exactly ONE token exchange, matching legacy's own per-`Read`
`requester()` construction). This never changes emitted record data; it only means a long-lived
process issuing many separate `Read` calls re-exchanges the token more often than legacy's in-process
cache would have. Documented as an ACCEPTABLE deviation (§5 meta-rule): never data-changing, and the
engine provides no mechanism for the connector-defined layer to cache state across separate `Read`
invocations.

`base_url` defaults to the North America (NA) endpoint (`https://advertising-api.amazon.com`),
matching legacy's `regionBaseURL["NA"]` fallback. Legacy also derives this from a `region` config enum
(NA/EU/FE) at runtime; this bundle requires the resolved `base_url` directly for non-NA regions
instead, since the engine's `spec.json` `"default"` mechanism only materializes one fixed literal, not
a region-derived choice between three URLs (documented scope narrowing). Set `base_url` to
`https://advertising-api-eu.amazon.com` or `https://advertising-api-fe.amazon.com` for EU/FE accounts
(and `token_url` to the matching `https://api.amazon.co.uk/auth/o2/token` /
`https://api.amazon.co.jp/auth/o2/token`). `token_url` defaults to
`https://api.amazon.com/auth/o2/token`, matching legacy's `regionTokenURL["NA"]`.

## Streams notes

Seven streams:

- `profiles` — `GET v2/profiles`. Unscoped (no Scope header). Records at the JSON array root
  (`records.path: ""`).
- `campaigns` — `GET v2/sp/campaigns`. Profile-scoped.
- `ad_groups` — `GET v2/sp/adGroups`. Profile-scoped.
- `portfolios` — `GET v2/portfolios`. Profile-scoped.
- `keywords` — `GET v2/sp/keywords`. Profile-scoped.
- `product_ads` — `GET v2/sp/productAds`. Profile-scoped.
- `negative_keywords` — `GET v2/sp/negativeKeywords`. Profile-scoped.

All five paginate identically with `pagination.type: offset_limit` (`limit_param: count`,
`offset_param: startIndex`, `page_size: 100`), matching legacy's `harvest` loop's
`startIndex`/`count` offset pagination with a short-page stop — `connsdk.OffsetPaginator` implements
the exact same "stop once a page returns fewer than the requested page size" rule.

Every v2 entity endpoint returns a top-level JSON array (no envelope object), hence `records.path: ""`
on every stream (the connsdk root-array convention, also used by e.g. `acuity-scheduling`).

**Every stream renames its raw camelCase fields to the schema's snake_case names via
`computed_fields`** (`profileId`→`profile_id`, `campaignId`→`campaign_id`, etc., matching legacy's
`*Record` mapper functions in `streams.go` field-for-field) — plain schema projection matches by exact
key only, so without these renames every field would silently drop from parity. `profiles` additionally
flattens the nested `accountInfo` object (`accountInfo.marketplaceStringId` →
`marketplace_string_id`, `.type` → `account_type`, `.name` → `account_name`, `.id` → `account_id`),
matching legacy's `profileRecord`'s type-asserted flatten.

These are full-refresh streams (no `incremental` block, no `x-cursor-field`) — legacy's own comment
notes there is no reliable updated-at cursor on the v2 entity endpoints, and `InitialState` never
persists anything beyond the stream marker.

## Write actions & risks

None. This connector is `capabilities.write: false`; no `writes.json` is shipped, matching legacy's
`Write` always returning an "amazon-ads connector is read-only" error.

## Known limits

- Sponsored Brands and Sponsored Display campaigns are not modeled because their record shapes differ
  from the Sponsored Products v2 entity surface carried by this bundle; they remain explicit
  `api_surface.json` exclusions rather than guessed schemas.
- Sponsored Products write endpoints use batch-array mutation bodies and partial-success semantics
  that the current write dialect cannot express without a `WriteHook`; reporting endpoints require
  async job creation/poll/download and conventionally require a `StreamHook`. No new hook packages
  are allowed for this shard, so those endpoints are excluded with typed reasons in
  `api_surface.json`.
- `base_url`/`token_url` region selection: only the North America endpoint is a spec default; EU/FE
  accounts must set both `base_url` and `token_url` explicitly (see "Auth setup").
- See "Auth setup" above for the AuthHook-folded conditional Scope header (a genuine Tier-2 need — no
  per-stream header override exists in the declarative dialect) and the re-exchange-per-`Read`-call
  deviation (never data-changing).
- `conformance`'s dynamic (fixture-replay) checks are skipped bundle-wide (`metadata.json`'s
  `conformance.skip_dynamic`): this connector's sole `auth` candidate is `mode: custom` with no
  `when`-gated non-custom fallback, so conformance's synthetic non-secret config (which can never
  populate a real `token_url`) cannot exercise it at all — every auth-resolving dynamic check would
  otherwise fail identically and uninformatively. The AuthHook's own unit tests
  (`internal/connectors/hooks/amazon-ads/hooks_test.go`) are the authoritative proof of the token
  exchange and conditional-header behavior.
