# Overview

SendPulse is a wave2 fan-out migration. This bundle reads SendPulse address books, campaigns, and
senders through the SendPulse REST API, migrating `internal/connectors/sendpulse` (the legacy
hand-written connector, which stays registered and unchanged until wave6's registry flip) at
capability parity.

## Auth setup

Provide `client_id`/`client_secret` secrets (SendPulse API app credentials); the bundle exchanges
them for a bearer token via OAuth2 client-credentials (`auth.mode: oauth2_client_credentials`)
against `token_url`, matching legacy's `connsdk.OAuth2ClientCredentials`. `token_url` defaults to
`https://api.sendpulse.com/oauth/access_token` (legacy's own hardcoded literal-append default);
`base_url` independently defaults to `https://api.sendpulse.com`. See Known limits for the
narrowed case where only one of the two is overridden.

## Streams notes

All 3 streams (`addressbooks`, `campaigns`, `senders`) share the same shape: `GET` against the
SendPulse list endpoint, records at the JSON body root (`records.path: "."`, matching legacy's
`recordsPath: ""` — SendPulse's list endpoints return a bare JSON array, no envelope object).
Pagination is `page_number` (`page_param: page`, `size_param: limit`, `start_page: 1`,
`page_size: 100`), matching legacy's `connsdk.PageNumberPaginator{PageParam: "page", SizeParam:
"limit", StartPage: 1, PageSize: pageSize}` at its default `page_size` value (100 — legacy's
`defaultPageSize`). `max_pages` is intentionally left UNDECLARED in `base.pagination` (unbounded,
per `PaginationSpec.MaxPages`'s `<= 0` = unbounded rule) rather than baked in at legacy's own
`defaultMaxPages: 1` — see Known limits for why. Neither `page_size` nor `max_pages` is exposed as
a runtime-configurable `spec.json` property: the `page_number` paginator's `size_param` query
value and its own short-page stop threshold both come from the SAME `PaginationSpec.PageSize` JSON
literal, and `mergeQuery` (`read.go`) lets the paginator's own query entry unconditionally
overwrite any same-keyed `stream.Query` entry — a `{{ config.page_size }}`-templated `limit` query
param would therefore be dead code, silently discarded every request, not a genuine override; and
`PaginationSpec.MaxPages` has no templating mechanism at all. None of the 3 streams are
incremental in legacy (no `cursor_field`/`incremental` handling anywhere in `sendpulse.go`), so
none of the `streams.json` entries declare an `incremental` block. Primary keys:
`addressbooks`/`campaigns` use `id` (integer), `senders` uses `email` (string) — matching legacy's
declared `PrimaryKey`.

## Write actions & risks

None. This connector is read-only, matching legacy's `Write` stub (`connectors.ErrUnsupportedOperation`).

## Known limits

- **Dynamic (fixture-replay) conformance checks are marked `skip_dynamic` at the bundle level**
  (`metadata.json`'s `conformance` block). `oauth2_client_credentials` auth requires a real
  `token_url` to POST a token request against; conformance's synthetic non-secret config value
  (the literal string `"synthetic-conformance-value"`) is not a resolvable URL, so the token
  exchange fails before any declarative stream/check request is ever issued — every auth-resolving
  dynamic check (`check_fixture`, `read_fixture_nonempty:*`, `pagination_terminates`,
  `records_match_schema`, `cursor_advances`) would otherwise fail identically and
  uninformatively, for a reason that has nothing to do with this bundle's actual read/pagination/
  schema-projection shape. Static checks (`spec_schema_valid`, `stream_schemas_valid`,
  `interpolations_resolve`, `docs_present`, `fixtures_present`, `secret_redaction`, etc.) are
  unaffected and still run. This bundle has no Tier-2 `AuthHook` (its auth is fully declarative
  `oauth2_client_credentials`), so there is no `paritytest/sendpulse` package for this wave; the
  read/pagination/schema shape is proven by structural review against legacy
  `internal/connectors/sendpulse` instead.
- **`max_pages` is left unbounded (undeclared in `base.pagination`) rather than baked in at
  legacy's own default of 1.** Legacy's `defaultMaxPages = 1` is itself just a config-overridable
  default (`req.Config.Config["max_pages"]`), not load-bearing single-page-only behavior; since
  `PaginationSpec.MaxPages` cannot be wired to config at all (no templating mechanism — see
  `docs/migration/conventions.md`'s pagination table), the two faithful options were "hardcode 1"
  (silently narrows every real sync to a single page with no override, an accepted-input-behavior
  change for any caller that previously passed a larger `max_pages`) or "leave unbounded" (widens
  a fresh full sync to exhaust the stream via the short-page stop rule alone, matching the OTHER
  common legacy default shape used elsewhere in this codebase, e.g. `elasticemail`). This bundle
  chooses unbounded so real syncs are not artificially truncated, and so the required 2-page
  conformance fixture (`fixtures/streams/*/page_{1,2}.json`) can prove genuine pagination
  termination via the short-page stop rule rather than the `max_pages` cap. Documented as a
  parity-deviation ledger candidate: an operator who explicitly relied on legacy's default
  single-page-per-sync behavior (with no `max_pages` override) now receives every available page
  instead. Never changes individual record DATA for any given page, only how many pages are
  fetched.
- `token_url`'s default is a fixed literal (`https://api.sendpulse.com/oauth/access_token`), not
  a `base_url`-derived value. Legacy derives `token_url` from whatever `base_url` resolves to at
  runtime (`strings.TrimRight(base, "/") + "/oauth/access_token"`), so a caller overriding
  `base_url` alone (leaving `token_url` unset) gets a token endpoint under the SAME custom host in
  legacy, but would still hit the fixed default host here. The engine's `spec.json` `"default"`
  materialization mechanism fills only a literal per property, with no cross-property derivation
  (conventions.md §3, the sentry/chargebee derived-default case) — this is a documented,
  accepted config-surface narrowing: a caller who overrides `base_url` for a test/proxy setup must
  also override `token_url` explicitly to point at the same host. Not exercised by any of this
  bundle's fixtures (both default to the real SendPulse hosts).
- Full SendPulse API surface (SMTP/transactional email sending, SMS, chatbots, push
  notifications, web push, contact-level mutations) is out of scope for wave2; see
  `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B capability
  expansion"}` entries. Only the 3 legacy-parity read streams are implemented.
