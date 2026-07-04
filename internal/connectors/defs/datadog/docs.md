# Overview

Datadog reads monitors, dashboards, dashboard lists, users, SLOs, SLO corrections, scheduled
downtimes, notebooks, organizations, hosts, Synthetics tests/locations/variables, and API/
application keys, and writes monitor/dashboard/downtime/notebook/SLO/user/event/Synthetics-test/
API-key mutations, through the Datadog API v1 — the full 5-stream legacy-parity surface of
`internal/connectors/datadog` (the legacy hand-written connector, which stays registered and
unchanged until wave6's registry flip) plus, as of this Pass B full-surface expansion, every other
practical v1 catalog resource and dialect-expressible mutation. See `api_surface.json` for the
endpoint-by-endpoint accounting against Datadog's own published OpenAPI v1 spec
(`github.com/DataDog/datadog-api-client-go/.generator/schemas/v1/openapi.yaml`).

## Auth setup

Provide `api_key` and `application_key` secrets; both are sent as raw header values —
`DD-API-KEY: <api_key>` and `DD-APPLICATION-KEY: <application_key>` — matching legacy's
`DefaultHeaders` map exactly (neither is a Bearer/Basic scheme, so `streams.json`'s `base.headers`
declares them directly rather than via an `auth` candidate; both secrets are required, so an absent
value is always a hard error per the engine's secrets-in-headers rule, matching legacy's own
explicit `errors.New(...)` checks). Neither value is ever logged.

## Streams notes

The original 5 streams (`monitors`, `dashboards`, `users`, `slo`, `downtimes`) are unchanged from
the wave2 migration; see the per-stream pagination/incremental notes already documented below.

New in this Pass B expansion (9 additional streams, all read-only catalog resources):

- **`dashboard_lists`** (`GET /api/v1/dashboard/lists/manual`, records at `dashboard_lists`) —
  unpaginated, matching the real endpoint's shape (no `limit`/`offset`/page params documented).
- **`notebooks`** (`GET /api/v1/notebooks`, records at `data`) — JSON:API envelope
  (`{id, type, attributes: {...}}`); `computed_fields` lifts `name`/`created`/`modified` out of
  `attributes` and `author_handle` out of the nested `attributes.author.handle`. Pagination is
  `offset_limit` with `offset_param: start`, `limit_param: count` (the real endpoint's own
  parameter names — an index-of-first-result offset and a result-count limit, not Datadog's more
  common `page`/`page_size` naming).
- **`organizations`** (`GET /api/v1/org`, records at `orgs`) — unpaginated (every managed
  organization for the account is returned in one call); primary key is `public_id` (Datadog
  organizations have no numeric `id` field).
- **`hosts`** (`GET /api/v1/hosts`, records at `host_list`) — pagination is `offset_limit` with
  `offset_param: start`, `limit_param: count` (the same offset/limit shape as `notebooks`, distinct
  param names from either).
- **`slo_corrections`** (`GET /api/v1/slo/correction`, records at `data`) — JSON:API envelope;
  `computed_fields` lifts `slo_id`/`category`/`description`/`start`/`end`/`duration`/`timezone`/
  `created_at`/`modified_at` out of `attributes`. Pagination is `offset_limit` with
  `offset_param: offset`, `limit_param: limit` (the real endpoint's own documented param names).
- **`synthetics_tests`** (`GET /api/v1/synthetics/tests`, records at `tests`) — pagination is
  `page_number` with `page_param: page_number`, `size_param: page_size`, `start_page: 0` (the real
  endpoint's own docs state "Starts at zero" for `page_number`, the same 0-based convention already
  established for `monitors`/`slo`/`users`). Primary key is `public_id` (Synthetics tests have no
  numeric `id` field, unlike monitors/dashboards).
- **`synthetics_locations`** (`GET /api/v1/synthetics/locations`, records at `locations`) —
  unpaginated; returns every public and private test-execution location.
- **`synthetics_variables`** (`GET /api/v1/synthetics/variables`, records at `variables`) —
  unpaginated; the raw API's `value.value`/`value.secure` fields (the variable's actual stored
  value and whether it is marked secure) are deliberately NOT declared in this stream's schema —
  see Known limits.
- **`api_keys`** (`GET /api/v1/api_key`, records at `api_keys`) — unpaginated; primary key is `key`
  (the API key's own value IS its identifier in this API, not a separate opaque id).
- **`application_keys`** (`GET /api/v1/application_key`, records at `application_keys`) —
  unpaginated; primary key is `hash` (a fixed-length hash of the key value, the closest thing to an
  id this resource has — the real key VALUE itself is never returned by the list endpoint, only by
  the create response).

`GET /api/v1/events` (Datadog's event-stream read) and `GET /api/v1/tags/hosts` (a tag-to-hostnames
map) are NOT migrated as streams — both are genuine `ENGINE_GAP`s, not scoping choices; see Known
limits.

## Write actions & risks

20 write actions now cover every dialect-expressible Datadog v1 mutation:

- **Monitors**: `create_monitor`/`update_monitor`/`delete_monitor`.
- **Dashboards**: `create_dashboard`/`update_dashboard`/`delete_dashboard` (update replaces the
  dashboard's full widget layout — the real API's `PUT` is a whole-resource replace, not a patch).
- **Dashboard lists**: `create_dashboard_list`/`update_dashboard_list`/`delete_dashboard_list`.
- **Downtimes**: `create_downtime`/`update_downtime`/`cancel_downtime` (`cancel_downtime` maps to
  the real API's `DELETE /api/v1/downtime/{downtime_id}`, which Datadog itself calls "cancel", not
  "delete" — downtime records are never truly erased, only deactivated).
- **Notebooks**: `create_notebook`/`update_notebook`/`delete_notebook` (update replaces the full
  cell/time definition — the real API's `PUT` is a whole-resource replace).
- **SLOs**: `create_slo`/`update_slo`/`delete_slo`.
- **Users**: `create_user` (invites a new user), `update_user`, `disable_user` (Datadog has no true
  user delete — accounts are disabled, never erased, matching the real API's `DELETE` semantics
  exactly).
- **Events**: `create_event` (posts a custom event into the event stream) — the corresponding read
  (`GET /api/v1/events`) is out of scope (see Known limits), but posting a NEW event needs no time
  window and is fully dialect-expressible.
- **Synthetics**: `create_synthetics_api_test`/`update_synthetics_api_test` (API-type tests only —
  browser/mobile tests require a recorded step sequence from Datadog's own visual test recorder,
  with no practical declarative-record equivalent; see `api_surface.json`).
- **API keys**: `create_api_key`/`update_api_key`/`delete_api_key`.

`capabilities.write` is now `true` (previously `false`); `metadata.json`'s `risk.write`/
`risk.approval` document per-action risk tiers — every delete/cancel action and every write with a
live-alerting or access-control side effect (monitor/downtime/SLO/user updates, downtime/user
creation) requires approval; pure-creation/rename actions with no live-alerting impact are
low-risk.

## Known limits

- A `site`-derived `base_url` (legacy's `datadogBaseURL` builds `https://api.<site>` from a `site`
  config value, e.g. `datadoghq.eu`, when `base_url` itself is unset) is not modeled: the engine's
  `spec.json` `"default"` mechanism only materializes a fixed literal, not one derived from another
  config field (conventions.md §3, the sentry/chargebee derived-default case). Set `base_url`
  directly to the regional host (e.g. `https://api.datadoghq.eu`) instead of a bare `site` value.
- **`GET /api/v1/events` is an `ENGINE_GAP`, not a scoping choice**: the real endpoint requires BOTH
  `start` and `end` as mandatory POSIX-timestamp query parameters. `end`'s real-world value is
  always "now" at request time — there is no fixed or config-derivable value for it — and the
  engine's template `Vars` environment has no resolvable "current time" reference at all (`grep`
  confirms no `time.Now()`-equivalent anywhere in `engine/{interpolate,read}.go`), the same gap
  documented for the datascope bundle's windowed `answers`/`notifications` streams in this same
  wave. `create_event` (posting a NEW event) needs no time window and is fully implemented as a
  write action; only the READ side is blocked.
- **`GET /api/v1/tags/hosts` is an `ENGINE_GAP`, not a scoping choice**: the real response shape is
  `{"tags": {"<tag-name>": ["host1", "host2", ...]}}` — a map whose VALUES are ARRAYS of hostnames,
  not objects. The engine's `records.keyed_object` flag explodes a keyed object's OBJECT-valued
  entries into records; it has no equivalent for a keyed object whose values are arrays, and there
  is no other declarative way to flatten an arbitrary-cardinality tag-to-hostnames map into stable
  schema properties without a `RecordHook` (a 3rd hook interface this Tier-1 bundle does not have,
  and does not need for anything else).
- **`synthetics_variables`' raw `value.value`/`value.secure` fields are deliberately not declared
  in the read-side schema**: the real API's list response DOES include the variable's actual
  stored value (Datadog itself marks some variables `secure`/masks them server-side, but the field
  is present in the raw JSON either way) — emitting a field that may carry a live credential value
  into a destination warehouse is exactly the record-data credential-exfiltration risk the engine's
  `computed_fields`' secrets exclusion (conventions.md §3) is designed to prevent at the
  computed-fields layer; this bundle extends the same caution to plain schema projection for this
  one field by simply never declaring it, rather than declaring and then hoping a downstream
  consumer treats it carefully. `create`/`update`/`delete` for Synthetics variables are correctly
  NOT implemented as writes this pass for the same reason (see `api_surface.json`) — a future
  increment modeling a write-only-credential write action (the same pattern the dbt bundle's
  `create_ssh_tunnel`/`update_ssh_tunnel` `private_key` field already establishes) would be the
  correct way to add variable writes without this exposure.
- The v1-vs-v2 API boundary: Datadog's v2 API is a separate, vastly larger (~1383-operation)
  surface spanning entirely distinct product areas this connector's `docs_url`/`base_url` never
  targeted — Security Monitoring, Cloud Cost Management, Case Management, LLM Observability,
  Incidents, Teams, Static Analysis, Status Pages, Feature Flags, App Builder, On-Call, and
  identity/org-administration surfaces (Org Groups, Roles, Service Accounts, Restriction Policies).
  This bundle stays v1-only, matching its own `spec.json` `base_url` default
  (`https://api.datadoghq.com`, the same host both API versions share, but this bundle's `streams`/
  `writes` paths are exclusively `/api/v1/...`) and legacy's own v1-only implementation; the v2
  surface is out of scope for this pass, not individually enumerated in `api_surface.json` (which
  only lists the target v1 surface this connector actually addresses, the same convention the dbt
  bundle uses for its v2-vs-v3 boundary in this same wave).
- Full per-endpoint reasoning for every excluded v1 endpoint (integration credential management,
  usage/billing metering, log-pipeline administration, metrics/timeseries query surfaces, public
  dashboard sharing, Synthetics browser/mobile test authoring, and more) is in `api_surface.json`'s
  `excluded` entries — every one carries a specific category and reason, no blanket bucket.
