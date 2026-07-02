# Overview

Clazar is a cloud go-to-market platform (AWS/Azure/GCP marketplace co-sell, listings, private
offers, contracts). This bundle reads Clazar buyers, listings, contracts, opportunities, and
private offers through the Clazar REST API (`https://api.clazar.io`) using OAuth2 client
credentials. It is a wave2 fan-out migration of `internal/connectors/clazar` (the hand-written
connector it migrates); the legacy package stays registered and unchanged until wave6's registry
flip. Clazar is read-only here — legacy only supports full-refresh reads and exposes no
obviously-safe reverse-ETL writes; `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Auth setup

Provide Clazar OAuth2 client-credentials `client_id`/`client_secret` secrets (`mode:
oauth2_client_credentials`); the engine POSTs a `grant_type=client_credentials` token request to
`token_url` (`{{ config.base_url }}/authenticate/`, matching legacy's `base + "/authenticate/"`)
and applies the resulting bearer token to every request, refreshing before expiry
(`connsdk.OAuth2ClientCredentials`, the same authenticator legacy itself uses). Neither secret is
ever logged. `base_url` defaults to `https://api.clazar.io` and may be overridden for tests/proxies
(the override also reshapes the derived token endpoint, matching legacy's `clazarBaseURL` +
`clazarTokenPath` composition).

## Streams notes

All 5 streams (`buyers`, `listings`, `contracts`, `opportunities`, `private_offers`) share the
identical shape: `GET` against the Clazar list endpoint, `response_format=common` sent on every
request (legacy's own static query value), records at `results`, primary key `["id"]`, incremental
cursor field `last_modified_at`. Pagination is `page`/`page_size` (`pagination.type: page_number`,
`page_param: page`, `size_param: page_size`, `start_page: 1`, `page_size: 100` — legacy's own
default and max), matching legacy's `connsdk.PageNumberPaginator{PageParam: "page", SizeParam:
"page_size", StartPage: 1, PageSize: pageSize}` exactly (page-size stop threshold: a page
returning fewer than `page_size` records ends the stream). Incremental reads send
`last_modified_at_after` (`incremental.request_param`) computed either from the sync's persisted
cursor or, on a fresh sync, from the RFC3339 `start_date` config value — identical to legacy's
`incrementalLowerBound` (`cursor, then start_date, then absent`).

Legacy exposes `page_size`/`max_pages` as config-driven overrides (`clazarPageSize`/
`clazarMaxPages`, default 100, max 100). The engine's `PaginationSpec.PageSize`/`MaxPages` fields
are static JSON integers (no `{{ config.* }}` templating support), so these cannot be wired as
runtime-configurable knobs — this bundle declares neither `page_size` nor `max_pages` in
`spec.json` at all (a declared-but-unwireable key is worse than an absent one, per
conventions.md F6), matching bitly's/sendpulse's identical documented limitation. `page_size: 100`
(legacy's own default) is pinned statically in `streams.json`'s base pagination block.

## Write actions & risks

None. Clazar is a read-only source connector (legacy's own package doc: "Clazar is read-only
here... Capabilities.Write is false"); this bundle ships no `writes.json`, matching legacy's
`Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **Dynamic (fixture-replay) conformance checks are marked `skip_dynamic` at the bundle level**
  (`metadata.json`'s `conformance` block). `oauth2_client_credentials` auth's `token_url` is
  derived from `{{ config.base_url }}/authenticate/`; conformance's `withReplayURL` only
  overrides `b.HTTP.URL` (the base request URL used for stream/check paths), never
  `RuntimeConfig.Config["base_url"]` itself, so the `token_url` template still resolves to the
  synthetic non-secret value (`"synthetic-conformance-value/authenticate/"`), an unreachable
  non-URL — the OAuth token exchange fails before any declarative stream/check request is ever
  issued, so every auth-resolving dynamic check (`check_fixture`, `read_fixture_nonempty:*`,
  `pagination_terminates`, `records_match_schema`, `cursor_advances`) would otherwise fail
  identically and uninformatively, for a reason that has nothing to do with this bundle's actual
  read/pagination/incremental/schema-projection shape. Static checks (`spec_schema_valid`,
  `stream_schemas_valid`, `interpolations_resolve`, `docs_present`, `fixtures_present`,
  `secret_redaction`, etc.) are unaffected and still run. This bundle has no Tier-2 `AuthHook`
  (auth is fully declarative `oauth2_client_credentials`), so there is no `paritytest/clazar`
  package for this wave; the read/pagination/incremental/schema shape is proven by structural
  review against legacy `internal/connectors/clazar` instead. Matches sendpulse's identical
  documented precedent.
- Full Clazar API surface (listing/private-offer/contract writes) is out of scope; see
  `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}`
  entries. Only the 5 legacy-parity read streams are implemented.
