# Overview

Secoda is a read-only declarative migration of `internal/connectors/secoda` (legacy Go connector).
It reads Secoda catalog metadata (tables, documents, collections, questions) through the Secoda API.
This bundle is capability-parity with legacy; legacy stays registered and unchanged until wave6's
registry flip.

## Auth setup

Provide a Secoda API key via the `api_key` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_key>`) and is never logged. There is no unauthenticated fallback mode —
legacy hard-errors when `api_key` is unset (`secoda connector requires secret api_key`), matching
this bundle's `required: ["api_key"]`.

## Streams notes

All 4 streams (`tables`, `documents`, `collections`, `questions`) share the identical shape: `GET`
against the Secoda list endpoint, records at `results`, primary key `["id"]`, and `page_number`
pagination (`page`/`page_size` query params, `page_size: 100` matching legacy's `defaultPageSize`).
Every stream declares `"projection": "passthrough"`: legacy's `Read` calls `connsdk.Harvest` directly
with no `mapRecord`-style projection function at all, so the raw record passes through unfiltered —
matching that exactly. None of the 4 streams declares an `incremental` block or an `x-cursor-field`:
although legacy's `streams()` catalog metadata lists `updated_at` as a `CursorFields` entry for
manifest purposes, legacy's real (non-fixture) `Read` path never actually filters or advances a sync
by it — every read is a full stream read — so declaring `x-cursor-field` here would misrepresent
`incremental_append` capability this connector does not have (the sync-mode derivation rule ties
`x-cursor-field`'s meaning to a backing `incremental` block, which legacy has no equivalent of).

Pagination is capped at `max_pages: 1` (legacy's own `defaultMaxPages`), matching legacy's default:
by default only the first page is ever fetched, regardless of how many records the API reports.
This is why the fixtures ship only a single page per stream (`fixtures/streams/<name>/page_1.json`)
rather than the usual 2-page proof — with `max_pages: 1` fixed at the base level, no bundle
configuration ever requests a second page, so there is nothing for a second fixture page to prove;
`pagination_terminates` still exercises the real short-circuit (exactly 1 request made, exactly 1
fixture page consumed).

## Write actions & risks

None. Legacy `Write` always returns `connectors.ErrUnsupportedOperation`; `metadata.json` declares
`capabilities.write: false` and no `writes.json` file exists, matching legacy exactly.

## Known limits

- `page_size` config validation (legacy's 1-500 numeric-range check) is not reproduced at the bundle
  level; the engine treats `page_size` as an opaque value substituted directly into the `page_size`
  query param via the pagination block, sent to Secoda as-is rather than rejected client-side the way
  legacy's `strconv.Atoi` range check would. This never changes emitted record DATA for any
  legacy-valid input; it only narrows client-side input validation, out of scope for wave2 fan-out
  (Pass B).
- Legacy's `max_pages` config (a non-negative integer, defaulting to `1`) has no bundle-level
  spec property — the hard cap is instead a fixed literal in `streams.json`'s
  `base.pagination.max_pages`, matching legacy's own default value exactly but not exposing it as a
  runtime-configurable override the way legacy's `maxPages(raw string)` helper does. A caller needing
  a different page-count ceiling than legacy's default has no declarative equivalent to set one. Out
  of scope for wave2 fan-out (Pass B).
