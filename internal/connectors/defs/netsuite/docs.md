# Overview

NetSuite is a Tier-2 (AuthHook) migration of `internal/connectors/netsuite`. It reads a
conservative, allow-listed set of NetSuite REST Record API resources — customers, vendors,
inventory items, and sales orders — authenticating with OAuth 1.0a Token-Based Authentication
(TBA), an HMAC-SHA256 request-signing scheme the declarative dialect cannot express (no
`auth.mode` computes a signature over the outgoing request's own method/URL/query/timestamp/nonce
— this is the `AUTH_COMPLEX` signature-auth Tier-2 trigger named in `conventions.md` §1's Tier-2
table, "signature auth (SigV4, HMAC)"). This bundle targets capability parity with
`internal/connectors/netsuite` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip.

**Tier justification**: legacy is a `connsdk`-based HTTP connector (`connsdk.Requester` +
`connsdk.RecordsAt`), not a non-REST protocol — it is not SQL/queue/filesystem-native, so Tier 3
does not apply despite the catalog's `runtime_kind: native_go` label (that label reflects legacy's
hand-written-Go implementation status, not a genuine non-HTTP protocol; NetSuite's REST Record API
is ordinary JSON-over-HTTPS). The only reason this bundle is not pure Tier 1 is the OAuth 1.0a
HMAC-SHA256 signing scheme, which needs to compute a signature over the request's own method, URL,
query parameters, a per-request timestamp, and a per-request nonce — exactly the "signature auth"
Tier-2 escape hatch, not a Tier-3 native-protocol concern. Everything else (streams, pagination,
records extraction, schema) is fully declarative.

## Auth setup

Provide five values: `realm` (NetSuite account id, e.g. `123456` or `123456_SB1`), and four
secrets — `consumer_key`, `consumer_secret`, `token_key`, `token_secret` (all `x-secret` in
`spec.json`, never logged). `hooks/netsuite/hooks.go` implements `AuthHook`, porting legacy
`netsuite.go`'s `oauth1`/`oauthSignature` field-for-field: on every outgoing request it builds the
OAuth 1.0a parameter set (`oauth_consumer_key`, `oauth_token`, `oauth_signature_method:
HMAC-SHA256`, `oauth_timestamp`, `oauth_nonce`, `oauth_version: 1.0`), merges in the request's own
query parameters, computes the HMAC-SHA256 signature over the canonical base string
(`METHOD&percent-encoded-base-url&percent-encoded-sorted-param-string`) keyed by
`percent-encode(consumer_secret)&percent-encode(token_secret)`, and sets a single `Authorization:
OAuth realm="...", oauth_consumer_key="...", ...` header carrying every OAuth parameter plus the
computed signature. `consumer_secret`/`token_secret` never appear on the wire themselves (only the
HMAC digest does) and are never logged; no error path in the hook interpolates a secret value into
an error string.

`base_url` is **required** (unlike legacy, which derives it from `realm` when unset) — see Known
limits for why the engine cannot express that derivation, and set it to
`https://<realm-with-underscores-as-hyphens-lowercased>.suitetalk.api.netsuite.com/services/rest/
record/v1` (the exact value legacy's own `baseURL` function computes; `hooks/netsuite/hooks.go`'s
`BaseURLFromRealm` helper reproduces this derivation for any caller-side tooling that wants to
compute it, though the engine's own read/check path does not call it, since `streams.json`'s
`base.url` is a plain `{{ config.base_url }}` template).

The bundle's `base.auth` declares exactly one candidate: `{"mode": "custom", "hook": "netsuite"}` —
legacy has no alternate auth path, so there is no `when`-gated fallback to declare (same shape as
gmail's and nexus-datasets' sole custom-mode candidate).

## Streams notes

Four streams — `customers` (`/customer`), `vendors` (`/vendor`), `items` (`/inventoryItem`),
`sales_orders` (`/salesOrder`) — all sharing the IDENTICAL record shape and mapper legacy uses (a
single `record()` function serves all four `streamEndpoints` entries): primary key `id`, cursor
field `last_modified_date`, and fields `entity_id`/`name`/`email`/`status`. All four streams
reference the same `schemas/customer_record.json` schema for this reason (mirrors legacy's single
shared `fields` slice reused across all four `connectors.Stream` catalog entries). Pagination is
`offset_limit` (`limit`/`offset` query params, `page_size` from `config.page_size`, default 100)
with the engine's standard short-page stop — matches legacy's `harvest` loop exactly (offset
advances by the number of records returned each page until a page returns zero records or the
response's `hasMore` field is false; see Known limits for the `hasMore`-vs-short-page nuance).
Records are extracted from the page body's `items` array (`records.path: "items"`), NetSuite's
documented REST Record API list envelope shape. The shared legacy fallback chains for
`entity_id` (`entityId` then `tranId`), `name` (`companyName` then `name` then `title`), and
`status` (`status` then `entityStatus`) are modeled with `coalesce` computed fields.

No `incremental` block is declared for any stream, matching legacy exactly: legacy's `Read` never
applies a server-side `lastModifiedDate` filter (there is no `request_param`/`param_format` wiring
in `netsuite.go`'s `harvest`), so adding one here would be new, behavior-changing filtering legacy
never performed. `last_modified_date` is still declared as `x-cursor-field` in the schema (matching
legacy's `CursorFields` catalog declaration) for catalog-shape parity; every sync is a full read of
every page, exactly as legacy behaves today.

## Write actions & risks

None. NetSuite is a read-only source connector (`capabilities.write: false`); this bundle ships no
`writes.json`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`base_url` is required; legacy's realm-derived default is not automatically applied.** Legacy's
  `baseURL` function derives the host from `realm` (lowercased, underscores to hyphens) when
  `base_url` is unset. The engine's `streams.json` `base.url` is a plain `{{ config.base_url }}`
  template resolved via `Interpolate`, which hard-errors on an absent key (no absent-key-falsy
  tolerance outside `when` conditions, `conventions.md` §3) — and spec-default materialization only
  fills in a literal per-key default, not a value derived from another config key
  (`conventions.md`'s "derived default" guidance, the same gap nylas's `api_server`-to-host
  derivation and sentry/chargebee's `hostname`/`site`-to-URL derivations hit). `base_url` is
  therefore declared `required` here with no default; an operator must compute and set it
  themselves (`hooks/netsuite/hooks.go`'s exported `BaseURLFromRealm` helper reproduces legacy's
  exact derivation for any caller-side tooling). This is a documented config-surface narrowing
  (every legacy-accepted `realm`-implies-`base_url` shape has an operator-reachable equivalent), not
  a data-shape regression once configured.
- **`hasMore`-vs-short-page double stop signal.** Legacy's `harvest` stops when `!hasMore(resp.Body)
  || len(records) == 0` — i.e. it reads the body's `hasMore` boolean field in addition to the
  short-page stop. The engine's `offset_limit` paginator implements only the short-page stop
  (a page returning fewer than `page_size` records stops pagination) with no `stop_path`-equivalent
  for `offset_limit` (`stop_path` is only wired for the two `cursor` pagination variants, per
  `conventions.md` §3's pagination table). This can cause, at most, one harmless extra request on
  the rare page where a full-size page happens to exactly exhaust the underlying NetSuite result
  set while `hasMore` is already false — the following request then returns an empty page and stops
  normally. It never omits, duplicates, or reorders any record for any input legacy itself would
  accept (same acceptable-deviation shape as jamf-pro's `totalCount` early-stop, `conventions.md`
  §5 item 13).
- **`TestConformance/netsuite`'s dynamic (fixture-replay) checks are `skip_dynamic`'d** for the
  same reason as gmail/nexus-datasets: the sole auth candidate is `mode: custom`, and
  conformance's synthetic non-secret config can never carry a real `consumer_secret`/`token_secret`
  that would make a signed request meaningful against the replay server. See `metadata.json`'s
  `conformance.reason`. The hook's own unit tests (`internal/connectors/hooks/netsuite/
  hooks_test.go`) and the pre-existing legacy test suite (`internal/connectors/netsuite/
  netsuite_test.go`, unchanged, still passing against the read-only legacy package) remain the
  authoritative correctness bar for the OAuth 1.0a signing path.
