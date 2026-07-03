# Overview

Google Ads reads accessible customers and a small, deliberately allow-listed set of GAQL search
resources (`campaigns`, `ad_groups`) through the Google Ads REST API. This is a Tier-2 migration of
`internal/connectors/google-ads` (the hand-written legacy connector, which despite its
`internal/connectors/<name>/` location and `runtime_kind: native_go` inventory label is a plain
`connsdk`-HTTP connector with no SQL/queue/protocol-native behavior — the label predates this
convention's tier ladder). The legacy package stays registered and unchanged until wave6's registry
flip. Read-only; arbitrary GAQL is not accepted, matching legacy's conservative design intent
(package doc comment: "Google Ads reads are deliberately allow-listed ... arbitrary GAQL is not
accepted here").

Auth (Bearer + `developer-token` + optional `login-customer-id`) is fully declarative
(`streams.json`'s `base.auth`/`base.headers`); every stream read is dispatched through
`internal/connectors/hooks/google-ads/hooks.go`'s `StreamHook`, for two independent
engine-dialect reasons documented per stream below (never an auth limitation).

## Auth setup

Provide a Google OAuth 2.0 access token (Google Ads read scope) via the `access_token` secret —
used only for Bearer auth (`Authorization: Bearer <access_token>`), applied by the fully
declarative `bearer` auth mode. Provide the mandatory Google Ads `developer_token` secret — sent as
the `developer-token` header on every request via `streams.json`'s `base.headers` (a `secrets.*`
header reference is always a hard error when unset, matching legacy's own
`googleAdsValidateSecrets` requirement that both secrets be present). An optional
`login_customer_id` config value (a manager/MCC account id) is sent as the `login-customer-id`
header; when unset it is omitted entirely, not sent empty (`login_customer_id` is declared but not
in `spec.json`'s `required[]`, so the engine's declared-optional-config header-omission rule
applies — matches legacy's own `strings.TrimSpace(a.loginCustomerID) != ""` gate exactly). No
secret is ever logged.

## Streams notes

`accessible_customers` (`GET customers:listAccessibleCustomers`, no body, no pagination) lists
every customer resource name accessible to the OAuth token. Its real response shape,
`{"resourceNames": ["customers/123", ...]}`, is a bare JSON array of **scalar strings**, not
objects — `connsdk.RecordsAt` (the engine's only record-extraction primitive) silently drops any
array element that does not decode as a JSON object, yielding **zero** records for this shape via
the purely declarative path (the identical gap documented for `ip2whois`'s `nameservers` field,
conventions.md's parity-deviation ledger entry 12). The `StreamHook` ports legacy's
`readAccessibleCustomers` verbatim: splits each resource name on `/` to derive `customer_id`
(the trailing segment), matching legacy's own inline split exactly.

`campaigns`/`ad_groups` (`POST customers/{customer_id}/googleAds:search`) run a fixed
allow-listed GAQL query (`SELECT campaign.id, campaign.name, campaign.status,
campaign.resource_name FROM campaign` / the `ad_group` equivalent) and page via a JSON request
body carrying `{"query", "pageSize", "pageToken"}`, reading the next `pageToken` from the response
body's `nextPageToken` field. The engine's declarative read path (`engine/read.go`'s
`readOneSequence`) always issues its request with a **nil** body — `StreamSpec.Body`
(`engine/bundle.go`) is declared in the dialect but never read anywhere in the read path — and
every one of the engine's 6 pagination types advances by adding a query parameter, never a body
field, so an in-body `pageToken` cannot be expressed in `streams.json` alone regardless of body
support. This is the identical shape already solved for `stigg` (GraphQL-over-HTTP) and
`google-search-console`'s `search_analytics_by_*` streams (POST body carrying `startRow`); the
`StreamHook` ports legacy's `search` function verbatim (POST body construction, `results` record
extraction, `nextPageToken` continuation, `pageSize`/`max_pages` config resolution). `customer_id`
is required at read time for these two streams only (legacy: `search` returns an error when
`customer_id` is empty), never for `accessible_customers` — the hook enforces this exactly like
legacy's own `Read` dispatch, since the engine dialect has no per-stream-required-config mechanism.

Every stream's record shape (`id`/`name`/`status`/`resource_name` for the GAQL streams;
`customer_id`/`resource_name` for `accessible_customers`) is ported field-for-field from legacy's
own `mapRecord`/`fixtureRecord` functions — schema-mode projection (default) matches legacy's
actual field-by-field emission (§8 rule 1); legacy never emits verbatim raw records.

## Write actions & risks

None. `capabilities.write` is `false`; legacy's `Write` always returns
`connectors.ErrUnsupportedOperation`.

## Known limits

- Only 3 streams are implemented (`accessible_customers`, `campaigns`, `ad_groups`), matching
  legacy's own conservative allow-list — arbitrary GAQL, and the full Google Ads resource/reporting
  surface, are deliberately out of scope (both here and in legacy). See `api_surface.json`'s
  `excluded: {category: out_of_scope, ...}` entries.
- All three streams are `StreamHook`-dispatched (`streams.json`'s per-stream
  `conformance.skip_dynamic` markers); `conformance`'s declarative fixture-replay dynamic checks
  cannot exercise a POST-body or scalar-array-response read at all — the hook's own unit tests
  (`internal/connectors/hooks/google-ads/hooks_test.go`) are the authoritative parity proof for
  every read path, porting legacy's `google_ads_test.go` coverage (pagination, auth header
  application, fixture-mode catalog/write-capability checks) verbatim.
- Legacy's `page_size` validation bounds the value to `[1, 10000]` (`googleAdsMaxPageSize`) and
  hard-errors outside that range; the `StreamHook` preserves this exact bound (ported verbatim from
  `googleAdsPageSize`/`intConfig`), unlike a purely declarative bundle which would have no
  range-validation mechanism at all for a config value.
