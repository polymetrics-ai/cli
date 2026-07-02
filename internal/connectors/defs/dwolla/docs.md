# Overview

Dwolla is a payments platform exposing a HAL+JSON REST API. This bundle reads Dwolla customers,
events, exchange partners, and business classifications through `https://api.dwolla.com` using
OAuth2 client-credentials auth. It is read-only, migrated from `internal/connectors/dwolla` (the
hand-written connector this bundle replaces at capability parity); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide `client_id` and `client_secret` secrets; the engine's declarative
`oauth2_client_credentials` auth mode mints and caches a Bearer token via the OAuth2
client-credentials grant against `{{ config.base_url }}/token` (Dwolla hosts the token endpoint at
`/token` on the same host as the API, matching legacy's `dwollaTokenURL` derivation), refreshing
automatically before expiry — matching legacy's `connsdk.OAuth2ClientCredentials`. Both secrets
flow only into the token exchange and are never logged. Every request also sends
`Accept: application/vnd.dwolla.v1.hal+json` (Dwolla's HAL media type, matching legacy's
`dwollaHALAccept`).

Set `base_url` to `https://api-sandbox.dwolla.com` for Dwolla's sandbox environment, or leave unset
for production (`https://api.dwolla.com`, the default). See Known limits for the narrowed
`environment` config surface.

## Streams notes

Four top-level list streams, all sharing Dwolla's HAL pagination shape: records live under
`_embedded.<key>` and the next page is an absolute URL at `_links.next.href`
(`pagination.type: next_url`, `next_url_path: "_links.next.href"`), matching legacy's `harvest`
loop exactly. The initial request for every stream sends `limit=25` (legacy's
`dwollaDefaultPageSize`).

- `customers` (`GET /customers`, records at `_embedded.customers`).
- `events` (`GET /events`, records at `_embedded.events`).
- `exchange_partners` (`GET /exchange-partners`, records at `_embedded.exchange-partners` — the
  HAL embed key uses a hyphen, matching legacy's `embedKey: "exchange-partners"`; dotted-path
  extraction splits only on `.`, so the hyphenated segment is unaffected).
- `business_classifications` (`GET /business-classifications`, records at
  `_embedded.business-classifications`; reference data with no `created` timestamp, matching
  legacy's own catalog which omits `CursorFields` for this stream).

Every other stream's primary key is `["id"]` and incremental cursor field is `created`, matching
legacy's uniform catalog — but no stream declares an `incremental` block: Dwolla's list endpoints
accept no server-side `updated_since`-style filter parameter, matching legacy's own `InitialState`
(always an empty cursor; `start_date` config accepted by legacy's own comment but never actually
consumed as a request filter either).

## Write actions & risks

None. Dwolla is read-only in both legacy and this bundle (`capabilities.write: false`); no
`writes.json` file is shipped. Legacy's own comment explains why: Dwolla is an upstream source
with no reverse-ETL surface.

## Known limits

- **Dynamic (fixture-replay) conformance checks are marked `skip_dynamic` at the bundle level**
  (`metadata.json`'s `conformance` block). `oauth2_client_credentials` auth's `token_url` is
  derived from `{{ config.base_url }}/token`; conformance's `withReplayURL` only overrides
  `b.HTTP.URL` (the base request URL used for stream/check paths), never
  `RuntimeConfig.Config["base_url"]` itself, so the `token_url` template still resolves to the
  synthetic non-secret value (`"synthetic-conformance-value/token"`), an unreachable non-URL — the
  OAuth token exchange fails before any declarative stream/check request is ever issued, so every
  auth-resolving dynamic check would otherwise fail identically and uninformatively. Static checks
  (spec/schema validity, `interpolations_resolve`, docs/fixtures presence, secret redaction) still
  run and pass. This bundle has no Tier-2 `AuthHook` (auth is fully declarative
  `oauth2_client_credentials`), so there is no `paritytest/dwolla` package for this wave; the
  read/pagination/schema-projection shape is proven by structural review against legacy
  `internal/connectors/dwolla` instead. Matches `clazar`'s and `sendpulse`'s identical documented
  precedent.
- Legacy accepted an `environment` config value (`api`/`api-sandbox`) as an alternative to an
  explicit `base_url` override, deriving the production/sandbox host from it. The engine's
  `spec.json` `"default"` materialization only supports a fixed literal default, not one derived
  from another config value at read/check time — the same limitation `docs/migration/
  conventions.md` documents for sentry's `hostname`-derived base URL. This bundle narrows the
  config surface to `base_url` only (defaulting to production); a caller who needs the sandbox
  host sets `base_url` to `https://api-sandbox.dwolla.com` directly instead of `environment:
  api-sandbox`. This narrows accepted CONFIGURATION surface only, never emitted record data.
  `funding-sources`/`transfers` and any per-customer-scoped resource are out of scope for this
  wave, matching legacy's own "intentionally out of this first cut" scoping; see
  `api_surface.json`'s `excluded` entries.
- `metadata.json` declares no `rate_limit` block: legacy enforces no client-side rate limiting for
  Dwolla, so none is added here either (matches legacy's real, lack-of, throttling behavior).
