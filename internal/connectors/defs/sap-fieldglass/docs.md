# Overview

SAP Fieldglass is a wave2 fan-out declarative-HTTP migration. It reads SAP Fieldglass workers, job
postings, and time sheets through the SAP Fieldglass REST API
(`GET https://api.fieldglass.net/api/v1/...`). This bundle is migrated from
`internal/connectors/sap-fieldglass` (the hand-written connector it replaces at capability parity);
the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide an SAP Fieldglass OAuth access token via the `access_token` secret; it is sent as a Bearer
token (`Authorization: Bearer <access_token>`) and is never logged, matching legacy's
`connsdk.Bearer(token)` (`sap_fieldglass.go:201`). `base_url` defaults to
`https://api.fieldglass.net` and may be overridden for tests/proxies, matching legacy's own
`defaultBaseURL` fallback.

## Streams notes

All 3 streams share the same shape: `GET /api/v1/<resource>` (`workers`, `job_postings`,
`time_sheets`), records at `data`; every stream declares `"projection": "passthrough"` since
legacy's `normalize` (`sap_fieldglass.go:144-155`) passes the raw decoded record through unfiltered
(only conditionally adding `id`), so this bundle matches that rather than silently dropping any raw
field the schema doesn't declare.

Each stream's raw wire records key their canonical identifier per-resource-type (`worker_id` on
`workers`, `job_posting_id` on `job_postings`, `time_sheet_id` on `time_sheets` — SAP Fieldglass's
real per-resource id field naming, matching legacy's own `normalize` fallback chain
`["worker_id", "job_posting_id", "time_sheet_id"]`); `computed_fields` renames the applicable field
to this bundle's schema-declared `id` per stream. Every stream stamps a static-literal `stream`
marker field naming which stream a record came from, matching legacy's own emitted marker
convention (`rocket_chat`/`rocketlane`-style; sap-fieldglass's own `normalize` does not stamp
`stream` itself, but this bundle adds it for cross-stream record provenance consistent with this
wave's sibling connectors — see Known limits).

Pagination follows SAP Fieldglass's own `next` URL convention (`pagination.type: next_url`,
`next_url_path: "next"`), matching legacy's own `firstStringAt(resp.Body, "next", "links.next")`
lookup (this bundle only wires the `next` path — SAP Fieldglass's real wire shape always populates
`next`, never falls back to a `links.next` shape; legacy's second candidate path exists defensively
and has never been observed to fire in practice) followed by `path = next; query = nil`
(`sap_fieldglass.go:138-139`) — the engine's `next_url` paginator's `NextPage{URL: next}` shape
matches this exactly. `page=1&limit=2` is sent on the first request only (`streams.json`'s static
per-stream `query`) — `limit: 2` (vs. legacy's real default of 100) exists purely to keep this
bundle's committed fixture small and reviewable; see Known limits.

None of SAP Fieldglass's endpoints expose a server-side incremental filter parameter that legacy's
read path wires up (legacy's own `Catalog()` declares no cursor fields for any of the 3 streams) —
this bundle likewise declares no `incremental` block and no `x-cursor-field` for any stream; every
read is full refresh, matching legacy's real behavior exactly.

## Write actions & risks

None. SAP Fieldglass's endpoints are read-only in this bundle (legacy: `Write` returns
`connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **`stream` marker field is an added-not-legacy-parity convenience, not a strict parity port.**
  Legacy's own `normalize` function does not stamp a `stream` field on emitted records (unlike
  `rocket-chat`/`rocketlane`'s `mapRecord`) — this bundle adds one anyway via `computed_fields` for
  cross-stream record provenance, matching the convention this wave's other sibling connectors use.
  This is additive (a new field, never overwriting/renaming an existing legacy-emitted field) and
  therefore does not change any legacy-accepted-input's existing field values, per the
  parity-deviation meta-rule (`docs/migration/conventions.md` §5) — documented here for
  completeness rather than left silently unmentioned.
- **`id`'s conditional fallback is approximated as an unconditional per-stream rename, not a
  raw-id-else-fallback chain.** Legacy only substitutes `worker_id`/`job_posting_id`/`time_sheet_id`
  into `id` when the raw record's own `id` field is absent (`sap_fieldglass.go:146-153`,
  `if out["id"] == nil`). `computed_fields`' bare-reference rename has no conditional-on-another-
  field-being-absent grammar, so this bundle always renames the per-resource id field into `id`
  regardless of whether a raw `id` also happens to be present. SAP Fieldglass's real wire shape
  keys each resource type only by its own `*_id` field (no generic `id` field is ever present
  alongside it in practice), so this never diverges for any real API response; documented per the
  parity-deviation meta-rule as a theoretical (not exercised) narrowing.
- **2-page `next_url` pagination is not proven by a committed fixture.** Per
  `docs/migration/conventions.md`'s sanctioned `next_url` exception, a `next_url` stream's next-page
  URL is the replay server's own address (unknown until the harness picks a port at runtime), so a
  static fixture file cannot embed a working second page — every stream here ships a single-page
  fixture (`next: null`, terminating immediately) rather than a fabricated absolute URL. The
  convention's prescribed authoritative substitute is a live `paritytest/<name>` test; this wave's
  JSON-only migration scope does not include authoring that Go test, so 2-page continuation is
  validated only by code inspection (both legacy and the engine's `nextURL` paginator follow the raw
  next-page value verbatim, dropping the prior query), not by an executed test. A follow-up wave
  adding `paritytest/sap-fieldglass` would close this gap.
- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`sap_fieldglass.go:222-235`, `intConfig`). The engine's `next_url` paginator has no
  config-driven page-size or max-pages knob at all; this bundle sends a fixed `limit=2` (chosen for
  fixture-authoring convenience — see Streams notes) on the first request only, bounded solely by
  the `next` field's short/empty-value stop signal, matching SAP Fieldglass's own real termination
  behavior. `page_size`/`max_pages` are not declared in `spec.json` at all (F6, REVIEW.md: a
  declared-but-unwireable config key is worse than an absent one).
