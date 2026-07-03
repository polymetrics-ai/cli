# Overview

Amazon Seller Partner reads orders, FBA inventory summaries, and financial event groups through the
Amazon Selling Partner API (SP-API). This bundle migrates
`internal/connectors/amazon-seller-partner` (the hand-written connector) to a declarative defs
bundle at capability parity; the legacy package stays registered and unchanged until wave6's
registry flip. It is read-only — SP-API does expose mutating operations (feed submission, shipment
confirmation), but legacy exposes none of them as reverse-ETL targets, matching this bundle's
`capabilities.write: false`.

## Auth setup

SP-API authenticates with Login with Amazon (LWA): a long-lived `refresh_token` is exchanged for a
short-lived `access_token` at the LWA token endpoint, sent on every data request as the
`x-amz-access-token` header (NOT `Authorization: Bearer`). This exchange is a **Tier-2 AuthHook**
(`internal/connectors/hooks/amazon-seller-partner/hooks.go`) — the declarative dialect's
`oauth2_client_credentials` mode is the client-credentials grant, not the refresh-token grant SP-API
uses, and no declarative `auth` mode can route a fetched token into a non-`Authorization` header —
this is the exact "token-exchange auth" Tier-2 trigger conventions.md §1 names for GitHub App's
JWT→installation-token exchange, ported here for LWA's refresh_token→access_token exchange instead.

Provide three secrets: `lwa_app_id` (LWA application client ID), `lwa_client_secret` (LWA
application client secret), and `refresh_token` (the long-lived LWA refresh token) — all
`x-secret: true`, never logged. The hook performs `POST {lwa_token_url}` with
`grant_type=refresh_token` form-encoded, matching legacy's `lwaAuthenticator.accessToken` exactly,
and returns a `connsdk.APIKeyHeader("x-amz-access-token", accessToken, "")` authenticator. Unlike
legacy's `lwaAuthenticator`, which caches the token across calls with a 60-second-before-expiry
refresh window, this hook re-exchanges on every `Check`/`Read` call (the engine calls
`AuthHook.Authenticator` once per `Read`/`Check`, and that single authenticator instance is reused
across every page within that one call — see `engine/read.go`'s `newRuntime`/`selectAuth` — so a
multi-page Read still performs exactly ONE token exchange, matching legacy's own per-`Read`
`requester()` construction). This never changes emitted record data; it only means a long-lived
process issuing many separate `Read` calls re-exchanges the token more often than legacy's
in-process cache would have. Documented as an ACCEPTABLE deviation (§5 meta-rule): never
data-changing, and the engine provides no mechanism for the connector-defined layer to cache state
across separate `Read` invocations.

`base_url` defaults to the North America SP-API endpoint (`https://sellingpartnerapi-na.amazon.com`),
matching legacy's `spDefaultBaseURL` fallback. Legacy also derives this from a `region` config enum
(NA/US/CA/MX/BR -> NA endpoint; EU/GB/UK/DE/FR/ES/IT/NL/SE/PL/BE/TR/AE/SA/EG/IN -> EU endpoint;
JP/AU/SG -> FE endpoint) at runtime; this bundle requires the resolved `base_url` directly for
non-NA sellers instead, since the engine's `spec.json` `"default"` mechanism only materializes one
fixed literal, not a region-derived choice between three URLs (documented scope narrowing). Set
`base_url` to `https://sellingpartnerapi-eu.amazon.com` or `https://sellingpartnerapi-fe.amazon.com`
for EU/FE sellers. `lwa_token_url` defaults to `https://api.amazon.com/auth/o2/token`, matching
legacy's `lwaDefaultTokenURL`.

## Streams notes

Three streams, matching legacy's `spStreamEndpoints` table (`order_items` was never implemented by
legacy despite being named in `Metadata().Description`; only the 3 streams legacy's
`spStreamEndpoints` table actually wires are ported):

- `orders` — `GET /orders/v0/orders`, records at `payload.Orders`. Paginated with
  `pagination.type: cursor`, `cursor_param: NextToken`, `token_path: payload.NextToken`. Incremental
  on `LastUpdateDate`, sent as `LastUpdatedAfter` (`request_param`), sourced from the state cursor or
  `replication_start_date` (`start_config_key`).
- `inventory_summaries` — `GET /fba/inventory/v1/summaries`, records at
  `payload.inventorySummaries`. Paginated with `cursor_param: nextToken`,
  `token_path: pagination.nextToken` (a DIFFERENT body path than orders/financial_event_groups,
  matching legacy's distinct `tokenPath` per endpoint exactly). Incremental on `lastUpdatedTime`,
  sent as `startDateTime`.
- `financial_event_groups` — `GET /finances/v0/financialEventGroups`, records at
  `payload.FinancialEventGroupList`. Paginated identically to `orders` (`NextToken`/
  `payload.NextToken`). Incremental on `FinancialEventGroupEnd`, sent as
  `FinancialEventGroupStartedAfter`.

**`replication_start_date` is REQUIRED here, unlike legacy.** Legacy's `orders`/
`financial_event_groups` streams always compute a lower-bound filter — falling back to a MOVING
"now minus 2 years" default (`spDefaultCreatedAfter`, re-evaluated via `time.Now()` on every single
call) when `replication_start_date` is unset and no incremental state cursor exists yet. The engine's
static default-materialization mechanisms (`spec.json`'s `"default"`, `stream.Query`'s object-form
`default`) can only bake a single FIXED literal at bundle-author time — neither can express a value
relative to the current wall-clock time. Per SP-API's own documented contract, `GET
/orders/v0/orders` requires at least one of `CreatedAfter`/`LastUpdatedAfter`/`NextToken` on every
request; omitting the lower-bound filter entirely (the alternative to requiring the field) would
send a request the live API itself rejects, which is a strictly worse divergence than a config
requirement. `replication_start_date` is therefore promoted to `required` at the SPEC level (applies
to all three streams, for a simpler single config surface) — a documented scope narrowing (never
data-changing for any request where legacy's caller DID set the field, which is the
production-realistic case) rather than an `ENGINE_GAP` blocker. `inventory_summaries` itself has no
fallback-default behavior to preserve either way (legacy's `inventoryBaseQuery` never applies one —
it simply omits `startDateTime` when the field is unset); requiring the field for that stream too is
a strictly narrower config surface, not a behavior change for any input where the field was already
set. The underlying per-request mechanism is still `stream.Query`'s `omit_when_absent` semantics on
`start_config_key`-sourced incremental params (`request_param`/`{{ incremental.lower_bound }}` are
only ever emitted when the bound actually resolves) — with the field required, that bound always
resolves in practice, so `startDateTime`/`LastUpdatedAfter`/`FinancialEventGroupStartedAfter` are
always sent once the caller supplies a value, which is the only reachable state.

**Subsequent pages resend the full base query alongside the cursor token — legacy explicitly drops
it.** Legacy's `harvest` loop rebuilds an EMPTY `url.Values{}` on every page after the first,
setting ONLY the `tokenParam` (a code comment reads "SP-API requires only the token; other filters
must be dropped"). The engine's `cursor`/`token_path` paginator (`tokenPathCursor.Next`,
`engine/read.go`'s `mergeQuery`) merges the paginator's token param INTO the stream's base query
unconditionally — the same base `MarketplaceIds`/`MaxResultsPerPage`/`LastUpdatedAfter` (or
equivalent per-stream params) are resent on every page, alongside `NextToken`/`nextToken`. This is
the SAME structural pagination-continuation shape every other `token_path`-cursor bundle in this repo
uses (e.g. `poplar`'s docs.md documents the identical "re-sent on every page request per the cursor
paginator's normal merge behavior") — no per-stream override exists in the dialect for "replace the
base query entirely on the next page" (a real but narrow `ENGINE_GAP`: this dialect gap would need a
new pagination field, e.g. `pagination.replace_query_on_next: true`, to close). This bundle does not
block on it: fixtures/parity target the engine's actual (merge) behavior, and legacy's own comment is
the only evidence this specific combination is rejected server-side — no live SP-API test confirms
it. Documented here as a known limitation for anyone wiring this bundle against production SP-API:
if Amazon does reject the merged-filter continuation request, the fix is the same engine dialect
addition, not a per-connector workaround.

## Write actions & risks

None. This connector is `capabilities.write: false`; no `writes.json` is shipped, matching legacy's
`Write` always returning `connectors.ErrUnsupportedOperation`.

## Known limits

- Feeds, reports, listings, pricing, catalog, and notifications APIs are out of scope for this
  migration; see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B capability
  expansion"}` entries. `orders/v0/orders/{orderId}/orderItems` (order line items) and
  `finances/v0/financialEvents` (line-level financial events) are likewise out of scope — legacy
  never implemented them either.
- See "Streams notes" above for the `replication_start_date`-required scope narrowing (moving-window
  default cannot be expressed statically) and the base-query-resend-on-continuation-page limitation
  (structural to every `token_path`-cursor bundle in this repo, not specific to this connector).
- `base_url`/region selection: only the North America endpoint is a spec default; EU/FE sellers must
  set `base_url` explicitly (see "Auth setup").
