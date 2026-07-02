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
table), primary key `["id"]`, incremental cursor field `updated_at`.

Pagination follows Zendesk's cursor convention (`pagination.type: cursor`
with `token_path: meta.after_cursor`): the next page's `page[after]` value is
read from the response body's `meta.after_cursor`, and pagination stops when
that value is absent/null — which is Zendesk's real termination signal
whenever `meta.has_more` is false (legacy's own fixture and live-API
behavior always emit `after_cursor: null` on the final page; see
`zendesk_support_test.go`'s recorded fixture). Every request sends
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
- **Documented parity deviation**: this bundle declares
  `incremental.request_param: "updated_at[gte]"` (`cursor_field: updated_at`,
  `start_config_key: start_date`) so the engine's InitialState/state-cursor/
  start_date plumbing is exercised, per the pilot spec's "start_date-raised"
  incremental-parity requirement. Legacy itself sends NO server-side
  incremental filter query parameter at all — `harvest()` always requests
  the full, unfiltered collection every sync; `start_date` is mentioned only
  in a doc comment on `InitialState`, never wired to any query key. This is
  strictly MORE filtering than legacy ever performed for the identical
  config (a fresh sync with no cursor/start_date still sends no lower-bound
  param on either side, so first-sync behavior is unchanged), and it never
  changes the emitted DATA for any record legacy itself would return in that
  scenario — it only narrows sync SCOPE when a start_date or persisted
  cursor is supplied, which is the intended incremental-sync behavior this
  bundle adds on top of legacy's always-full-scan baseline. See
  `docs/migration/conventions.md`'s parity-deviation ledger for the
  acceptance rule this satisfies (never changes accepted-input behavior).
- All fixtures (`fixtures/streams/**`, `fixtures/check.json`) represent
  Zendesk's real wire shape, including `meta.has_more`/`meta.after_cursor`
  exactly as the API returns them (a `null` `after_cursor` on the final
  page, not an omitted key).
