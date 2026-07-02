# Overview

Northpass LMS is a wave2 fan-out declarative-HTTP migration of
`internal/connectors/northpass-lms` (the hand-written legacy connector this bundle migrates; the
legacy package stays registered and unchanged until wave6's registry flip). It reads Northpass LMS
people, courses, course enrollments, and groups through the Northpass REST API v2
(`https://api.northpass.com/v2`), a JSON:API-flavoured API (list responses carry records under
`data` as `{id,type,attributes}`, paginated via a `links.next` absolute URL). Read-only.

## Auth setup

Provide a Northpass API key via the `api_key` secret; it is sent as the `X-Api-Key` header
(matching legacy's `connsdk.APIKeyHeader("X-Api-Key", secret, "")`) and is never logged.
`base_url` defaults to `https://api.northpass.com/v2` and may be overridden for tests/proxies.

## Streams notes

All 4 streams (`people`, `courses`, `course_enrollments`, `groups`) share the same shape: `GET`
against the Northpass collection endpoint, records extracted from the response's `data` array
(each a JSON:API `{id, type, attributes}` object). `id`/`type` project straight from the raw
top-level keys via schema projection; every other field lives one level down at
`attributes.<field>` and is surfaced via a `computed_fields` rename (e.g. `"email": "{{
record.attributes.email }}"`), matching legacy's `attributesOf(item)` flattening helper
byte-for-byte. `course_enrollments`' `percentage` field is a bare single-reference
`computed_fields` entry (`{{ record.attributes.percentage }}`), so it receives the engine's typed
extraction and preserves its native JSON integer type rather than being stringified.

No stream declares an `incremental` block — legacy never implements `InitialState` or exposes a
cursor field for any of these 4 objects (`northpass-lms/streams.go`'s doc comment: "Only
full-refresh sync is supported upstream, so no cursor fields are advertised"), so every sync
(fresh or resumed) reads the full, unfiltered collection, matching legacy exactly.

Pagination is `next_url` (`next_url_path: links.next`): the first request sends `limit=100`
(legacy's `northpassDefaultPageSize`) as a static per-stream query value; every subsequent page's
absolute URL (already carrying its own `limit`/pagination state) is followed directly, matching
legacy's `harvest` loop (`northpass_lms.go:140-181`), which explicitly clears the merged query
once it starts following an absolute `links.next` URL. The engine's `next_url` paginator enforces
a same-host SSRF guard by default; Northpass's own `links.next` URLs are always same-origin as
`base_url` in production, so this is not exercised differently from legacy.

## Write actions & risks

None. Legacy `northpass-lms` is read-only (`Write` returns `connectors.ErrUnsupportedOperation`);
`metadata.json` declares `capabilities.write: false` and this bundle ships no `writes.json`.

## Known limits

- Full Northpass API surface (certificates, quizzes, course content, webhooks) is out of scope
  for wave2; see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B
  capability expansion"}` entries. Only the 4 legacy-parity streams are implemented.
- **Fixtures are single-page** (`fixtures/streams/<stream>/page_1.json`, each with `links.next:
  null`), matching klaviyo's identical `next_url`-pagination-golden shape: every stream in this
  bundle shares the `next_url` type at the base level, so `conformance`'s `pagination_terminates`
  check exercises the first stream (`people`)'s single page trivially (1 fixture page served
  exactly once). A genuine 2-page `next_url` fixture is not authorable here (the second page's
  absolute URL is the replay server's own runtime-assigned address, unknown until the harness
  picks a port — conventions.md §4's sanctioned `next_url` exception). Real 2-page
  `links.next`-follow correctness (the engine issuing a request to the literal absolute URL found
  in the prior response body, with the merged static query correctly NOT re-appending `limit` a
  second time in a conflicting way) is proven by the engine's own `next_url` paginator unit tests
  (`internal/connectors/engine/paginate_test.go`) rather than a connector-specific
  `paritytest/northpass-lms` suite, which this wave2 fan-out migration does not create (out of
  scope per this wave's mandate — JSON+docs.md only).
- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`northpassPageSize`/`northpassMaxPages`). The engine's `next_url` paginator has no
  config-driven page-size knob at all (it never reads `PaginationSpec.PageSize`, unlike
  `page_number`/`offset_limit`) and no `MaxPages`-equivalent stop mechanism beyond the paginator's
  own empty-`links.next` stop signal, so neither is declared in `spec.json` (a
  declared-but-unwireable config key is worse than an absent one, per conventions.md §3). This
  bundle sends legacy's own default (`limit=100`) as a static per-stream query literal, matching
  stripe's `limit=100`-via-static-query precedent.
