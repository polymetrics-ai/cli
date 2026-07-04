# Overview

Grafana is a Tier-1 declarative-HTTP wave2 fan-out migration. It reads Grafana dashboards, folders,
data sources, organization users, and provisioned alert rules through the Grafana instance REST
API. This bundle migrates `internal/connectors/grafana` (the hand-written connector); the legacy
package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide the Grafana instance URL as `base_url` (e.g. `https://your-grafana.grafana.net`; per-tenant,
no fixed default — matching legacy's own no-default `url`/`base_url` config) and a Grafana service
account token or API key as the `api_key` secret. Auth is Bearer (`Authorization: Bearer <api_key>`)
and the key is never logged.

## Streams notes

- `dashboards` (`GET /api/search?type=dash-db`) and `folders` (`GET /api/search?type=dash-folder`)
  use `page_number` pagination (`page`/`limit`, `page_size: 1000` matching legacy's
  `grafanaDefaultPageSize`), stopping on a short (or empty) page — legacy's own
  `count < pageSize` rule.
- `datasources` (`GET /api/datasources`), `org_users` (`GET /api/org/users`), and `alert_rules`
  (`GET /api/v1/provisioning/alert-rules`) declare `pagination: {"type": "none"}` at the stream
  level (wholesale-replacing the base's `page_number` default) — legacy's `harvest` never paginates
  these three endpoints (`endpoint.paginated: false`), fetching each in one request.
- Every stream's response is a top-level JSON array, so `records.path` is `""` for all five.
- No `incremental` block is declared for any stream — legacy never sends a date-filter query
  parameter on any Grafana endpoint; every sync is a full, unfiltered fetch. `x-primary-key` still
  matches legacy's catalog `PrimaryKey` (`uid` for dashboards/folders/datasources/alert_rules,
  `userId` for org_users) as informational/dedup metadata.
- `page_size` is not exposed as a `spec.json` config property: the engine's `PaginationSpec.PageSize`
  field is a plain literal integer, not a templatable reference, so there is no mechanism to make it
  config-driven without diverging from the declared literal at load time. Declaring a
  `spec.json.page_size` property that no template anywhere in the bundle consumes would be dead
  config (`conventions.md` §3's dead-config rule) — it is therefore omitted, and the fixed literal
  `1000` (legacy's own default) is used unconditionally, matching legacy's own default-if-unset
  behavior for every caller that didn't override `page_size` (legacy's config override path is
  itself narrowed here, not the runtime behavior most callers actually exercised).

## Write actions & risks

None. Grafana is a read-only source in this connector (legacy `Capabilities.Write` is `false`); no
`writes.json` file is present.

## Known limits

- Full Grafana API surface (dashboard/datasource/alert-rule create/update/delete, annotations,
  teams, service accounts, snapshots, etc.) is out of scope for wave2; see `api_surface.json`'s
  `api_surface.json` concrete exclusion entries. Only the 5
  legacy-parity read streams are implemented.
- Legacy's config-driven `page_size` override (1-5000) is dropped; the bundle always requests
  `limit=1000` for the two paginated streams — see Streams notes above.
- No incremental sync mode is derived for any stream, matching legacy's real unfiltered-every-sync
  behavior.
