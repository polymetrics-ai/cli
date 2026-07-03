# Overview

RKI COVID reads public Germany COVID-19 metrics derived from Robert Koch-Institut reports via the
corona-zahlen.org JSON API (`https://api.corona-zahlen.org`). It is read-only and requires no
credentials — every endpoint is public. This is a pure **Tier-1 declarative bundle**
(`docs/migration/conventions.md` §1): legacy (`internal/connectors/rki-covid/rki_covid.go`) is a
plain `connsdk`-based HTTP JSON reader with no auth, no pagination, and no protocol-native logic —
every behavior it has is expressible in `streams.json`/`spec.json`/schemas alone. The legacy package
stays registered and unchanged until wave6's registry flip.

## Auth setup

None. `streams.json`'s `base.auth` is `[{"mode": "none"}]`, matching legacy's credential-free
`connsdk.Requester` construction (`rki_covid.go`'s `requester` sets only `Client`/`BaseURL`/
`UserAgent`, no `Auth`).

## Streams notes

Legacy defines 5 streams (`endpoints` map, `rki_covid.go:114-120`), each a distinct GET endpoint on
`api.corona-zahlen.org`, with no pagination on any of them:

- **germany** (`GET /germany`) — a single whole-object response (Germany-wide current metrics), no
  wrapping `data` array. `records.path: ""` selects the whole response body as one record
  (`connsdk.RecordsAt` treats a bare JSON object at the given path as exactly one record). The raw
  payload has no natural id field, so `computed_fields.id` is the static literal `"germany"` —
  matching legacy's `mapRecord` fallback chain (`id`/`ags`/`abbreviation`/`name`/`date` all absent on
  this shape, so `mapRecord` falls all the way through to `out["id"] = stream`).
- **states** (`GET /states`) — `data` is a JSON OBJECT keyed by state abbreviation (`"BW"`, `"BY"`,
  ...), not an array; `records.keyed_object: true` (`key_field: "abbreviation"`) explodes each value
  into its own record (conventions.md §3's keyed-object flatten dialect). Each state object already
  carries its own `id` field in the raw API, so no `computed_fields.id` override is declared —
  `projection: "passthrough"` copies it (and every other raw field) forward verbatim, matching
  legacy's `mapRecord` which never overwrites an already-present `id`.
- **districts** (`GET /districts`) — same keyed-object shape as `states`, keyed by AGS district code
  (`key_field: "ags"`). District objects have no top-level `id`, only `ags`; `computed_fields.id:
  "{{ record.ags }}"` (a bare single reference — typed extraction preserves `ags`'s native JSON
  string type, matching legacy's `first()` returning the raw, unstringified value) reproduces
  legacy's fallback exactly.
- **cases_history** / **deaths_history** (`GET /history/cases`, `GET /history/deaths`) — `data` is
  an array of `{cases|deaths, date}` objects, no id field at all; `computed_fields.id: "{{
  record.date }}"` reproduces legacy's fallback (`ags`/`abbreviation`/`name` are all absent on this
  shape, so `first()` resolves `date`). Both streams declare bare `incremental.cursor_field: "date"`
  with **no** `request_param` — matching legacy's published `CursorFields: []string{"date"}`
  (`rki_covid.go:127-128`) while staying behaviorally identical to legacy's real always-full-sync
  behavior: legacy's `Read` never consults incoming state/cursor at all (grep confirms no
  state-read anywhere in the file), so this bundle's `incremental` block exists purely for
  manifest/derived-sync-mode parity (conventions.md §8 rule 2), never wired to an actual server-side
  filter.

Every stream declares `projection: "passthrough"` (conventions.md §8 rule 1): legacy's `mapRecord`
copies every raw field verbatim (`for k, v := range rec { out[k] = v }`) before only fixing up
`id`/`stream` — this is not a field-built `connectors.Record{...}` mapping, so schema-mode
projection (which would silently drop any undeclared field) would be a parity violation. Every
stream also stamps a static-literal `stream` computed field (`"germany"`/`"states"`/`"districts"`/
`"cases_history"`/`"deaths_history"`), matching legacy's own `out["stream"] = stream` marker.

The optional `days` config value is sent as a `days` query parameter on **every** stream's request
when set (`stream.Query`'s `omit_when_absent` dialect) — this reproduces legacy's real (if slightly
odd) behavior verbatim: `rki_covid.go:79-82` builds one shared `query := url.Values{}` with `days`
set unconditionally before the single `r.Do(...)` call, so `days` is sent on `germany`/`states`/
`districts` requests too, not only the two history endpoints, even though those three endpoints do
not document a `days` parameter. This bundle intentionally reproduces that exact behavior rather than
"fixing" it (§5's meta-rule: never diverge from an accepted-input legacy behavior).

## Write actions & risks

None. `capabilities.write: false`, no `writes.json` — matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation` unconditionally (`rki_covid.go:108-110`).

## Known limits

- **`page_size` config is NOT migrated (F6, dead config).** Legacy's `pageSize(req.Config)`
  (`rki_covid.go:83-85`) validates `config.page_size` is an integer in `[1, 1000]` but its return
  value is discarded (`if _, err := pageSize(req.Config); err != nil`) — it is never sent as a query
  parameter or used to bound any request. `page_size` is therefore genuinely dead config with zero
  wire effect (searxng's F6 precedent, conventions.md §2): declaring it in `spec.json` with no
  template anywhere consuming it would be worse than omitting it. The one observable legacy behavior
  this drops is validation-only: legacy rejects an out-of-range `page_size` value even though it does
  nothing with it; this bundle has no equivalent no-op validation surface to reject against. No
  accepted-input record-emitting behavior changes for any config legacy would actually use
  successfully.
- **No pagination on any stream**, matching legacy exactly (`rki_covid.go` issues exactly one
  `r.Do(...)` call per `Read`, no loop). `streams.json` declares `pagination.type: "none"` at the
  base level; no 2-page fixture is required (conventions.md §4 only mandates one when pagination is
  declared).
- **`cases_history`/`deaths_history`'s `x-cursor-field` is manifest-only**, matching legacy's own
  always-full-sync behavior — see "Streams notes" above.
- **Fixture record shapes (states/districts/history) are recorded-real-shape best-effort, not
  captured from a live call.** This environment had no outbound network access to
  `api.corona-zahlen.org` at migration time; fixtures reproduce the documented, publicly-known
  corona-zahlen.org response shapes (keyed `data` objects for `states`/`districts`, array `data` for
  the history endpoints) with synthetic values, following the same field names legacy's own
  `fields()` catalog declares (conventions.md §4's "recorded-real-shape, sanitized" rule, applied
  here as "documented-shape, sanitized" given the live-capture constraint).
