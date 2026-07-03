# Overview

Apple Search Ads (Apple Ads) reads campaigns, ad groups, targeting keywords, and ads through the
Apple Search Ads Campaign Management API. This bundle migrates
`internal/connectors/apple-search-ads` (the hand-written connector) to a declarative-bundle-plus-hook
shape at capability parity; the legacy package stays registered and unchanged until wave6's registry
flip. It is read-only — Apple's Campaign Management API does expose mutating operations
(create/update campaigns, ad groups, keywords), but legacy exposes none of them as reverse-ETL
targets, matching this bundle's `capabilities.write: false`.

## Auth setup

Apple Search Ads authenticates with a standard OAuth2 `client_credentials` grant: `client_id`/
`client_secret` are exchanged at Apple's token endpoint (`token_refresh_endpoint`, defaulting to
`https://appleid.apple.com/auth/oauth2/token`) for a short-lived access token, sent on every data
request as `Authorization: Bearer <access_token>`. Unlike amazon-ads/amazon-seller-partner's
refresh_token grant, this is the engine's native `oauth2_client_credentials` auth mode — **no
AuthHook is needed**; `streams.json` `base.auth` declares it directly with `scopes: "searchadsorg"`
(sent as the form-encoded `scope` parameter on the token request, matching legacy's
`clientCredentialsAuth.accessToken`'s `form.Set("scope", ...)`). Legacy's default token endpoint
literal additionally carries `grant_type`/`scope` as URL query params
(`https://appleid.apple.com/auth/oauth2/token?grant_type=client_credentials&scope=searchadsorg`) —
this bundle's `token_refresh_endpoint` default omits them since the engine's
`OAuth2ClientCredentials` authenticator always sends both as form fields regardless, which Apple's
token endpoint accepts identically whether or not the URL itself repeats them.

Provide `client_id` and `client_secret` (both `x-secret: true`, never logged) and `org_id` (NOT
secret — an organization identifier, not a credential). `org_id` is sent on every Campaign Management
API request as the `X-AP-Context: orgId=<org_id>` header, declared directly in `streams.json`
`base.headers` (a single global header, unconditional — unlike amazon-ads's per-endpoint-conditional
Scope header, every Apple Search Ads endpoint this bundle reads is org-scoped, so no Tier-2 hook is
needed for this header).

## Streams notes

Four streams, matching legacy's `appleStreamEndpoints` table exactly. Apple Search Ads exposes two
access patterns for these streams, both returning the same `{data:[...], pagination:
{totalResults,startIndex,itemsPerPage}}` envelope:

- `campaigns` — `GET /campaigns` with `offset`/`limit` query params.
- `adgroups` — `POST /adgroups/find`, org-wide (not scoped to one campaign in the request), with a
  `{pagination:{offset,limit}}` selector BODY.
- `keywords` — `POST /targetingkeywords/find`, same body-selector shape.
- `ads` — `POST /ads/find`, same body-selector shape.

**All 4 streams are read through a single Tier-2 `StreamHook`**
(`internal/connectors/hooks/apple-search-ads/hooks.go`), not just the 3 POST `.../find` streams.
`campaigns` alone COULD be expressed fully declaratively (`method: GET`, `pagination.type:
offset_limit`) — `streams.json` still declares that complete shape for manifest/catalog purposes —
but the 3 `.../find` streams cannot: `engine/bundle.go`'s `StreamSpec.Body` field exists but
`engine/read.go`'s declarative read path never reads it back out (the declarative path always issues
a nil body), so a selector-body-carried pagination request has no declarative expression at all. This
is the same class of Tier-2 trigger conventions.md §1 names ("sub-resource fan-out reads" /
body-carried pagination state) that monday's and plaid's hooks document; folding `campaigns` into the
same hook too (rather than splitting it out as the one declarative-eligible stream) matches monday's
own documented precedent of routing every stream through one `StreamHook` for consistency, even where
not every individual stream strictly required it.

The hook's `harvest` loop (ported verbatim from legacy's `apple_search_ads.go`) stops pagination on
the first short page (fewer records than the requested page size) OR once the running total reaches
the response's `pagination.totalResults` field, whichever comes first — matching legacy's exact dual
stop condition.

**Every raw camelCase field is renamed to the schema's snake_case name via `computed_fields`**
(`campaignId`→`campaign_id`, `servingStatus`→`serving_status`, etc., matching legacy's `*Record`
mapper functions in `streams.go` field-for-field) — this is manifest/documentation parity for the
declarative shape `streams.json` still carries; the ACTUAL record shape at runtime comes from the
hook's own Go-level `*Record` mapper functions (`campaignRecord`/`adGroupRecord`/`keywordRecord`/
`adRecord`), a direct verbatim port of legacy's.

`modification_time` is declared as `x-cursor-field`/`incremental.cursor_field` on every stream
(matching legacy's `appleStreams()`'s published `CursorFields: []string{"modification_time"}`), but —
matching legacy exactly — this is manifest-only: `incremental` declares NO `request_param`, and
legacy's `Read` never consults `req.State` (its `InitialState` always returns just the stream marker),
so both connectors always perform a full stream read regardless of any incremental state passed in.
This is the identical pattern monday's `boards`/`items` streams document (a published cursor field for
sync-mode-derivation/manifest purposes, with no server-side filter actually wired).

## Write actions & risks

None. This connector is `capabilities.write: false`; no `writes.json` is shipped, matching legacy's
`Write` always returning an "apple-search-ads connector is read-only" error.

## Known limits

- Creative sets, geo/demographic targeting, spend/performance reports, and every write/mutating
  endpoint are out of scope for this migration; see `api_surface.json`'s `excluded: {category:
  out_of_scope, reason: "Pass B capability expansion"}` entries. Legacy never implemented them
  either.
- `conformance`'s dynamic (fixture-replay) checks are skipped bundle-wide (`metadata.json`'s
  `conformance.skip_dynamic`) for two independent, additive reasons: (1) `oauth2_client_credentials`'s
  `token_url` is a separate declared `config.token_refresh_endpoint` property that conformance's
  replay harness never overrides (it only rewrites `b.HTTP.URL`), so the token exchange always
  targets the synthetic non-secret value — an unreachable non-URL — before any declarative
  stream/check request is ever issued (matches box's/clazar's/kyriba's identical documented
  precedent); (2) 3 of the 4 streams are `StreamHook`-handled with body-carried pagination that has
  no declarative-fixture-replay equivalent at all. The `StreamHook`'s own unit tests
  (`internal/connectors/hooks/apple-search-ads/hooks_test.go`) are the authoritative proof of the
  read/pagination/record-mapping behavior for all 4 streams. Static checks (spec/schema validity,
  interpolations_resolve, docs/fixtures presence, secret redaction) still run and pass.
