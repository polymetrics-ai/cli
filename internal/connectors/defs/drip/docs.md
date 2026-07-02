# Overview

Drip is an email-marketing platform. This bundle reads Drip subscribers, campaigns, broadcasts,
and accounts through the Drip REST API (`https://api.getdrip.com/v2`) using HTTP Basic auth. It is
read-only, migrated from `internal/connectors/drip` (the hand-written connector this bundle
replaces at capability parity); the legacy package stays registered and unchanged until wave6's
registry flip.

## Auth setup

Provide a Drip API key via the `api_key` secret; it is sent as the HTTP Basic username with a
blank password (`Authorization: Basic base64(api_key:)`), matching legacy's `connsdk.Basic(secret,
"")`, and is never logged. Provide a `account_id` config value to scope `subscribers`/
`campaigns`/`broadcasts` (Drip's account-scoped resources); the `accounts` stream itself is
account-agnostic and does not use it.

## Streams notes

Four streams: `subscribers`, `campaigns`, `broadcasts` (account-scoped, path
`/{account_id}/<resource>`) and `accounts` (global, path `/accounts`, no `account_id` prefix,
matching legacy's `endpointPath`'s `accountScoped: false` branch). The three account-scoped
streams share the base-level `page_number` pagination (`page`/`per_page` query params, 100 records
per page, stopping on a short page — matches legacy's `harvest`'s `len(records) < pageSize` /
`meta.total_pages` combination for the common case where Drip's list responses are exactly
`page_size` long except on the final page). The `accounts` stream overrides pagination to `none`
at the stream level: Drip's `/accounts` endpoint returns a single unpaginated array with no
`meta.total_pages` field at all, and legacy's own harvest loop treats a missing `meta.total_pages`
as "a single page of results" for exactly this endpoint.

Every stream's primary key is `["id"]` and incremental cursor field is `created_at`, matching
legacy's uniform `dripStreams()` catalog — but no stream declares an `incremental` block here: Drip's
list endpoints accept no `updated_since`-style server-side filter parameter, matching legacy's own
`InitialState` always starting with an empty cursor (full refresh only).

## Write actions & risks

None. Drip is read-only in both legacy and this bundle (`capabilities.write: false`); no
`writes.json` file is shipped.

## Known limits

- Full Drip API surface (subscriber writes, events, forms, workflows) is out of scope for this
  wave; see `api_surface.json`'s `excluded: {category: out_of_scope}` entries.
- `metadata.json` declares no `rate_limit` block: legacy enforces no client-side rate limiting for
  Drip, so none is added here either (matches legacy's real, lack-of, throttling behavior).
- `page_size`/`max_pages` are not exposed as runtime-configurable spec properties (unlike legacy's
  `dripPageSize`/`dripMaxPages` config-driven overrides): the engine's `PaginationSpec.PageSize`/
  `MaxPages` fields are static bundle-authored integers, not `{{ }}`-templated from
  `config.*` — there is no runtime override mechanism for either at the engine level (matches the
  same limitation documented in searxng's `docs.md`/ledger item 4). `page_size` is fixed at 100
  (legacy's own default `dripDefaultPageSize`); `max_pages` is unbounded (legacy's own default when
  unset). A declared-but-unwireable spec property was intentionally omitted rather than declared
  dead (F6, REVIEW.md).
- Legacy tolerated Drip's `meta.total_pages` field being absent (treating that as "a single page of
  results"); this bundle's `page_number` paginator instead stops purely on a short page
  (`recordCount < page_size`). For any stream whose LAST page happens to return exactly
  `page_size` records with no further page, legacy would issue one more (empty) request and stop
  there, while this bundle stops one request earlier without ever seeing an empty page. Both
  converge on the identical final record set for every input; only the harmless extra empty
  request legacy issues is not reproduced. Documented here as a benign pagination-loop-count
  difference, not an emitted-data deviation.
