# Overview

SpaceX API is a wave2 fan-out migration from `internal/connectors/spacex-api` (the legacy
hand-written connector this bundle replaces at capability parity), expanded to full documented
surface in Pass B. It reads public SpaceX launch, rocket, core, capsule, crew, Dragon capsule
type, history event, payload, Starlink, launchpad, landpad, ship, roadster, and company data
through the public SpaceX API v4/v5. Read-only, no-auth, matching the searxng golden's read-only
no-auth pattern.

## Auth setup

None. The SpaceX API is public and requires no credentials — `streams.json` `base.auth` declares
`[{"mode":"none"}]` unconditionally, matching legacy's `requester` (no `Auth` set on the
`connsdk.Requester`).

## Streams notes

All 13 streams (`launches`, `rockets`, `capsules`, `cores`, `crew`, `dragons`, `history`,
`payloads`, `starlink`, `launchpads`, `landpads`, `ships`, `roadster`, `company` — 11 legacy-parity
plus `cores`/`dragons`/`history` newly added in Pass B full-surface expansion) are simple `GET`
reads with no pagination (this API's list endpoints return every record in one response, matching
legacy's `readRecords` for its own 11 streams — no page-advance loop) and no query parameters at
all — matching legacy's `readRecords(ctx, r, resource, endpoint.recordsPath, emit)` call, which
passes a nil `url.Values`. Every stream's `records.path` is `"."` (body root): every stream except
`roadster` and `company` returns a top-level JSON array; `roadster` and `company` return a
top-level JSON object (`single_object: true`, both singleton resources with no list wrapper).
`cores`/`dragons`/`history` follow the identical shape as every other list stream (`GET`, no query
params, `records.path: "."`, `passthrough` projection) — see `docs/{cores,dragons,history}/v4/
all.md` in the upstream docs tree.

All 13 streams declare `"projection": "passthrough"`. Legacy's `Read` emits the raw API record
verbatim (`emit(connectors.Record(rec))`, `spacex_api.go:127`, inside `readRecords`) with no
field-building/filtering step for its own 11 streams — `streams()`'s three-field `Fields` list
(`spacex_api.go:109`) is consumed only by `Catalog`, never by `Read`; the 3 newly added streams
follow the same passthrough convention for consistency with every sibling stream in this bundle.
Every real SpaceX API field beyond each schema's narrow `id`/`name`/`date_utc`-shaped properties
(e.g. `launches`' `rocket`/`success`/`links`/`fairings`, `rockets`' `height`/`mass`/`engines`,
`cores`' `block`/`reuse_count`/`rtls_attempts`, `dragons`' `heat_shield`/`thrusters`, `history`'s
`links.article`) survives to the emitted record exactly as the live API would return it. Declaring
the default `"schema"` projection mode here would silently narrow every emitted record to the
schema's declared properties — so `passthrough` is required, matching conventions.md §8 rule 1
(legacy's raw `emit(record)` with no `mapRecord` field-building is the mechanical signal to use
`passthrough`).

## Write actions & risks

None. SpaceX API is read-only in both legacy and this bundle (`capabilities.write: false`, no
`writes.json`).

## Known limits

- Legacy also supports two ad hoc request-shaping config knobs not modeled here: an `options`
  config value that redirects the `launches` stream to a `launches/<option>` sub-resource (e.g.
  `launches/upcoming`, `launches/past`), and an `id` config value that appends `/<id>` to any
  stream's resource path for a single-record-by-id lookup. Neither is expressible in this
  dialect: `streams.json`'s `path` templating has no absent-key-falsy tolerance for path segments
  (unlike `stream.Query`'s opt-in `omit_when_absent` object form) — conditionally appending a path
  segment only when a config value happens to be set would require the key to always resolve,
  which defeats the "optional" behavior legacy provides. Both knobs are intentionally not
  declared in `spec.json` (a declared-but-unwireable key is worse than an absent one, per the
  searxng `subreddit` precedent) — the full-list `launches` stream already contains every record
  the `/v4/launches/{id}` and `options`-scoped sub-resource endpoints would return, so no read
  capability is lost, only the convenience of narrowing the request server-side. See
  `api_surface.json`'s `duplicate_of` entries for `/v5/launches/{id}`, `/v5/launches/latest`,
  `/v5/launches/next`, `/v5/launches/upcoming`, `/v5/launches/past`.
- **`launches` stays on the legacy `/v4/launches` route, not the current documented `/v5/launches`
  route.** The upstream project moved `launches` to a `v5` path (`docs/launches/v5/README.md`: the
  only documented v4→v5 change is `crew` becoming an array of per-member objects instead of a
  bare-string-id array — every other field is identical); this bundle intentionally keeps the `v4`
  path unchanged from the pre-existing legacy connector's own behavior (`config.base_url` defaults
  to `.../v4`, and `streams.json`'s `launches.path` is the relative `/launches`, resolved under
  that base) to avoid an unrequested behavior change to the emitted `crew` field's shape. Both
  routes are recorded in `api_surface.json` (`/v4/launches` `covered_by` the stream, `/v5/launches`
  `excluded: duplicate_of`) so the newer route is not silently missing from the surface record.
  Migrating to `v5` (and widening `crew`'s schema type to match the new object-array shape) is a
  deliberate, separately-reviewable follow-up, not a Pass B surface-coverage gap.
- Every resource's `POST /<resource>/query` MongoDB-query-DSL search endpoint (arbitrary
  `{query, options}` filter/sort/pagination body) is out of scope: it is a read operation (not a
  mutation, so it cannot become a `writes.json` action) whose body has no fixed field set this
  dialect's templated `stream.query` can express, and the corresponding list stream already
  returns the full population any query would filter a subset of. See `api_surface.json`.
- The live `api.spacexdata.com` origin was unreachable at the time of this Pass B review
  (Cloudflare 525 SSL-handshake-failed on every route) — the expanded surface (`cores`, `dragons`,
  `history`) and their fixtures are sourced from the still-published GitHub docs tree and
  `models/*.js` Mongoose schema source, not a live probe, and use synthetic fixture data
  throughout.
