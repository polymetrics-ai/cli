# Overview

SolarWinds Service Desk is a wave2 fan-out declarative-HTTP migration, expanded to full practical
surface coverage in Pass B. It reads SolarWinds ITSM (Samanage) incidents, problems, changes,
change catalogs, releases, solutions, catalog items, configuration items, users, sites,
departments, roles, groups, categories, hardware/mobile/other/software assets, printers,
contracts, purchase orders, vendors, audits, and risks through the SolarWinds ITSM (Samanage) REST
API (`GET https://api.samanage.com/<resource>.json`), and writes delete actions for every resource
with a documented delete-by-id endpoint. This bundle is capability-parity migrated from
`internal/connectors/solarwinds-service-desk` (the hand-written connector it migrates); the legacy
package stays registered and unchanged until wave6's registry flip.

**`metadata.json`'s `docs_url` was corrected this pass**: the prior
`documentation.solarwinds.com/.../swsd-api.htm` URL 404s live as of this review (`DOCS_UNREACHABLE`
territory). The real, working, machine-readable API reference is
`https://apidoc.samanage.com/` (a ReDoc-rendered OpenAPI 3 document; the raw spec is served at
`https://apidoc.samanage.com/redoc/schema/resolved_schema.json`) — this is the source of truth for
every endpoint/schema claim in this docs.md and `api_surface.json`.

## Auth setup

Provide a SolarWinds Service Desk API key via either the `api_key_2` secret or the `api_key`
secret; both are sent as a Bearer token. `base.auth` declares two `when`-gated bearer candidates in
declared order (conventions.md §3's dual-auth-ordering pattern): `api_key_2` first (gated `when:
{{ secrets.api_key_2 }}`), `api_key` second (gated `when: {{ secrets.api_key }}`) — reproducing
legacy's own `firstSecret(cfg, "api_key_2", "api_key")` precedence exactly
(`solarwinds_service_desk.go:141`: `api_key_2` wins when both are configured). Every request also
sends a fixed `Accept: application/vnd.samanage.v1.1+json` header, matching legacy's
`acceptHeader` constant (`solarwinds_service_desk.go:18`, `connsdk.Requester.Accept`). Neither
secret is ever logged.

## Streams notes

All 24 streams (`incidents`, `users`, `departments`, `categories`, `problems`, `changes` — 6
legacy-parity — plus `change_catalogs`, `releases`, `solutions`, `catalog_items`,
`configuration_items`, `sites`, `roles`, `groups`, `hardwares`, `mobiles`, `other_assets`,
`softwares`, `printers`, `contracts`, `purchase_orders`, `vendors`, `audits`, `risks`, newly added
in Pass B) share the same shape: `GET /<resource>.json`, records at the response body's ROOT array
(`records.path: ""`, matching legacy's `streamEndpoints[...].recordsPath == ""` and the real API's
flat-array JSON list convention — see the schema-shape note below), and `pagination.type: "none"`
— legacy itself performs **no automatic pagination loop** for its own 6 streams, and this bundle
extends that exact same unconditional-single-request shape to every newly added stream for
consistency, rather than inventing a pagination loop the legacy connector never had.

`incidents` additionally forwards `config.start_date` as the `updated_after` query param via the
opt-in optional-query dialect (`omit_when_absent: true`) — present only when `start_date` is
configured, omitted entirely otherwise, matching legacy's own `copyConfig(q, cfg, "start_date",
"updated_after")` (`solarwinds_service_desk.go:150`, only wired for the `incidents` stream). All 24
streams forward `config.page`/`config.per_page` verbatim as `page`/`per_page` query params
(likewise `omit_when_absent`), matching legacy's own unconditional `copyConfig(q, cfg, "page",
"page")`/`copyConfig(q, cfg, "per_page", "per_page")` pattern extended uniformly to the new
streams. No stream declares an `incremental` block: legacy's own catalog declares no
`CursorFields` for any of its 6 streams, and `updated_after` is a raw pass-through param, not an
engine-computed incremental lower bound — matching that, no schema (old or new) declares
`x-cursor-field`.

**Response-envelope judgment call for the newly added streams**: the live OpenAPI spec's schema
for several new list endpoints (e.g. `solutions`, `hardwares`, `contracts`) nests each array
item's properties under a singular resource-named key in its documented `items` schema (e.g.
`{"solution": {...}}`) — but this exactly mirrors the SAME nesting the spec ALSO shows for
`incidents`' XML representation (`xml: {wrapped: true}`), while `incidents`' real, already-verified
JSON list response (this bundle's own working `incidents` stream, and legacy before it) is
genuinely FLAT, not wrapped. This is an OpenAPI-authoring artifact from a schema shared between an
XML and a JSON representation, not a real per-resource JSON response difference. All 18 new
streams therefore use the same flat, unwrapped `records.path: ""` shape as the 6 proven legacy
streams, consistent with the well-documented community understanding of Samanage's JSON API
convention (community API clients / Postman collections uniformly document flat JSON list
responses). If a future live-credentialed verification finds any specific new resource's JSON list
response genuinely IS wrapped, that resource's `records.path` needs a one-line fix — flagged here
rather than silently assumed.

All 24 streams declare `"projection": "passthrough"`. Legacy's `Read` emits the raw API record
verbatim (`emit(connectors.Record(rec))`, `solarwinds_service_desk.go:117`, inside `readRecords`)
with no field-building/filtering step for its own 6 streams; the 18 newly added streams follow the
same passthrough convention for consistency with every sibling stream in this bundle. Every real
Samanage field beyond each schema's narrow `id`/`name`/`created_at`/`updated_at` properties (e.g.
`priority`, `state`, `requester`, `assignee`, custom fields) survives to the emitted record exactly
as the live API would return it. Declaring the default `"schema"` projection mode here would
silently narrow every emitted record to the schema's declared properties — so `passthrough` is
required, matching conventions.md §8 rule 1.

## Write actions & risks

20 delete write actions, newly added in Pass B (SolarWinds Service Desk was entirely read-only
before this pass): `delete_incident`, `delete_problem`, `delete_change`, `delete_change_catalog`,
`delete_release`, `delete_solution`, `delete_catalog_item`, `delete_configuration_item`,
`delete_user`, `delete_site`, `delete_department`, `delete_role`, `delete_group`,
`delete_category`, `delete_hardware`, `delete_mobile`, `delete_other_asset`, `delete_contract`,
`delete_purchase_order`, `delete_vendor` — one per resource with a documented `DELETE
/<resource>/{id}` endpoint (`body_type: "none"`, `path_fields: ["id"]`), except `memberships` (no
GET/list endpoint exists to source a valid membership id from; see Known limits). Each is a
**permanent, irreversible external deletion** of the named record on the connected SolarWinds
Service Desk account; external approval is required for every one. None declares
`delete.missing_ok_status`: the live API's own OpenAPI spec documents `404` as a genuine
"Not Found" error response for an unknown id, not an explicitly-safe idempotent-delete signal, so
an unknown id is treated as a real per-record failure, matching the API's own documented contract.

**create/update (POST/PUT) mutations are NOT implemented — `ENGINE_GAP` blocker.** Every one of
this API's create/update request bodies wraps the payload under a single resource-named root key
(e.g. `POST /incidents` expects `{"incident": {...}}`, `PUT /solutions/{id}` expects `{"solution":
{...}}`, confirmed directly from the live OpenAPI spec's `requestBody` schemas). This dialect's
`write.go` has no mechanism to nest a constructed JSON body under a declared root key — `body_type:
"json"` sends every non-`path_fields` record field flush at the top level, and `"form"`/`"none"`
have no analogous namespacing either. **This is the identical engine gap already filed against
`internal/connectors/defs/statuspage`'s mutation endpoints in this program** (see that bundle's
`docs.md`/`api_surface.json`): the two viable fixes are (1) a Tier-1 `writes.json` field (e.g.
`body_envelope: "incident"`) that wraps the constructed body under that key before sending — which
would unblock every Rails-conventioned API hitting this exact shape across the whole migration
program, meeting the §6 recurrence threshold on its own — or (2) an orchestrator-approved new
`hooks/solarwinds-service-desk/` package implementing `WriteHook` to wrap the body. This pass's
constraints forbid creating a new hook package, so every affected create/update endpoint is
recorded in `api_surface.json` as `excluded: {category: "out_of_scope", reason:
"ENGINE_GAP: ..."}`, naming the exact blocker per endpoint (not a blanket placeholder).

## Known limits

- **No automatic pagination is a legacy limitation this bundle reproduces exactly, not a bundle
  narrowing.** Legacy's own `Read` never loops or reads a next-page signal for its 6 streams; this
  bundle extends the identical single-request shape to all 18 newly added streams rather than
  inventing a pagination loop legacy never had for any resource. A result set larger than one API
  page is genuinely truncated by this bundle today, exactly as by legacy.
- **`page`/`per_page`/`start_date` are forwarded with no local validation**, matching legacy's own
  `copyConfig` (a plain pass-through with no parsing/bounds-check). An invalid value now surfaces as
  the API's own error response instead of a local validation error.
- **All create/update mutations are blocked by the request-body-envelope `ENGINE_GAP`** — see
  Write actions & risks above. This is the primary scope limitation of this Pass B pass; every
  delete-by-id endpoint IS implemented (deletes carry no request body and are unaffected by the
  gap).
- **`memberships` has no read stream and no write action**: the live API documents `POST
  /memberships` (relates a user to a group) and `DELETE /memberships/{id}`, but there is no `GET
  /memberships` list endpoint at all to discover a valid id from, and `POST /memberships` itself
  takes `group_id`/`user_ids` as bare QUERY parameters with no request body — this dialect's
  `writes.json` has no declared-query-parameter mechanism (only `path`/`path_fields`/`body_type`/
  `body_fields`), so there is no field to carry those values into the query string even though the
  body-envelope `ENGINE_GAP` above does not itself block this one action.
- **`/attachments` (file upload) is out of scope**: it requires a `multipart/form-data` file body,
  not a JSON/form record this dialect's write actions can express (`binary_payload` category).
- Several sub-resource endpoints scoped to a dynamic `{object_type}` path segment (tasks, comments,
  time_tracks, purchases — attachable to incidents/problems/changes/etc. generically) and a handful
  of niche relationship-management actions (asset links, hardware warranties, contract line items,
  change-catalog-triggered change requests, catalog-item-triggered service requests) are
  intentionally out of scope this pass as a breadth-vs-cost triage call — see `api_surface.json`'s
  per-endpoint `out_of_scope` reasoning.
- Legacy's own hard-error message when neither secret is configured names only `api_key_2`
  (`"solarwinds-service-desk connector requires secret api_key_2"`,
  `solarwinds_service_desk.go:143`) even though either secret satisfies the requirement — a
  pre-existing legacy wording quirk. This bundle's engine-native "no auth spec matched" error text
  differs but the same input (neither secret set) fails identically; this is a parity-neutral
  error-message difference, not an accepted-input-behavior change (§5 meta-rule).
