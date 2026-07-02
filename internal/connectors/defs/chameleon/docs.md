# Overview

Chameleon is a wave2 fan-out migration. This bundle reads Chameleon surveys, tours, launchers,
tooltips, and segments through the Chameleon v3 REST API, migrating
`internal/connectors/chameleon` (the legacy hand-written connector, which stays registered and
unchanged until wave6's registry flip) at capability parity. Chameleon's upstream API has no
documented cursor pagination, so this bundle reads a bounded limit/offset loop that stops on the
first short page, exactly like legacy. This connector is read-only.

## Auth setup

Provide the Chameleon account secret via the `api_key` secret; it is sent as the
`X-Account-Secret` header (`auth.mode: api_key_header`), matching legacy's
`connsdk.APIKeyHeader(chameleonSecretHeader, secret, "")` with no value prefix. `base_url`
defaults to `https://api.chameleon.io/v3` (legacy's `chameleonDefaultBaseURL`).

## Streams notes

All 5 streams (`surveys`, `tours`, `launchers`, `tooltips`, `segments`) share the same shape: `GET`
against the Chameleon list endpoint, records at the stream's own top-level JSON field
(`surveys`/`tours`/`launchers`/`tooltips`/`segments` respectively — Chameleon's list endpoints
return `{"<field>":[...]}` envelopes, matching legacy's per-endpoint `fieldPath`), offset/limit
pagination (`pagination.type: offset_limit`, `limit_param: limit`, `offset_param: offset`,
`page_size: 50` — matches legacy's `chameleonDefaultPageSize`), stopping on the first short page
(fewer than 50 records), matching legacy's `len(records) < pageSize` rule exactly (Chameleon has
no documented next-page token). Every stream declares `incremental.cursor_field: updated_at`
(matching legacy's declared `CursorFields`) with NO `request_param` and NO `client_filtered` —
legacy's own `harvest` never sends any incremental filter to the API and never client-side filters
either (every sync, incremental or full, walks every page from offset 0); the bare `cursor_field`
declaration exists only so the engine derives `incremental_append` sync-mode eligibility (matching
legacy's own published catalog capability), with the actual read remaining an unfiltered full walk
on every sync, exactly as legacy behaves. Primary key for every stream is `id` — matching legacy's
declared `PrimaryKey`.

`spec.json` intentionally does NOT declare a `limit`/`max_pages` runtime-configurable property
(unlike legacy, which accepts a `config.limit` page-size override and a `config.max_pages`
override): `PaginationSpec.PageSize`/`MaxPages` are read exclusively from `streams.json`'s static
`pagination` JSON literal, never from a `config.*`-templated value (F6, `conventions.md`: a
declared-but-unwireable spec property is worse than an absent one). See Known limits.

## Write actions & risks

None. This connector is read-only, matching legacy's `Write` stub (`connectors.ErrUnsupportedOperation`).

## Known limits

- `limit` (page size)/`max_pages` runtime overrides are not exposed (see Streams notes above) —
  every read uses the fixed `page_size: 50`/unbounded-pages shape baked into `streams.json`. This
  never changes any single emitted record's DATA, only how many requests a sync issues and at what
  page size — parity-deviation ledger candidate, ACCEPTABLE under the meta-rule.
- Full Chameleon API surface (users, events, microsurvey responses, and any mutation of
  experiences) is out of scope for wave2; see `api_surface.json`'s
  `excluded: {category: out_of_scope}` entries. Only the 5 legacy-parity read streams are
  implemented.
