# Overview

Ubidots reads devices, variables, variable values (dots), device groups, device types, dashboards,
and events from the Ubidots Industrial API (`GET {base_url}/api/v2.0/<resource>/`, plus the
`v1.6`-only values endpoint), and writes device/variable lifecycle mutations and new variable data
points. This bundle originally migrated the hand-written `internal/connectors/ubidots` legacy
package to a declarative Tier-1 defs bundle at capability parity (4 read streams, read-only); this
revision is a Pass B full-surface expansion that researches Ubidots' real documented API and adds
every practical additional list stream and dialect-expressible mutation on top of that parity
baseline. The legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Requires one secret: `token` (Ubidots API token), sent as the `X-Auth-Token` header on every
request via `streams.json` `base.auth`'s `api_key_header` mode — matching legacy's
`connsdk.APIKeyHeader("X-Auth-Token", token, "")` exactly (no prefix). `base_url` defaults to
`https://industrial.api.ubidots.com` (legacy's `defaultBaseURL`), overridable for tests or proxies.

## Streams notes

Four original parity streams share the identical `page_number` pagination shape (legacy's own
`page_size`/`page` query convention) and the identical record mapping (`id`, `label`, `name`,
`created_at`): `devices` (`GET api/v2.0/devices/`), `variables` (`GET api/v2.0/variables/`),
`dashboards` (`GET api/v2.0/dashboards/`), `events` (`GET api/v2.0/events/`) — all read records from
the paginated envelope's `results` array, matching Ubidots' real Django REST Framework-style list
response shape (`count`/`next`/`previous`/`results`). Pagination sends
`page_size=<page_size>&page=<n>` and stops on a short page (fewer than `page_size` records),
matching legacy's `harvest` loop exactly; `page_size` defaults to 100 (legacy's `defaultPageSize`)
and is a fixed bundle-authored value (see Known limits).

No stream declares `x-cursor-field`: legacy's own `streams()` catalog declares no `CursorFields` for
any of these four streams either (`ubidots.go:212-220`), so there is nothing to reproduce — this is
not a scope narrowing, it mirrors legacy's own manifest exactly. `created_at` is projected with a
typed coalesce computed field that preserves legacy's `first(item, "created_at", "createdAt")`
fallback behavior.

**Pass B additions** (new in this revision, researched against `docs.ubidots.com/reference`):
- `device_groups` (`GET api/v2.0/device_groups/`) and `device_types` (`GET api/v2.0/device_types/`)
  — same `page_number`/`results` shape and record mapping as the four parity streams.
- `variable_values` (`GET api/v1.6/variables/{id}/values/`) — Ubidots' variable "dots" (data points)
  live under the older `v1.6` path (the `v2.0` API has no values sub-resource of its own); this
  stream uses the `fan_out` dialect (`ids_from.request` against `api/v2.0/variables/`, `records_path:
  results`, `id_field: id`) to enumerate every variable id first, then reads that variable's values
  sub-resource once per id, `stamp_field: variable_id` tagging every emitted dot with its source
  variable. Each dot's real wire shape is `{value, timestamp, context}` (`timestamp` a Unix
  milliseconds integer, matching Ubidots' documented Dot object exactly — see `dev.ubidots.com`'s
  "Devices, variables, and Dots" guide). No cursor metadata is declared because legacy published no
  cursor fields for Ubidots.

## Write actions & risks

Seven actions, all newly added in this Pass B revision (legacy shipped none —
`capabilities.write` flips to `true`):

- `create_device` (`POST api/v2.0/devices/`) — creates a new device; low-risk, no approval required.
- `update_device` (`PATCH api/v2.0/devices/{id}/`) — updates one or more device fields; no approval
  required.
- `delete_device` (`DELETE api/v2.0/devices/{id}/`) — **destructive**: permanently removes a device
  and cascades to all of its variables and stored values; approval required. Idempotent
  (`missing_ok_status: [404]`).
- `create_variable` (`POST api/v2.0/variables/`) — creates a new variable under an existing device;
  no approval required.
- `update_variable` (`PATCH api/v2.0/variables/{id}/`) — updates one or more variable fields; no
  approval required.
- `delete_variable` (`DELETE api/v2.0/variables/{id}/`) — **destructive**: permanently removes a
  variable and its entire value history; approval required. Idempotent (`missing_ok_status: [404]`).
- `create_variable_value` (`POST api/v1.6/variables/{variable_id}/values/`) — injects a new data
  point (dot) into an existing variable; `path_fields: ["variable_id"]`, body restricted to
  `value`/`timestamp`/`context` via `body_fields` (matching the Dot object's real writable fields);
  no approval required.

Bulk/range value-delete and multi-device bulk-delete endpoints are deliberately excluded (see
`api_surface.json`) as `destructive_admin` — see Known limits.

## Known limits

- **`page_size`/`max_pages` are not exposed as runtime config.** Legacy accepts `config["page_size"]`
  (1-1000, default 100) and `config["max_pages"]` (default 1; `"all"`/`"unlimited"` for unbounded)
  as caller-overridable values. The engine's `PaginationSpec.PageSize`/`MaxPages` fields are plain
  JSON integers fixed at bundle-authoring time in `streams.json`'s `base.pagination` block — there is
  no templated/config-driven override mechanism for either field. Declaring `page_size`/`max_pages`
  as `spec.json` properties that no template in the bundle ever consumes would be dead config (F6,
  REVIEW.md; see also searxng's identical precedent), so neither is declared. This bundle bakes in
  legacy's own DEFAULT values instead: `page_size: 100`, `max_pages: 1` — reproducing the exact
  behavior a caller who never overrides either config key already gets from legacy.
- **`variable_values` declares no cursor or `incremental` block.** Every read re-fans-out across all
  variables and re-reads each variable's full first page of values. A future capability-expansion
  pass could add real incremental support once a stable request-side timestamp filter is confirmed
  against Ubidots' `v1.6` values endpoint.
- **Organizations, users, tokens, UbiFunctions, Pages, plugins, and dashboard-widget CRUD are
  out of scope** (see `api_surface.json`'s `excluded` entries) — none of these were part of legacy's
  surface, and each is either elevated-scope multi-org admin, non-data account introspection, or
  code/UI-configuration deployment rather than a syncable data record.
- **Bulk/range value deletes and multi-device bulk-delete are excluded as `destructive_admin`.**
  Ubidots exposes both a whole-variable value-history delete and a range-scoped value delete; both
  are irreversible mass time-series data loss with no single-record-shaped equivalent this dialect
  can express safely, so neither is modeled as a write action.
