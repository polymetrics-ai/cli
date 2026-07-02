# Overview

Imagga is a wave2 fan-out declarative-HTTP migration, **PARTIAL**: only the account-scoped
`usage` stream is ported at Tier 1. Legacy's 4 per-image detection streams (`tags`, `categories`,
`colors`, `faces_detections`) each issue **one HTTP request per configured image URL**
(`imaggaImages`, `imagga.go:243-261`, iterated in `Read`, `imagga.go:126-135`) — a client-side,
config-driven multi-request fan-out the declarative engine dialect has no mechanism to express
(pagination follows a server-provided next-page signal; there is no "iterate this stream once per
comma-separated config value" primitive). This is documented as an `ENGINE_GAP` blocker below and
in `api_surface.json`; the legacy connector at `internal/connectors/imagga`
(`imagga.go`/`streams.go`) stays authoritative for those 4 streams until a Tier-2 `StreamHook`
closes the gap in a follow-up wave. The `usage` stream itself has no such fan-out (a single
account-scoped request, no image), so it is fully ported here.

## Auth setup

Provide `api_key` and `api_secret` secrets; they are sent as HTTP Basic auth (`api_key` as
username, `api_secret` as password — `streams.json`'s `base.auth`, `mode: basic`), matching
legacy's `connsdk.Basic(key, secret)` (`imagga.go:228`). `base_url` defaults to
`https://api.imagga.com/v2` and may be overridden for tests/proxies (legacy's own `imaggaBaseURL`
validates scheme+host the same way; the engine's base-URL resolution has no equivalent runtime
validation, but every parity/conformance fixture only ever points at an httptest server, so this
is not exercised differently on either side).

## Streams notes

`usage` is the only ported stream: `GET /usage` returns a single object at the `result` key
(`records.path: "result"`, `single_object: true`), matching legacy's `usageRecords` mapper
(`streams.go:221-233`) which flattens the account-scoped result object into one record. This
bundle uses typed `computed_fields` (bare `{{ record.<path> }}` references) to rename
`result.total` → `requests` and copy `monthly_processed`/`monthly_limit`/`daily_processed`
verbatim, preserving their native integer type (per conventions.md's typed-extraction rule); the
static-literal `period: "current"` computed field stamps legacy's synthetic single-row marker
(legacy's `usageRecords` hardcodes `"period": "current"` identically, `streams.go:227`). `usage`
is full-refresh only, matching legacy (no `CursorFields` published for this stream in
`streams.go:65-69`).

## Write actions & risks

None. Imagga is a read-only image-analysis API with no reverse-ETL surface (legacy's own package
doc: "Imagga is a read-only analysis API with no reverse-ETL surface"); `capabilities.write` is
`false` and this bundle ships no `writes.json`, matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation`.

## Known limits

- **`ENGINE_GAP` — the 4 per-image detection streams (`tags`, `categories`, `colors`,
  `faces_detections`) are NOT ported.** Legacy's `Read` iterates every configured image URL
  (`imaggaImages`, sourced from the comma-separated `image_urls` config, the single
  `img_for_detection` config, or a hardcoded sample-image fallback, `imagga.go:243-261`) and issues
  one request per image (`readOne`, `imagga.go:141-168`), fanning each response's nested result
  array (`result.tags`/`result.categories`/`result.colors[_scope]`/`result.faces`) out into
  records stamped with that image's URL. The declarative dialect's `records.path` extraction
  (`connsdk.RecordsAt`) DOES support fanning ONE response's array into multiple records (this
  bundle's own `usage` stream and every other Tier-1 goldens' list streams rely on exactly this),
  so the nested-array-flattening half of legacy's behavior is not the blocker. The blocker is
  strictly the OUTER fan-out: no `PaginationSpec` type follows a client-side, config-driven list of
  distinct request parameter values (pagination follows a server-emitted next-page signal —
  `token_path`/`last_record_field`/`next_url`/`Link` header/offset — never a config value split on
  a delimiter), and `stream.Query`'s optional-query dialect (conventions.md §3) has no "repeat this
  request once per list element" primitive either. A single-fixed-image workaround (declaring one
  static `image_url` config value and reading it as an ordinary single-request stream) was
  considered and rejected: it would silently drop legacy's accepted multi-image `image_urls`
  config input to a single image, an accepted-input-behavior change the conventions.md §5 meta-rule
  forbids — this is exactly the `StreamHook` interface's documented purpose ("sub-resource fan-out
  reads", conventions.md §1's Tier-2 table), so it is reported here as a blocker rather than
  faked. A follow-up Tier-2 wave should add `internal/connectors/hooks/imagga/hooks.go` implementing
  `StreamHook.ReadStream` for these 4 streams, iterating the configured image list and emitting
  each response's flattened records — the schema/PK shape below is the target parity contract for
  that hook to satisfy.
  - `tags`: primary key `["image_url", "tag"]`, fields `image_url`/`tag`/`confidence`
    (`streams.go:73-79`, `tagsRecords`, `streams.go:124-140`).
  - `categories`: primary key `["image_url", "category"]`, fields
    `image_url`/`category`/`confidence` (`streams.go:81-87`, `categoriesRecords`,
    `streams.go:142-158`).
  - `colors`: primary key `["image_url", "html_code", "color_scope"]`, fields
    `image_url`/`color_scope`/`html_code`/`closest_palette_color`/`percent`/`r`/`g`/`b`
    (`streams.go:89-100`, `colorsRecords`, `streams.go:160-192` — three color-scope groups:
    `overall`/`foreground`/`background`).
  - `faces_detections`: primary key `["image_url", "face_index"]`, fields
    `image_url`/`face_index`/`confidence`/`x1`/`y1`/`x2`/`y2` (`streams.go:102-112`,
    `facesRecords`, `streams.go:194-218`).
- **Legacy's fixture-mode-only `fixture: true` marker is not modeled.** Legacy's `readFixture` path
  (only reached when `config.mode == "fixture"`) stamps a `fixture: true` field onto every
  fixture-mode record (`imagga.go:205`). This bundle's `usage` schema and fixtures target the live
  record shape only; the engine's own conformance/fixture-replay harness supplies the
  credential-free test affordance legacy's fixture mode existed for.
