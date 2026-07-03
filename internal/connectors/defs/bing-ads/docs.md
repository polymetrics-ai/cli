# Overview

Bing Ads (Microsoft Advertising) is a **Tier-3 native package** migration
(`internal/connectors/native/bing-ads/`) of `internal/connectors/bing-ads` (legacy, read-only
reference until the wave6 registry flip). It reads Microsoft Advertising accounts, users,
campaigns, ad groups, and ads through the v13 Customer Management and Campaign Management REST
services. Every legacy operation is an HTTP `POST` to a `.../Query`-style endpoint with a small
JSON body whose response wraps records in a named array (`{"AccountsInfo":[...]}`,
`{"Campaigns":[...]}`); auth is the Microsoft identity platform OAuth 2.0 **refresh-token grant**.
This package is engine-vs-legacy parity-tested against `internal/connectors/bing-ads`. Read-only:
legacy `bing_ads.go:173-175` always returns `ErrUnsupportedOperation` from `Write`, and this package
declares `capabilities.write: false` with no `writes.json` to match.

This bundle still ships `spec.json`/`streams.json`/`schemas/*.json`/`api_surface.json`/`docs.md` so
identity/spec/docs/schema stay uniform with every other connector (`native/bing-ads`'s
`connector.go` embeds `engine.Base`, built from this bundle, purely to serve
`Name()`/`Metadata()`/`Definition()`) — but `streams.json`'s `base`/`streams[]` entries are
**documentation/schema-reference only**: `Check`/`Catalog`/`Read` are hand-written Go
(`internal/connectors/native/bing-ads/{connection,reader,cataloger}.go`), never routed through
`engine.Check`/`engine.Read`. `metadata.json` does not set `capabilities.dynamic_schema: true`
(unlike the postgres Tier-3 golden) because Bing Ads' 5-stream schema is statically known ahead of
time, not discovered at runtime from the live API — `streams.json` is present specifically so its
real schema/path/records-shape documentation has a home, even though it is never executed.

## Why Tier 3, not Tier 2

This connector was FIRST attempted as a Tier-2 hook set (AuthHook + StreamHook, mirroring gmail's
refresh-token AuthHook and monday's POST-body StreamHook almost exactly) and escalated per
conventions.md §1/§6 rule 2 ("exceeding 400 [lines], OR needing a 3rd hook interface... escalates
to Tier 3 rather than stretching a hook package"): the combined hook set needed **3** hook
interfaces (AuthHook for the OAuth refresh grant, StreamHook for the POST-body routing/per-service
base URL/per-account fan-out, and CheckHook for a real POST-body health check — a bounded
`AccountsInfo/Query` matching legacy's own `Check`) and **~570 lines**, both past the Tier-2 caps.
Three independent gaps compound onto this single connector (any one alone would likely have stayed
within Tier-2 caps, as monday's single GraphQL-body gap did):

1. `StreamSpec.Body` is declared in `engine/bundle.go` but deliberately unwired in the declarative
   read path (`engine/read.go` always sends `nil` as the request body) — the same gap monday's
   GraphQL POST hit.
2. Campaign-scoped streams (`campaigns`/`ad_groups`/`ads`) target a SECOND base URL (Campaign
   Management, `config.campaign_base_url`) that the engine's single `base.url` template cannot
   express per-stream.
3. `campaigns`' per-account-id fan-out injects the id into the JSON **body** (`{"AccountId":
   "<id>"}`), not a query param or path variable — `FanOutSpec.Into` supports only
   `query_param`/`path_var` (S4 engine mini-wave item 2), never a body field.

See the `ENGINE_GAP` entry in Known limits below for gap 3 specifically (the other two are
structural Tier-3 triggers, not expressible via any future dialect addition without collapsing the
Customer/Campaign service split entirely).

## Auth setup

Provide `client_id`, `developer_token`, and `refresh_token` secrets (all required); `client_secret`
and `tenant_id` are optional secrets. `connection.go`'s `oauthRefreshAuth` ports legacy `auth.go`'s
identically-named type almost verbatim: it POSTs `grant_type=refresh_token` + `refresh_token` +
`client_id` [+ `client_secret`] + `scope` (fixed to `https://ads.microsoft.com/msads.manage
offline_access`, matching legacy's `defaultScope`) to `token_url` (default
`https://login.microsoftonline.com/<tenant_id>/oauth2/v2.0/token`, tenant defaulting to `common`),
caches the resulting access token until 60 seconds before its declared expiry, and sets
`Authorization: Bearer <access_token>` on every request. `token_url` MUST resolve to an `https://`
URL in production — `resolveBaseURL`'s scheme guard (http-or-https, matching legacy exactly) bounds
`base_url`/`campaign_base_url`/`token_url` overrides alike (parity with legacy's `resolveBaseURL`/
`tokenURL`, bing_ads.go:296-350, which validate all three identically). `developer_token` is sent
as the `DeveloperToken` header on every request — never logged.

## Streams notes

Five streams, all primary-keyed on `Id`, full_refresh only (legacy's own doc comment,
streams.go:117-121, states the upstream catalog only advertises full_refresh, matching legacy
exactly — no stream publishes a cursor field): `accounts` (Customer Management, `AccountsInfo/Query`,
conditionally includes `CustomerId` in the POST body when `config.customer_id` is set), `users`
(Customer Management, `User/Query`, a static `{"UserId": null}` body), and
`campaigns`/`ad_groups`/`ads` (Campaign Management service, scoped by `AccountId`/`CampaignId`/
`AdGroupId` respectively in the POST body). `campaigns` fans out over `config.account_ids` (a
comma-separated list; falls back to a single id from `customer_account_id`/`account_id` when unset,
exactly matching legacy's `accountIDList`), issuing one `QueryByAccountId` POST per account id and
concatenating the results with NO account-origin marker stamped onto the record (legacy's
`campaignRecord` never adds one — see the parity-deviation ledger entry below). `Check()` issues a
bounded `AccountsInfo/Query` POST (mirroring legacy's `Check`, bing_ads.go:97-103), confirming the
OAuth exchange, `DeveloperToken`, and connectivity without mutating anything.

A `mode=fixture` config (`cfg.Config["mode"]=="fixture"`) short-circuits all network access,
mirroring legacy's identical `fixtureMode` gate and every other Tier-1/2/3 connector's
credential-free conformance affordance: `Check` succeeds trivially, and `Read` emits the same
deterministic per-stream fixture records legacy's `readFixture` generates (see `cataloger.go`).

## Write actions & risks

None — Bing Ads is read-only here. `capabilities.write: false`, no `writes.json` file, matching
legacy's `ErrUnsupportedOperation` (`bing_ads.go:173-175`).

## Known limits

- **`ENGINE_GAP` — no declarative `fan_out` body-field target**: `FanOutSpec.Into`
  (`engine/bundle.go`) supports only `query_param` (the resolved id added as a URL query parameter)
  or `path_var` (referenceable in a stream's `path` template) — never a JSON request-body field.
  Bing Ads' `campaigns` stream needs exactly that (`{"AccountId": "<id>"}` in a POST body), which is
  why this connector could not collapse to a Tier-1 `fan_out` block even after solving the
  POST-body-in-general gap. A future `FanOutSpec.Into.BodyField` addition, paired with a
  declarative per-stream `Body` template (once `StreamSpec.Body` itself is wired into the read
  path — currently dead, see monday's identical finding), would let this collapse to a Tier-1 or
  thin Tier-2 shape; deferred per conventions.md §6 ("`ENGINE_GAP`s recur >=3 times -> the
  orchestrator extends the engine in a mini wave-0 increment" — monday's GraphQL-in-body pagination
  and this connector's body-field fan-out are two independent instances of the same underlying
  "POST body needs runtime-templated content" gap; a third instance would meet the threshold).
- **Tier-2 escalation, not a Tier-1/Tier-2 workaround**: see "Why Tier 3, not Tier 2" above — this
  is a structural decision (3 hook interfaces + ~570 lines under the Tier-2 attempt), not a
  correctness gap being silently worked around.
- **`account_id`/`account_ids`/`customer_id`/`campaign_id`/`ad_group_id` are plain (non-secret)
  `spec.json` string properties, matching legacy's own plain-config treatment** — none of these
  identify a credential; they are read/write scoping values.

## Parity-deviation ledger entry

`campaigns`' cross-account fan-out concatenates every configured account's campaigns into one flat
stream with no account-origin marker, exactly matching legacy's `campaignRecord`
(bing_ads/streams.go) — this is not a deviation, just documented here since it is easy to assume a
fan-out implies a stamped origin field. Preserving legacy's exact emitted-record shape, including
the absence of a field legacy never emitted, is the correct behavior per conventions.md §5's
meta-rule, not a gap.
