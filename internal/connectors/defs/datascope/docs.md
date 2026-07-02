# Overview

DataScope is a wave2 fan-out migration. This bundle reads DataScope locations, form answers, lists,
and notifications from the DataScope external REST API, migrating
`internal/connectors/datascope` (the legacy hand-written connector, which stays registered and
unchanged until wave6's registry flip) at full-refresh capability parity. Server-side incremental
date-window filtering (`answers`/`notifications`' `start`/`end` request params) is narrowed to
full-refresh-only — see Known limits.

## Auth setup

Provide `api_key` as a secret; it is sent as the raw `Authorization` header value with no
Bearer/Basic prefix (`auth.mode: api_key_header`, `header: Authorization`, no `prefix`), matching
legacy's `connsdk.APIKeyHeader("Authorization", secret, "")` exactly. Never logged.

## Streams notes

All 4 streams (`locations`, `answers`, `lists`, `notifications`) share the same shape: `GET`
against a DataScope endpoint, records at the response root (`records.path: "."` — every DataScope
list endpoint returns a bare top-level JSON array, matching legacy's `field_path: []`/root-selector
convention), pagination `offset_limit` (`limit_param: limit`, `offset_param: offset`), stopping on
a short page — identical to legacy's `connsdk.OffsetPaginator{LimitParam: "limit", OffsetParam:
"offset", PageSize: pageSize}`. No `computed_fields` are needed for any of the 4 streams: legacy's
mappers (`datascopeLocationRecord`/`datascopeAnswerRecord`/`datascopeListRecord`/
`datascopeNotificationRecord`) all copy fields directly off the raw item with no renames, so plain
schema projection reproduces them exactly. `streams.json`'s `pagination.page_size: 2` is a
deliberately small static value chosen purely to keep the required 2-page fixture
(`fixtures/streams/locations/{page_1,page_2}.json`) compact — matching the identical
auth0/aviationstack/criteo-marketing precedent in this repo; it has no bearing on a live
deployment.

`answers`/`notifications` are `windowed` in legacy: a `start`/`end` datetime-range request pair,
formatted in DataScope's native `dd/mm/yyyy HH:MM` layout, is added when a persisted cursor or the
`start_date` config resolves (`windowStart`/`normalizeDatascopeTime`); `end` is always
`time.Now().UTC()` at request time. **Neither the custom date format nor a "current wall-clock
time" query value is expressible in this engine's declarative dialect** — see Known limits. This
bundle therefore always omits `start`/`end` (the exact behavior legacy itself falls back to on a
fresh sync with no `start_date` configured — `windowStart` returns `""`, adding no window params at
all), so every read is an unfiltered full sync; no `streams.json` entry declares an `incremental`
block. This never changes emitted record DATA for a full sync — every record legacy would return
without a configured window is still returned identically — it only removes the ability to
narrow a subsequent sync to a date range server-side.

## Write actions & risks

None. This connector is read-only, matching legacy's `Write` stub (`connectors.ErrUnsupportedOperation`).

## Known limits

- **`answers`/`notifications`' incremental date-window filtering is not modeled (documented scope
  narrowing, not a silent behavior change).** Two independent gaps compound here: (1) DataScope's
  window params use a non-standard `dd/mm/yyyy HH:MM` datetime format; the engine's
  `incremental.param_format` dialect supports only `rfc3339`/`unix_seconds`/`date`
  (`2006-01-02`)/`github_date_range` — none matches DataScope's layout, and there is no
  custom-format escape hatch. (2) Legacy's `end` parameter is always the CURRENT wall-clock time at
  request time (`time.Now().UTC()`), not a value derived from config/state/incremental lower
  bound — the engine's template `Vars` environment has no "current time" reference at all (`grep`
  confirms no `time.Now()`-equivalent anywhere in `engine/{interpolate,read}.go`), so even a
  hypothetical custom `param_format` could not produce `end` on its own. Both gaps would need new
  engine dialect surface (a custom/pluggable date-format parameter, plus a resolvable
  "current time" reference) to close correctly; approximating either one (e.g. hardcoding a format
  string transform in a `computed_fields`-style template, which does not exist for query params
  anyway, or omitting `end` while still sending a wrong-shaped `start`) would silently diverge from
  legacy's real request shape rather than matching it, so this bundle narrows scope instead of
  faking the window. `x-cursor-field` is still declared on both streams' schemas
  (`created_at`) matching legacy's catalog metadata, but — matching the narrowed behavior — no
  `incremental` block is declared, so only full-refresh sync modes apply.
- Full DataScope API surface (form/location mutation, dispatch/task creation, user management) is
  out of scope for wave2; see `api_surface.json`'s `excluded: {category: out_of_scope, reason:
  "Pass B capability expansion"}` entries.
