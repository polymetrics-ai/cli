# Overview

SurveyCTO is a Tier-1 declarative-HTTP migration. It reads SurveyCTO forms, submissions,
datasets, and cases through the SurveyCTO API (`GET https://<server>.surveycto.com/api/v2/...`).
This bundle targets capability parity with `internal/connectors/surveycto` (package `surveycto`,
the hand-written connector it migrates); the legacy package stays registered and unchanged until
wave6's registry flip. SurveyCTO is read-only both in legacy and here (legacy's `Write` always
returns `connectors.ErrUnsupportedOperation`).

**Tier justification**: legacy is a pure `connsdk`-based HTTP connector — a `connsdk.Requester`
with `connsdk.Basic` auth, `connsdk.OffsetPaginator`/`connsdk.Harvest` for pagination, and plain
per-record field copies (`copyRecord`). No signature auth, no async job polling, no multipart/XML
bodies, no sub-resource fan-out beyond a single required path substitution (`form_id`) — nothing
that needs a Go hook. This is a clean Tier-1 declarative bundle.

## Auth setup

Provide a SurveyCTO username and password/API key via the `username` and `password` secrets; both
are sent as HTTP Basic auth credentials (`Authorization: Basic base64(username:password)`),
matching legacy's `connsdk.Basic(username, password)` (`surveycto.go:119`). Neither is logged.
Legacy marks only `password` as sensitive in its own error text, but per conventions.md §2's
x-secret discipline this bundle marks BOTH `username` and `password` as `x-secret: true` — an
account identifier paired with Basic auth is still credential-shaped.

## Streams notes

`base_url` is **required** in this bundle, narrowing legacy's config surface: legacy accepts
EITHER an explicit `base_url` OR a bare `server_name`, deriving
`https://<server_name>.surveycto.com/api/v2` in Go when `base_url` is unset
(`surveycto.go:122-145`). This derivation is a function of another config value (not a fixed
literal), which the engine's `spec.json` `"default"` materialization mechanism cannot express (it
only fills in a literal default for an ABSENT key — see conventions.md §3's `spec_json_default`
section and its explicit carve-out for derived defaults like sentry's `hostname`-based URL). This
bundle keeps `server_name` in `spec.json` as descriptive-only config (documented, not wired into
any template) and requires callers to pass the fully-derived `base_url` directly. This is a
documented config-surface narrowing (conventions.md §3), not a data/behavior change for any config
shape that already sets `base_url`.

All 4 streams share the same envelope shape (a JSON object keyed by the stream's collection name
— `{"forms": [...]}`, `{"submissions": [...]}`, etc., legacy's own `endpoint.recordsPath` per
stream) and the same pagination (`connsdk.OffsetPaginator{LimitParam: "limit", OffsetParam:
"offset", PageSize: pageSize}`, `surveycto.go:99`), so `base.pagination` declares
`type: offset_limit` once for all streams. `page_size` defaults to 100 (legacy's
`defaultPageSize`) and is bounded to 1-1000 (legacy's `maxPageSize`) at the application layer in
legacy; this bundle does not re-express that bound (the engine's `offset_limit` paginator takes a
plain integer page size with no client-side range validation) — an out-of-range `page_size` is a
documented, narrow deviation (out-of-range values are simply sent as-is rather than legacy's
hard validation error), never a data change for any legacy-valid config.

`forms` and `datasets` are both simple, non-scoped list endpoints (`GET /forms`, `GET /datasets`)
that legacy maps identically (`formFields()`/`copyRecord("id", "title", "version")` for both) —
this bundle keeps them as two distinct schemas/streams (matching legacy's own two distinct catalog
entries and endpoints), even though their shape happens to coincide.

`submissions` is scoped to one form: the path template `/forms/{{ config.form_id }}/submissions`
substitutes the required `form_id` config value (urlencoded by `InterpolatePath`'s per-segment
default, matching legacy's own `strings.ReplaceAll(endpoint.resource, "{form_id}",
url.PathEscape(formID))`); an absent `form_id` hard-errors on both sides (legacy: `"surveycto
stream requires config form_id for path..."`; engine: an unresolved `config.form_id` path-template
key — same failure classification, different literal text, per conventions.md §5's precedent for
config-validation-error parity).

None of the 4 streams declare an `incremental` block: legacy's `submissions` catalog entry
publishes `CursorFields: []string{"submissionDate"}` but `InitialState`/the read path never sends
any server-side incremental filter parameter for it — full refresh only on every read, matching
legacy exactly (conventions.md §8 rule 2: `x-cursor-field` present in the schema is kept as an
informational/soft cursor marker for downstream use, but no `incremental` block is declared since
legacy sends no server-side filter).

`cases`' primary key is `caseid` (legacy's own catalog: `PrimaryKey: []string{"caseid"}`, not
`id` — cases' raw wire shape uses `caseid`, unlike every other stream), preserved verbatim here.

## Write actions & risks

None. SurveyCTO is read-only both in legacy and here: legacy's own `Write` method returns
`connectors.ErrUnsupportedOperation` unconditionally. `capabilities.write` is `false` and this
bundle ships no `writes.json`.

## Known limits

- **`base_url`/`server_name` derivation is not modeled** (see Streams notes above) — `base_url` is
  required directly; `server_name` is accepted but never wired into any template.
- **`page_size`/`max_pages` range validation is not re-implemented** — legacy hard-errors on an
  out-of-range `page_size` (not 1-1000) or a negative `max_pages`; the engine's declarative
  pagination has no equivalent client-side validation. An operator who already passes a
  legacy-valid `page_size`/`max_pages` sees identical behavior; an out-of-range value that legacy
  would have rejected is instead sent as-is rather than erroring — narrower validation, never a
  data change for any legacy-accepted input.
- **`max_pages` is not modeled as a runtime override.** The engine's `offset_limit` paginator has
  no `MaxPages`-equivalent config-driven knob wired into `streams.json`; pagination is bounded only
  by the short/empty-page stop signal (an empty page from `/forms`, `/submissions`, `/datasets`, or
  `/cases`), matching legacy's own real termination behavior for any config that never sets
  `max_pages` (the common case) and differing only in the rare case an operator relied on the hard
  request-count cap to stop a sync early.
