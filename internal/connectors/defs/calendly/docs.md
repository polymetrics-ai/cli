# Overview

Calendly is a wave1-pilot declarative-HTTP migration (P-4), expanded to full documented API-surface
coverage in Pass B. It reads Calendly scheduled events (and their invitees), event types,
organization memberships, groups, routing forms and submissions, webhook subscriptions, availability
schedules, activity log entries, and the current user, and creates/cancels bookings, manages
webhooks/memberships/invitations, and creates event types/shares, through the Calendly v2 REST API.
This bundle targets capability parity with `internal/connectors/calendly` (the hand-written
connector it migrates, which was read-only); the legacy package stays registered and unchanged until
wave6's registry flip. Pass B research verified this bundle against Calendly's real published
OpenAPI 3.1 spec (35 method+path endpoints across 30 paths,
`github.com/api-evangelist/calendly/openapi/calendly-scheduling-api-openapi.yml`); see
`api_surface.json` for the full endpoint-by-endpoint disposition.

## Auth setup

Provide a Calendly personal access token or OAuth token via the `api_key` secret; it is used only
for Bearer auth (`Authorization: Bearer <api_key>`) and is never logged.

Every organization-scoped list stream (`scheduled_events`, `event_types`,
`organization_memberships`, `groups`, `routing_forms`, `webhook_subscriptions`,
`group_relationships`, `activity_log_entries`) also requires the `organization_uri` config value —
the authenticated user's Calendly organization URI (e.g.
`https://api.calendly.com/organizations/AAAAAAAAAAAAAAAA`). Resolve it once by calling
`GET https://api.calendly.com/users/me` with your API key and reading
`resource.current_organization`, then configure it here. See "Known limits" below for why this
differs from legacy's behavior. Two more streams need their own single-resource config value, for
the identical reason: `user_availability_schedules` needs `user_uri` (a specific user's URI —
defaults to the authenticated user if you resolve it the same way as `organization_uri`, via
`resource.uri` instead of `resource.current_organization`), and `routing_form_submissions` needs
`routing_form_uri` (one specific form's URI, resolvable from the `routing_forms` stream's own `uri`
field). Both are optional at the spec level (they gate only their own stream, not the whole
connector) but reads of that specific stream hard-error without them.

## Streams notes

- `scheduled_events` (records at `collection`, cursor field `start_time`): Calendly's scheduled
  (booked) events for the organization. Incremental reads send `min_start_time` (RFC3339, sent
  verbatim — `param_format` defaults to `rfc3339`) computed from the persisted state cursor or,
  on a fresh sync, from the `start_date` config value — this is the ONLY stream that receives a
  server-side lower-bound filter, matching legacy's `harvest`/`incrementalLowerBound` exactly
  (legacy only sets `min_start_time` when `endpoint.resource == "scheduled_events"`; every other
  stream ignores an unknown filter param, so legacy never sends one for them either, and this
  bundle matches that by simply not wiring `incremental.request_param` for those streams).
- `event_types` (cursor field `updated_at`), `organization_memberships` (cursor field
  `updated_at`, computed_fields flatten the nested `user` object into `user`/`user_name`/
  `user_email`), `groups` (cursor field `updated_at`): all three declare an `x-cursor-field` for
  manifest-surface parity with legacy's published `CursorFields`, but — matching legacy exactly —
  receive NO server-side incremental filter and are NOT `client_filtered` either; legacy simply
  never applies any filter for these three streams' `updated_at` cursor, so a full page is always
  requested regardless of prior sync state. This bundle intentionally does not set
  `client_filtered: true` for these streams — doing so would be NEW behavior legacy never had.
- `users` (`single_object: true`, records at `resource`, no incremental — matches legacy, which
  publishes no `CursorFields` for this stream): the authenticated Calendly user (`GET /users/me`).
- **New in Pass B:**
  - `routing_forms` (cursor field `updated_at`) and `routing_form_submissions` (cursor field
    `updated_at`, requires `routing_form_uri` — see Auth setup): routing-form definitions and their
    submitted responses.
  - `webhook_subscriptions` (cursor field `updated_at`): registered webhook endpoints, scoped to the
    organization (`scope=organization` sent as a static query value — the API also supports
    `scope=user`, not modeled as a separate stream in this wave).
  - `user_availability_schedules` (no incremental, unpaginated — Calendly's own response has no
    `pagination` block for this endpoint; requires `user_uri` — see Auth setup): a user's named
    recurring/date-override availability windows.
  - `group_relationships` (cursor field `updated_at`): user-to-group membership links.
  - `activity_log_entries` (cursor field `occurred_at`): THE ONLY new stream with a genuine
    server-side incremental filter — `incremental.request_param: min_occurred_at` sends the
    resolved lower bound directly (Calendly's own documented filter parameter for this endpoint).
  - `invitees` (cursor field `updated_at`): a `fan_out` stream over `scheduled_events` — Calendly has
    no "list every invitee across the account" endpoint; invitees are always scoped to one scheduled
    event (`GET /scheduled_events/{event_uuid}/invitees`). `fan_out.ids_from.request` re-lists
    `scheduled_events` (organization-scoped, same as the `scheduled_events` stream itself) to collect
    every event's `uri` as the fan-out id (`id_field: "uri"`), then `into.path_var` threads each
    event's URI into `invitees`' own `path` as `{{ fanout.id }}`; `stamp_field:
    scheduled_event_id` stamps that same event URI onto every emitted invitee record. Because the
    engine's `fan_out.ids_from.request` has no `query` field of its own (only `path`), the
    organization/count query parameters are embedded directly as a literal query string inside the
    `path` template itself (`/scheduled_events?organization={{ config.organization_uri | urlencode
    }}&count={{ config.page_size }}`) rather than a separate `query` block — `InterpolatePath`
    treats `{{ }}` markers exactly the same wherever they appear in the template string, and
    `connsdk.Requester`'s URL resolution parses the resulting literal `?...` suffix as an ordinary
    query string, so this is a correct, supported (if unusual-looking) use of the existing dialect,
    not a workaround.

Pagination is Calendly's absolute-URL `pagination.next_page` cursor (`pagination.type: next_url`,
`next_url_path: pagination.next_page`) — a `null`/absent `next_page` stops pagination immediately
(SPEC wave1-pilot §4 N3: calendly returns ABSOLUTE next-page URLs, so the engine's same-host SSRF
guard never blocks legitimate pagination here; no `allow_cross_host` override is needed). The
`users` single-object stream (`GET /users/me`) declares an explicit stream-level
`"pagination": {"type": "none"}` override rather than inheriting the collection streams'
`next_url` paginator, matching its real (unpaginated, single-object) response shape.

`count={{ config.page_size }}` is sent on every organization-scoped list request. `page_size` is
NOT required: leaving it unset resolves to spec.json's declared `"default": "100"`, materialized
into `RuntimeConfig.Config` at runtime by the engine's config-default mechanism
(`engine/read.go`'s `materializeConfigDefaults`) before query templating runs — matching legacy's
`calendlyPageSize` default-100 fallback (`calendly.go:363-376`) exactly, for the identical (unset)
input, without a hard error.

## Write actions & risks

This bundle adds write support beyond legacy (which was read-only); `capabilities.write` is now
`true` and `writes.json` declares 8 actions:

- **`cancel_scheduled_event`** (`POST /scheduled_events/{{ record.uuid }}/cancellation`,
  `path_fields: ["uuid"]`): cancels a real scheduled event. **Risk: notifies invitees; approval
  required.**
- **`create_invitee`** (`POST /invitees`, nested `invitee` object): books a new meeting on an event
  type at a specific `start_time`. **Risk: books a real slot and notifies the invitee; approval
  required.**
- **`create_webhook_subscription`**/**`delete_webhook_subscription`** (`POST /webhook_subscriptions`
  requires `url`/`events`/`organization`/`scope`; `DELETE /webhook_subscriptions/{{ record.uuid }}`).
  **Risk: a new subscription receives live invitee/routing-form event payloads at an operator-supplied
  URL; approval required.**
- **`remove_organization_membership`** (`DELETE /organization_memberships/{{ record.uuid }}`).
  **Risk: destructive — revokes a real user's access to the organization; approval required.**
- **`invite_user_to_organization`** (`POST /organizations/{{ record.organization_uuid
  }}/invitations`, `body_fields: ["email"]` since the org uuid lives only in the path). **Risk:
  sends a real invitation email; approval required.**
- **`create_one_off_event_type`** (`POST /one_off_event_types`, requires
  `name`/`host`/`duration`/`date_setting`). **Risk: publishes a new publicly-bookable event type;
  approval required.**
- **`create_share`** (`POST /shares`, requires `event_type`). **Risk: creates a new shareable
  booking link with its own spot limit; approval required.**

An in-place webhook update (`PATCH /webhook_subscriptions/{uuid}`) is not published by Calendly's
API at all (only create/list/delete exist) — there is nothing to model beyond create+delete.

## Known limits

- **Organization scoping is config-driven, not auto-discovered (documented parity deviation,
  conventions.md §5 ledger).** Legacy resolves the `organization` query param DYNAMICALLY on every
  single read by first calling `GET /users/me` and reading `resource.current_organization`
  (`calendly.go`'s `currentUser`/`scopeQuery`). The engine's declarative dialect has no mechanism
  to chain one request's response into a later request's query params (`read.go`'s
  `buildInitialQuery` only resolves `config.*`/`secrets.*`/`record.*`/`cursor` templates against
  the READ REQUEST's own inputs, never a prior response body) — expressing this would require a
  `StreamHook` (Tier 2), which SPEC wave1-pilot §5.2 does not call for calendly. This bundle
  instead asks the operator to configure `organization_uri` once (the exact, per-account-invariant
  value legacy would have discovered via `/users/me` at read time). Every subsequent request both
  connectors send is byte-identical given the same organization URI — this never changes any
  emitted record's DATA for any input legacy itself would accept, it only changes WHEN/HOW the
  (invariant) organization URI is supplied. If a future wave needs true auto-discovery (e.g. to
  support switching organizations without a config update), that is a `StreamHook` escalation, not
  an engine dialect change.
- **The `id` primary-key convenience field IS reproduced** (gap-loop cycle-1 fix, REVIEW-B.md
  finding 1/adjudication 1 — this item previously documented `id` as a NOT-reproduced deviation;
  that was superseded by the `last_path_segment` engine filter and is corrected here rather than
  left stale). Legacy derives `id` from `uri`'s trailing path segment (`idFromURI`) on every
  record; every stream here does the identical derivation via
  `"id": "{{ record.uri | last_path_segment }}"` in `computed_fields`, and every schema declares
  `x-primary-key: ["id"]` — matching legacy's published primary key exactly, byte-for-byte, for
  every input legacy itself would accept.
- `event_types`/`organization_memberships`/`groups` publish an `x-cursor-field` (`updated_at`,
  matching legacy's published `CursorFields`) but have NO server-side incremental filtering and
  are NOT `client_filtered` — matching legacy's actual (lack of) filtering behavior for these
  three streams exactly, not a bundle-authoring gap.
- **`user_availability_schedules` and `routing_form_submissions` are `conformance: {skip_dynamic:
  true}`** at the stream level: both require a config value (`user_uri`/`routing_form_uri`
  respectively) naming one specific real Calendly resource that `conformance`'s synthetic
  non-secret config value (`"synthetic-conformance-value"` for every declared spec property) cannot
  meaningfully represent — the same class of limitation `organization_uri` itself would have if this
  bundle's very first (wave1-pilot) fixtures had not already been authored against it. Both streams'
  declarative shape (an ordinary `collection`+pagination or `single_object`-equivalent stream) is
  identical to every other stream in this bundle and is proven correct by static validation
  (`connectorgen validate`'s `checkInterpolations`) plus the shared, already-proven pattern; there is
  no hook or engine-side gap here, purely a fixture-authoring/conformance-harness limitation for a
  per-resource-specific required filter value.
- **`invitees` is `conformance: {skip_dynamic: true}`** at the stream level: it is a `fan_out`
  stream, and `conformance`'s dynamic (fixture-replay) checks assume one fixed request path per
  stream — they have no mechanism to replay a two-phase "list ids, then fan out" sequence. The
  `fan_out` mechanism itself (id-listing request, `path_var` threading, `stamp_field`) is not new —
  it is the same engine mechanism other already-migrated bundles use — and is proven here by static
  validation of `fan_out.ids_from.request.path`'s interpolation (identical `ResolveCheck` coverage
  `stream.path` gets) plus the shared, already-proven pattern.
- Full Calendly v2 API surface beyond what is listed above (single-resource detail-fetch duplicates,
  GDPR data-compliance deletion, pending-invitation management, point-in-time availability
  calculations) is out of scope; see `api_surface.json`'s per-endpoint `excluded` categories and
  reasons.
- Legacy's fixture-mode credential-free read path (`readFixture`, deterministic synthetic records
  keyed off `mode: fixture`) is a legacy-only affordance and is NOT part of this bundle; this
  bundle's own `fixtures/` directory serves the same credential-free-testing purpose for
  `conformance`'s dynamic checks via bundle-level fixture replay, not a runtime `mode` branch.
