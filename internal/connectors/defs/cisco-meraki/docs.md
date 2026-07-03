# Overview

Cisco Meraki reads Meraki Dashboard API v1 organizations, and per-organization networks,
devices, and admins. This bundle is a full capability-parity migration of the legacy hand-written
connector (`internal/connectors/cisco-meraki`), which stays registered and unchanged until
wave6's registry flip. Read-only: the Meraki Dashboard API exposes configuration/state with no
obvious safe reverse-ETL write target for these streams, matching legacy's own `Capabilities.Write
= false`.

## Auth setup

Provide a Meraki Dashboard API key via the `api_key` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_key>`) and is never logged. `base_url` defaults to
`https://api.meraki.com/api/v1` and may be overridden for tests/proxies (validated as an absolute
http/https URL with a host).

## Streams notes

`organizations` (`GET /organizations`, top-level JSON array response, `records.path: ""`) lists
every organization the API key can access; primary key `id`.

`organization_networks`/`organization_devices`/`organization_admins` are **org-scoped**: legacy's
own read path first lists every accessible organization, then reads the per-organization sub-resource
(`/organizations/{organizationId}/networks|devices|admins`) once per organization id, stamping
`organizationId` onto every emitted record. This bundle reproduces that exact pattern with the
engine's `stream.fan_out` dialect: `ids_from.request` issues a preliminary, fully-paginated
`GET /organizations` listing (the SAME endpoint the `organizations` stream itself reads, extracting
`id` off each record); `into.path_var` makes the resolved organization id referenceable in the
stream's own `path` as `{{ fanout.id }}` (e.g. `/organizations/{{ fanout.id }}/devices`);
`stamp_field: organizationId` writes the current organization id onto every emitted record of that
stream, after projection — exactly matching legacy's `harvest`'s manual `record["organizationId"]
= orgID` stamp. Pagination, page size, and any future per-organization `max_pages` cap are
independent per organization (a fresh Link-header paginator per org), mirroring legacy's own
per-org `harvest` call.

All 4 streams share the same pagination shape: Meraki's RFC 5988 `Link: <url>; rel="next"` header
convention (`pagination.type: link_header`), consumed by the engine's SSRF-guarded
`linkHeaderPaginator` (same-host-only by default). `perPage` is sent as `{{ config.page_size }}`
on the FIRST request of each sub-sequence only — once a `Link` header is present, its absolute
next-page URL (which already carries `startingAfter`) is followed verbatim and no base query is
re-applied, identical to legacy's own `harvest` (`query = url.Values{}` once `page.URL != ""`).
`page_size` defaults to 1000 (Meraki's own page-size ceiling and legacy's
`merakiDefaultPageSize`/`merakiMaxPageSize`).

None of these 4 resources exposes a legacy-recognized incremental cursor field (configuration/state
snapshots, not event streams) — matching legacy's own catalog, which publishes no `CursorFields`
for any stream. All 4 streams are full-refresh only; primary keys are `id` (`organizations`,
`organization_networks`, `organization_admins`) or `serial` (`organization_devices`, matching
legacy's own choice of the hardware serial as the stable device identifier).

## Write actions & risks

None. `capabilities.write` is `false`; no `writes.json` is shipped. Legacy itself implements no
write path for Meraki (`Write` is a stub returning `ErrUnsupportedOperation`).

## Known limits

- Legacy's config-overridable `max_pages` (a per-read hard page-count cap, defaulting to
  unlimited) has no engine-dialect equivalent: `PaginationSpec.MaxPages` is a static integer
  declared in `streams.json`, not a templated/config-driven value. This bundle declares no
  `max_pages` (absent = unbounded), matching legacy's own default (unlimited) behavior for every
  caller that never overrides it; a caller that previously set a numeric `max_pages` override has
  no equivalent knob here. `page_size` (which legacy also allows overriding) IS fully
  config-driven via `spec.json`'s `page_size` property (default 1000), since it flows through an
  ordinary per-request query parameter rather than a paginator-construction-time field.
- `fixtures/streams/organization_networks|organization_devices|organization_admins/page_1.json`
  each ship two fixture organization ids' worth of preliminary `/organizations` listing data plus
  one page of the sub-resource itself, to exercise the fan-out path under conformance's replay
  harness; see `conformance`'s fixture-replay behavior for fan_out-declared streams.
