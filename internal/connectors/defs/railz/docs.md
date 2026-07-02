# Overview

Railz is a wave2 fan-out declarative-HTTP migration. It reads Railz businesses, connections,
customers, invoices, and bills through the Railz REST API (`GET https://api.railz.ai/v1/...`).
This bundle migrates `internal/connectors/railz` (the hand-written connector); the legacy package
stays registered and unchanged until wave6's registry flip.

## Auth setup

Railz accepts a pre-issued Bearer access token; OAuth token exchange is intentionally left to
callers on both sides (legacy's own package doc: "OAuth token exchange is intentionally left to
callers so the package stays dependency-free"). Provide either the `access_token` secret
(preferred, checked first) or the `api_key` secret (fallback), matching legacy's own
`firstSecret(cfg, "access_token", "api_key")` precedence exactly (`railz.go:143`,
`railz.go:211-218`). `streams.json`'s `base.auth` declares two `when`-gated bearer candidates in
that exact declaration order — `access_token` first, `api_key` second — reproducing legacy's
first-match-wins precedence (`docs/migration/conventions.md` §3's dual-auth-ordering rule); both
secrets are marked `x-secret: true` and never logged. `base_url` defaults to
`https://api.railz.ai/v1` and may be overridden for tests/proxies.

## Streams notes

All 5 streams share the same shape: `GET` against the Railz list endpoint, records at the `data`
key. Pagination is `offset_limit` (`limit`/`offset` query params, base default `page_size: 100`
matching legacy's `railzDefaultPageSize`) — a page shorter than `limit` is the last page, matching
legacy's own `len(records) < pageSize` stop condition (`railz.go:117-120`) exactly. The
`businesses` stream declares a stream-level pagination override (`page_size: 2`) purely to keep its
2-page conformance fixture small (`docs/migration/conventions.md` §4); the other 4 streams keep the
base's real default and ship single-page fixtures.

`businesses` and `connections` publish a decorative `incremental.cursor_field: created_at` (no
`request_param`) matching legacy's identical `CursorFields` catalog declaration — legacy never
filters server-side either; every read is a full refresh on all 5 streams on both sides.

Each stream's record shape uses Railz's `*Uuid`-suffixed field names (`businessUuid`,
`connectionUuid`, `customerUuid`, `invoiceUuid`, `billUuid`, `vendorUuid`) via `computed_fields`
renames to legacy's `id`/`business_id`/`customer_id`/`vendor_id` output field names, matching
legacy's `businessRecord`/`basicRecord`/`customerRecord`/`invoiceRecord` mappers' PRIMARY field
name in each case (see Known limits for the secondary/tertiary fallback names those mappers also
try).

## Write actions & risks

None. Railz's legacy connector implements no writes (`Write` returns
`connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **Same-or-alternate-key field fallbacks are not modeled.** Legacy's record mappers each read
  several fields via a small `first(item, keys...)` same-or-alternate-key helper: e.g.
  `businessRecord`'s `id` falls back `businessUuid` → `uuid` → `id`
  (`railz.go:178`); `invoiceRecord`'s `id` falls back `invoiceUuid`/`billUuid` → `uuid` → `id`
  (`railz.go:187`); similar chains exist for `business_id`, `customer_id`, `vendor_id`, `name`,
  and `status`/`created_at` across the other mappers. The engine's `computed_fields` dialect has no
  coalesce/fallback filter — only rename, join, static-literal, and typed bare-reference copy — an
  `ENGINE_GAP` for expressing a same-or-alternate-key fallback declaratively (conventions.md §5's
  agilecrm/searxng precedent for this exact class of gap). Only the PRIMARY key name in each
  fallback chain is modeled here (`businessUuid`, `connectionUuid`, `customerUuid`, `invoiceUuid`,
  `billUuid`, `vendorUuid`, `name`, `totalAmount`) — a hypothetical account whose Railz responses
  use only an alternate key name for one of these fields would see that field come through as
  `null` here where legacy would have populated it. Documented scope narrowing, not a silent
  divergence: no fixture or live Railz response encountered during this migration exercised the
  alternate key names, and legacy's own fallback order always tries the primary name first.
- **`page_size`/`max_pages` config overrides are not modeled.** Legacy exposes `page_size` (1-500,
  default 100) and `max_pages` (0/all/unlimited or a positive integer cap) as config-driven
  overrides (`railzPageSize`/`railzMaxPages`, `railz.go:238-260`). The engine's `offset_limit`
  paginator has no config-driven page-size or request-count-cap knob at all (matches the
  stripe/adobe-commerce-magento/agilecrm precedent for this exact gap class). `page_size`/
  `max_pages` are therefore not declared in `spec.json`; this bundle sends Railz's own default
  (`limit=100`, unbounded pages) as the static pagination block.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (reached only
  when `config.mode == "fixture"`, a credential-free conformance-harness affordance) stamps
  synthetic fields (`uuid`, `businessUuid: "biz_%d"`, etc.) that are not part of the LIVE record
  shape; this bundle's schemas and fixtures target the live path only (`harvest`), matching
  conventions.md's instruction to ignore legacy's fixture-mode-only fields.
