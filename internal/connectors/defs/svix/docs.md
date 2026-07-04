# Overview

Svix is a Tier-1 declarative-HTTP connector for the Svix REST API v1
(`https://api.svix.com/api/v1`; real published OpenAPI 3.1 spec at
`https://api.svix.com/api/v1/openapi.json`). This is a Pass B full-surface expansion against that
spec: 7 read streams (`applications`, `endpoints`, `event_types`, `messages`, `background_tasks`,
`connectors`, `operational_webhook_endpoints`) and 16 write actions covering the
application/endpoint/event-type/connector/operational-webhook-endpoint lifecycle plus outgoing
message creation — the core surface an outbound-webhooks-as-a-service account actually operates on
day to day. See `api_surface.json` for the full endpoint-by-endpoint disposition of the other ~100
documented operations (message-attempt/delivery-log detail, endpoint secret/header/transformation
sub-resources, and the newer Stream/Poller/Ingest beta product lines).

**Tier justification**: a plain declarative-HTTP bundle. Auth is a static Bearer token (Svix's own
OpenAPI spec declares `HTTPBearer`), pagination is the engine's `cursor` type end to end, and every
sub-resource stream is an ordinary single-level `fan_out` over `applications` — nothing needs a Go
hook.

## Auth setup

Provide a Svix API key via the `api_key` secret; it is sent as a Bearer token (`Authorization:
Bearer <api_key>`) and is never logged. `base_url` defaults to `https://api.svix.com/api/v1`. Svix
is multi-region (US/EU/CA/AU/IN, per the OpenAPI spec's `servers` list); an account provisioned
outside the default region must override `base_url` with the matching regional host (e.g.
`https://api.eu.svix.com/api/v1`).

## Streams notes

All 7 streams share the base's `cursor` pagination (`cursor_param: iterator`, `token_path:
iterator`) — every Svix list endpoint returns the identical `{data, iterator, done}` envelope.
Svix's real stop signal is the boolean `done` field, not merely an empty `iterator` (the OpenAPI
spec marks `iterator` `nullable` but the value on a final page is not guaranteed to be an empty
string — Svix's own SDKs treat `done` as authoritative). The engine's `stop_path` mechanism cannot
express this directly: `stop_path` stops on a FALSY body value (i.e. it names a "more pages exist"
field), while Svix's `done` means the opposite (`true` = no more pages) — there is no negation
option in the pagination dialect for a stop-when-truthy field. This bundle instead relies on the
`token_path` cursor paginator's built-in same-token-twice loop guard (`tokenPathCursor`,
`docs/migration/conventions.md` §3): if a final page repeats its own `iterator` value (the
documented Svix behavior), the second identical request is detected and pagination stops safely
even without inspecting `done`. Every stream's fixtures set `iterator: ""` on the final page (the
common case, and the one this repo's own conformance harness needs to terminate cleanly), so this
gap does not affect fixture-replay correctness; it is a real, documented gap for the corner case
where Svix returns a non-empty repeating iterator on the last page (`ENGINE_GAP`: the pagination
dialect has no stop-when-truthy variant).

`applications` (`GET /app`) lists every application. Its emitted schema is intentionally narrowed
to legacy's fixed projection (`id`, `name`, `created_at`), with `created_at` filled from either the
raw `created_at` or Svix's `createdAt` camelCase field. Legacy published no cursor field for this
stream, so the schema declares none.

`endpoints` (`GET /app/{app_id}/endpoint`) and `messages` (`GET /app/{app_id}/msg`) both `fan_out`
over every application id (`ids_from.request`: `GET /app`, `records_path: data`, `id_field: id`),
stamping `app_id` onto every emitted record.

`event_types` (`GET /event-type`) sends `with_content=true` and `include_archived=true` statically
so the emitted records carry the full JSON-schema payload (`schemas`) and include archived
(soft-deleted) event types, matching what an operator managing event-type lifecycle needs to see.

`background_tasks` (`GET /background-task`) and `connectors` (`GET /connector`) and
`operational_webhook_endpoints` (`GET /operational-webhook/endpoint`) are plain top-level list
streams with no fan-out.

Message-attempt/delivery-log data (`/app/{app_id}/attempt/endpoint/{endpoint_id}`,
`/app/{app_id}/endpoint/{endpoint_id}/msg`, `/app/{app_id}/msg/{msg_id}/endpoint`, and related
per-attempt detail) is nested two-to-three levels deep (app_id, then endpoint_id or msg_id, then
attempt_id) — the engine's `fan_out` dialect resolves exactly one level of parent ids per stream;
a second nested fan-out level cannot be expressed in Tier 1 without inventing a `StreamHook`, which
`docs/migration/conventions.md` §1 reserves for genuinely un-expressible shapes, not a convenience
shortcut. This is a real, documented `requires_elevated_scope` gap (`api_surface.json`), not a
silently-dropped capability.

## Write actions & risks

- `create_application`/`update_application`/`delete_application` — application lifecycle.
  `delete_application` irreversibly removes an application and all its endpoints, messages, and
  delivery history; approval required.
- `create_endpoint`/`update_endpoint`/`delete_endpoint` — webhook delivery endpoint lifecycle on an
  application. `create_endpoint` immediately starts receiving future events matching its filters;
  `update_endpoint` changing `url` redirects all future deliveries; `delete_endpoint` stops all
  future deliveries to that endpoint. Approval required for the destructive delete; the create/update
  mutations are lower-risk (no approval required) since they only affect future deliveries, not
  already-sent data.
- `create_event_type`/`update_event_type`/`delete_event_type` — event type definition lifecycle.
  `delete_event_type` is Svix's own soft-delete (archive), not a hard delete.
- `send_message` — sends a REAL outgoing webhook message that Svix immediately attempts to deliver
  to every matching endpoint on the application; approval required (this is the highest-risk action
  in this bundle — it is not a metadata mutation, it triggers real external HTTP delivery attempts).
- `create_connector`/`update_connector`/`delete_connector` — payload-transformation-template
  lifecycle; `update_connector` changes the payload shape delivered to every endpoint using that
  connector.
- `create_operational_webhook_endpoint`/`update_operational_webhook_endpoint`/
  `delete_operational_webhook_endpoint` — account-level operational-event (e.g.
  `message.attempt.exhausted`) webhook endpoint lifecycle, parallel to the per-application `endpoint`
  actions above but scoped to Svix's own account-level ops events rather than application messages.

All updates use `PUT` (full-body replace), matching Svix's own `EndpointUpdate`/`ApplicationIn`-
shaped full-replacement semantics, rather than `PATCH` (Svix's partial-update variant) — the
engine's write dialect sends the record's full field set by default, which is the `PUT` contract;
the `PATCH` operations for each of these resources are excluded in `api_surface.json` as
`duplicate_of` since the `PUT` action is a strict superset of what a partial-update caller needs.

## Known limits

- **Nested fan-out (message attempts / per-endpoint message lists) is not modeled** — see Streams
  notes above (`requires_elevated_scope`, 2-3 levels of path nesting the single-level `fan_out`
  dialect cannot express).
- **`done`-as-stop-signal is not directly wired** — pagination relies on the `token_path`
  paginator's same-token-twice loop guard instead of Svix's own `done` boolean; see Streams notes.
  This never causes a real sync to loop forever (the loop guard is unconditional), but a
  pathological API response repeating a non-final page's iterator would not be distinguished from
  a genuine last page by this mechanism alone.
- Endpoint/operational-webhook-endpoint HMAC secrets, custom headers, and per-endpoint inline
  transformation JS are excluded (`requires_elevated_scope`/`duplicate_of`) — none of these are
  ordinary list-shaped or create/update/delete-shaped resources this dialect's stream/write model
  fits cleanly; see `api_surface.json`.
- The newer Stream/Poller/Ingest-source product surface (inbound webhook receiving and a
  separately-versioned event-streaming product) is entirely out of scope for this pass — it is a
  distinct product line from the classic outgoing-webhooks surface this bundle covers.
