# Overview

TVmaze Schedule is a wave2 fan-out read-only, credential-free declarative-HTTP migration. It reads
public TVmaze broadcast and web schedules through the TVmaze API
(`GET https://api.tvmaze.com/schedule` / `GET https://api.tvmaze.com/web/schedule`). This bundle is
capability-parity migrated from `internal/connectors/tvmaze-schedule` (the hand-written connector
it migrates); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

No credentials are required: TVmaze's schedule endpoints are a fully public API, matching legacy
(no `secrets`/`auth` reference anywhere in `trustpilot.go`'s TVmaze sibling). `spec.json` declares
no `x-secret` field and `streams.json`'s `base` has no `auth` block at all (mode defaults to
`none`).

## Streams notes

Both streams (`schedule`, `web_schedule`) hit their respective TVmaze endpoint with two optional
query parameters, `country` and `date`, matching legacy's conditional `query.Set` calls
(`tvmaze_schedule.go:76-82`, only set when non-empty). Both are declared with the opt-in
`omit_when_absent` query-object form so a caller who leaves them unset gets the identical
"send neither param, let TVmaze apply its own defaults" behavior legacy has — TVmaze's own
documented behavior is to default `country` to the request's inferred geolocation and `date` to
today. Records are extracted from the bare top-level JSON array (`records.path: "."` against an
array-shaped body, exactly like legacy's `connsdk.RecordsAt(resp.Body, ".")`). `show_id`/`show_name`
are derived via `computed_fields` reaching into the raw nested `show` object
(`{{ record.show.id }}` / `{{ record.show.name }}`), matching legacy's `episodeRecord`'s
`show, _ := item["show"].(map[string]any)` destructure; the bare-single-reference shape preserves
`show.id`'s native integer type (typed extraction), matching the schema's `"integer"` type.

Primary key is `["id"]` on both streams; `airdate` is declared as `x-cursor-field` for manifest
parity, matching legacy's `CursorFields: []string{"airdate"}` — neither legacy nor this bundle
declares an `incremental` block, so no request-time filtering happens on either side (both are
always full-stream reads). Neither stream paginates: legacy issues exactly one request per read
(`tvmaze_schedule.go:83`), so no `pagination` block is declared on either stream.

## Write actions & risks

None. Legacy `Write` always returns `connectors.ErrUnsupportedOperation`; `metadata.json` declares
`capabilities.write: false` and no `writes.json` file exists, matching legacy exactly.

## Known limits

- **`web_schedule`'s distinct response shape is not separately modeled.** Legacy uses the
  identical `episodeRecord` mapper for both `schedule` and `web/schedule` responses (TVmaze's own
  web-schedule endpoint returns the same episode-with-nested-show shape as the broadcast schedule,
  just scoped to web-only shows); this bundle's `web_schedule` schema is therefore intentionally
  identical to `schedule`'s, matching legacy's shared mapper exactly, not an oversight.
- No pagination fixture pages are shipped beyond a single page per stream since neither stream
  paginates at all (matching legacy, which issues exactly one request per `Read` call) — there is
  no second page to prove.
