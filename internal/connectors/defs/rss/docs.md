# Overview

RSS reads channel metadata and feed items from any RSS 2.0 feed URL. It is read-only and requires
no credentials. Legacy (`internal/connectors/rss/rss.go`) is a `connsdk`-based HTTP reader whose
response body is **RSS/XML**, decoded via `encoding/xml`
(`rss.go`'s `rssDocument`/`rssChannel`/`rssItem` structs) — not JSON. This is a **Tier-2 hooks
migration** (`docs/migration/conventions.md` §1/§6): the engine's declarative read path
(`internal/connectors/engine/read.go`) only ever decodes a response body as JSON
(`connsdk.RecordsAt`/`extractRecords` both assume a JSON envelope); XML body decoding is one of the
conventions.md §1 Tier-2 table's explicitly named legitimate triggers ("multipart/XML bodies").
`internal/connectors/hooks/rss/hooks.go` implements `StreamHook` (both streams: `items`, `channel`)
and `CheckHook` (the health check also decodes XML, via the same `load` helper), porting
`rss.go`'s `load`/`channelRecord`/`itemRecord` logic verbatim, reusing `rt.Requester` (the engine's
already-built HTTP client/headers/auth plumbing) exactly as monday's StreamHook does. The legacy
package stays registered and unchanged until wave6's registry flip.

## Auth setup

None. `streams.json`'s `base.auth` is `[{"mode": "none"}]`, matching legacy's credential-free
`connsdk.Requester` construction. `base.headers` declares `Accept: application/rss+xml,
application/xml, text/xml, */*`, matching legacy's `r := connsdk.Requester{..., Accept: "application/rss+xml, application/xml, text/xml, */*"}`
(`rss.go:119`) exactly — this is a declarative header (not something the hook needs to set itself),
since `hooks/rss/hooks.go` reuses the SAME engine-built `rt.Requester` that already carries it.

## Streams notes

Legacy defines 2 streams, both derived from ONE GET request against the configured feed URL
(`rss.go`'s `load`, `Read`): `items` (every `<channel><item>` element) and `channel` (the channel's
own metadata, one record). `hooks/rss/hooks.go`'s `ReadStream` fetches the feed exactly once per
call via `rt.Requester.Do(ctx, http.MethodGet, "", nil, nil)` (an empty path resolves to
`streams.json`'s `base.url`, i.e. `config.feed_url`), decodes the body with `encoding/xml` into the
identical `rssDocument`/`rssChannel`/`rssItem` shape legacy uses, and emits records via the same
`itemRecord`/`channelRecord` mapping functions (id fallback chains ported verbatim: an item's id is
`guid`, falling back to `link`, falling back to `title`; a channel's id is `link`, falling back to
`title`). `published_at`/`updated_at` are the raw `<pubDate>`/`<lastBuildDate>` text, never
reformatted, matching legacy exactly.

`items` declares bare `incremental.cursor_field: "published_at"` with no `request_param` — matching
legacy's published `CursorFields: []string{"published_at"}` (`rss.go:73`) while staying behaviorally
identical to legacy's real always-full-sync behavior (legacy's `Read` never consults incoming
state/cursor at all).

### Declarative path (`streams.json`) vs. the live StreamHook/CheckHook path

`streams.json` still declares complete stream/schema metadata for both streams (the catalog/manifest
surface: stream names, schemas, PK/cursor fields) even though `hooks/rss/hooks.go`'s `StreamHook`
recognizes and handles both stream names unconditionally (`handled=true` always) and its `CheckHook`
handles every `Check()` call — the declarative fallback is **never exercised by production
traffic**. `metadata.json` carries a bundle-level `"conformance": {"skip_dynamic": true, "reason":
"..."}` marker (`docs/migration/conventions.md` §4/§6, the gmail-shaped bundle-level variant, chosen
over per-stream markers because rss's `Check()` is ALSO XML-based, not just its reads — a
declarative fixture-replay `check_fixture` check has no faithful way to exercise an XML-decoding
CheckHook either). This Skips every auth-dependent dynamic check outright (`check_fixture`, every
`read_fixture_nonempty:<stream>`, `pagination_terminates`, `records_match_schema`,
`cursor_advances`) — `fixtures/**` are retained purely as documentation (satisfying the static
`fixtures_present` check, which only requires the FIRST declared stream to have at least one fixture
page) with an explicit human-readable note in each file's body explaining they are never replayed.
The authoritative substitute is `internal/connectors/hooks/rss/hooks_test.go` (unit tests against a
real `httptest.Server` serving actual RSS/XML) and `internal/connectors/paritytest/rss/parity_test.go`
(drives the real hook-dispatched connector against the SAME `httptest.Server` fixture legacy's own
test uses, asserting byte-for-byte record parity).

## Write actions & risks

None. `capabilities.write: false`, no `writes.json` — matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation` unconditionally (`rss.go:110-112`).

## Known limits

- **`encoding/xml` decoding is unreachable from the declarative path (ENGINE_GAP, documented,
  non-blocking).** The engine's declarative read/check paths (`engine/read.go`) only ever decode a
  JSON response body; there is no XML-body dialect equivalent to `records.path`/`RecordsAt`. This
  was the pre-identified Tier-2 trigger for this connector (conventions.md §1's Tier-2 table
  explicitly lists "multipart/XML bodies... response decompression/CSV parsing" as legitimate
  StreamHook triggers), not a blocker: `hooks/rss/hooks.go` implements the real XML decode entirely
  within the sanctioned hook seam, reusing `rt.Requester` exactly as the declarative path itself
  would for a JSON API. If 3+ connectors need XML-body decoding, the ENGINE_GAP recurrence rule
  (conventions.md §6) promotes this to a real engine feature; monday's POST-body GraphQL gap and
  this XML gap are different shapes of the same underlying limitation (the declarative path only
  ever sends no body and only ever decodes JSON) and do not obviously combine into one fix, so they
  are tracked as separate occurrences.
- **The declarative `streams.json` path is never live-dispatched** (see "Declarative path" above)
  — the bundle-level `conformance.skip_dynamic` marker names
  `hooks/rss/hooks_test.go`/`paritytest/rss/parity_test.go` as the authoritative substitute;
  conformance's dynamic (fixture replay) checks Skip entirely rather than exercising a JSON-shaped
  replay that could never match rss's real XML wire format.
- **No incremental filtering, matching legacy exactly.** `published_at` is published as
  `x-cursor-field` for manifest-surface parity, but legacy never filters or advances reads by it;
  every read is a full feed read.
- **Legacy's `mode: fixture` credential-free affordance is NOT part of this bundle.** Legacy's
  `readFixture`/`fixtureMode` (`rss.go:154-167`) emit synthetic records without any network call when
  `config.mode == "fixture"` — this is a legacy-only testing convenience; parity is asserted against
  legacy's LIVE (httptest-driven) read path only, matching the wave1-pilot convention (monday's
  identical note).
- **`base_url` alias fallback is not modeled.** Legacy's `feedURL` helper checks `feed_url`, then
  `base_url`, then the default `https://xkcd.com/rss.xml`. The declarative `base.url` field can
  reference only one resolved config value, so this bundle declares `feed_url` with the legacy
  default and intentionally does not declare an ignored `base_url` alias.
