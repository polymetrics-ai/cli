# Overview

QuickBooks is a Tier-2 (AuthHook + StreamHook) migration of `internal/connectors/quickbooks`
(read-only reference, quarantined `ENGINE_GAP` in `docs/migration/quarantine.json`). It reads
QuickBooks Online customers, invoices, payments, accounts, and vendors via the v3 Query API's
single shared `query` endpoint, authenticating with the real QuickBooks OAuth 2.0
**refresh-token grant** (`client_id`/`client_secret`/`refresh_token` -> short-lived access token) —
the catalog's documented auth shape (`website/.enrich/enr/source-quickbooks.json`), which legacy's
simplified reference implementation approximates with a pre-issued static `access_token` secret
only. This bundle is engine-vs-legacy parity-tested against `internal/connectors/quickbooks`; the
legacy package stays registered and unchanged until wave6's registry flip. Read-only: legacy
`quickbooks.go:155-157` always returns `ErrUnsupportedOperation` from `Write`, and this bundle
declares `capabilities.write: false` with no `writes.json` to match.

## Auth setup

Provide four secrets: `client_id`, `client_secret`, `refresh_token` (long-lived, up to ~100 days;
never logged), and `realm_id` (the QuickBooks company/Realm ID, treated as secret-adjacent since it
identifies a specific company and is validated as a path-safe segment). `hooks/quickbooks/hooks.go`
implements `AuthHook`, mirroring `hooks/gmail/hooks.go`'s `oauthRefreshAuth` almost verbatim: it
POSTs `grant_type=refresh_token` + `refresh_token` + `client_id` + `client_secret` to `token_url`
(default Intuit's OAuth 2.0 bearer-token endpoint, config-overridable), caches the resulting access
token until 60 seconds before its declared expiry (QuickBooks access tokens are valid for one hour),
and sets `Authorization: Bearer <access_token>` on every request.

`token_url` MUST resolve to an `https://` URL (THREAT-MODEL.md Delta 2, same guard as gmail's hook):
the hook fails closed on a non-https or unparseable override rather than sending the refresh
token/client secret to an attacker-chosen endpoint.

The bundle's `base.auth` declares exactly one candidate: `{"mode": "custom", "hook": "quickbooks",
...}` — there is no static-token fallback to declare (unlike legacy's simplified `access_token`
secret, which this migration does not carry forward since it is not the real, catalog-documented
auth shape).

## Streams notes

Five streams, all primary-keyed on `id`: `customers`, `invoices`, `payments`, `accounts`, `vendors`
— every one reads the identical `v3/company/{realmId}/query` endpoint with a different `SELECT *
FROM <Entity>` query, exactly like legacy's shared `harvest` helper. **This is entirely
`StreamHook`-handled, not declarative**: QuickBooks' pagination state (`STARTPOSITION <n>
MAXRESULTS <size>`) is embedded inside the single `query` query-string VALUE itself, not sent as
independent query parameters — the engine's 6 declarative pagination types (`page_number`,
`offset_limit`, `cursor` x2, `next_url`, `link_header`) all assume independently-named page/offset
query parameters and cannot express "increment a number baked into the middle of another param's
string value" (`docs/migration/quarantine.json`'s original `quickbooks` blocker; still unresolved by
the S3/S4 engine mini-waves, which added `incremental.lower_bound` query-vars, `fan_out`,
`keyed_object`, 0-indexed `start_page`, and `param_format` date-only parsing — none of which
introduce an interpolatable "current pagination offset" reference usable inside `stream.Query`).
`hooks/quickbooks/hooks.go`'s `ReadStream` ports legacy's `harvest` loop verbatim: build the
`SELECT * FROM <Entity> STARTPOSITION <start> MAXRESULTS <pageSize>` query string, issue the GET via
`rt.Requester` (already carrying the AuthHook-resolved bearer token), decode
`QueryResponse.<Entity>`, map each record, and stop when the page returns fewer than `page_size`
records (a short/empty final page) or `max_pages` is reached — identical stop conditions to
legacy's `harvest`.

Record mapping is also hook-side Go (matching legacy's `qbCustomer`/`qbInvoice`/`qbPayment`/
`qbAccount`/`qbVendor` field-for-field, including `CustomerRef.value` unwrapping via `refValue`), not
declarative `computed_fields` — QuickBooks' PascalCase wire fields and the read path both live in the
same hook, so there is no separate declarative projection step to keep in sync. `streams.json`'s
`records.path` (`QueryResponse.<Entity>`) is documented for schema/tooling consistency but is never
actually read by the declarative fallback path in this bundle (StreamHook always returns
`handled=true`).

## Write actions & risks

None — QuickBooks is read-only here. `capabilities.write: false`, no `writes.json` file, matching
legacy's `ErrUnsupportedOperation` (`quickbooks.go:155-157`).

## Known limits

- **No incremental sync mode**: legacy's QuickBooks connector has no incremental cursor/filter
  logic at all (`start_date` is declared in both legacy's catalog config and this bundle's
  `spec.json` but is never wired into a query filter on either side) — full_refresh only. `start_date`
  is retained in `spec.json` purely as forward-compatible config surface (matches legacy's own
  unused field; see gmail's identical precedent).
- **`sandbox` is informational only**: `base_url` is the operative environment switch (default
  production; override to `https://sandbox-quickbooks.api.intuit.com` for a sandbox company) —
  legacy has no separate `sandbox` boolean branch either; the config field exists only to mirror the
  real catalog's documented shape (`website/.enrich/enr/source-quickbooks.json`).
- **Auth deviates from the simplified legacy reference implementation, not from the real
  QuickBooks product**: legacy's `internal/connectors/quickbooks/quickbooks.go` uses a pre-issued
  static `access_token` secret with no refresh logic at all — a deliberate simplification for the
  read-only reference connector. This bundle instead implements QuickBooks' actual, catalog-documented
  OAuth 2.0 refresh-token grant (`client_id`/`client_secret`/`refresh_token`), since that is the auth
  scheme the real API requires and the one a production credentials layer would present. This is a
  documented, deliberate widening of the auth surface versus legacy, not a parity deviation in the
  §5 sense (no accepted legacy INPUT behavior changes — legacy accepted no refresh-grant input at
  all to diverge from).
- **`TestConformance/quickbooks`'s dynamic (fixture-replay) checks are `skip_dynamic`'d for two
  independent, compounding reasons**: (1) its sole auth candidate is `mode: custom` with no
  `when`-gated fallback, and conformance's synthetic config can never carry a real `https` `token_url`
  the AuthHook's fail-closed guard would accept; (2) every stream is `StreamHook`-handled with an
  in-query-string pagination shape no declarative fixture-replay check can drive independently of
  the hook. `internal/connectors/paritytest/quickbooks` (a live `httptest.Server` proving the real
  `STARTPOSITION`/`MAXRESULTS` request shape and stop conditions) and
  `internal/connectors/hooks/quickbooks/hooks_test.go` (the `AuthHook`'s refresh-grant behavior) are
  the authoritative correctness bar this marker names.
