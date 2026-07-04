# Overview

Thrive Learning is a wave2 fan-out declarative-HTTP migration of
`internal/connectors/thrive-learning` (the hand-written legacy connector this bundle migrates; the
legacy package stays registered and unchanged until wave6's registry flip), since Pass B expanded
to the full documented Thrive Public API surface at `docs.thrivelearning.com`. It still reads the 3
original legacy-parity streams (`users`, `content`, `completions`) through the legacy connector's
own simplified endpoint (`GET https://api.thrivelearning.com/{users,content,completions}`) exactly
as before, and additionally reads the REAL, currently-documented Thrive Public API v1
(`https://public.api.learn.link/rest/v1`): activities, contents, learning completions, assignments
and their enrolments, audiences and their members/managers, tags, CPD
categories/entries/requirement-summaries, and skill levels. Still read-only
(`capabilities.write: false`) — see Known limits for why the real API's documented write surface
could not be safely added this pass.

**Important discovery (see `api_surface.json`'s `scope` note for the full account):** the legacy
connector's 3 endpoints do not appear anywhere in the real, currently-documented Thrive API at all —
`api.thrivelearning.com` is not a host the real docs ever mention, and the real API's own `/users`,
`/contents`, `/completions`-equivalent resources (`/rest/v1/users`, `/rest/v1/contents`,
`/rest/v1/learning/completions`) have entirely different field shapes, pagination conventions
(`page`/`perPage`, not `page`/`limit`), and response envelopes. This bundle does NOT touch the 3
original streams (they are legacy parity, unaffected by this discovery) and instead adds every new
Pass B read resource against the real API on its own real host, exactly as documented.

## Auth setup

Provide a tenant username via the `username` config value and an API password via the `password`
secret; both are sent via HTTP Basic auth (`auth: [{"mode": "basic", ...}]`), matching legacy's
`connsdk.Basic(username, password)` (`thrive_learning.go:121`) AND the real, currently-documented
API's own recommended Basic-auth shape (tenant-id as username, API secret as password —
`docs.thrivelearning.com/apidocs/authentication`) — the same credentials authenticate both the
legacy-parity streams and every new Pass B stream, no separate credential is needed. `password` is
never logged (`x-secret: true`). This bundle sources `password` from `secrets.password` only;
legacy's `secret()` helper additionally falls back to `cfg.Config["password"]` when the secret is
unset — that fallback is not modeled here (a genuine, narrow scope reduction: a caller relying on a
config-plaintext password rather than the secret store would need to migrate to the secret,
documented in Known limits). `base_url` (legacy's own simplified endpoint host) defaults to
`https://api.thrivelearning.com` and may be overridden for tests/proxies. The new Pass B streams'
real-API host (`https://public.api.learn.link`) is NOT config-templated (see Known limits'
absolute-URL-path note for why) — it is a literal string baked into each new stream's `path`,
matching conventions.md's squarespace/defillama absolute-URL precedent exactly. The real API also
documents an OAuth 2.0 client-credentials alternative for new integrations; this bundle models only
the Basic-auth shape, matching legacy and the docs' own "also supported for existing integrations"
framing.

## Streams notes

**Legacy-parity streams (unchanged from wave2, still targeting `config.base_url`):** `users` (`GET
/users`, records at `items`), `content` (`GET /content`, records at `items`), `completions` (`GET
/completions`, records at `items`); every emitted field matches the raw API's own field names
exactly. All three declare primary key `["id"]`, `"projection": "passthrough"` (legacy's
`connsdk.Harvest` callback emits every raw field verbatim, `thrive_learning.go:96-98` — no
`mapRecord`-style filtering step), and share `config.start_date`'s optional `updated_since` query
param (`"query": {"updated_since": {"template": "{{ config.start_date }}", "omit_when_absent":
true}}`) and the legacy `page`/`limit` page-number pagination (`page_size: 100`). None of this
changed in Pass B.

**New Pass B streams (13), all targeting the real, currently-documented API on its own absolute
host via a literal `https://public.api.learn.link/rest/v1/...` string in `stream.path`
(conventions.md's squarespace/defillama absolute-URL-path precedent, since a stream cannot
override `base.url` itself, only its own `path` — the host is a hardcoded literal, not
`{{ config.* }}`-templated, per the Known limits note below):**

- `activities` — `GET /rest/v1/activities`, records at `results`, `page`/`perPage` pagination
  (`page_size: 1000`, the endpoint's own documented max), primary key `["id"]`, `x-cursor-field:
  date`.
- `contents_v1` — `GET /rest/v1/contents`, records at `results`, same pagination shape, primary key
  `["id"]`, `x-cursor-field: updatedAt`. Named `contents_v1` (not `content`/`contents`) to avoid
  colliding with the legacy-parity `content` stream, which is a genuinely different, simpler
  resource shape on a genuinely different host.
- `learning_completions` — `GET /rest/v1/learning/completions`, records at `.` (this endpoint's
  own documented response is a BARE array, unlike most other v1 list endpoints), same pagination
  params. Named `learning_completions` (not `completions`) for the same collision-avoidance reason
  as `contents_v1`.
- `assignments` — `GET /rest/v1/assignments`, records at `.` (bare array), primary key `["id"]`,
  `x-cursor-field: updatedAt`.
- `assignment_enrolments` — `GET /rest/v1/assignments/{assignmentId}/enrolments`, a `fan_out`
  stream (`ids_from.request` over the `assignments` endpoint's own bare-array response, `into:
  {path_var: assignmentId}`, `stamp_field: assignment_id`) since enrolments are only reachable
  per-assignment with no top-level list endpoint — the sanctioned dialect mechanism for exactly
  this shape (conventions.md §3 "Sub-resource fan-out").
- `audiences` — `GET /rest/v1/audiences`, records at `results`, primary key `["id"]`,
  `x-cursor-field: updatedAt`.
- `audience_members` / `audience_managers` — both `fan_out` streams over `audiences` (`into:
  {path_var: audienceId}`, `stamp_field: audience_id`), matching `assignment_enrolments`'
  reasoning. Neither resource has a globally-unique per-record id (`AudienceMember`/
  `AudienceManager`'s own `userId` is only unique within one audience's listing); `[audience_id,
  userId]` is the genuine composite primary key, matching this migration's ip2whois `[domain,
  role]` composite-key precedent (conventions.md ledger item 11). `audience_managers`' documented
  response is an unpaginated bare array (`pagination.type: none`); `audience_members` uses the
  standard `results`/pagination envelope.
- `tags` — `GET /rest/v1/tags`, records at `results`, `pagination.type: none` (the documented
  endpoint takes no page/pagination parameters at all, unlike every other v1 list endpoint),
  primary key `["id"]`.
- `cpd_categories` — `GET /rest/v1/cpdCategories`, records at `results`, primary key
  `["categoryId"]` — this resource's own documented identifier field name, not `id`.
- `cpd_entries` — `GET /rest/v1/cpdEntries`, records at `results`, primary key `["logEntryId"]` —
  again the resource's own documented identifier field name.
- `cpd_requirements` — `GET /rest/v1/cpdRequirementSummaries`, records at `results`, primary key
  `["audienceRequirementId"]`.
- `skill_levels` — `GET /rest/v1/skills/levels`, records at `.` (a small, fixed, unpaginated bare
  array of `{name, isEnabled, value}`), primary key `["value"]` (the only field guaranteed unique
  per documented level).

Every new stream declares a per-stream `"conformance": {"skip_dynamic": true, "reason": "..."}`
marker: an absolute-URL `stream.path` is resolved by `connsdk.Requester.resolveURL` as an
ALREADY-ABSOLUTE request target and never substitutes the conformance replay server's origin for
it (conventions.md's squarespace `contacts` stream precedent), so a declarative fixture replay
would dial the real internet instead of the test double. Every new stream still ships a real
fixture (`fixtures/streams/<stream>/page_*.json`, matching the published OpenAPI response shape
field-for-field) proving the intended request/response shape for documentation and any future
`paritytest/<name>` authoring, even though the dynamic replay checks themselves don't exercise it.

## Write actions & risks

None. `capabilities.write` stays `false`; this bundle ships no `writes.json`. See Known limits for
the full account of why the real API's documented write surface (v2 user lifecycle, completion
recording, assignment/audience lifecycle, tag/skill management) is a verified `ENGINE_GAP` rather
than an implemented-but-untested write path.

## Known limits

- **The real API's documented write surface cannot be safely conformance-verified with this
  engine version, and is therefore not implemented at all (`ENGINE_GAP`, fully detailed in
  `api_surface.json`'s `scope` note and every affected endpoint's own `excluded.reason`).**
  `StreamSpec` has a per-stream `conformance.skip_dynamic` escape hatch (used by all 13 new read
  streams above) precisely for the "this stream's real endpoint lives on a different host than
  `base.url`" shape — but `WriteAction` has NO equivalent marker at all, and
  `checkWriteRequestShape` unconditionally runs `engine.Write` against a local capture server built
  from `bundle.HTTP.URL`. A write action whose own `path` is an absolute URL (the only mechanism
  available to target a different host than `base.url`, identical to what the read streams use)
  bypasses that capture server's origin entirely and dials the literal host baked into the path —
  LIVE, in CI, on every conformance run. This was verified empirically during authoring: a first
  attempt at implementing `create_user`/`update_user`/`suspend_user`/`delete_user` (v2 user
  lifecycle) plus 8 other real-API writes produced genuine HTTP round trips to
  `https://public.api.learn.link` (a 109-second test run, real 401 Unauthorized responses) instead
  of exercising a test double. Shipping that would mean every future `connectorgen validate`/
  conformance run for this bundle silently depends on a real external service being reachable and
  behaving a specific way — a correctness and reliability risk conventions.md's meta-rule forbids
  papering over. The fix is an engine-level `WriteAction`-scoped (or bundle-level, mirroring
  `Metadata.Conformance`) `skip_dynamic` marker; until that lands, every real-API write endpoint is
  documented as `excluded: {category: out_of_scope, reason: "ENGINE_GAP: ..."}` in
  `api_surface.json` rather than shipped unverified.
- **A second, independent write-body `ENGINE_GAP` would ALSO block the audience
  members/managers add-and-replace endpoints even if the above were fixed**: their documented
  request body is a bare top-level JSON array (of user emails/refs/ids, or of
  `AudienceManagerInput` objects), not a JSON object. `engine/write.go`'s body construction
  (`buildJSONBody`/`buildBodyFieldsPayload`) only ever builds and sends a JSON object — there is no
  dialect mechanism to emit a bare array as the whole request body. Wrapping the array in a
  synthetic object key (e.g. `{"user_refs": [...]}`) would silently diverge from what the real API
  actually accepts, so this is documented as a second, compounding gap on those 4 specific
  endpoints, not approximated.
- **`PATCH /rest/v1/audiences/{audienceId}` (update an audience) would additionally not be
  implemented even once the above gaps are fixed.** The published OpenAPI spec tags this operation
  `Coming Soon` alongside its `Audiences` tag, signaling it is documented ahead of general
  availability; shipping a write action against a not-yet-confirmed-live endpoint risks a false
  green after a future spec revision removes the tag without changing the shape underneath it.
- **The real API's own `/rest/v1/users` (and its `users/ref/{ref}`/`user/{id}` single-lookup
  variants) is not modeled as a stream** even though it is a real, richer, differently-shaped user
  resource than the legacy-parity `users` stream — adding a second users-shaped stream on the same
  conceptual resource would be a confusing duplicate of intent rather than a distinct resource; see
  `api_surface.json`'s `duplicate_of` exclusion.
- **`cpdUserLogSummaries` is not modeled as a stream.** Its documented `CPDUserSummary` schema has
  no unique per-record identifier at all (`userId` alone is not unique across different
  caller-chosen date ranges) — it is an aggregate query result, not a discrete syncable entity.
- Full Thrive Learning API surface beyond what's listed in `api_surface.json` (courses/pathways,
  groups, certificates on the legacy connector's own simplified host — legacy itself never
  implemented these) remains out of scope; this is parity with legacy on that host, not a
  reduction.
- **`page_size`/`max_pages` are not runtime-configurable on the legacy-parity streams.** Legacy
  exposes both as config-driven overrides (`thrive_learning.go:190-212`: `pageSize(cfg)` any
  positive integer defaulting to 100, `maxPages(cfg)` any non-negative integer defaulting to
  0/unbounded, both erroring loudly on a non-numeric or negative value). The engine's `page_number`
  paginator constructor reads `PaginationSpec.PageSize` as a static bundle-level integer from
  `streams.json` and has no `MaxPages`-equivalent knob at all, so neither is wireable from
  `config.*` without inventing Go. This bundle hardcodes `page_size: 100` (legacy's own default) and
  leaves pagination unbounded (matching legacy's own `max_pages` default of 0/unbounded) —
  matching every input that does not explicitly override either value (the common case). Neither
  `page_size` nor `max_pages` is declared in `spec.json` (F6, REVIEW.md: a declared-but-unwireable
  key is worse than an absent one). New Pass B streams hardcode `page_size: 1000`, the real API's
  own documented max, for the same reason.
- **The real-API host is a hardcoded literal, not a configurable spec property.** An earlier
  authoring attempt added an `api_base_url` spec property and templated it into every new stream's
  `path` (`{{ config.api_base_url }}/rest/v1/...`) so a caller could override it for a staging
  tenant. This broke under conformance: `runtimeConfigForEngine` fills every declared spec property
  with the literal string `"synthetic-conformance-value"`, so the templated path resolved to
  `"synthetic-conformance-value/rest/v1/..."` — a value with NO `http(s)://` prefix, which
  `connsdk.Requester.resolveURL` then (correctly, per its own contract) treats as a RELATIVE path
  against whatever `base.url`/replay-server origin is active, silently defeating the entire
  absolute-URL-path mechanism the skip_dynamic markers depend on. Reverting to a hardcoded literal
  host (squarespace/defillama's exact precedent, neither of which offers a config override either)
  fixed this at the cost of no staging-host override; a caller needing a staging tenant must use a
  local proxy/rewrite in front of this bundle, or wait for the `WriteAction`-level skip_dynamic gap
  above to be closed so the pattern can be revisited end-to-end.
- **Password config-fallback is not modeled.** Legacy's `secret(cfg, "password")` helper falls back
  to a plaintext `cfg.Config["password"]` value when `cfg.Secrets["password"]` is unset
  (`thrive_learning.go:214-221`) — a narrow legacy affordance for callers that never migrated the
  password into the secret store. This bundle's `auth` spec reads only `secrets.password` (the
  dialect has no config-fallback-if-secret-absent primitive, and x-secret discipline treats a
  password as inherently secret-shaped regardless of where a caller happens to store it); a caller
  relying on the plaintext-config fallback must move the value to the `password` secret. This is a
  documented, narrow scope reduction, not a data-shape deviation — no request output changes for
  any caller already using the secret.
- No incremental cursor is modeled on any new stream (matching the legacy-parity streams' own
  no-incremental behavior): `x-cursor-field` is declared where a genuine timestamp field exists
  (catalog-hint value only), but no `incremental` block is declared on any stream, since wiring a
  real server-side `updatedAtFrom`/`updatedSince`-style incremental filter for these new streams
  is a distinct, larger increment out of scope for this pass.
