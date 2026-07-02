# Overview

New York Times is a wave2 fan-out declarative-HTTP migration. It reads the NYTimes Most Popular
(viewed, emailed, shared) article feeds through the NYTimes Developer APIs
(`GET https://api.nytimes.com/svc/mostpopular/v2/...`). This bundle targets capability parity with
`internal/connectors/nytimes` (the hand-written connector it migrates) for the Most Popular
streams only; the legacy package stays registered and unchanged until wave6's registry flip, and
continues to serve the `archive` stream (see "Known limits").

## Auth setup

Provide an NYTimes Developer API key via the `api_key` secret; it is sent as the `api-key` query
parameter (`pagination.type` is not used here — no pagination exists on these endpoints), matching
legacy's `connsdk.APIKeyQuery("api-key", secret)` exactly, and is never logged.

## Streams notes

3 streams (`most_popular_viewed`, `most_popular_emailed`, `most_popular_shared`) each issue a
single `GET` request to `/mostpopular/v2/<metric>/{{ config.period }}.json` with no pagination —
every result is returned inline in one response, matching legacy's `readMostPopular` (no loop, no
next-page token). `period` (`1`, `7`, or `30` days) defaults to `"7"`, matching legacy's
`nytimesPeriod` fallback. Records live at `results`; primary key `["id"]`; `published_date` is
carried as a cursor field for catalog purposes only — legacy applies no server-side date filter on
these endpoints (no `incremental` block is declared here, matching that exactly).

## Write actions & risks

None. NYTimes is a read-only source connector (`capabilities.write: false`); this bundle ships no
`writes.json`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **The `archive` stream is BLOCKED, not migrated (`ENGINE_GAP`).** Legacy's `readArchive`
  (`nytimes.go:173-217`) iterates one calendar month at a time from `start_date` to `end_date`
  (inclusive), issuing a separate `GET /archive/v1/{year}/{month}.json` request per month and
  accumulating records across the whole window — a genuine multi-request, config-driven fan-out
  loop over a computed sequence of path segments. The declarative dialect has no primitive for
  "iterate N path-templated requests derived from two config values" (`stream.path` is a single
  template resolved once per `Read` invocation's pagination loop, not a generator of N request
  variants); this is exactly the sub-resource/multi-request fan-out class `conventions.md` §1
  reserves for a Tier-2 `StreamHook` (or, if that undersells the complexity, Tier-3). Per this wave's
  hard rule (JSON + docs.md only, no Go), no hook was authored; `archive` is not declared in this
  bundle's `streams.json` at all and the legacy `internal/connectors/nytimes` package remains the
  only implementation for it. A follow-up wave with hooks authoring in scope should add a
  `StreamHook` that reproduces `readArchive`'s month-iteration loop verbatim.
- **`most_popular_shared`'s optional `share_type` path-segment breakdown is not modeled (documented
  scope narrowing, not a blocker).** Legacy's `readMostPopular` builds
  `/mostpopular/v2/shared/{period}/{share_type}.json` (an extra path segment) when a `share_type`
  config value is set, falling back to `/mostpopular/v2/shared/{period}.json` (identical shape to
  `viewed`/`emailed`) when it is unset. `stream.path` is a single fixed template with no per-segment
  conditional/omit-when-absent tolerance (unlike `stream.Query`'s opt-in object form — path
  interpolation has no equivalent), so only legacy's own no-`share_type` fallback (the common,
  default-config case) is implemented here; the `share_type`-present variant is out of scope. This
  is strictly a narrowing (a subset of legacy's accepted configs), never a data-shape change for any
  config this bundle does accept.
- Full NYTimes surface (Books, Top Stories, Article Search, and the `archive` API itself) is out of
  scope for this wave; see `api_surface.json`'s `excluded`/blocked entries.
