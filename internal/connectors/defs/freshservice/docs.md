# Overview

Freshservice is a wave2 fan-out declarative-HTTP migration. It reads Freshservice tickets, agents,
requesters, assets, and problems through the Freshservice REST API v2
(`GET https://<domain>/api/v2/<resource>`). This bundle is capability-parity migrated from
`internal/connectors/freshservice` (the hand-written connector it migrates); the legacy package
stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Freshservice API key via the `api_key` secret; it is sent as the username of HTTP Basic
auth with the literal password `"X"` (`Authorization: Basic base64(<api_key>:X)`), matching
legacy's `connsdk.Basic(secret, freshserviceBasicPassword)` (`freshservice.go:232`) exactly, and is
never logged. `domain_name` (e.g. `acme.freshservice.com`) is required and combined with the fixed
`/api/v2` path segment to build the base URL, matching legacy's domain-derived
`freshserviceBaseURL` fallback path.

## Streams notes

All 5 streams (`tickets`, `agents`, `requesters`, `assets`, `problems`) share `page_number`
pagination: `page` + `per_page` query params, `page_size: 100` (legacy's own default and max page
size, `freshserviceDefaultPageSize`/`freshserviceMaxPageSize`), primary key `["id"]`, cursor field
`updated_at`. A page shorter than 100 records signals the last page, matching legacy's own
`connsdk.PageNumberPaginator` contract exactly (`freshservice.go:152-163`).

`tickets` additionally sends `updated_since` (the resolved incremental lower bound, RFC3339,
`param_format` default) via the opt-in optional-query dialect
(`{"template": "{{ incremental.lower_bound }}", "omit_when_absent": true}`) — present only when
the incremental lower bound resolves (a persisted cursor, or the `start_date` config on a fresh
sync), omitted entirely on an unfiltered read. This matches legacy's own stream-specific gating
(`freshservice.go:142-146`: "Tickets support server-side filtering by updated_since; other streams
would reject the param so it is only applied to tickets"). The other 4 streams (`agents`,
`requesters`, `assets`, `problems`) declare no `incremental` block and never send `updated_since`,
matching legacy exactly — only `tickets` supports this server-side filter upstream.

## Write actions & risks

None. Freshservice is read-only (`capabilities.write: false`, no `writes.json`), matching legacy's
`Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`page_size` is not runtime-configurable per the engine dialect.** Legacy exposes `page_size` as
  a config-driven override (`freshservicePageSize`, `freshservice.go:288-301`, clamped 1-100). The
  engine's `page_number` paginator's `page_size` is a fixed value baked into `streams.json`'s
  `base.pagination` block (conventions.md §3: no runtime config-driven page-size override mechanism
  exists for any pagination type). This bundle bakes in legacy's own default (100), matching
  legacy's out-of-the-box behavior for every caller that never overrode `page_size`; a caller that
  previously set a non-default `page_size` would see a documented, out-of-scope config surface
  narrowing here (`spec.json`'s `page_size` property is declared for documentation/future-wiring
  purposes only and is not consumed by any template today).
- **`max_pages` is not modeled.** Legacy's hard request-count cap override
  (`freshserviceMaxPages`, `freshservice.go:303-316`) has no equivalent knob on the `page_number`
  paginator; `MaxPages` is a bundle-level engine field this connector does not set, so pagination is
  bounded only by the short-page stop signal, matching legacy's own unbounded-unless-configured
  default behavior.
- The `assets` stream's `impact` field is typed `string` here, matching legacy's own declared
  catalog field type (`freshserviceAssetFields()`) even though legacy's own fixture-mode stub
  happens to stamp an integer value for that field — the live Freshservice API returns `impact` as
  a string label on real asset records, and legacy's fixture-mode data is a fixture-only
  inconsistency this bundle does not reproduce (fixture-mode-only fields are out of scope per
  standard migration practice; see the bitly/searxng precedent in `docs/migration/conventions.md`).
