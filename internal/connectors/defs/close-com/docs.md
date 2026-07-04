# Overview

Close.com (Close CRM) is a wave2 fan-out declarative-HTTP migration, expanded to a broad slice of
the documented Close API surface in Pass B (Close's full OpenAPI 3.1 spec, `https://developer.
close.com/openapi.json`, publishes 297 operations — see `api_surface.json` for the itemized
per-endpoint disposition). It reads Close leads, contacts, opportunities, activities, users, tasks,
lead statuses, opportunity statuses, pipelines, roles, groups, and lead/contact/opportunity custom
field definitions; it writes leads, contacts, opportunities, and tasks (create/update/delete)
through the Close REST API (`https://api.close.com/api/v1/...`). This bundle targets capability
parity with `internal/connectors/close-com` (the hand-written connector it migrates) for its
original 5 streams; the legacy package stays registered and unchanged until wave6's registry flip.
The Pass B streams/writes (`tasks`, `lead_statuses`, `opportunity_statuses`, `pipelines`, `roles`,
`groups`, `custom_fields_lead`/`custom_fields_contact`/`custom_fields_opportunity`, and every
`writes.json` action) are new coverage beyond legacy's own scope — legacy never implemented them —
so there is no parity constraint on their record shape.

## Auth setup

Provide a Close API key via the `api_key` secret; it is sent as the HTTP Basic auth username with
an empty password (`"mode": "basic", "username": "{{ secrets.api_key }}", "password": ""`),
matching legacy's `connsdk.Basic(secret, "")` exactly (`close_com.go:242`). It is never logged.
`base_url` defaults to `https://api.close.com/api/v1` and may be overridden for tests/proxies
(legacy's own `closeBaseURL` validates scheme+host the same way; the engine's base-URL resolution
has no equivalent runtime validation, but every conformance fixture only ever points at an
httptest server, so this is not exercised differently on either side).

## Streams notes

All five streams (`leads`, `contacts`, `opportunities`, `activities`, `users`) are simple list
endpoints under Close's singular, trailing-slashed resource paths (`/lead/`, `/contact/`,
`/opportunity/`, `/activity/`, `/user/`), records at the top-level `data` key. Pagination is
Close's `_skip`/`_limit` offset convention (`pagination.type: offset_limit`, `limit_param: _limit`,
`offset_param: _skip`, `page_size: 100` matching legacy's `closeDefaultPageSize`) — the engine's
`offset_limit` paginator stops on a short page (fewer than `page_size` records), which is
DATA-equivalent to legacy's own `has_more != "true" || len(records)==0` stop rule (`close_com.go:
170-176`) for every real Close response shape: a full page always means more records remain, a
short page always means none do (see Known limits for the one theoretical edge case).

Legacy declares a `CursorFields: ["date_updated"]` on every original stream (`streams.go:35` etc.)
and an `InitialState` returning an empty starting cursor (`close_com.go:98-106`), but never actually
wires `date_updated` into any request parameter or client-side filter anywhere in the harvest loop
(`close_com.go:144-180`) — every sync is a full, unfiltered `_skip`/`_limit` walk of the entire
resource regardless of any prior cursor. This bundle mirrors that split: the five legacy schemas
carry schema-level `x-cursor-field: "date_updated"` metadata, but no stream declares an
`incremental` block. None of the Pass B streams below declare an `incremental` block either —
Close's API offers no server-side updated-since filter on any of them.

**Pass B streams**: `tasks` reuses the same `_skip`/`_limit` offset pagination and `data` envelope as
`leads`/`contacts`. `lead_statuses`, `opportunity_statuses`, `pipelines`, `roles`, and `groups` are
all `pagination.type: none` — Close's own OpenAPI spec declares no `_limit`/`_skip` query parameters
on any of these five endpoints (organization-scoped configuration objects Close does not paginate).
`custom_fields_lead`/`custom_fields_contact`/`custom_fields_opportunity` paginate like the CRM-core
streams. **Schema-derivation caveat**: Close's own published OpenAPI spec (`https://developer.close.
com/openapi.json`) declares EMPTY (`{"type":"object","properties":{}}`) response-body schemas for
every one of these Pass B resources — unlike `Lead`/`Contact`/`Opportunity`, which have fully-typed
schemas. Every Pass B schema in this bundle was instead derived from real, cross-referenced facts in
the same spec: `tasks`' fields come from the `CreateTask` discriminated-union request body (the
`_type: "lead"` variant, the most general-purpose task type: `_type`/`lead_id`/`text`/`contact_id`/
`assigned_to`/`due_date`/`is_complete`/etc., all independently documented, required fields
`["_type","lead_id","text"]`); `opportunity_statuses`' `type` field is grounded in the independently
enumerated `OpportunityStatusType` schema (`won`/`lost`/`active`, matching the already-parity-proven
`opportunities` stream's own `status_type` field); `lead_statuses`' `id`/`label` come from the
`InlineLeadStatusPayload` reference schema; `groups`' full shape (`id`/`name`/`organization_id`/
`members[].user_id`) comes from the fully-typed `Group`/`GroupMember` schemas (one of the few Pass B
resources with a complete spec). `pipelines` (`id`/`name`/`organization_id`/`statuses[]`) and `roles`
(`id`/`name`/`organization_id`) use the minimal fields directly confirmed by the spec's query
parameters and Close's own documented organization-scoped-resource convention; a live capture
against a real Close account would be needed to confirm every additional field these two endpoints
return, but the declared fields are all real and none are guessed beyond what the spec/docs support.

## Write actions & risks

`create_lead`/`update_lead`/`delete_lead`, `create_contact`/`update_contact`/`delete_contact`,
`create_opportunity`/`update_opportunity`/`delete_opportunity`, and
`create_task`/`update_task`/`delete_task` are new Pass B writes (legacy never implemented any Close
write path — legacy's own package doc states "no obviously-safe reverse-ETL writes... it exposes no
reverse-ETL writes, so Capabilities.Write is false"; this bundle now supersedes that for these 4
resources). Every action is a live external mutation against the real Close organization; `risk` on
each action requires approval. `create_task` is scoped to Close's `_type: "lead"` task variant only
(the general-purpose case) — Close's `CreateTask` body is a discriminated union keyed by `_type`
(`lead`/`outgoing_call`/etc.), each variant selecting different required fields on the SAME `/task/`
endpoint; only one variant is dialect-expressible per write action (`record_schema` is a single flat
object, not a union), so `outgoing_call` and other task variants are excluded (see
`api_surface.json`). `capabilities.write` is now `true`.

Custom field definitions, lead/opportunity statuses, pipelines, roles, and groups have no write
action in this bundle — all five are workspace-schema/access-control configuration objects with no
demonstrated write demand (deleting or renaming any of them has wide blast radius across every
record referencing them); see `api_surface.json` for the itemized reason on every excluded
create/update/delete endpoint.

## Known limits

- **Declared `CursorFields`/`InitialState` are not modeled as an `incremental` block.** See Streams
  notes above — legacy declares the scaffolding for incremental sync (`CursorFields`,
  `InitialState`) but the harvest loop never actually filters by it, so the bundle preserves only
  the schema-level cursor metadata and does not declare a request-side or client-side incremental
  filter.
- **The engine's `offset_limit` paginator stops on a short PAGE, not on legacy's own `has_more`
  flag.** These two stop rules are DATA-equivalent for every real Close response shape (Close's own
  `has_more` flag directly tracks whether the requested page was full) — the one theoretical edge
  case (the true last page happens to contain EXACTLY `page_size` records, with `has_more: false`)
  still terminates correctly on both sides: legacy stops immediately on `has_more: false`, while the
  engine's paginator would issue one further request that returns zero records and stop via the
  same short/empty-page rule — an extra request, never an extra or missing RECORD.
- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`closePageSize`/`closeMaxPages`, `close_com.go:275-303`). The engine's `offset_limit`
  paginator's `PageSize`/`MaxPages` fields are plain JSON values in `streams.json`, not templated
  against `config.*` — there is no mechanism in this dialect to wire a runtime config value into
  either field. This bundle ships legacy's own default (`page_size: 100`, `max_pages` unbounded) as
  a static value.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`, a credential-free conformance-harness affordance,
  `close_com.go:185-225`) stamps extra fields onto every fixture-mode record — `connector` (a
  static "close-com" marker), `stream` (which stream the record came from), and `previous_cursor`
  (echoing `req.State["cursor"]` when set) — none of which are part of the live record shape. This
  bundle's schemas and fixtures target the live path only; the engine's own conformance/fixture-
  replay harness provides the credential-free test affordance this bundle needs.
