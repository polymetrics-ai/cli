# Overview

OnceHub is a wave2 fan-out declarative-HTTP migration. It reads OnceHub bookings, contacts,
booking pages, users, and event types through the OnceHub REST API
(`GET https://api.oncehub.com/v2/...`). This bundle targets capability parity with
`internal/connectors/oncehub` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide an OnceHub API key via the `api_key` secret; it is sent as the `API-Key` header, matching
legacy's `connsdk.APIKeyHeader("API-Key", secret, "")` exactly, and is never logged.

## Streams notes

All 5 streams (`bookings`, `contacts`, `booking_pages`, `users`, `event_types`) are `GET` list
endpoints whose responses carry records at `data`, matching legacy's `oncehubStreamEndpoints`
table exactly. Pagination follows OnceHub's `after=<last record id>` + `Link: rel="next"` header
convention: the next page's cursor is the `id` field of the LAST record on the current page, and
the response's `Link` header signals whether another page exists. This bundle declares
`pagination.type: link_header`: the engine follows the exact absolute URL OnceHub's own API
supplies in the `Link: <url>; rel="next"`
header (`engine/paginate.go`'s `linkHeaderPaginator`), which OnceHub's own pagination
implementation constructs as `<base>?after=<last id>` — mechanically equivalent to legacy's own
independent re-derivation of `after` from the last record's `id` field, since both are ultimately
driven by the same server-side pagination state. `limit` (and, for `bookings`,
`last_updated_time.gt`) are re-applied onto every followed page URL via the engine's query-merge
behavior (`read.go`'s `mergeQuery`, `connsdk`'s `resolveURL` Del+Add semantics — see bitly's
identical documented pattern in `docs/migration/conventions.md`), reproducing legacy's own
`cloneValues(base)` + `after` reconstruction on every page request. Pagination stops when the
response carries no `Link: rel="next"` header, exactly matching legacy's `hasNextLink` check —
critically, legacy's own test (`oncehub_test.go`'s `TestReadPaginatesAndAuthenticates`) shows the
LAST page still returning a record with a valid `id` but no Link header, and legacy correctly stops
there rather than continuing to page on cursor availability alone; `link_header`'s stop signal
(header presence, not a body field or "last record has an id") reproduces this exactly, which a
`cursor`+`last_record_field`+`stop_path` declaration could NOT (`stop_path` only ever reads the
response BODY, never a header — there is no dialect primitive for a header-sourced stop signal on
that paginator variant).

Only `bookings` supports OnceHub's `last_updated_time.gt` incremental filter (matching legacy's
`streamEndpoint.incremental` flag, which is `true` only for `bookings`): `incremental.cursor_field:
last_updated_time`, `request_param: last_updated_time.gt`, `start_config_key: start_date` — the
lower bound is the sync's persisted cursor, falling back to the RFC3339 `start_date` config value
on a fresh sync, identical to legacy's `incrementalLowerBound`. The other 4 streams declare no
`incremental` block, matching legacy exactly (their `streamEndpoint.incremental` is `false`, so
`Read` never sends the filter for them).

## Write actions & risks

None. OnceHub is a read-only source connector (`capabilities.write: false`); this bundle ships no
`writes.json`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **The fixture-replay harness cannot exercise `link_header`'s real 2-page continuation.**
  `fixtures/streams/**` (`conformance/replay.go`'s `fixtureResponse` shape) has no field for
  declaring HTTP response headers — only `status` and `body` — so a fixture page can never carry
  the `Link: <url>; rel="next"` header OnceHub's real API sends, and `pagination_terminates`/
  `records_match_schema` can only ever observe the paginator's natural single-page stop (no Link
  header present = no next page, exactly like a real last-page response). This is a structural
  fixture-format limitation affecting any `link_header` bundle in this repo (see buildkite's/
  guru's identical documented limit), not an oncehub-specific shortcut. Every stream fixture here
  is a single, representative page; the 2-page Link-header-following codepath itself
  (`internal/connectors/engine/paginate.go`'s `linkHeaderPaginator`) is exercised by the shared
  engine's own `paginate_test.go` coverage. Per hard-rule scope for this migration wave, no Go
  (hooks/paritytest) was authored to work around this.
- **`max_pages` is not runtime-enforced beyond the engine's generic hard cap.** Legacy exposes
  `max_pages` as a config-driven request-count override read fresh on every `Read` call
  (`oncehubMaxPages`). The `link_header` paginator type has no per-request page-count field of its
  own in `PaginationSpec` beyond the generic `MaxPages` hard cap (`read.go`'s `readDeclarative`
  loop enforces it when set to a positive integer, matching legacy's 0/all/unlimited-means-unbounded
  semantics for the bounded case; the same non-wired gap class buildkite/bitly/guru document for
  their own non-`page_number` paginators).
- **Legacy's fixture-mode-only field (`previous_cursor`) is not modeled.** Legacy's `readFixture`
  path (only reached when `config.mode == "fixture"`) stamps `previous_cursor` from
  `req.State["cursor"]` onto fixture-mode records — not part of the live API shape. This bundle's
  schemas and fixtures target the live record shape only; the engine's own conformance/fixture-replay
  harness provides the credential-free test affordance legacy's fixture mode was built for.
