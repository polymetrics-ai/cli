# Overview

AppsFlyer is a Tier-2 (declarative bundle + `StreamHook`) migration of
`internal/connectors/appsflyer`. Legacy's read path is a plain connsdk-HTTP GET against
AppsFlyer's Pull API raw-data export endpoints, but the response body is **`text/csv`, not
JSON** (`emitCSV` decodes it via `encoding/csv`) — a documented Tier-2 trigger
(`docs/migration/conventions.md` §1's Tier-2 table: "response decompression/CSV parsing").
`internal/connectors/hooks/appsflyer/hooks.go` implements `StreamHook` for both streams, porting
legacy's `emitCSV`/`snake`/`reportQuery`/`reportPath` logic verbatim; the declarative
`streams.json` still declares both streams' shape (path, check) for identity/documentation/
`connectorgen validate` coverage, but the declarative record-extraction path is never reached in
practice (`streams[].conformance.skip_dynamic` marks both). This bundle is engine-vs-legacy
parity-tested against `internal/connectors/appsflyer` (the hand-written connector it migrates);
the legacy package stays registered and unchanged until the registry flip.

## Auth setup

Provide an AppsFlyer Pull API token via the `api_token` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_token>`) and is never logged, matching legacy's
`connsdk.Bearer(token)` (`appsflyer.go:107`). `base_url` defaults to `https://hq1.appsflyer.com`
and may be overridden for tests/proxies. `app_id` is required (path-substituted into every
request, `url.PathEscape`-safe via `InterpolatePath`'s default urlencode filter).

## Streams notes

Both streams (`installs_report`, `in_app_events_report`) hit
`GET /api/raw-data/export/app/{{ config.app_id }}/<report>/v5` and return a CSV body whose header
row names every column; `hooks/appsflyer/hooks.go`'s `ReadStream` decodes it via `encoding/csv`,
snake-cases each header (identical to legacy's `snake()` — including legacy's specific
`apps_flyer` -> `appsflyer` correction so `AppsFlyer ID` becomes `appsflyer_id`, not
`apps_flyer_id`), and emits one record per data row with every column present (verbatim
passthrough — `streams.json` declares `"projection": "passthrough"` for documentation honesty,
though the declarative projection path is never actually invoked since the hook calls `emit`
directly). `from`/`to` query params are derived from `start_date`/`end_date` (date-only portion,
matching legacy's `firstDate` truncation at the first space), defaulting `to` to `start_date` when
`end_date` is unset (legacy's `first(end_date, start_date)`); an optional `timezone` config value
is sent verbatim when set. No incremental cursor is modeled — legacy's own catalog declares no
`CursorFields` for either stream, and `from`/`to` are static per-read config values, not a
persisted-cursor-driven incremental filter.

## Write actions & risks

None. Legacy `appsflyer.Write` always returns `connectors.ErrUnsupportedOperation`;
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **No `x-primary-key` is declared on either schema.** Legacy's own `Catalog()` declares no
  `PrimaryKey` for either stream (a raw CSV export row has no natural unique identifier AppsFlyer
  guarantees is stable/unique across report re-exports) — this bundle matches that shape exactly
  rather than inventing one.
- **Dynamic (fixture-replay) conformance is skipped for both streams
  (`streams[].conformance.skip_dynamic`).** The standard `fixtures/streams/<stream>/page_N.json`
  envelope's `response.body` is JSON (`json.RawMessage`), and `conformance`'s replay server always
  serves it with `Content-Type: application/json` — there is no way to represent AppsFlyer's real
  `text/csv` wire shape in that envelope. The hook's CSV-decode behavior is instead covered by
  `internal/connectors/hooks/appsflyer/hooks_test.go`, which drives a real `httptest.Server`
  serving actual CSV bytes and asserts header-row snake-casing, auth, query construction
  (`from`/`to`/`timezone`), and per-row record emission — this is the authoritative proof for the
  skipped behavior, per conventions.md's `skip_dynamic` rule ("the reason MUST name the
  authoritative substitute that actually proves the skipped behavior"). `check_fixture` (which
  never parses the response body) is unaffected and still runs normally.
- **`fixtures/streams/<stream>/page_1.json` is a structural "shadow" fixture, never actually
  replayed.** `conformance`'s STATIC `fixtures_present` check (design §E.2 "first stream
  mandatory") requires at least one fixture page file to exist for the bundle's first declared
  stream regardless of any `skip_dynamic` marker (that marker only affects DYNAMIC checks); this
  bundle ships a page_1.json per stream whose `response.body` is a JSON string containing sample
  CSV text (mirroring monday's identical all-streams-skipped shadow-fixture shape), purely to
  satisfy that structural presence check — it is never read by any dynamic replay path since both
  streams are hook-handled.
- **Fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached when
  `config.mode == "fixture"`) synthesizes fields directly in Go (`fixture: true`,
  `appsflyer_id: "af_fixture_N"`, etc., `appsflyer.go:161-171`) and never talks to a real
  AppsFlyer API; this bundle's own hook-driven tests provide the equivalent credential-free
  affordance, so no fixture-mode config branch is modeled here.
