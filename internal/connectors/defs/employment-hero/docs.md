# Overview

Employment Hero is a wave2 fan-out migration. It reads Employment Hero organisations, employees,
leave requests, and teams through the Employment Hero REST API
(`https://api.employmenthero.com/api/v1`). This bundle migrates
`internal/connectors/employment-hero` (Go package `employmenthero`); the legacy package stays
registered and unchanged until wave6's registry flip. The API is full-refresh only and read-only:
there is no obviously-safe reverse-ETL write surface, matching legacy's `Capabilities.Write: false`.

## Auth setup

Provide an Employment Hero API token via the `api_key` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_key>`) and is never logged.

## Streams notes

`organisations` is the root stream (`GET /organisations`) and supplies organisation ids for the
other three streams. `employees`, `leave_requests`, and `teams` are org-scoped substreams
(`GET /organisations/{organization_id}/employees|leave_requests|teams`); the organisation id is
resolved from the required-for-those-streams `organization_id` config value (interpolated directly
into the stream `path`), matching legacy's `organizationID` resolution. All four streams share
Employment Hero's `page_index`/`items_per_page` page-number pagination convention
(`pagination.type: page_number`, `page_param: page_index`, `size_param: items_per_page`,
`start_page: 1`, `page_size: 100` — legacy's real default `items_per_page`), records at
`data.items`. Every object exposes a string `id`; the API offers only full-refresh syncs (no
incremental cursor), matching legacy's empty `CursorFields`.

Documented scope narrowing: legacy's `organizationID` resolver accepts EITHER a single
`organization_id` config value OR the first non-empty entry of a comma-separated
`organization_configids` list, as a convenience fallback mirroring the upstream catalog's
multi-org config shape. The engine's path-templating dialect has no string-split/first-of-list
filter, so only the single `organization_id` form is wired here; `organization_configids` is not
declared in `spec.json` at all (a declared-but-unwireable key is worse than an absent one, per
searxng's precedent). This never changes the emitted record DATA for any input legacy itself would
accept via `organization_id` — it narrows an alternate CONFIG-SHAPE convenience, not accepted
output.

`items_per_page` is likewise not declared in `spec.json`: `streams.json`'s `pagination.page_size`
is a fixed JSON literal the `page_number` paginator constructor reads once at bundle-authoring
time, with no runtime config-driven override mechanism in this dialect (the same class of gap
documented in `docs/migration/conventions.md`'s searxng worked example) — declaring a spec
property no template consumes would be dead config (F6).

## Write actions & risks

None. Employment Hero is read-only; `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- Only the 4 legacy-parity read streams are implemented; the full Employment Hero surface
  (timesheets, pay runs, expenses, documents, and any write endpoints) is out of scope for this
  wave — see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B capability
  expansion"}` entries.
- Legacy's `organization_configids` comma-list fallback for resolving the org-scoped streams'
  organisation id is not modeled (see Streams notes above) — only the single `organization_id`
  config value is wired.
- Legacy's `max_pages` config override (accepting `0`/`all`/`unlimited` for unbounded, or a
  positive integer cap) has no equivalent in this dialect: `PaginationSpec.MaxPages` is a fixed
  bundle-authored literal, not runtime-config-driven, and this bundle leaves it unset
  (unbounded), matching legacy's own default (`max_pages` unset/`all`/`unlimited` client-side).
- `items_per_page` runtime override is not modeled for the same reason (`pagination.page_size` is
  a fixed literal); the bundle's declared `page_size: 100` reproduces legacy's real default.
