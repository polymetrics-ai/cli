# Overview

Ticketmaster is a wave2 fan-out declarative-HTTP migration. It reads events, venues, attractions,
and classifications from the Ticketmaster Discovery API
(`GET https://app.ticketmaster.com/discovery/v2/...`). This bundle targets capability parity with
`internal/connectors/ticketmaster` (the hand-written connector it migrates); the legacy package
stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Ticketmaster Discovery API key via the `api_key` secret; it is sent as the `apikey` query
parameter (`api_key_query` auth mode), matching legacy's `baseQuery`'s
`url.Values{"apikey": {key}}` (`ticketmaster.go:191-196`), and is never logged.

## Streams notes

All four streams share the identical Ticketmaster Discovery envelope:
`GET /<resource>.json` returns `{"_embedded":{"<resource>":[...]}}`; records live at
`_embedded.<resource>` for every stream (legacy's `recordsPath` values, uniform per stream).
Pagination is `page_number` (`page`/`size`, static `page_size: 200` matching legacy's
`defaultPageSize`). `start_page: 0` is declared because Ticketmaster's real page index is
zero-based (legacy's `harvestPages` loop starts at `page := 0` and sends `page=0` as its first
request, `ticketmaster.go:121-127`); the engine honors an explicit `start_page: 0` verbatim
(`startPageOrDefault` treats only a nil `*int` as unset, `engine/paginate.go`), so this bundle's
first live request is `page=0`, matching legacy exactly. Fixture query params carry the same
zero-based page numbers (`events` page_1.json = `page=0`, page_2.json = `page=1`).

Every stream also declares the optional `keyword`/`countryCode`/`locale` query filters (legacy's
`Read` sets these unconditionally from `config.keyword`/`config.country_code`/`config.locale`
before calling `harvestPages` with the requested stream's spec — `ticketmaster.go:89-93` — so all
four streams, not just events, receive the same filter set). Each is declared with
`omit_when_absent: true` so an unset filter is left off the request entirely rather than causing an
unresolved-key error, matching legacy's own `if v := ...; v != ""` conditional-set behavior.

`venues.city`/`venues.country` and `classifications.segment`/`genre`/`subGenre` are real
Ticketmaster wire-shape nested objects (confirmed against the Discovery API v2 reference), not
flat strings — legacy's own `Fields` catalog metadata mistakenly typed them `"string"`, but legacy's
`Read` never actually projects/flattens records (`harvestPages` emits `connectors.Record(rec)`
straight from `connsdk.RecordsAt`, no `mapRecord` step), so the real emitted data has always
contained these nested objects; this bundle's schemas correct the type to match the real emitted
data rather than propagating legacy's inaccurate catalog-metadata typing.

Because that same evidence line establishes a verbatim, no-field-building emit (no `mapRecord` step
of any kind, for any of the four streams), all four declare `"projection": "passthrough"` so every
raw field the Discovery API returns reaches the emitted record, matching legacy exactly rather than
silently dropping any field not listed in `schemas/*.json` (a schema-mode projection would have done
exactly that). The schemas stay a documentation surface describing the known/expected shape and are
not widened to `additionalProperties: true`, matching the pingdom/searxng precedent for this rule.

None of the four streams declare an `incremental` block: legacy's `Read`/`harvestPages` never
applies a cursor-based filter parameter — every read is a full paginated sweep, matching legacy's
true behavior exactly.

## Write actions & risks

None. Ticketmaster is read-only (`capabilities.write` is `false`); this bundle ships no
`writes.json`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`page_size`/`max_pages` config overrides are not modeled.** Legacy exposes `page_size`
  (default 200, positive integer) and `max_pages` (0/absent = unbounded, non-negative integer cap)
  as config-driven overrides (`pageSize`/`maxPages`, `ticketmaster.go:218-239`). The engine's
  `page_number` paginator has no config-driven page-size or request-count-cap knob (mirrors the
  aha/thinkific-courses precedent from this same wave); `page_size`/`max_pages` are therefore not
  declared in `spec.json`, and this bundle sends Ticketmaster's own default (`size=200`) as a
  static pagination-block value with no page cap.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path stamps a
  `fixture: true` marker field with no live-path equivalent (`ticketmaster.go:183`). This bundle's
  schemas and fixtures target the live record shape only; the engine's own
  `internal/connectors/conformance` fixture-replay harness provides the credential-free test
  affordance this bundle needs.
