# Overview

Granola is a Tier-1 declarative-HTTP fan-out migration: it reads Granola meeting notes list
metadata AND full per-note detail (summary, transcript, attendees, calendar event, folders) through
the Granola public API. This bundle migrates both streams of `internal/connectors/granola` (the
hand-written connector) — `notes` and `detailed_notes`; the legacy package stays registered and
unchanged until wave6's registry flip. `detailed_notes`' sub-resource fan-out, previously blocked
(`ENGINE_GAP`, no declarative fan-out mechanism), is now expressed via `streams.json`'s
`fan_out.ids_from.request` dialect (S4 engine mini-wave item 2); see Streams notes.

## Auth setup

Provide the Granola public API key (`grn_` prefix) as the `api_key` secret; it flows into Bearer
auth (`Authorization: Bearer <api_key>`) and is never logged. `base_url` defaults to
`https://public-api.granola.ai/v1`.

## Streams notes

`notes` (`GET /notes`) uses the base's `cursor` pagination: `cursor_param: cursor`,
`token_path: cursor` (the next page's cursor, read from the response body's `cursor` field), and
`stop_path: hasMore` — Granola's own `{"notes":[...],"hasMore":bool,"cursor":"..."}` envelope.
Legacy's stop condition is `hasMore != "true" || cursor == ""`; the engine's `stop_path` mechanism
(falsy on any value other than the literal string `"true"`) plus the paginator's own
stop-on-empty-token default reproduce this exactly. `limit` sends `config.page_size` (default `30`,
matching legacy's `granolaDefaultPageSize`/`granolaMaxPageSize`, both capped at 30 upstream).

**Incremental sync is genuinely stateful here, unlike this wave's other bundles** — legacy's
`incrementalLowerBound` reads `req.State["cursor"]` first (the persisted incremental cursor from a
prior sync), falling back to the `start_date` config value only on a fresh sync. This is expressed
with a proper `incremental` block: `cursor_field: created_at`, `request_param: created_after`,
`param_format: rfc3339`, `start_config_key: start_date`. The engine computes the exact same
lower-bound precedence (state cursor, falling back to `start_config_key`) via
`incremental.lower_bound`/`buildInitialQuery`.

`owner_name`/`owner_email` are derived via `computed_fields` from the raw nested `owner.{name,email}`
object (legacy's `ownerFields` helper), flattening it to match the schema's top-level fields.

`start_date` only accepts RFC3339 in this bundle (`spec.json`'s `format: date-time`). Legacy also
accepted a bare `YYYY-MM-DD` and normalized it to RFC3339 in code
(`time.Parse("2006-01-02", startDate)`); the engine's `rfc3339` `param_format` sends the configured
value verbatim with no normalization step, so a bare-date `start_date` would be sent un-normalized
to the API (`created_after=2026-01-15` instead of legacy's `2026-01-15T00:00:00Z`). Rather than risk
that silent divergence, this bundle narrows `start_date`'s accepted input to RFC3339 only — the
bare-`YYYY-MM-DD` convenience is dropped (documented config-surface narrowing, matching the
established pattern in other bundles' `start_date` fields, e.g. gitlab's).

`detailed_notes` (`GET /notes/{id}`) fans out from the paginated `notes` list to a per-note detail
fetch, assembling the richer record (`summary`, `transcript`, `attendees`, `calendar_event`,
`folders`) from each individual response — legacy's `harvestDetailed`/`notesSeq`
(`internal/connectors/granola/granola.go:172-257`). This bundle reproduces the shape with
`stream.fan_out`: `ids_from.request` issues a preliminary, fully-paginated `GET /notes` listing
(the SAME endpoint and pagination the `notes` stream itself reads — base's `cursor` pagination,
since `detailed_notes` declares no stream-level override — extracting `id` off each record at
`records_path: "notes"`); `into.path_var` makes the resolved note id referenceable in the stream's
own `path` as `{{ fanout.id }}` (`/notes/{{ fanout.id }}`); the static query param
`include=transcript` matches legacy's own detail-fetch query
(`r.DoJSON(ctx, http.MethodGet, "notes/"+url.PathEscape(id), url.Values{"include":
[]string{"transcript"}}, ...)`, `granola.go:184`). Pagination of the id-listing phase and the
detail-fetch requests are independent, mirroring legacy's own per-note follow-up call.
`owner_name`/`owner_email` are derived the same way as `notes`' identical computed_fields. No
`stamp_field` is declared: unlike a stitched sub-resource (trello's boards→lists, google-tasks'
tasklists→tasks), each `detailed_notes` record already carries its own real `id` from the detail
response itself — there is no parent id to stamp onto it.

**`detailed_notes` has no `incremental` block, unlike `notes`** (documented parity deviation,
ACCEPTABLE per conventions.md §5's meta-rule; also see §8 rule 2's incremental truth table). Legacy
applies `created_after` filtering only to the underlying LIST phase of `harvestDetailed` (the same
`notesSeq` the `notes` stream itself walks) — the per-note detail GET carries no incremental filter
of its own. The engine's `fan_out.ids_from.request` mechanism has no `query`/`incremental` field at
all (`FanOutIDsRequest` is `path`/`records_path`/`id_field` only — see `bundle.go`); its preliminary
id-listing request is not routed through `buildInitialQuery`, so there is no way to apply
`created_after` to that phase declaratively. Meanwhile, `stream.Query`/`incremental` on
`detailed_notes` itself would apply to the PER-NOTE DETAIL fetch instead (every per-id sub-sequence
request runs through the stream's own `buildInitialQuery`) — sending `created_after` to
`GET /notes/{id}` would be a genuinely wrong request shape, not what legacy does at all. Declaring
`incremental` here was therefore rejected as an active-behavior-changing mistake, not merely
omitted; `x-cursor-field: created_at` stays on the schema for downstream dedup/sort parity with
legacy's published `CursorFields`, matching the §8 rule 2 "neither → no incremental block, keep
`x-cursor-field` in schemas only" outcome for a mechanism the engine genuinely cannot reproduce
faithfully. Every `detailed_notes` sync therefore re-lists and re-fetches every note on every run —
see Known limits.

## Write actions & risks

None. Granola is a read-only source in this connector (legacy `Capabilities.Write` is `false`); no
`writes.json` file is present.

## Known limits

- **`detailed_notes` has no incremental narrowing (documented parity deviation, ACCEPTABLE)**: as
  explained in Streams notes, the engine's `fan_out.ids_from.request` id-listing phase has no
  `query`/`incremental` field to carry `created_after` on, and applying `incremental` to the stream
  itself would incorrectly send that filter to the per-note detail fetch instead of the listing
  phase (not what legacy does). Every `detailed_notes` sync re-lists every note (via the same
  `/notes` endpoint the `notes` stream reads, unfiltered) and re-fetches every note's detail on
  every run, rather than narrowing to notes created since the last sync the way legacy's
  `harvestDetailed` does. This never produces WRONG data (every note's current detail is always
  re-emitted in full), only a strictly larger read volume per sync than legacy's incrementally
  narrowed one — downstream dedup on `id`/`x-primary-key` still produces the correct final state.
  Closing this fully would need either an engine extension to `fan_out.ids_from.request` (an
  optional `query`/`incremental`-forwarding field for the listing phase specifically) or a
  `StreamHook` (Tier 2); out of scope for this Tier-1 JSON-only pass.
- `start_date`'s bare-`YYYY-MM-DD` convenience is dropped (see Streams notes above); only RFC3339 is
  accepted.
- Full Granola API surface beyond `/notes` and `/notes/{id}` (if any) is out of scope for wave2.
