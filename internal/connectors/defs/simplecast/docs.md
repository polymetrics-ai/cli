# Overview

Simplecast is a wave2 fan-out declarative-HTTP migration. It reads Simplecast podcasts and episodes
through the Simplecast REST API (`GET https://api.simplecast.com/...`). This bundle targets
capability parity with `internal/connectors/simplecast` (the hand-written connector it migrates);
the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Simplecast OAuth access token via the `access_token` secret; it is sent as a Bearer token
(`Authorization: Bearer <access_token>`) and is never logged, matching legacy's
`connsdk.Bearer(token)` (`simplecast.go:134`). `base_url` defaults to `https://api.simplecast.com`
and may be overridden for tests/proxies.

## Streams notes

Both streams (`podcasts`, `episodes`) are `GET` list endpoints with records at the top-level
`collection` key, primary key `["id"]`, incremental cursor field `updated_at`. The first request
sends `page=1`/`limit=100` (legacy's own `simplecastDefaultPageSize` default); pagination follows
Simplecast's own absolute-URL convention (`pagination.type: next_url`,
`next_url_path: "pages.next.href"`) — the next page's fully-qualified URL is read from the response
body's `pages.next.href` field, matching legacy's own `connsdk.StringAt(resp.Body,
"pages.next.href")` call, and pagination stops when that value is null/absent (legacy's
`strings.TrimSpace(nextURL) == ""` stop condition).

**Wire-request-shape divergence from legacy on page 2+ (bitly precedent,
`docs/migration/conventions.md`'s ledger item for bitly)**: legacy explicitly resets to an empty
`url.Values{}` once it follows an absolute next-page URL (`simplecast.go:96-101`: `q` stays `nil`
for every page after the first), sending `page`/`limit` only on the initial request. The engine's
`readDeclarative` merges `stream.Query` into EVERY page request regardless of whether that page's
URL came from `page.URL` (an absolute next-page URL) or was built from the stream's own `path`
(`engine/read.go`'s `mergeQuery(baseQuery, page.Query)`), and `connsdk.Requester.resolveURL`
re-applies that merged query onto the absolute URL (Del+Add, replacing any same-named param already
present). This is a data-safe divergence only because Simplecast's own `pages.next.href` URL already
carries `page`/`limit` values consistent with the merged query the engine re-applies — verified
against the pilot's own fixture/legacy test shape (`simplecast_test.go`'s `pages.next.href` always
embeds the same `limit` the client requested). If Simplecast's `next` URL ever diverged from the
static `page`/`limit` this bundle declares, the effective request would differ from legacy's; today
it does not.

**Sanctioned single-page fixture exception (`next_url` pagination, conventions.md §4)**: both
streams ship a single-page fixture because a `next_url` stream's next-page URL is the replay
server's own address, unknown until the harness picks a port at runtime — a static fixture file
cannot embed the correct absolute URL for a second page. `fixtures/streams/{podcasts,episodes}/
page_1.json` set `pages.next: null`, satisfying `fixtures_present`/`read_fixture_nonempty` and
`pagination_terminates` (which exercises this bundle's only non-marker-excluded stream and expects
exactly 1 request for exactly 1 fixture page).

Legacy's shared `simplecastRecord` mapper derives `title` via `first(item, "title", "name")` — both
streams' real wire shape uses `title` directly (confirmed by legacy's own fixture/test shape), so
plain schema projection (an exact-key-match copy) reproduces the identical value with no
`computed_fields` rename needed; the `name` fallback is never exercised on real Simplecast responses.

## Write actions & risks

None. Simplecast's read endpoints have no obviously-safe reverse-ETL writes modeled;
`capabilities.write` is `false` and this bundle ships no `writes.json`, matching legacy's `Write`
returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`pageSize`/`maxPages` helpers, `simplecast.go:214-238`). The engine's `next_url`
  paginator has no analogous config-driven page-size/max-pages knob (it never reads
  `PaginationSpec.PageSize`/`MaxPages`), and `stream.Query` templating has no absent-key-falsy
  tolerance, so a `{{ config.page_size }}` template would hard-error whenever `page_size` is unset —
  the common case. This bundle sends Simplecast's own default (`limit=100`) as a static per-stream
  query literal (matching stripe's/bitly's static-query precedent) and does not declare
  `page_size`/`max_pages` in `spec.json` at all (F6, REVIEW.md). Pagination is bounded only by the
  short/empty-`pages.next.href` stop signal, matching Simplecast's own real termination behavior.
- See "Streams notes" above for the documented, verified-benign query-re-application divergence on
  page 2+ next-URL requests, and the sanctioned single-page conformance fixture for `next_url`
  pagination.
