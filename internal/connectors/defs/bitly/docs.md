# Overview

Bitly is a wave1-pilot declarative-HTTP migration (PLAN.md P-3). It reads Bitly organizations,
groups, campaigns, and bitlinks through the Bitly v4 REST API
(`GET https://api-ssl.bitly.com/v4/...`). This bundle is engine-vs-legacy parity-tested against
`internal/connectors/bitly` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Bitly OAuth access token via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`) and is never logged, matching legacy's
`connsdk.Bearer(secret)` (`bitly.go:239`). `base_url` defaults to `https://api-ssl.bitly.com/v4`
and may be overridden for tests/proxies (legacy's own `bitlyBaseURL` validates scheme+host the
same way; the engine's base-URL resolution has no equivalent runtime validation, but every
parity/conformance fixture only ever points at an httptest server, so this is not exercised
differently on either side).

## Streams notes

`organizations`, `groups`, and `campaigns` are simple, non-paginated list endpoints (`GET
/organizations`, `/groups`, `/campaigns`); records live at the top-level key matching the stream
name (`organizations`/`groups`/`campaigns`). None of the three core list endpoints expose an
incremental cursor field (legacy `bitly.go:36`'s own comment: "none expose an incremental cursor
field in these core list endpoints, so reads are full refresh") — this bundle declares no
`incremental` block for any stream, matching legacy exactly.

`bitlinks` is scoped to one Bitly group: the path template `/groups/{{ config.group_guid }}/bitlinks`
substitutes the required `group_guid` config value (urlencoded by `InterpolatePath`'s
per-segment default, matching legacy's own `url.PathEscape(guid)` in `resolveResource`); an absent
`group_guid` hard-errors on both sides (legacy: `"bitly bitlinks stream requires config
group_guid"`; engine: an unresolved `config.group_guid` path-template key — same failure
classification, different literal text, per conventions.md §5's precedent for config-validation
parity). Records live at the `links` key. Pagination follows Bitly's own `pagination.next`
**absolute** URL convention (`pagination.type: next_url`, `next_url_path: "pagination.next"`) —
SPEC.md's N3 carried flag ("relative next-page URLs fail closed") does not bite here: Bitly's real
wire shape always emits a fully-qualified absolute URL (confirmed against legacy's own
`bitly_test.go:64`'s `TestReadBitlinksPaginates` fixture, which serves `srv.URL +
"/groups/g1/bitlinks?search_after=tok2"`, and legacy's own comment at `bitly.go:180-183`). The
engine's `next_url` paginator's same-host SSRF guard (THREAT-MODEL §3) passes cleanly since the
`next` URL Bitly returns is always same-origin as `base_url` in production. `size=50` (bitly's
own default page size, `bitlyDefaultPageSize`) is declared as a static per-stream `query` value
(`streams.json`'s `"query": {"size": "50"}`) and is, in fact, **re-sent on every page**: the
engine's `readDeclarative` merges `stream.Query` into EVERY page request
(`engine/read.go`'s `mergeQuery(baseQuery, page.Query)`), and `connsdk.Requester.resolveURL`
re-applies that merged query onto the absolute next-page URL (Del+Add, replacing any same-named
param already present in the URL) — unlike legacy, which explicitly resets to an empty
`url.Values{}` once it follows an absolute next-page URL (`bitly.go:180-183`), sending `size` only
on the first request. This is a **wire-request-shape divergence from legacy**, verified benign in
DATA terms only because Bitly's own `pagination.next` URL already carries the identical `size=50`
value the engine re-applies (the replace is idempotent) — so the effective query on every page is
value-identical to what the next URL already carries, not a behavior change an operator or Bitly's
API would observe differently. If Bitly's `next` URL ever omitted or diverged from `size`, this
bundle's request would differ from legacy's; today it does not.

## Write actions & risks

None. Bitly's core list endpoints have no obviously-safe reverse-ETL writes (legacy's own package
doc: "no obviously-safe reverse-ETL writes, so the connector advertises Write=false");
`capabilities.write` is `false` and this bundle ships no `writes.json`, matching legacy's
`Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`, a credential-free conformance-harness affordance) stamps
  extra fields — `connector` (a static "bitly" marker), `fixture` (always `true`), and
  `previous_cursor` (echoing `req.State["cursor"]` when a prior cursor happens to be set) — onto
  every fixture-mode record (`bitly.go:212-217`). None of these are part of the LIVE record shape;
  this bundle's schemas and parity suite target the live path only (`bitly.go`'s `harvest`
  function), per SPEC.md §5.2's explicit instruction to ignore legacy's fixture-mode-only fields.
  The engine's own conformance/fixture-replay harness (`internal/connectors/conformance`) provides
  the credential-free test affordance this bundle needs, so no fixture-mode equivalent is needed
  here.
- **`page_size` is not runtime-configurable.** Legacy exposes `page_size`/`max_pages` as
  config-driven overrides on the `bitlinks` stream (`bitlyPageSize`/`bitlyMaxPages`,
  `bitly.go:286-314`). The engine's `next_url` paginator has no analogous config-driven page-size
  knob (it never reads `PaginationSpec.PageSize`, unlike `page_number`/`offset_limit`), and
  `stream.Query` templating has no absent-key-falsy tolerance (conventions.md §3), so a
  `{{ config.page_size }}` template would hard-error whenever `page_size` is unset — the common
  case. This bundle therefore sends bitly's own default (`size=50`) as a static per-stream query
  literal, matching stripe's `limit=100` static-query precedent (`docs/migration/conventions.md`),
  and does not declare `page_size` in `spec.json` at all (F6, REVIEW.md: a declared-but-unwireable
  config key is worse than an absent one). `max_pages` (legacy's hard request-count cap override)
  is likewise not modeled; the engine's `next_url` paginator has no `MaxPages`-equivalent knob
  wired to a config value either — pagination is bounded only by the short/empty-page stop signal
  (an empty `pagination.next` value), matching Bitly's own real termination behavior.
