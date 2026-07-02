# Overview

GlassFrog is a wave2 fan-out declarative-HTTP migration. It reads GlassFrog circles, roles,
people, projects, and assignments through the GlassFrog API v3
(`GET https://api.glassfrog.com/api/v3/...`). This bundle migrates
`internal/connectors/glassfrog` (the hand-written connector); the legacy package stays registered
and unchanged until wave6's registry flip.

## Auth setup

Provide a GlassFrog API token via the `api_key` secret; it is sent as the `X-Auth-Token` header
and is never logged, matching legacy's `connsdk.APIKeyHeader("X-Auth-Token", secret, "")`
(`glassfrog.go:231`). `base_url` defaults to `https://api.glassfrog.com/api/v3` and may be
overridden for tests/proxies (legacy's own `glassfrogBaseURL` validates scheme+host the same way;
the engine has no equivalent runtime validation, but every parity/conformance fixture only ever
points at an httptest server, so this is not exercised differently on either side).

## Streams notes

All five streams (`assignments`, `circles`, `people`, `projects`, `roles`) share GlassFrog's
uniform page/per_page pagination (`pagination.type: page_number`, `page_param: page`,
`size_param: per_page`, `start_page: 1`, `page_size: 100`, matching legacy's
`glassfrogDefaultPageSize`/`glassfrogMaxPageSize` of 100) and its resource-named nested-array
envelope (e.g. `{"circles":[...]}`, `records.path` set to the stream name). None of GlassFrog's
list endpoints expose an incremental cursor field (legacy declares no `CursorFields` for any
stream), so no stream declares an `incremental` block — full refresh only, matching legacy exactly.

GlassFrog nests foreign-key ids under a `links` object rather than as top-level scalar fields.
`assignments.person_id`/`assignments.role_id` are derived via `computed_fields` bare single-reference
templates (`{{ record.links.person }}`, `{{ record.links.role }}`), and `circles.supported_role_id`
via `{{ record.links.supported_role }}` — each a bare `{{ record.<path> }}` reference with no filter
stage, so the engine's typed extraction preserves the raw integer type exactly as legacy's own
`nestedObject(item, "links")["person"]` (an untyped `any` that is always a JSON number on the wire)
does. Legacy's `linksField` helper design and this bundle's schema both type these fields
`integer`.

`page_size`/`max_pages` runtime config-driven overrides that legacy exposes
(`glassfrogPageSize`/`glassfrogMaxPages`, `glassfrog.go:264-292`) are not modeled: `page_number`'s
`PageSize` is fixed per bundle (no per-request config lookup in `paginate.go`'s
`newPaginator`), matching the `page_size: 100` legacy default exactly; `max_pages` (legacy's hard
request-count cap override) has no engine-side per-connector config override either, so this
bundle relies solely on the short/empty-page stop signal, the same termination behavior any
`max_pages=0`/unlimited legacy run would exhibit.

## Write actions & risks

None. GlassFrog is read-only (legacy's own `Capabilities.Write: false`, `Write` returns
`connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **Full GlassFrog API surface (checklist items, metrics, action items, organization profile) is
  out of scope for wave2.** See `api_surface.json`'s `excluded: {category: out_of_scope, reason:
  "Pass B capability expansion"}` entries. Only the 5 legacy-parity streams are implemented.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`, a credential-free conformance-harness affordance) synthesizes
  its own deterministic records directly in Go rather than shaping a real API response; this
  bundle's schemas and fixtures target the LIVE record shape only (`glassfrog.go`'s `harvest`
  function and `mapRecord` functions), per convention — the engine's own
  `internal/connectors/conformance` fixture-replay harness provides the credential-free test
  affordance this bundle needs.
- **`page_size`/`max_pages` are not runtime-configurable** — see Streams notes above; this is a
  spec-surface narrowing (both keys are absent from `spec.json`, matching the searxng precedent of
  not declaring a config key no template in the bundle actually consumes), not a data-emission
  change for any single read.
