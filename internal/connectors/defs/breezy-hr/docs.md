# Overview

Breezy HR is a wave2 fan-out declarative-HTTP migration. It reads Breezy HR positions (job
openings) and hiring pipelines for a configured company through the Breezy v3 REST API. This
bundle targets capability parity with `internal/connectors/breezy-hr` (the hand-written connector
it migrates) for the `positions` and `pipelines` streams only; the legacy package stays registered
and unchanged until wave6's registry flip, and remains authoritative for the `candidates` stream
this bundle does not cover (see Known limits).

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

Neither stream declares an incremental cursor: Breezy's public API exposes no updated-since filter
for positions/pipelines, matching legacy (`breezyStreams()` declares no `CursorFields` for either;
full-refresh only).

## Write actions & risks

None. Breezy HR is read-only in this connector (`capabilities.write: false`), matching legacy
exactly (`Write` returns `connectors.ErrUnsupportedOperation`).

## Known limits

- **`candidates` is not modeled as a stream in this bundle (ENGINE_GAP).** Legacy's `readCandidates`
  (`breezy-hr/breezyhr.go:146`) is a genuine sub-resource fan-out: it first fully paginates
  `positions` to collect every position id, then issues one
  `GET /position/{position_id}/candidates` request per id, stamping the enclosing `position_id` onto
  every emitted candidate row (falling back to the raw record's own `position_id` field only if the
  caller didn't supply one). The engine's declarative read path drives exactly one paginated request
  sequence per stream against a single templated `path` — there is no primitive for "read stream A
  in full first, then issue one independent request sequence per discovered id from A, merging the
  results into stream B." This is one of conventions.md's named Tier-2 `StreamHook` triggers
  ("sub-resource fan-out reads", e.g. issue -> comments per issue), and Tier-2/Tier-3 escape hatches
  are out of scope for this wave's fan-out task — legacy stays authoritative for `candidates` until a
  future capability-expansion wave implements it via a `StreamHook`.
- **`firstString(item, "_id", "id")`'s `id` fallback is not modeled.** Legacy's `_id`-then-`id`
  fallback (used for both positions' `position_id` and pipelines' `id`) is defensive: Breezy's real
  API consistently returns `_id` for every object in both endpoints, never a bare `id` — the
  `computed_fields` dialect has no conditional/fallback reference syntax (a template is a fixed
  reference or filter chain, never an "A-or-B" expression), so this bundle maps directly from
  `{{ record._id }}`. This never changes emitted data for any real Breezy API response; it only
  differs from legacy for a hypothetical response shape legacy's own dead-code branch anticipated
  but the real API has never been observed to send.
- Full Breezy v3 API surface (candidates in aggregate across all companies, custom fields,
  interview scheduling) is out of scope for this wave; see `api_surface.json`'s `excluded` entry for
  `candidates`.
