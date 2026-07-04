# Overview

Statsig is a Tier-1 declarative-HTTP connector reading and managing feature gates, dynamic
configs, experiments, segments, target apps, tags, keys, holdouts, layers, users, audit logs, and
environments through the Statsig Console API (`https://statsigapi.net/console/v1/...`, API version
`20240601`). This bundle was Pass-B full-surface expanded against the real published OpenAPI
specification (`https://api.statsig.com/openapi/20240601.json`, linked from
`https://docs.statsig.com/console-api/all-endpoints-generated`) — see `api_surface.json` for the
per-endpoint disposition of all 302 documented method+path pairs. It originally targeted capability
parity with `internal/connectors/statsig` (the hand-written connector it migrates); the legacy
package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Statsig Console API key via the `api_key` secret; it is sent as the `STATSIG-API-KEY`
header with no prefix, matching legacy's `connsdk.APIKeyHeader("STATSIG-API-KEY", key, "")`
(`statsig.go:114`). Never logged. `base_url` defaults to `https://statsigapi.net/console/v1` and
may be overridden for tests/proxies. The real API also accepts an optional
`STATSIG-API-VERSION` header (currently optional, will become required per Statsig's own docs) —
not sent by this bundle since the API treats its absence as "use latest" today; add it as a static
`streams.json` header if Statsig ever makes it mandatory.

## Streams notes

12 GET list/detail streams, all sharing the real API's pagination shape (`page`/`limit` query
params, `data` records array, page-number pagination) EXCEPT `environments` (single-page,
non-paginated, nested at `data.environments`):

- `feature_gates` (`/gates`), `dynamic_configs` (`/dynamic_configs`), `experiments`
  (`/experiments`), `segments` (`/segments`) — the original 4 legacy-parity streams.
- `target_apps` (`/target_app`), `tags` (`/tags`), `keys` (`/keys`), `holdouts` (`/holdouts`),
  `layers` (`/layers`), `users` (`/users`), `audit_logs` (`/audit_logs`) — new Pass-B streams
  covering the rest of the Console API's top-level catalog resources.
- `environments` (`/environments`) — new Pass-B stream; the real endpoint is NOT a paginated list
  (no `page`/`limit` params documented) and its response body nests the array at
  `data.environments`, not a bare `data` array like every other resource — `records.path:
  "data.environments"` and `pagination: {"type": "none"}` are declared per-stream to override the
  base pagination block accordingly.

None of the streams declare a cursor field (matching legacy's original 4 — no `CursorFields` on any
`connectors.Stream`, and the Console API documents no `updatedAfter`/similar filter on any list
endpoint) — full refresh only for every stream.

**Pagination correction (parity deviation, RESOLVED)**: the pre-Pass-B bundle declared
`pagination.type: offset_limit` (`offset`/`limit` query params), ported verbatim from legacy's
`connsdk.OffsetPaginator{LimitParam:"limit", OffsetParam:"offset"}`. The real, documented Statsig
Console API accepts only `page`/`limit` on every list endpoint (confirmed against the OpenAPI spec's
`parameters` for `/gates`, `/dynamic_configs`, `/experiments`, `/segments`, etc. — no `offset`
parameter is documented anywhere in the Console API surface). Sending `offset` against the real API
would be silently ignored (the server has no such parameter to bind it to), so every read beyond
page 1 would have silently returned page 1's data forever — a genuine correctness bug, not a
stylistic preference. Pagination for all 12 streams is now `page_number` (`page_param: page,
size_param: limit, start_page: 1`), matching the real wire contract; this changes the QUERY STRING
sent on the wire but not any emitted record's data for any input the real API accepts (in fact it
makes multi-page reads work at all against the live API, where they previously would not have
paginated past page 1).

`keys`' primary key is `key` (the API key value/identifier itself — the Console API's own `id` path
parameter for `/keys/{id}` operations); this is a normal record identifier field returned by the
list endpoint, not a bearer credential minted by this connector, but authors should still treat
`keys` stream output with the same care as any other credential-identifier data (see Known limits).

## Write actions & risks

20 write actions across 8 resources, every one a plain single-record JSON-body CRUD mutation the
engine's declarative dialect expresses directly (`body_type: json`, `path_fields: ["id"]` for
update/delete):

- **Gates**: `create_gate`, `update_gate` (PATCH partial update), `delete_gate`.
- **Dynamic configs**: `create_dynamic_config`, `update_dynamic_config`, `delete_dynamic_config`.
- **Segments**: `create_segment`, `delete_segment` (no update endpoint exists on the real API for
  the top-level segment record — `PATCH /segments/{id}/add_ids` and friends are ID-list membership
  operations, excluded, see `api_surface.json`).
- **Tags**: `create_tag`, `update_tag`, `delete_tag`.
- **Target apps**: `create_target_app`, `update_target_app`, `delete_target_app`.
- **Holdouts**: `create_holdout`, `delete_holdout` (update excluded — see Known limits).
- **Layers**: `create_layer`, `delete_layer` (update excluded — see Known limits).
- **Keys**: `create_key`, `delete_key` (deactivate/rotate/update excluded — see Known limits).

All 20 are `capabilities.write: true`; every action's `risk` field flags it as an external mutation
requiring approval, and every `delete_*` action additionally flags irreversibility and declares
`delete.missing_ok_status: [404]` (idempotent delete — the real API documents a 404 response for
every delete endpoint used here).

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`statsig.go`'s `pageSize`/`maxPages`, bounded 1-1000 / a non-negative integer or
  `all`/`unlimited`). The engine's `page_number` paginator has no config-driven page-size or
  max-pages knob (`PaginationSpec.PageSize`/`MaxPages` are static bundle JSON, never templated), so
  this bundle uses legacy's own default (`page_size: 100`) as a fixed bundle value and does not
  declare `page_size`/`max_pages` in `spec.json` at all (a declared-but-unwireable config key is
  worse than an absent one, per `docs/migration/conventions.md` F6). Pagination is unbounded by
  default (reads every page until a short page), matching legacy's own default of `maxPages=0`
  (unbounded) when `max_pages` is unset.
- **No update action for holdouts/layers/keys.** The real API documents `PATCH
  /holdouts/{id}` (pass-percentage and attached gate/experiment/layer ID list), `PATCH/POST
  /layers/{id}` (parameter set and target-app bindings), and `PATCH /keys/{id}` (scopes/environments/
  target-app bindings for a LIVE credential). Each reshapes production traffic allocation or
  credential scope in ways too high-blast-radius for a plain declarative partial-update this pass —
  see `api_surface.json`'s `destructive_admin`/`out_of_scope` entries for the exact per-endpoint
  reasoning. Create/delete are modeled for all three; update is deliberately not.
- **No write actions for experiments.** Experiment creation/update requires a full experimental-design
  payload (variant groups, allocation, primary/secondary metrics, targeting gate, statistical
  settings) that is an authoring-workflow form, not a plain catalog record shape — see
  `api_surface.json`'s exclusion entries for `POST/PATCH/DELETE /experiments{,/{id}}`. The
  `experiments` stream remains read-only.
- **Rule-level, review-workflow, override, and analytics sub-resources are out of scope.** Every
  gate/dynamic-config/experiment/holdout/layer's targeting rules are read/written as part of that
  resource's own full record (the parent create/update write action replaces the whole `rules`
  array); there is no rule-level sub-CRUD (`POST/PATCH/DELETE
  /gates/{id}/rule(s)/{ruleID}`, `.../dynamic_configs/{id}/rule/{ruleId}`), no change-review
  workflow (`.../reviews`), no per-user override management (`.../overrides`), and no pulse/exposure
  analytics (`.../pulse_results`, `.../cumulative_exposures`, etc.) — see `api_surface.json` for the
  full per-endpoint breakdown (categories: `out_of_scope`, `destructive_admin`,
  `requires_elevated_scope`, `binary_payload`, `duplicate_of`, `non_data_endpoint`).
- **Segment ID-list membership (`add_ids`/`remove_ids`/`id_list`) is excluded as
  `binary_payload`.** These endpoints accept/return large newline-delimited ID-list bodies, not the
  JSON catalog records this dialect's write/record model expresses.
- **Key lifecycle actions (`deactivate`/`rotate`/`update`) are excluded as `destructive_admin`.**
  These mutate or invalidate a LIVE production credential in place; `create_key`/`delete_key` are
  the only key mutations modeled, both already flagged high-risk in `writes.json`.
- Full API-surface disposition (every one of the 302 documented Console API v1 method+path pairs) is
  recorded in `api_surface.json`.
