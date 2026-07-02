# Overview

Pretix is a wave2 fan-out declarative-HTTP migration. It reads pretix organizers, events, items,
and orders through the pretix REST API v1 (`GET https://pretix.eu/api/v1/...`, or a self-hosted
instance's equivalent base URL). This bundle targets capability parity with
`internal/connectors/pretix` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a pretix API token via the `api_token` secret. Legacy sends it as
`Authorization: Token <api_token>` (`pretix.go:166`, `connsdk.APIKeyHeader("Authorization", token,
"Token ")`) — this bundle uses `api_key_header` mode (not `bearer`, which the engine hard-codes to
a `Bearer ` prefix) with an explicit `"prefix": "Token "` to reproduce the exact header value.
`base_url` defaults to `https://pretix.eu/api/v1` and may be overridden for self-hosted pretix
instances, tests, or proxies, matching legacy's own `baseURL` scheme+host validation (the engine
has no equivalent runtime validation, but every fixture/conformance path only ever points at a
replay server, so this is not exercised differently on either side).

## Streams notes

`organizers` is a top-level list endpoint (`GET /organizers/`). `events`, `items`, and `orders` are
scoped beneath a required `organizer` config value (events/items/orders) and a required `event`
config value (items/orders only), matching legacy's `resourcePath` path templates exactly
(`pretix.go:203-219`) — an absent `organizer`/`event` fails the same way on both sides (legacy:
"pretix connector requires config organizer/event for this stream"; engine: an unresolved
`config.organizer`/`config.event` path-template key — same failure classification, different
literal text, per `conventions.md` §5's precedent for config-validation parity).

Every stream reads records from the `results` array and paginates via pretix's own `next` absolute
URL field (`pagination.type: next_url`, `next_url_path: "next"`). The first request sends
`page_size=100` (legacy's `defaultPageSize`); `page` itself is deliberately NOT declared as a
static per-stream query value (unlike `page_size`), because the engine re-applies every
`stream.Query` entry on EVERY page request (`read.go`'s `mergeQuery`), including when following an
absolute `next_url` — `resolveURL`'s query merge REPLACES (not adds to) any same-named param
already present on the URL (`Del` then `Add`). A static `page: "1"` would therefore silently force
every subsequent page's URL back to `page=1`, an actual pagination-breaking bug (an infinite loop
on page 1), not a benign idempotent re-send like `page_size` (whose value never changes across
pages). Pretix's Django-REST-Framework-style pagination defaults to `page=1` when the param is
omitted entirely, so omitting it reproduces legacy's exact first request while staying safe on
every subsequent page — subsequent requests follow the recorded absolute `next` URL verbatim (which
already carries pretix's own correct page number), with `page_size=100` harmlessly re-applied
(same value on every page, so the replace is a no-op in practice), matching legacy's own `first`
flag intent of "send page/page_size once" for any real pretix response's next-URL page numbering.
Pretix's own docs confirm `next`/`previous` are always fully-qualified absolute URLs in every list
response, so the engine's same-host SSRF guard on `next_url` pagination passes cleanly in
production.

Legacy's `mapRecord` (`pretix.go:179-189`) applies a generic multi-key fallback chain
(`first(item, "slug", "code", "id", "ID")` for `id`; `first(item, "created", "created_at")` for
`created_at`; `first(item, "modified", "updated_at", "date_from")` for `updated_at`) shared across
all 4 streams. The engine's `computed_fields` dialect has no multi-key coalesce/fallback filter, so
this bundle expresses the SAME resulting values per stream directly, using each stream's actual
real pretix wire shape to determine which single key legacy's fallback chain would land on:
- `organizers`/`events`: pretix's real API never returns `slug`+`code`+`id` together — only
  `slug` — so `id: "{{ record.slug }}"` reproduces legacy's fallback exactly for real records.
- `orders`: pretix's real order objects carry only `code` (no `slug`/`id`) — `id: "{{ record.code
  }}"` reproduces legacy's fallback exactly.
- `items`: pretix's real item objects carry only `id` (no `slug`/`code`) — plain schema
  projection (no computed_fields override needed) copies `id` through directly, matching legacy's
  fallback landing on its 3rd candidate.

**`created_at` is never modeled and always resolves absent, matching legacy exactly.** Pretix's
real API responses for every one of these 4 resources (confirmed against pretix's own
documentation) never emit a `created` or `created_at` field at all — legacy's `first(item,
"created", "created_at")` therefore always returns `nil` in production, despite being present in
`mapRecord`'s shape. This bundle does not declare a `created_at` computed_fields entry or schema
property for any stream (an absent-source computed field is silently skipped per `conventions.md`
§3, so declaring one that can never resolve would add no value); this is data-identical to
legacy's always-nil field, not a narrower parity — every consumer preserving the DATA sees no
`created_at` value on either side.

**`updated_at` is modeled ONLY for `events`, matching legacy's actual (not advertised) behavior.**
Legacy's `updated_at` fallback chain checks `modified`, then `updated_at`, then `date_from`.
Pretix's real wire shapes: `events` carries `date_from` (the chain's 3rd candidate resolves) —
modeled here as `"updated_at": "{{ record.date_from }}"`. `organizers` and `items` carry none of
`modified`/`updated_at`/`date_from` — the field always resolves absent on both sides, so it is not
declared. **`orders` carries `last_modified`, which is NOT one of legacy's 3 checked keys** — this
is a real legacy gap (the fallback chain omits pretix's actual field name for orders), so legacy's
`updated_at` for orders is always `nil` in production despite the stream's catalog metadata
advertising `CursorFields: []string{"updated_at"}` (`pretix.go:175`). Per the meta-rule (never
diverge from legacy's actual accepted-input behavior), this bundle does NOT map `last_modified` to
`updated_at` for orders — doing so would be MORE correct than legacy, which is exactly the kind of
silent behavior change the migration forbids. `orders`' schema declares no `x-cursor-field` and
this bundle declares no `incremental` block for it (or any stream): legacy's `Read` implements no
incremental filtering logic whatsoever (no state-based cursor request param, no client-side
filtering) despite the catalog advertising cursor fields — every read is a full-refresh scan on
both sides.

**Pagination fixtures are single-page**, per `conventions.md` §4's sanctioned `next_url` exception:
every one of pretix's 4 streams uses `next_url` pagination, whose next-page URL is the replay
server's own runtime-assigned address — a static fixture file cannot embed it. `pagination_terminates`
exercises the first stream (`organizers`) against its single fixture page and confirms exactly one
request is served. A live 2-page proof (the exception's recommended `paritytest/pretix`-style test)
is out of scope for this wave (hard rule: JSON + docs.md only, no Go/paritytest packages) — see
Known limits.

## Write actions & risks

None. Pretix's endpoints have no reverse-ETL writes modeled by legacy (`Write` returns
`connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **No live 2-page pagination proof for this wave.** As documented above, all 4 streams use
  `next_url` pagination with single-page fixtures (the sanctioned `conventions.md` §4 exception).
  The exception's recommended live `paritytest/pretix`-style test (driving a real `httptest.Server`
  through 2 pages) is out of scope under this wave's hard rule prohibiting new Go/paritytest
  packages; a future wave should add one to fully prove multi-page `next` URL following, mirroring
  legacy's own `pretix_test.go`'s `TestReadEventsAuthenticatesPaginatesAndMaps` 2-page fixture.
- **`created_at` is not modeled for any stream.** Confirmed data-identical to legacy (which always
  emits `nil` for this key against real pretix API responses across all 4 resources) — see Streams
  notes above.
- **`updated_at` is modeled only for `events`.** `organizers`/`items` never resolve it (matching
  legacy's always-nil behavior); `orders`' real `last_modified` field is deliberately NOT mapped to
  `updated_at` because legacy's own fallback chain never checks that key name — mapping it would
  silently diverge from legacy's actual (buggy) accepted behavior. See Streams notes above for the
  full reasoning.
- **Legacy's `raw` escape-hatch field is not modeled.** `mapRecord` stamps a full copy of the
  source item onto every record under the key `raw` (`pretix.go:187`). The engine's
  `computed_fields` dialect has no whole-record reference primitive (only dotted `record.<path>`
  paths are addressable, never the bare record itself), so this cannot be expressed. This is a
  scope-narrowing, not a data change to any of the other named fields.
- **Legacy's fixture-mode-only synthetic records are not modeled.** Legacy's `readFixture` path
  (only reached when `config.mode == "fixture"`) emits synthetic records with slightly different
  field shapes (e.g. an object-valued `name` on organizers/events) than the live path; this bundle
  targets the live `harvest`/`mapRecord` path only, matching every other wave1/wave2 bundle's
  documented precedent for ignoring legacy's fixture-mode-only behavior.
