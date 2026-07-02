# Overview

Aircall is a wave2 fan-out declarative-HTTP migration. It reads Aircall calls, users, contacts,
numbers, and teams through the Aircall REST API (`GET https://api.aircall.io/v1/...`). This bundle
targets capability parity with `internal/connectors/aircall` (the hand-written connector it
migrates); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide `api_id` and `api_token` secrets; they are sent as HTTP Basic auth (`api_id:api_token`,
base64-encoded) and never logged, matching legacy's `connsdk.Basic(id, token)`
(`aircall.go:257`). `base_url` defaults to `https://api.aircall.io/v1` (legacy's
`aircallDefaultBaseURL`) and may be overridden for tests/proxies.

## Streams notes

All five streams share Aircall's `meta.next_page_link` envelope: `GET /<resource>` returns
`{"<resource>":[...],"meta":{"next_page_link":<url|null>,...}}`, records live at the
resource-named top-level key (identical to the resource segment: `calls`/`users`/`contacts`/
`numbers`/`teams`). Pagination is `next_url` (`next_url_path: meta.next_page_link`) — Aircall
returns a fully-qualified absolute next-page URL, matching legacy's own "follow next_page_link
verbatim" behavior (`aircall.go:190-193`) and the engine's `next_url` paginator's same-host SSRF
guard (THREAT-MODEL §3), which passes cleanly in production since Aircall's own `next_page_link`
is always same-origin as `base_url`. `per_page=50` (Aircall's own default, `aircallDefaultPageSize`)
is a static per-stream query value re-sent on every page request the engine issues; see Known
limits for the same "re-sent vs. legacy's reset-to-empty-then-follow" divergence already documented
on bitly's bundle (this wave's `next_url` sibling pilot).

`calls` and `contacts` are the two streams Aircall's own API supports a `from` (unix-seconds) lower
bound filter on (legacy's `harvest`, `aircall.go:156-158`: `if fromUnix != "" && (endpoint.resource
== "calls" || endpoint.resource == "contacts")`); this bundle expresses that exact same-branch
gating via `incremental.request_param: "from"` (declared only on `calls`', cursor field
`started_at`, and `contacts`', cursor field `created_at`, `incremental` blocks) with
`param_format: unix_seconds` (matching legacy's `toUnixSeconds` conversion of the RFC3339
`start_date`/state cursor) — the engine's `buildInitialQuery` sends `request_param` only when the
formatted lower bound is non-empty, identical to legacy's own `if fromUnix != ""` gate. `users`,
`numbers`, and `teams` declare no `incremental` block and no `from` query key at all, matching
legacy's `harvest` never setting `from` for those three resources (the `base.Set("from", ...)` call
is conditioned on the resource name, so `users`/`numbers`/`teams` never receive it either way).

## Write actions & risks

None. Aircall's API has no obvious safe reverse-ETL surface (legacy's own package doc: "no obvious
safe reverse-ETL surface"); `capabilities.write` is `false` and this bundle ships no `writes.json`,
matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`next_url` fixtures are single-page, per the sanctioned exception (conventions.md §4).** A
  `next_url` stream's next-page URL is the replay server's own runtime address, unknown until the
  harness picks a port — a static fixture file cannot embed the correct absolute URL for a second
  page. Every stream in this bundle ships a single-page fixture (satisfies `fixtures_present`/
  `read_fixture_nonempty`); `pagination_terminates` passes on the first stream (`calls`) with its
  single page (`hits == len(pages) == 1`), which is not a 2-page pagination proof but is not a
  false failure either. Real 2-page `next_url`-following correctness for THIS bundle's exact request
  shape is proven by legacy's own existing test
  (`internal/connectors/aircall/aircall_test.go`'s `TestReadPaginatesAndAuthenticates`, which drives
  a real 2-page `httptest.Server` and asserts the second page is requested via the served
  `next_page_link`), plus the engine's own generic `next_url` paginator unit tests
  (`internal/connectors/engine/paginate_test.go`'s `TestNewPaginatorNextURLFollowsAbsoluteURL` and
  siblings) and read-path integration test
  (`internal/connectors/engine/read_test.go`'s `TestReadNextURLPaginationSetsBaseHostFromRequester`).
  This wave does not add a new `paritytest/aircall` package (out of scope per this wave's JSON-only
  mandate); a future wave adding hand-written parity suites should follow bitly's/calendly's
  `TestParity<Name>_..._TwoPagePagination` pattern for aircall specifically.
- **`per_page` is re-sent on every page request, unlike legacy's reset-to-empty-then-follow.** The
  engine's `readDeclarative` loop merges `stream.Query` into every page request regardless of
  pagination type, and `connsdk.Requester.resolveURL` re-applies that merged query onto the absolute
  next-page URL (replacing any same-named param already present). Legacy instead resets to an empty
  `url.Values{}` once it follows an absolute next-page URL (mirrors bitly's identical, already-ledgered
  divergence in this same wave — see bitly's `docs.md`). This is benign in DATA terms only because
  Aircall's own `next_page_link` already carries the identical `per_page` value the engine
  re-applies (the replace is idempotent); if Aircall's `next_page_link` ever diverged from
  `per_page`, this bundle's request would differ from legacy's — today it does not.
- **`per_page`/`max_pages` config overrides are not modeled.** Legacy exposes `per_page` (1-50,
  default 50) and `max_pages` (0/all/unlimited or a positive integer cap) as config-driven overrides
  (`aircallPageSize`/`aircallMaxPages`). The engine's `next_url` paginator has no config-driven
  page-size or request-count-cap knob at all (mirrors bitly's identical, already-ledgered
  limitation); `per_page`/`max_pages` are therefore not declared in `spec.json`, and this bundle
  sends Aircall's own default (`per_page=50`) as a static per-stream query literal.
