# Overview

PayPal Transaction is a read-only source connector. It reads PayPal transaction history, account
balances, catalog products, and customer disputes through the PayPal REST API
(`https://api-m.paypal.com`), authenticating with the OAuth 2.0 client-credentials grant. This
bundle migrates `internal/connectors/paypal-transaction` (the hand-written legacy connector); the
legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

PayPal authenticates with the OAuth 2.0 client-credentials grant, but its token endpoint
authenticates the CLIENT with HTTP Basic (`Authorization: Basic base64(client_id:client_secret)`)
rather than the standard form-encoded `client_id`/`client_secret` the engine's declarative
`oauth2_client_credentials` auth mode always sends
(`engine/auth.go`'s `buildOAuth2ClientCredentials` / `connsdk.OAuth2ClientCredentials.accessToken`
POST a `grant_type`/`client_id`/`client_secret`/`scope` form body unconditionally). Neither built-in
declarative auth mode fits this shape — `"basic"` sends Basic credentials on every data request
rather than exchanging them once for a Bearer token, and `"oauth2_client_credentials"` sends the
wrong client-authentication shape — so `hooks/paypal-transaction/hooks.go`'s `AuthHook` (mode
`custom`, hook `paypal-transaction`) ports legacy's `basicTokenAuth` verbatim: POST
`grant_type=client_credentials` (form body, no client_id/client_secret in the body) to
`{{ base_url }}/v1/oauth2/token` with HTTP Basic `client_id:client_secret`, caching the returned
`access_token` (refreshing 60s before its declared `expires_in`) and sending it as
`Authorization: Bearer` on every subsequent data request. This is the identical
legitimate Tier-2 `AuthHook` trigger `hooks/jamf-pro/hooks.go` documents for its own
Basic-credential token exchange (conventions.md §1's hook table: "token-exchange auth").

Provide `client_id` and `client_secret` (both secrets, never logged) and a required `start_date`
(RFC3339 lower bound for the `transactions` reporting stream). `base_url` defaults to the
production host (`https://api-m.paypal.com`); see Known limits for why `is_sandbox` is not modeled
as a separate config toggle.

## Streams notes

- **`transactions`**: `GET /v1/reporting/transactions`, `page`/`page_size` page-number pagination
  (`pagination.type: page_number`, `start_page: 1`, `page_size: 100`, legacy's own default),
  records at `transaction_details`. `start_date`/`end_date` are sent via `stream.Query`:
  `start_date` templates `{{ incremental.lower_bound }}` (the engine's resolved incremental lower
  bound — the state cursor if a prior sync ran, else the required `start_date` config value,
  formatted `rfc3339` — see conventions.md §3), and `end_date` is an `omit_when_absent` optional
  param (present iff `config.end_date` is set). `fields=all` is a fixed static query param, matching
  legacy's `baseQuery`. Every field is hoisted out of the raw `{transaction_info:{...}}` envelope via
  `computed_fields` (a rename/nested-path extraction for every field, since the raw shape nests
  everything one level down and PayPal's monetary values are themselves nested `{currency_code,
  value}` objects) — byte-for-byte matching legacy's `transactionRecord` mapper.
- **`balances`**: `GET /v1/reporting/balances`, no pagination, records at `balances`.
  `computed_fields` flattens the three nested `{currency_code,value}` monetary objects
  (`total_balance`/`available_balance`/`withheld_balance`) into `total_value`/`available_value`/
  `withheld_value` plus `total_currency_code`, matching legacy's `balanceRecord` exactly.
- **`products`**: `GET /v1/catalogs/products`, `page`/`page_size` page-number pagination with
  `page_size: 20` (legacy: "products endpoint caps page_size at 20", a stream-level pagination
  override — conventions.md §3's "stream-level pagination replaces base-level wholesale"), records
  at `products`. No `computed_fields` needed: every raw field name already matches legacy's output
  verbatim (`id`/`name`/`description`/`type`/`category`/`create_time`).
- **`disputes`**: `GET /v1/customer/disputes`. PayPal's real wire pagination for this endpoint is a
  HATEOAS `links:[{rel,href}]` ARRAY (matched by `rel=="next"`), not a bare string path — the
  engine's only body-path next-page-URL pagination type (`pagination.type: next_url`) reads a single
  dotted string path (`connsdk.StringAt`/`selectPath`), which has no array-element-matching-by-
  sibling-field grammar; a fixed numeric array index would only ever read one array position, never
  "whichever element has `rel==next`". `hooks/paypal-transaction/hooks.go`'s `StreamHook` ports
  legacy's `harvestPageToken` loop verbatim: the first request sends `page_size=50`, each subsequent
  page follows the `rel=="next"` entry's absolute `href` until no such entry is found (an empty or
  absent `links` array). `streams.json` still declares `disputes`' `path`/`records`/`schema` for
  documentation, even though the hook fully overrides the read (`handled=true`); it carries a
  stream-level `conformance.skip_dynamic` marker naming this exact gap (the bundle-level marker
  below already supersedes it functionally, since the AuthHook forces every dynamic check to skip
  regardless).

Every `computed_fields` entry above is either a bare single `{{ record.<path> }}` reference (the
engine's typed-extraction rule preserves the raw JSON type — `primary` stays a real boolean, not a
stringified one) or a nested dotted-path extraction; none use a filter chain.

`check` issues a single bounded `GET /v1/reporting/balances` (the same probe legacy's own `Check`
uses), which also exercises the AuthHook's token exchange as a side effect of resolving auth for
that request — mirroring `jamf-pro`'s identical custom-auth `check` shape.

## Write actions & risks

None. PayPal Transaction is read-only (`capabilities.write: false`); legacy's own `Write` always
returns `connectors.ErrUnsupportedOperation` — there is no approved reverse-ETL write surface for
this data set.

## Known limits

- Only the 4 legacy-parity read streams are implemented; see `api_surface.json`. PayPal's much
  larger REST surface (payments/orders, invoicing, payouts, subscriptions/billing plans, identity,
  webhooks) is out of scope outside this connector's current declared surface.
- **`is_sandbox` is not modeled as a separate config toggle.** Legacy accepted an `is_sandbox`
  boolean config value as an alternative to an explicit `base_url` override, deriving
  `https://api-m.sandbox.paypal.com` from it (`isSandbox`/`paypalBaseURL`,
  `paypal_transaction.go:486-505`). The engine's `spec.json` `"default"` materialization only
  supports a fixed literal default, not one derived from another config value at read/check time —
  the same limitation `conventions.md` documents for sentry's `hostname`-derived base URL and
  dwolla's identical `environment`-derived base URL narrowing. This bundle narrows the config
  surface to `base_url` only (defaulting to production); a caller who needs the sandbox host sets
  `base_url` to `https://api-m.sandbox.paypal.com` directly. This narrows accepted CONFIGURATION
  surface only, never emitted record data.
- **`end_date`'s "now" default is not reproduced.** Legacy computes `time.Now().UTC()` client-side
  when `end_date` is unset (`baseQuery`, `paypal_transaction.go:174-179`) — the engine's
  `omit_when_absent` dialect can only OMIT an unresolved param, never compute a dynamic
  "current time" default (`spec.json`'s `"default"` mechanism is a fixed literal only). When
  `end_date` is left unset, this bundle omits the query param entirely rather than sending a
  client-computed timestamp; PayPal's own reporting endpoint applies its own server-side default in
  that case. Set `end_date` explicitly to reproduce a specific window.
- **`max_pages` is only wired for the `disputes` stream (StreamHook).** Legacy accepts a `max_pages`
  config override (0/all/unlimited = unbounded, default) uniformly across every stream
  (`paypalMaxPages`). The engine's `PaginationSpec.MaxPages` field (used by `transactions`/
  `products`' declarative `page_number` pagination) is a plain fixed JSON integer baked into
  `streams.json` — there is no templating/config-driven override mechanism for it, matching
  `linkedin-pages`/`jamf-pro`/`bamboo-hr`'s identical documented narrowing. `transactions`/
  `products` are therefore unbounded (matching legacy's own default of no cap) with no way to
  configure a smaller cap; only `disputes` (hook-driven, plain Go) reads `config.max_pages` directly,
  exactly like legacy.
- **`balances`' catalog-declared `as_of_time` cursor field is not modeled.** Legacy's catalog
  declares `CursorFields: []string{"as_of_time"}` for `balances`, but `balanceRecord` never actually
  emits a field of that name (the real `/v1/reporting/balances` response's `as_of_time` envelope key
  is never mapped into the record) — a pre-existing legacy catalog artifact that does not correspond
  to any real emitted field. This bundle's `balances` schema declares no `x-cursor-field` at all
  (conventions.md §8's incremental truth table: neither a real emitted cursor field nor a
  server-side filter exists, so no cursor marker is declared), rather than fabricate a phantom
  `as_of_time` property just to preserve a legacy catalog declaration that was never wired to real
  data.
- `transactions`/`products`' `page_number` paginator stops on a short/empty final page (fewer than
  `page_size` records), not on the response body's `total_pages` field the way legacy's own
  `harvestPageIncrement` additionally checks — the identical, ACCEPTABLE parity deviation
  `jamf-pro`/`bamboo-hr` document for their own `totalCount`-style early-stop signal: at most one
  harmless extra request on the rare page that exactly exhausts the total, never a
  record omission/duplication/reorder for any input legacy itself would accept.
- Bundle-level `conformance.skip_dynamic` (`metadata.json`): every fixture-replay dynamic check
  (`check_fixture`, `read_fixture_nonempty:*`, `pagination_terminates`, `records_match_schema`,
  `cursor_advances`) is skipped, since the AuthHook's real token exchange target
  (`config.base_url + "/v1/oauth2/token"`) can never resolve to the replay server's own address —
  conformance's synthetic non-secret config sets `base_url` to the literal string
  `"synthetic-conformance-value"`, and `withReplayURL` only overrides the bundle's declarative
  `HTTP.URL` field, not `RuntimeConfig.Config["base_url"]` the hook itself reads. Hook-covered,
  proven live by `internal/connectors/paritytest/paypal-transaction` (token exchange +
  `transactions`/`disputes` two-page pagination against a real `httptest.Server`, plus
  per-stream record-parity assertions for all 4 streams) and
  `internal/connectors/hooks/paypal-transaction/hooks_test.go`.
