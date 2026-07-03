# Overview

NASA Open APIs expose public astronomy and space data. This bundle reads Astronomy Picture of the
Day (`apod`), Near-Earth Objects (`neo_feed`, `neo_browse`), EPIC Earth imagery (`epic`), and Mars
rover photos (`mars_photos`) through `api.nasa.gov`. It is read-only, migrating
`internal/connectors/nasa` (the hand-written legacy connector, which stays registered and
unchanged until wave6's registry flip) at capability parity.

## Auth setup

Provide a NASA Open API key via the `api_key` secret; it is sent as the `api_key` query parameter
on every request and is never logged, matching legacy's `connsdk.APIKeyQuery("api_key", secret)`
wiring.

## Streams notes

- `apod` (`GET /planetary/apod`) returns a single top-level JSON object; `records.path: ""`
  treats that whole object as one record, matching legacy's `arrayPath: "."` behavior via the
  identical `connsdk.RecordsAt` semantics the engine reuses. Optional `start_date`/`end_date`/
  `count`/`thumbs` config values are sent only when set (`omit_when_absent`), matching legacy's
  `streamQuery` conditional query building exactly.
- `neo_feed` (`GET /neo/rest/v1/feed`) reads `records.path: "near_earth_objects"`. NASA's real
  wire shape nests NEOs under **date keys** (`{"near_earth_objects": {"2026-01-01": [...], ...}}`),
  not a flat array — `connsdk.RecordsAt` (used identically by both legacy and this bundle) treats
  a `map[string]any` node as a SINGLE pseudo-record rather than iterating into the nested arrays,
  so `neo_feed` degenerates to emitting exactly one record per read whose only key is a date
  string; every schema-declared field (`id`, `name`, etc.) is absent from that pseudo-record. This
  is legacy's own real, unmodified behavior (`nasa.go`'s `neoRecord`/`RecordsAt` combination), not
  a bug introduced by this migration — see Known limits.
- `neo_browse` (`GET /neo/rest/v1/neo/browse`) is the one paginated stream: NeoWs's browse
  endpoint returns a flat array under `near_earth_objects` (unlike `feed`'s date-keyed map) and
  paginates with `page.number`/`page.total_pages` metadata. This bundle models it as
  `pagination.type: page_number` (`page_param: "page"`, `size_param: ""` — no size query param is
  ever sent, matching legacy exactly, `start_page: 0`, `page_size: 20`) rather than reading
  `page.total_pages` directly: NeoWs's browse endpoint's own real page size is a fixed 20
  (undocumented as configurable), so every page except the last is exactly 20 records long — the
  engine's short-page (fewer than `page_size`) stop condition is behaviorally equivalent to
  legacy's `page+1 >= total_pages` check for every real NeoWs browse response, without needing a
  7th pagination type for a `page.total_pages`-driven stop signal. `pagination.max_pages: 5` caps
  the stream at 5 pages by default, matching legacy's `nasaDefaultMaxPages` — see Known limits for
  why this is a static bundle-authored bound rather than a config-driven one.
- `epic` (`GET /EPIC/api/natural`) returns a top-level JSON array; `records.path: ""` iterates it
  directly (the `[]any` branch of `connsdk.RecordsAt`), matching legacy's `arrayPath: "."`.
- `mars_photos` (`GET /mars-photos/api/v1/rovers/curiosity/photos`) sends `sol` (default `1000`,
  matching legacy's `nasaStreamSpecs["mars_photos"]` fallback) and flattens the nested
  `camera`/`rover` objects to their human-readable names via `computed_fields`
  (`{{ record.camera.full_name }}`, `{{ record.rover.name }}`) — see Known limits for the one
  `camera` fallback branch this does not model.

## Write actions & risks

None. The NASA Open APIs are read-only in this bundle (`capabilities.write: false`), matching
legacy's `Write` stub that always returns `connectors.ErrUnsupportedOperation`.

## Known limits

- **`neo_feed` emits a single degenerate pseudo-record per read, matching legacy exactly.**
  `neo_feed.json`'s schema omits `id` from `required` (unlike `neo_browse`, whose flat-array
  records genuinely carry `id`) because the real, correctly-reproduced emitted record for this
  stream never has an `id` key at all — this is not a schema-authoring shortcut, it is the
  honest reflection of legacy's own `RecordsAt`-on-a-map behavior for this specific endpoint
  shape. A future capability-expansion pass could properly flatten the date-keyed structure (would
  require a new dialect primitive: iterate values of a nested object-of-arrays rather than a
  single array), but that is a real behavior CHANGE from legacy, not a parity port, so it is out of
  scope here.
- **`mars_photos`'s `camera` field only models legacy's PRIMARY fallback branch.** Legacy's
  `nestedName(item["camera"], "full_name", "name")` tries `full_name` first, falling back to
  `name` if absent. `computed_fields` only supports a single template reference per field, with no
  multi-key fallback chain, so this bundle models only the `full_name` branch
  (`{{ record.camera.full_name }}`) — in practice NASA's Mars Photos API always includes both
  `camera.full_name` and `camera.name`, so the fallback branch is not reachable against the real
  API; documented here for completeness per the parity-deviation ledger.
- Full NASA Open API surface (DONKI space weather, Earth Landsat imagery, exoplanet archive, other
  rover cameras/rovers, TechPort, sounds) is out of scope for this wave; see `api_surface.json`'s
  `excluded` entries. Only the 5 legacy-parity streams are implemented.
- **`neo_browse`'s page cap is a static 5-page bound, not a runtime-configurable one.** Legacy's
  `nasaMaxPages` reads a `max_pages` config value (accepting an integer, `all`, or `unlimited`)
  and falls back to `nasaDefaultMaxPages = 5` when unset, explicitly "so a sync cannot run
  unbounded against the (large) NeoWs dataset." The engine's `PaginationSpec.max_pages`
  (`streams.json`'s `neo_browse.pagination.max_pages: 5`) is a static bundle-authored int, never
  templated against `config.*` — there is no dialect mechanism to make it operator-overridable per
  sync. This bundle therefore reproduces legacy's DEFAULT behavior exactly (bounded at 5 pages) but
  narrows legacy's accepted config surface: an operator can no longer opt into `all`/`unlimited`/a
  different integer at runtime. A `max_pages` config property is not declared in `spec.json` for
  this reason — a declared key with no template anywhere to consume it would be dead config (see
  searxng's identical `max_pages`-removal precedent). Widening this would require a
  `stream.pagination`-level templated-max-pages primitive; out of scope here as a genuine engine
  gap, not silently worked around.
