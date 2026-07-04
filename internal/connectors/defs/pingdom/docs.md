# Overview

Pingdom originated as a read-only declarative migration of `internal/connectors/pingdom` (legacy Go
connector), and is expanded to the full documented API 3.1 surface in Pass B. It reads 10 Pingdom
resources (checks, probes, actions, maintenance windows, maintenance occurrences, alerting
contacts, alerting teams, account credits, transaction checks, and reference data) and writes 11
practical mutations (check/contact/team/maintenance-window create/update/delete) through Pingdom's
REST API 3.1. The 5 original streams stay capability-parity with legacy; legacy stays registered
and unchanged until wave6's registry flip.

## Auth setup

Provide a Pingdom API 3.1 Bearer token via the `api_key` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_key>`) and is never logged. There is no unauthenticated fallback mode —
legacy hard-errors when `api_key` is unset (`pingdom connector requires secret api_key`), matching
this bundle's Bearer `auth` candidate and `api_key`'s `x-secret` presence check at request time.

## Streams notes

All 5 streams (`checks`, `probes`, `actions`, `maintenance`, `reference`) share the identical
`offset_limit` pagination shape (`limit`/`offset` query params, `page_size: 100` matching legacy's
`defaultPageSize`), and all pass through the RAW record with no field-filtering — legacy's `Read`
calls `connsdk.Harvest` directly with no `mapRecord`-style projection function at all, so every
stream declares `"projection": "passthrough"` to preserve that exact behavior. None of the five
streams declares an `incremental` block or an `x-cursor-field`: legacy never filters or advances
reads by any timestamp field on any of these resources (`lasttesttime`/`time` are ordinary
passthrough fields, not cursors) — every read is a full stream read, matching the sync-mode
derivation rule (an `x-cursor-field` with no backing `incremental` block would misrepresent
capability this connector does not have).

`checks`/`probes`/`actions`/`maintenance` extract records from their eponymous top-level array key
(`records.path: "checks"`, etc.). `reference` is different: legacy's `streamEndpoints["reference"]`
declares an EMPTY `recordsPath` (`""`), so `connsdk.RecordsAt` wraps the entire response body as a
single record (Pingdom's `/reference` endpoint returns one lookup bundle — check types, regions,
probe metadata — not a keyed collection). This bundle mirrors that exactly via
`"records": {"path": ".", "single_object": true}`; `schemas/reference.json` declares no
`x-primary-key` since there is no natural per-record identity for a singleton reference-data dump.

The 5 Pass-B-added streams all share the identical passthrough shape. `alerting_contacts`
(`GET /alerting/contacts`) and `alerting_teams` (`GET /alerting/teams`) declare
`"pagination": {"type": "none"}` at the stream level (overriding the base `offset_limit` spec
wholesale) — Pingdom's own OpenAPI 3.1 document declares no `limit`/`offset` query parameters for
either endpoint, and both return their full collection (`contacts[]`/`teams[]`) in one response.
`maintenance_occurrences` (`GET /maintenance.occurrences`) is likewise unpaginated in Pingdom's own
spec (its only query parameters are `maintenanceid`/`from`/`to` filters, none of which this bundle
wires); each occurrence record is one already-materialized instance of a recurring maintenance
window (`id`, `maintenanceid` linking back to the parent window, `from`/`to` Unix timestamps).
`credits` (`GET /credits`) is a second `single_object`-shaped stream like `reference`: Pingdom
returns one account-wide credits/limits bundle under a top-level `credits` key
(`"records": {"path": "credits", "single_object": true}`), not a collection — `schemas/credits.json`
likewise declares no `x-primary-key`. `tms_checks` (`GET /tms/check`) is Pingdom's transaction
(browser-script) check resource, a genuinely distinct business object from the uptime `checks`
stream (its own `id`/`name`/`type`/`interval`/`region` shape, `type` is `script` or `recording`
rather than an HTTP/TCP/DNS/etc. protocol) — it shares `checks`'/`probes`' `offset_limit`
pagination shape (Pingdom's spec declares the identical `limit`/`offset` parameters here).

## Write actions & risks

11 write actions: `create_check`/`update_check`/`delete_check`,
`create_contact`/`update_contact`/`delete_contact`, `create_team`/`update_team`/`delete_team`, and
`create_maintenance`/`delete_maintenance`. `create_check`'s `record_schema` models the common
HTTP-type check shape (`name`/`host`/`type`/`paused`/`resolution`/notification settings/`tags`);
Pingdom's real `POST /checks` body is a documented `oneOf` across 9 check types
(http/httpcustom/tcp/ping/dns/udp/smtp/pop3/imap), each with its own type-specific attributes on top
of the shared fields — the engine's draft-07 dialect has no `oneOf`, so only the shared-fields-plus-
HTTP-type shape is modeled here (the single most commonly used check type); see Known limits.
`update_contact`'s `record_schema` requires `name`+`paused`+`notification_targets` together because
Pingdom's `PUT /alerting/contacts/{contactid}` is a full-replacement update, not a partial patch —
this is Pingdom's own documented behavior, not an engine limitation. `delete_check`/`delete_contact`/
`delete_team`/`delete_maintenance` are NOT declared `missing_ok_status`-idempotent: Pingdom's OpenAPI
document only formally documents a `200` response for each delete operation (no documented `404`
shape to treat as idempotent-success), so this bundle does not invent that tolerance.

Per `metadata.json`'s `risk.approval`: `create_check`/`create_contact`/`create_team`/
`create_maintenance` require no approval (low-risk, non-destructive); every `update_*`/`delete_*`
action requires approval.

## Known limits

- `page_size` config validation (legacy's 1-25000 numeric-range check) is not reproduced at the
  bundle level; the engine treats `page_size` as an opaque value substituted directly into the
  `limit` query param, sent to Pingdom as-is rather than rejected client-side the way legacy's
  `strconv.Atoi` range check would. This never changes emitted record DATA for any legacy-valid
  input; it only narrows client-side input validation, out of scope for wave2 fan-out (Pass B).
- Legacy's `max_pages` config (a non-negative integer, or the keywords `all`/`unlimited` for
  unbounded) has no bundle-level equivalent — this bundle relies solely on the offset paginator's own
  short-page stop signal (no `MaxPages` hard-cap declared), matching legacy's own unbounded default
  (`maxPages() == 0` when `max_pages` is unset) exactly for the common case. Out of scope for wave2
  fan-out (Pass B).
- The 2-page conformance fixture (`fixtures/streams/checks/page_1.json` /`page_2.json`) uses a
  synthetic 100-record first page purely to exercise the real `offset_limit` short-page stop
  condition against the bundle's real `page_size: 100` default (the engine's `OffsetPaginator` stops
  only when a page returns fewer records than `page_size`); this is a fixture-authoring artifact
  proving pagination correctness, not a claim about Pingdom's actual typical result-set size.
- **`create_check` only models the HTTP check-type shape.** Pingdom's `POST /checks` accepts a
  `oneOf` across 9 check types (http/httpcustom/tcp/ping/dns/udp/smtp/pop3/imap), each type carrying
  its own additional type-specific attributes (e.g. `encryption`/`port`/`url` for http,
  `stringtosend`/`stringtoexpect` for tcp, `expectedip` for dns) on top of the shared
  name/host/type/paused/resolution/notification fields. The draft-07 dialect this engine uses has
  no `anyOf`/`oneOf`, so representing all 9 variants in one `record_schema` is not expressible;
  `create_check` models the shared fields plus the `http` type (Pingdom's default and most common
  check type). Creating a non-HTTP check type through this action is out of scope for this pass.
- **Bulk check/maintenance-window/maintenance-occurrence mutation endpoints are not modeled.**
  `PUT /checks`, `DELETE /checks`, `DELETE /maintenance`, and `DELETE /maintenance.occurrences` all
  operate on an explicit id-list request body (pause/resolution-change or delete for several
  records in one call) rather than a single record; this dialect's write actions are single-record
  mutations (`record.id`-keyed path/body), so a bulk id-list action has no natural single-record
  shape to declare. The per-record equivalents (`update_check`/`delete_check`/`delete_maintenance`)
  cover the same outcome one record at a time.
- **Transaction (TMS) check create/update/delete and in-place maintenance-window/occurrence update
  are out of scope.** `POST /tms/check`/`PUT /tms/check/{cid}` require authoring or editing a
  multi-step browser-script check definition — a payload shape this pass did not model;
  `DELETE /tms/check/{cid}` is accordingly also not covered (no safer alternative to pair it with).
  `PUT /maintenance/{id}`/`PUT /maintenance.occurrences/{id}` (in-place edits of an existing window
  or one already-materialized occurrence) are narrower than the create-and-replace lifecycle
  `create_maintenance`/`delete_maintenance` already cover and were deprioritized for this pass.
- **Reporting/analysis/on-demand-test endpoints are out of scope by design, not merely
  deprioritized.** `/analysis/{checkid}(/{analysisid})`, `/results/{checkid}`, `/single`,
  `/traceroute`, every `/summary.*` report, and `/tms/check/report/status`
  (`/tms/check/{check_id}/report/{status,performance}`) are all read-only diagnostic/aggregated-
  reporting endpoints or on-demand live-test triggers with no independent, stable business-object
  identity of their own — see `api_surface.json` for each endpoint's specific `non_data_endpoint`
  reasoning.
