# Overview

OnePageCRM is a read-only declarative-HTTP migration (wave2 fan-out) of `internal/connectors/onepagecrm`
(the hand-written connector it replaces at capability parity). It reads contacts, deals, actions,
companies, and users through the OnePageCRM v3 REST API. The legacy package stays registered and
unchanged until wave6's registry flip.

## Auth setup

Provide the OnePageCRM API user ID as the `username` config value and the OnePageCRM API key as the
`password` secret; both are required. They are sent as HTTP Basic auth (`username:password`),
matching legacy `onepagecrm.go`'s `connsdk.Basic(username, password)` exactly. The password is never
logged.

## Streams notes

All 5 streams (`contacts`, `deals`, `actions`, `companies`, `users`) share OnePageCRM's list-endpoint
shape: `GET /<resource>`, records under `data.<resource>` (`users` is the one exception — its records
live directly under `data`, matching legacy's `arrayPath: "data"` for that stream only). Every list
element is wrapped under a singular key (e.g. `{"contact": {...}}`); since the engine's schema
projection only matches top-level raw keys and has no per-element unwrap primitive, every field is
instead extracted via a `computed_fields` bare `{{ record.<wrap_key>.<field> }}` reference — the
engine's typed-extraction rule (a single bare `record.*` reference with no filter) copies the raw
JSON value straight through with its native type preserved (e.g. `deals.amount` stays a number,
`contacts.starred` stays a boolean), exactly matching legacy's `unwrap`+`mapRecord` behavior
field-for-field.

Pagination is `page`/`per_page` (`pagination.type: page_number`, `page_param: page`, `size_param:
per_page`, `start_page: 1`, `page_size: 100` matching legacy's default `onepagecrmDefaultPageSize`).

Documented parity deviation (stop condition): legacy's `harvest` loop stops primarily on the
response body's `data.max_page` field (falling back to a short-page heuristic only when `max_page` is
absent or unparsable). The engine's `page_number` paginator has no body-driven stop signal at all —
it stops purely on a short/empty page (`recordCount < page_size`). For every real OnePageCRM response
this bundle's pagination terminates at the exact same page a still-fully-populated final page (one
whose record count happens to equal `page_size`) causes one extra page request versus legacy (which
would stop immediately via `max_page`); that extra request returns an empty page and the engine stops
on the following iteration. No record is ever duplicated, dropped, or reordered by this difference —
it is a request-count/efficiency deviation, never an emitted-record-data deviation, so it is
ACCEPTABLE per `docs/migration/conventions.md`'s parity-deviation meta-rule.

## Write actions & risks

None. OnePageCRM is exposed read-only in legacy (`Capabilities.Write: false`); this bundle ships no
`writes.json`.

## Known limits

- OnePageCRM's write surface (creating/updating contacts, deals, actions) is out of scope; legacy
  itself never implemented it, so there is no parity gap, only an out-of-scope Pass B expansion (see
  `api_surface.json`).
- The pagination stop-condition deviation described above (body `max_page` vs. short-page heuristic)
  is an accepted, request-count-only difference — see Streams notes.
- `contacts`/`deals`/`actions`/`companies` declare `updated_at` as `x-cursor-field` for manifest
  parity with legacy's published `CursorFields`, but neither legacy nor this bundle actually issues a
  server-side incremental filter against it (legacy's OnePageCRM API integration performs full syncs
  only); `users` has no cursor field at all, matching legacy exactly.
