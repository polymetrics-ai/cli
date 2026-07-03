# Overview

Breezy HR is a wave2 fan-out declarative-HTTP migration. It reads Breezy HR positions (job
openings), hiring pipelines, and per-position candidates for a configured company through the
Breezy v3 REST API. This bundle now covers all 3 legacy streams (`positions`, `pipelines`,
`candidates`); the legacy package stays registered and unchanged until wave6's registry flip. The
`candidates` sub-resource fan-out read is expressed via the engine's `fan_out` dialect (S4 engine
mini-wave item 2) — the `ENGINE_GAP` that previously blocked this stream is closed.

## Auth setup

Provide a Breezy HR raw API key via the `api_key` secret; it is sent verbatim as the `Authorization`
header value (`auth: api_key_header`, no `Bearer`/other prefix — matches legacy's
`connsdk.APIKeyHeader("Authorization", secret, "")` with an empty prefix), and a `company_id`
secret naming the Breezy company to scope every request to. Both are never logged. The effective
base URL is `{{ config.base_url }}/company/{{ secrets.company_id }}` (`base_url` defaults to
`https://api.breezy.hr/v3`), matching legacy's `breezyBaseURL`'s root + `/company/<id>` composition
exactly — this is ordinary `base.url` templating, not the `spec.json` `default`-materialization
mechanism (which only fills in a single property's own missing value, not a value derived by
composing two properties together).

`company_id` is marked `x-secret: true` to match the legacy catalog's own field classification,
even though it is a path segment rather than a credential used for request signing/authentication —
matching conventions.md's guidance that the marker is about a field's *declared nature* in the
catalog, not whether this bundle's `auth` block itself consumes it as a secret.

## Streams notes

- `positions` — `GET /positions`, records at the response root (`records.path: "."`; Breezy returns
  a bare top-level JSON array, not an enveloped object). Paginated via `pagination.type:
  page_number` (`page_param: page`, `size_param: limit`, `start_page: 1`, `page_size: 100`),
  matching legacy's `harvestPositions` exactly (1-based pages, `limit=100`, stop on a short or empty
  page). `position_id` is computed from the raw API's `_id` field (`computed_fields`); `type` is
  computed from the raw nested `type.name` object field; `country_id`/`country_name` are computed
  from the raw nested `location.country.id`/`location.country.name` object fields — all three match
  legacy's `breezyPositionRecord`'s nested-object flattening exactly, expressed via
  `record.<dotted.path>` computed_fields instead of Go.
- `pipelines` — `GET /pipelines`, records at the response root, unpaginated (matches legacy's
  `readSimpleList`). `id` is computed from the raw API's `_id` field.
- `candidates` — a sub-resource fan-out over positions, matching legacy's `readCandidates`
  (`breezyhr.go:146`) exactly: `fan_out.ids_from.request` issues the SAME paginated `GET
  /positions` sequence the `positions` stream itself uses (`records_path: ""`, `id_field: "_id"`,
  reusing this stream's own `page_number` pagination override — conventions.md §3's "a fan-out
  id-listing request has no pagination block of its own; it reuses the surrounding stream's"), then
  `into.path_var: "position_id"` threads each discovered position id into `/position/{{ fanout.id
  }}/candidates`, and `stamp_field: "position_id"` writes it onto every emitted candidate record
  after projection — reproducing legacy's `pid` stamp (which always wins over the raw record's own
  `position_id` field, since legacy's `breezyCandidateRecord` only falls back to the raw field when
  no `positionID` argument is supplied, which never happens on the substream path). `id` is computed
  from the raw API's `_id` field; `stage` is computed from the raw nested `stage.name` object field
  — both match legacy's `breezyCandidateRecord` flattening exactly. `query: {"sort":
  "updated_date"}` matches legacy's `url.Values{"sort": []string{"updated_date"}}` verbatim.

None of the 3 streams declare an incremental cursor: Breezy's public API exposes no updated-since
filter for any of them, matching legacy (`breezyStreams()` declares no `CursorFields` for any
stream; full-refresh only).

## Write actions & risks

None. Breezy HR is read-only in this connector (`capabilities.write: false`), matching legacy
exactly (`Write` returns `connectors.ErrUnsupportedOperation`).

## Known limits

- **`candidates`'s per-position request is now paginated; legacy's was not.** Legacy's
  `readCandidates` issues exactly ONE unpaginated `GET /position/{id}/candidates` request per
  position (no `page`/`limit` query params at all) and takes whatever single page the API returns.
  The engine's `fan_out` dialect reuses the SAME `StreamSpec.Pagination` for both the preliminary
  id-listing request AND every per-id child sub-sequence (conventions.md §3) — there is no way to
  paginate the id-listing request (required to avoid silently dropping positions beyond page 1)
  while leaving the per-position candidate reads unpaginated. This bundle declares
  `page_number`/`page=1`/`limit=100` pagination on the `candidates` stream so the id-listing request
  correctly walks every page of `/positions`; the same spec then also (harmlessly, in the common
  case) paginates each per-position candidates read. Documented parity deviation (§5, ACCEPTABLE):
  for any position with 100 or fewer candidates (the page size), the paginator's short-page stop
  means exactly one request is issued per position — byte-identical to legacy. Only a position with
  more than 100 candidates would diverge, and only by emitting MORE candidate records (page 2+) than
  legacy's single-page read — never fewer or wrong ones, so this never changes emitted data for any
  input legacy itself would accept; it only surfaces additional true data for an edge case legacy's
  own unpaginated read would have silently truncated.
- **`firstString(item, "_id", "id")`'s `id` fallback is not modeled.** Legacy's `_id`-then-`id`
  fallback (used for positions' `position_id`, pipelines' `id`, and candidates' `id`) is defensive:
  Breezy's real API consistently returns `_id` for every object on all three endpoints, never a bare
  `id` — the `computed_fields` dialect has no conditional/fallback reference syntax (a template is a
  fixed reference or filter chain, never an "A-or-B" expression), so this bundle maps directly from
  `{{ record._id }}`. This never changes emitted data for any real Breezy API response; it only
  differs from legacy for a hypothetical response shape legacy's own dead-code branch anticipated
  but the real API has never been observed to send.
- Full Breezy v3 API surface (candidates in aggregate across all companies, custom fields,
  interview scheduling) is out of scope for this wave.
