# Overview

Imagga is a wave2 fan-out declarative-HTTP migration, **PARTIAL** (gap-closure pass): the
account-scoped `usage` stream plus 2 of legacy's 4 per-image detection streams (`tags`,
`categories`) are now ported at Tier 1, using the S4 engine mini-wave's `stream.fan_out` primitive
(`docs/migration/conventions.md` §3) to express legacy's client-side, config-driven "one request
per configured image URL" fan-out (`imaggaImages`, `imagga.go:243-261`, iterated in `Read`,
`imagga.go:126-135`). The remaining 2 detection streams (`colors`, `faces_detections`) are each
blocked by a SEPARATE, narrower gap that `fan_out` does not close — see Known limits. The legacy
connector at `internal/connectors/imagga` (`imagga.go`/`streams.go`) stays authoritative for those
2 streams until a follow-up wave closes the remaining gaps.

## Auth setup

Provide `api_key` and `api_secret` secrets; they are sent as HTTP Basic auth (`api_key` as
username, `api_secret` as password — `streams.json`'s `base.auth`, `mode: basic`), matching
legacy's `connsdk.Basic(key, secret)` (`imagga.go:228`). `base_url` defaults to
`https://api.imagga.com/v2` and may be overridden for tests/proxies (legacy's own `imaggaBaseURL`
validates scheme+host the same way; the engine's base-URL resolution has no equivalent runtime
validation, but every parity/conformance fixture only ever points at an httptest server, so this
is not exercised differently on either side).

## Streams notes

`usage` (`GET /usage`) returns a single object at the `result` key (`records.path: "result"`,
`single_object: true`), matching legacy's `usageRecords` mapper (`streams.go:221-233`) which
flattens the account-scoped result object into one record. This bundle uses typed
`computed_fields` (bare `{{ record.<path> }}` references) to rename `result.total` → `requests`
and copy `monthly_processed`/`monthly_limit`/`daily_processed` verbatim, preserving their native
integer type (per conventions.md's typed-extraction rule); the static-literal `period: "current"`
computed field stamps legacy's synthetic single-row marker (legacy's `usageRecords` hardcodes
`"period": "current"` identically, `streams.go:227`). `usage` is full-refresh only, matching
legacy (no `CursorFields` published for this stream in `streams.go:65-69`).

`tags` (`GET /tags`) and `categories` (`GET /categories/personal_photos`) are per-image detection
endpoints: legacy issues one request per configured image URL and fans each response's nested
result array out into records (`readOne`+`tagsRecords`/`categoriesRecords`,
`imagga.go:141-168`/`streams.go:124-158`). This is now expressed via `fan_out`:
`ids_from.config_key: "image_urls"` splits the comma-separated `image_urls` config value
(matching legacy's `imaggaImages` primary path) into one id per configured image;
`into.query_param: "image_url"` adds each resolved image URL as the `image_url` query parameter
on that sub-sequence's single request (matching legacy's `readOne`'s `query.Set("image_url",
image)`); `stamp_field: "image_url"` writes the same URL onto every emitted record (matching
legacy's `imageURL` parameter threaded through both mappers). `records.path: "result.tags"` /
`"result.categories"` selects the nested array directly (an ordinary dotted-path array
extraction — `connsdk.RecordsAt` fans it into one record per element with no special-casing
needed). Each raw tag/category's localized name is flattened via `computed_fields` with
`coalesce` (`record.tag.en` then `record.tag`, and `record.name.en` then `record.name`), matching
legacy's `localizedName` helper (`streams.go:206-215`) for both the localized object form and the
plain-string fallback. `confidence` uses a bare `{{ record.confidence }}` reference (typed
extraction, preserving its native number type). Neither stream is incremental, matching legacy (no
`CursorFields` for either).

`image_urls`'s `spec.json` `"default"` is Imagga's own sample image
(`https://imagga.com/static/images/categorization/child-476506_640.jpg`, matching legacy's
`imaggaDefaultImage`, `imagga.go:33`) — when the caller omits `image_urls` entirely, the engine's
`spec.json` default-materialization (conventions.md §3) fills it in before `fan_out` resolves ids,
reproducing legacy's "no image configured → analyze the sample image" fallback for that case. See
Known limits for the one input shape this does NOT reproduce (legacy's SECOND fallback key,
`img_for_detection`).

## Write actions & risks

None. Imagga is a read-only image-analysis API with no reverse-ETL surface (legacy's own package
doc: "Imagga is a read-only analysis API with no reverse-ETL surface"); `capabilities.write` is
`false` and this bundle ships no `writes.json`, matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation`.

## Known limits

- **`colors` is NOT ported (`ENGINE_GAP`).** Legacy's `colorsRecords` (`streams.go:160-192`) reads
  THREE sibling arrays nested under `result.colors` (`overall_colors`, `foreground_colors`,
  `background_colors`) and unions them into one record stream, tagging each with its
  `color_scope` label. The dialect's `records.path` extraction (`connsdk.RecordsAt`) selects
  exactly ONE dotted path per stream and fans that ONE array (or object) into records — there is
  no primitive to union multiple sibling arrays from the same page body into a single record
  stream, and `keyed_object` (S4 engine mini-wave item 3) explodes ONE object's VALUES, not
  several named array fields. Splitting `colors` into 3 separate streams (one per scope) was
  considered and rejected: it would change the stream catalog shape callers see (3 new stream
  names replacing 1), a bigger accepted-input/catalog-shape change than the meta-rule's parity bar
  allows for a same-named migrated stream. This is a genuine Tier-2 `StreamHook` trigger (or a new
  `records` dialect primitive), out of scope for this Tier-1-only gap-closure pass. Target parity
  contract for a follow-up hook: primary key `["image_url", "html_code", "color_scope"]`, fields
  `image_url`/`color_scope`/`html_code`/`closest_palette_color`/`percent`/`r`/`g`/`b`
  (`streams.go:89-100`).
- **`faces_detections` is NOT ported (`ENGINE_GAP`).** Legacy's `facesRecords`
  (`streams.go:194-218`) derives `face_index` — a component of the stream's primary key
  (`["image_url", "face_index"]`) — from the record's 0-based POSITION within the raw
  `result.faces[]` array (`for i, raw := range faces`), not from any field value on the record
  itself. `computed_fields`/`records` extraction resolve every field from the record's own
  content (or a config/fan-out reference); neither has an array-index accessor, and no other field
  on a raw Imagga face object (`confidence`, `coordinates`) is a usable stable identifier on its
  own (two faces can share identical coordinates in principle). Fabricating an alternative key
  (e.g. hashing `coordinates`) would silently diverge from legacy's actual PK derivation for any
  input with duplicate coordinate sets — the meta-rule (conventions.md §5) forbids this as a
  silent behavior change. This is a genuine Tier-2 `StreamHook` trigger (or a new dialect
  primitive exposing array position), out of scope for this Tier-1-only gap-closure pass. Target
  parity contract for a follow-up hook: primary key `["image_url", "face_index"]`, fields
  `image_url`/`face_index`/`confidence`/`x1`/`y1`/`x2`/`y2` (`streams.go:102-112`).
- **`img_for_detection`'s single-image fallback config key is not modeled (ACCEPTABLE, documented
  scope narrowing).** Legacy's `imaggaImages` (`imagga.go:243-261`) checks THREE sources in order:
  `image_urls` (comma-separated, primary), then `img_for_detection` (a single-image fallback key),
  then a hardcoded sample image. `fan_out.ids_from.config_key` reads exactly one named config key
  with no fallback chain to a second key. This bundle declares `image_urls` only (legacy's
  primary, first-checked key) plus its `spec.json` `"default"` (reproducing legacy's THIRD
  fallback, the hardcoded sample image, for the "nothing configured at all" case — see Streams
  notes). A caller who previously configured Imagga via `img_for_detection` alone (never setting
  `image_urls`) would need to migrate that config key name to `image_urls`; this is a documented
  config-surface narrowing, not a data-shape change for any caller using the canonical
  `image_urls` key or relying on the no-config default.
- **An explicit empty-string `image_urls` does not fall back to the default sample image
  (ACCEPTABLE, documented deviation).** The engine's `spec.json` default-materialization only
  fills a key that is genuinely ABSENT from `RuntimeConfig.Config`; an explicit empty string is
  preserved as-is, never overridden (conventions.md §3's `default` semantics), whereas legacy's
  `imaggaImages` treats a whitespace-only `image_urls` the same as an absent one and falls through
  to `img_for_detection`/the sample image. A caller who explicitly sets `image_urls: ""` (rather
  than omitting the key) sees zero fan-out ids (and therefore zero emitted records) here, versus
  legacy's sample-image fallback — a narrow corner case, not the common "key omitted entirely"
  path this bundle does reproduce exactly.
- **Legacy's fixture-mode-only `fixture: true` marker is not modeled.** Legacy's `readFixture` path
  (only reached when `config.mode == "fixture"`) stamps a `fixture: true` field onto every
  fixture-mode record (`imagga.go:205`). This bundle's schemas and fixtures target the live record
  shape only; the engine's own conformance/fixture-replay harness supplies the credential-free
  test affordance legacy's fixture mode existed for.
