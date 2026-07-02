# Overview

Calendly is a wave1-pilot declarative-HTTP migration (P-4). It reads Calendly scheduled events,
event types, organization memberships, groups, and the current user through the Calendly v2 REST
API. This bundle is engine-vs-legacy parity-tested against `internal/connectors/calendly` (the
hand-written connector it migrates); the legacy package stays registered and unchanged until
wave6's registry flip. Calendly is read-only: `capabilities.write` is `false` and no `writes.json`
is shipped, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Auth setup

Provide a Calendly personal access token or OAuth token via the `api_key` secret; it is used only
for Bearer auth (`Authorization: Bearer <api_key>`) and is never logged.

Every organization-scoped list stream (`scheduled_events`, `event_types`,
`organization_memberships`, `groups`) also requires the `organization_uri` config value — the
authenticated user's Calendly organization URI (e.g.
`https://api.calendly.com/organizations/AAAAAAAAAAAAAAAA`). Resolve it once by calling
`GET https://api.calendly.com/users/me` with your API key and reading
`resource.current_organization`, then configure it here. See "Known limits" below for why this
differs from legacy's behavior.

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

None — calendly is read-only in both legacy and this bundle (`capabilities.write: false`, no
`writes.json`).

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
- Full Calendly v2 API surface (invitees, webhooks, availability schedules, routing forms,
  cancellation) is out of scope for wave1-pilot; see `api_surface.json`'s
  `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}` entries.
- Legacy's fixture-mode credential-free read path (`readFixture`, deterministic synthetic records
  keyed off `mode: fixture`) is a legacy-only affordance and is NOT part of this bundle; this
  bundle's own `fixtures/` directory serves the same credential-free-testing purpose for
  `conformance`'s dynamic checks via bundle-level fixture replay, not a runtime `mode` branch.
