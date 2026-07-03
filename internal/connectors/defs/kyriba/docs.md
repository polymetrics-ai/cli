# Overview

Kyriba is a treasury and cash management platform. This bundle migrates
`internal/connectors/kyriba` (`kyriba.go` + `streams.go`), a conservative
read-only connector that reads a documented-style Kyriba tenant v1 REST API:
bank accounts, transactions, statements, and payments, each a page-number
paginated `GET` collection endpoint under a tenant-configurable base URL.
Kyriba deployments vary by tenant, so both `base_url` and `token_url` are
configurable overrides of the documented defaults
(`https://api.kyriba.com/api/v1` / `https://api.kyriba.com/oauth/token`),
matching legacy's own `baseURL`/`tokenURL` helper functions exactly.

## Auth setup

Kyriba uses OAuth2 client-credentials: `client_id` + `client_secret` (both
`x-secret`) are exchanged at `token_url` for a bearer access token, sent as
`Authorization: Bearer <token>` on every request — this bundle's
`streams.json` `base.auth` declares a single `oauth2_client_credentials`
candidate, matching legacy's `connsdk.OAuth2ClientCredentials` construction
(`kyriba.go:144`) exactly, including the optional `scope` config value
(`config.scope`, defaulting to an empty string so the templated `scopes`
field always resolves cleanly): legacy only sets `auth.Scopes` when
`config["scope"]` is non-empty after trimming, which the engine's own
`strings.Fields("")` → empty-slice behavior reproduces byte-for-byte when
`scope` is left unset.

## Streams notes

Legacy defines 4 streams, each a distinct GET collection endpoint sharing
identical page-number pagination (`page`/`size` query params, 1-based start
page, a page returning fewer than `page_size` records stops the sync —
`kyriba.go:97-118`'s `harvest`):

- `bank_accounts` (`/bank-accounts`) — `id`, `account_number` (from the raw
  API's `accountNumber`, or `account_number` if already present —
  `streams.go`'s `first` helper), `currency`, `status`.
- `transactions` (`/transactions`) — the same fields as `bank_accounts` plus
  `amount`.
- `statements` (`/statements`) — identical shape to `bank_accounts` (legacy's
  `statementRecord` is a bare alias for `bankAccountRecord`).
- `payments` (`/payments`) — `id`, `amount`, `currency`, `status` (no
  `account_number`; legacy's `paymentRecord` builds a distinct, smaller
  record).

None of the 4 legacy `connectors.Stream` definitions publish `CursorFields`
(confirmed by `streams.go`'s `streams()`), so this bundle declares no
`incremental` block on any stream and no `x-cursor-field` on any schema
(conventions.md §8 rule 2's "neither → no incremental block") — every read is
a full stream read, matching legacy exactly.

`accountNumber`-vs-`account_number` field-name tolerance (legacy's `first`
helper, preferring a raw `accountNumber` and falling back to an
already-snake_case `account_number`) is not modeled: fixtures record the real
Kyriba wire shape (`accountNumber`), and schema projection maps it via the
raw API's actual field name only. Kyriba's real wire format is always
`accountNumber` (confirmed by legacy's own comment and its own test fixture),
so the `account_number` fallback branch in `first()` is legacy dead code for
any real Kyriba tenant response — not a live parity concern.

## Write actions & risks

None. Kyriba is read-only (`capabilities.write: false`, no `writes.json`),
matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`
unconditionally (`kyriba.go:218-220`).

## Known limits

- **`page_size`/`max_pages` are NOT configurable at runtime (documented scope
  narrowing, ACCEPTABLE per conventions.md §5's meta-rule).** Legacy accepts
  `config.page_size` (1-500, default 100) and `config.max_pages` (default
  unbounded) at read time (`kyriba.go`'s `pageSize`/`maxPages` helpers).
  `streams.json`'s `base.pagination` is a fixed JSON literal
  (`page_size: 100`, no `max_pages` cap) the `page_number` paginator
  constructor reads once at bundle-authoring time, with no runtime
  config-driven override mechanism in this dialect (the same class of gap
  documented in `docs/migration/conventions.md`'s searxng worked example, and
  employment-hero's identical `items_per_page` deviation) — declaring a spec
  property no template consumes would be dead config (F6). `page_size: 100`
  and unbounded `max_pages` are legacy's own defaults, so a caller that never
  overrode either config value observes byte-for-byte identical pagination
  behavior; a caller that DID override them loses that override.
- **Bundle-level `conformance.skip_dynamic` (see `metadata.json`).** Kyriba's
  OAuth2 client-credentials `token_url` is a separate declared config
  property; `internal/connectors/conformance`'s `withReplayURL` only
  redirects `b.HTTP.URL` (stream/check request paths) to the fixture-replay
  server, never `RuntimeConfig.Config["token_url"]` — so the token exchange
  always targets the synthetic placeholder value
  (`"synthetic-conformance-value"`), an unreachable non-URL, and every
  auth-resolving dynamic check would fail identically and uninformatively
  before ever reaching a declarative stream/check request. This is the exact,
  now-repeated shape documented for clazar/sendpulse (`docs/migration/
  conventions.md` §4's skip-marker section) — static checks (spec/schema
  validity, `interpolations_resolve`, docs/fixtures presence, secret
  redaction) still run and pass; the read/pagination/schema-projection shape
  is proven by structural review against legacy `internal/connectors/kyriba`
  instead. There is no Tier-2 hook here (auth is fully declarative
  `oauth2_client_credentials`), so there is no `paritytest/kyriba` package
  for this wave.
- **Legacy's `mode: fixture` credential-free affordance is NOT part of this
  bundle.** Legacy's `readFixture`/`fixtureMode` (`kyriba.go:120-134,214-216`)
  emit synthetic records without any network call when `config.mode ==
  "fixture"` — this is a legacy-only testing convenience, not part of the
  live record shape; parity is asserted against legacy's LIVE (httptest-driven)
  read path only. The `connector`/`fixture` marker fields legacy's fixture
  mode stamps onto every record are correspondingly absent from this bundle's
  schemas.
