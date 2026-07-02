# Overview

Zendesk Support is a wave1-pilot migration of `internal/connectors/zendesk-support`
(the hand-written legacy connector this bundle migrates; the legacy package
stays registered and unchanged until wave6's registry flip). It reads Zendesk
Support tickets, users, organizations, groups, and satisfaction ratings
through the Zendesk Support REST API v2. This bundle is engine-vs-legacy
parity-tested in `internal/connectors/paritytest/zendesk-support`.

## Auth setup

Zendesk supports two credential shapes, both declared as `when`-gated
candidates on `streams.json`'s `base.auth` (first match wins):

1. **OAuth2 access token** (checked first, matching legacy's own precedence —
   `zendesk_support.go:272` checks `access_token` before `api_token`):
   provide `access_token` as a secret; sent as `Authorization: Bearer
   <access_token>`.
2. **API token** (Zendesk Admin Center > Apps and integrations > APIs >
   Zendesk API): provide `api_token` (secret) and `email` (secret) — the
   agent email address that owns the token. Sent via HTTP Basic as
   `Authorization: Basic base64("<email>/token:<api_token>")`, byte-for-byte
   identical to legacy's `connsdk.Basic(email+"/token", apiToken)`
   construction.

**Documented config-surface deviation (email)**: this bundle's `email`
property is secrets-only. Legacy additionally accepts the dotted secret keys
`credentials.api_token`/`credentials.email` (its multi-key `secret()` lookup,
`zendesk_support.go:271-287,366-378`) AND a plain, non-secret `config.email`
fallback when no `credentials.email`/`email` secret is present. The bare
`email`/`api_token` secret names are a reasonable canonicalization of the
dotted-key surface (every dotted-key value legacy would resolve is equally
reachable as the bare secret), but the `config.email` (non-secret) fallback
specifically is narrower here: this bundle requires `email` to be supplied as
a secret, not a plain config value. No request shape or emitted record
differs for any input using the canonical secret keys.

Provide `base_url` as your Zendesk account root, e.g. `https://acme.zendesk.com`
for subdomain `acme` — every stream `path` (and `check.path`) is declared as
`/api/v2/<resource>`, so `base_url` must NOT include `/api/v2` itself (the
`/api/v2` prefix is baked into each request path rather than `base.url`,
since a test/conformance harness that overrides `HTTP.URL` directly — e.g.
this bundle's own parity suite and `conformance`'s replay harness — replaces
the resolved base URL string wholesale, bypassing any suffix template that
would otherwise live on `base.url` itself). This is also the base URL
override mechanism for tests/proxies.

**Documented config-surface deviation**: legacy additionally accepts a bare
`subdomain` config value and derives `https://<subdomain>.zendesk.com/api/v2`
itself (`zendesk_support.go`'s `baseURL()`), falling back to a `base_url`
override only when set. The engine's declarative `url` template is a single
static string with no conditional branching between two config keys, so this
bundle requires `base_url` directly (matching every other wave1 pilot's
`base_url`-only config surface — stripe/bitly/calendly/etc.) rather than
declaring a `subdomain` property the dialect could never wire (conventions.md
§3: a declared-but-unwireable key is worse than an absent one — see
searxng's `subreddit` precedent). Every legacy-accepted `subdomain`-only
configuration is still reachable by an operator supplying the equivalent
`https://<subdomain>.zendesk.com` as `base_url` instead — no request shape or
emitted record ever differs between the two config styles for the same
effective account.

## Streams notes

All 5 streams (`tickets`, `users`, `organizations`, `groups`,
`satisfaction_ratings`) share the same shape: `GET` against the Zendesk
collection endpoint, records extracted from the response's own top-level key
(e.g. `tickets` for the tickets stream — legacy's `streamEndpoints` routing
table), primary key `["id"]`. Each stream's schema declares
`x-cursor-field: updated_at` for manifest-surface parity (documenting which
field callers could use to order/dedupe locally), but — see "Known limits"
below — no stream declares an `incremental` block, so no request-side
lower-bound filter is ever sent; this mirrors legacy exactly.

Pagination follows Zendesk's cursor convention (`pagination.type: cursor`
with `token_path: meta.after_cursor`, `stop_path: meta.has_more`): the next
page's `page[after]` value is read from the response body's
`meta.after_cursor`, but pagination stops as soon as `meta.has_more` is
falsy (any value other than the literal `true`), REGARDLESS of whether
`meta.after_cursor` is still populated. This exactly reproduces legacy's own
stop rule (`hasMore != "true" || nextCursor == ""` —
`zendesk_support.go:189` — has_more alone is sufficient to stop). Zendesk's
own cursor-pagination documentation warns that the cursor properties "may be
populated even when has_more is false", so relying on an absent/null cursor
alone (this bundle's pre-fix behavior) could issue one extra trailing
request, or loop, against a live account that populates a stale cursor on
its final page; `stop_path` closes that gap. Every request sends
`page[size]=100` via each stream's static `query` (matches legacy's default
`page_size`), not via `pagination.size_param`/`page_size`, which the
`cursor`+`token_path` paginator constructor never reads (only
`page_number`/`offset_limit` do) — mirrors stripe's `limit=100`-via-static-query
pattern (see conventions.md's F6 lesson).

## Write actions & risks

None. Legacy `zendesk-support` is read-only (`Write` returns
`connectors.ErrUnsupportedOperation`); `metadata.json` declares
`capabilities.write: false` and this bundle ships no `writes.json`.

## Known limits

- Full Zendesk Support API surface (macros, triggers, automations, views,
  help center, talk, chat, attachments) is out of scope for wave1; see
  `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B
  capability expansion"}` entries. Only the 5 legacy-parity streams are
  implemented.
- **No server-side incremental filter (matches legacy exactly)**: earlier
  drafts of this bundle declared `incremental.request_param:
  "updated_at[gte]"` on all 5 streams, speculatively mirroring the
  stripe/chargebee shape. That was wrong and has been removed: Zendesk's
  Support collection endpoints (`/api/v2/tickets`, `/users`, `/organizations`,
  `/groups`, `/satisfaction_ratings`) document no `updated_at[gte]`-style
  server-side filter, and legacy's own `harvest()` (`zendesk_support.go:152-
  195`) never sends one — `InitialState` always starts with an empty cursor,
  and `start_date` is mentioned only in a doc comment, never wired to any
  query key. Declaring the param would have been a real behavior CHANGE (new
  server-side filtering legacy never had), not parity — worse, since
  `updated_at[gte]` is not real Zendesk API surface, a live account would
  silently ignore it while the parity suite's own fixture servers "proved"
  it worked only because they honored the invented parameter. No stream
  declares an `incremental` block; every sync (fresh or resumed) reads the
  full, unfiltered collection, exactly like legacy. Each schema still
  declares `x-cursor-field: updated_at` for manifest-surface documentation
  only (the calendly-3-streams precedent: `organization_memberships`,
  `groups`, `users` there ship the same schema-only cursor-field annotation
  with no wired `incremental` block). A persisted state cursor is still
  accepted without error (never forwarded on the wire) so a caller that
  supplies one from a prior sync does not fail — see
  `paritytest/zendesk-support`'s `IncrementalConfigAcceptedWithoutServerFilter`
  and `StartDateConfigNeverSendsServerFilter` tests. True Zendesk incremental
  sync belongs to Pass B via the documented `/api/v2/incremental/*` export
  endpoints (a different endpoint shape entirely, with a `start_time` unix
  param), not this connector's collection endpoints.
- **Pagination stop-signal fix**: earlier drafts stopped pagination only on
  an absent/null `meta.after_cursor`, which happened to match every fixture
  and parity server (all of which null the cursor exactly when `has_more` is
  false) but was unverified against Zendesk's documented behavior that the
  cursor may still be populated on the final page. `stop_path:
  meta.has_more` (see "Streams notes" above) now reproduces legacy's real
  stop rule directly; regression-tested by
  `TestParityZendesk_HasMoreFalseWithNonNullCursorStopsPagination`.
- All fixtures (`fixtures/streams/**`, `fixtures/check.json`) represent
  Zendesk's real wire shape, including `meta.has_more`/`meta.after_cursor`
  exactly as the API returns them (a `null` `after_cursor` on the final
  page, not an omitted key).
