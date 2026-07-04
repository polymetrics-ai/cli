# Overview

Recreation.gov (RIDB) is a wave2 fan-out declarative-HTTP migration. It reads Recreation.gov RIDB
facilities, campsites, activities, organizations, and recreation areas through the public RIDB REST
API (`GET https://ridb.recreation.gov/api/v1/...`). This bundle targets capability parity with
`internal/connectors/recreation` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip. Read-only: legacy's `Write` always returns
`connectors.ErrUnsupportedOperation`, and this bundle declares `capabilities.write: false` with no
`writes.json` to match.

## Auth setup

Provide a RIDB API key via the `api_key` secret. It is sent on the `apikey` header with no prefix
(`apikey: <api_key>`), matching legacy's `connsdk.APIKeyHeader("apikey", key, "")`
(`recreation.go:148`) exactly via the engine's `api_key_header` auth mode (`header`/`value`/`prefix`
fields map 1:1 onto `connsdk.APIKeyHeader`'s three constructor arguments). `base_url` defaults to
`https://ridb.recreation.gov/api/v1` and may be overridden for tests/proxies (legacy's own `baseURL`
helper validates scheme+host the same way; the engine's base-URL resolution has no equivalent
runtime validation, but every fixture/conformance run only ever points at an httptest server, so
this is not exercised differently on either side).

## Streams notes

Five streams, all primary-keyed on `id`: `facilities`, `campsites`, `activities`, `organizations`,
`recareas`. Each hits a flat RIDB list endpoint (`/facilities`, `/campsites`, `/activities`,
`/organizations`, `/recareas`) whose records live under the top-level `RECDATA` array
(`connsdk.RecordsAt(resp.Body, "RECDATA")`, `recreation.go:104`, identical for every stream in
legacy). None of RIDB's raw field names match this bundle's schema field names directly (RIDB uses
PascalCase resource-prefixed keys like `FacilityID`/`FacilityName`/`CampsiteID`/`OrgID`), so every
stream uses `computed_fields` to rename the raw fields into legacy's own emitted shape — e.g.
`"id": "{{ record.FacilityID }}"`, `"name": "{{ record.FacilityName }}"` — a bare single-reference
`computed_fields` template, so the engine's typed extraction copies the raw JSON value verbatim
(RIDB's IDs are wire-format strings, matching legacy's own fixture-mode `strconv.Itoa` shape and
this bundle's `"string"`-typed `id` schema field). This mirrors legacy's own `mapRecord` functions
(`facilityRecord`/`campsiteRecord`/`activityRecord`/`organizationRecord`/`recAreaRecord`,
`recreation.go:178-192`) field-for-field: `facilities` and `recareas` map `id`/`name`/`updated_at`
(recareas has no `type` field, matching legacy); `campsites` maps `id`/`name`/`type`/`updated_at`;
`activities` and `organizations` map only `id`/`name`.

Pagination is offset+limit (`pagination.type: offset_limit`, `limit_param: limit`, `offset_param:
offset`, `page_size: 50` matching legacy's `recreationDefaultPageSize`) — the engine's
`OffsetPaginator` stops on a short page (fewer records returned than the page size), which
coincides with RIDB's own real termination behavior. Legacy's own stop condition additionally checks
the response body's `METADATA.RESULTS.TOTAL_COUNT` field directly (`offset >= total`,
`recreation.go:119`); the engine's `offset_limit` paginator has no equivalent body-driven
total-count stop-signal hook (unlike the `cursor` paginator's `stop_path`), so this bundle relies on
the short-page signal alone — see Known limits.

`facilities` and `campsites` declare `x-cursor-field: updated_at`, matching
legacy's own `Catalog` `CursorFields` declaration (`recreation.go:170-171`) for
informational/dedup-mode purposes only. `recareas` declares no `x-cursor-field`
because legacy's `recareas` stream (`recreation.go:174`) publishes no `CursorFields`
(unlike facilities/campsites). Per `docs/migration/conventions.md` §2,
`incremental_append` sync modes are gated on the presence of an `incremental` block, not on
`x-cursor-field` alone. None of the five streams expose a server-side incremental filter parameter
in legacy (`Read` never sends a date-filter query param — `harvest` only ever sends `limit`/`offset`),
so this bundle declares no `incremental` block for any stream, matching legacy exactly.

## Write actions & risks

None. Legacy's own `Metadata()` declares `Write: false`; `capabilities.write` is `false` and this
bundle ships no `writes.json`, matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation`.

## Known limits

- **`TOTAL_COUNT`-based stop signal is not independently checked.** Legacy's `harvest` stops when
  EITHER the page returned fewer records than `pageSize` OR `offset >= METADATA.RESULTS.TOTAL_COUNT`
  (`recreation.go:119`); the engine's `offset_limit` paginator (`connsdk.OffsetPaginator`) only
  implements the short-page stop rule and has no `stop_path`-equivalent hook to also read
  `METADATA.RESULTS.TOTAL_COUNT` directly (that mechanism exists only on the `cursor` paginator
  variant, per `docs/migration/conventions.md` §3's pagination table). In every real RIDB response, a
  short/empty final page and `offset >= TOTAL_COUNT` occur together (a full-size final page whose
  offset already reached the total would be a genuinely unusual/malformed response for this API), so
  this is not expected to diverge in practice; it is documented as an `ENGINE_GAP`-adjacent limitation
  rather than silently assumed identical.
- **`page_size` is not runtime-configurable.** Legacy exposes a config-driven `page_size` override
  (`recreationDefaultPageSize`/`recreationMaxPageSize`, `recreation.go:231-241`). The engine's
  `offset_limit` paginator's `PageSize` is a bundle-declared constant (`streams.json`'s
  `base.pagination.page_size: 50`), with no per-request config-driven override mechanism — a
  `{{ config.page_size }}` template on `page_size` is not expressible (`PaginationSpec.PageSize` is a
  plain int field, not a template). This bundle therefore fixes RIDB's own default (`limit=50`) and
  does not declare `page_size` in `spec.json` at all (a declared-but-unwireable config key is worse
  than an absent one, per the bitly/searxng/pagerduty F6 precedent).
- **`max_pages` is not modeled.** Legacy's hard request-count cap override (`recreation.go:243-253`)
  has no engine-side equivalent wired to a config value for `offset_limit`; pagination is bounded
  only by the short-page stop signal, matching RIDB's own real termination behavior.
- **Fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached when
  `config.mode == "fixture"`, a credential-free conformance-harness affordance) stamps every field as
  a `strconv.Itoa`-style string literal (`recreation.go:126-137`) for a fixed 2-record set; this is a
  test-only affordance, not part of the live record shape. The engine's own conformance/fixture-replay
  harness (`internal/connectors/conformance`) provides the credential-free test affordance this bundle
  needs, so no fixture-mode equivalent is needed here.
