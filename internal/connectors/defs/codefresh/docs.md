# Overview

Codefresh reads projects, pipelines, builds, runner agents, shared configuration contexts,
container images, registries, triggers, trigger events, and annotations through the Codefresh REST
API (`https://g.codefresh.io/api`), and can create/update/delete/run projects, pipelines, contexts,
and agents. This bundle started as a legacy-parity migration of `internal/connectors/codefresh`
(the hand-written connector it replaces, which stays registered and unchanged until the wave6
registry flip); this Pass B pass expanded it to the practical full documented surface after
reviewing the complete live OpenAPI spec (`https://g.codefresh.io/api/openapi.json`, 312 paths /
419 method+path operations — see `api_surface.json`).

## Auth setup

Provide a Codefresh API key via the `api_key` secret; it is sent verbatim (no `Bearer` prefix) as
the `Authorization` header (`base.auth`'s `api_key_header` mode with no `prefix`), matching
legacy's `connsdk.APIKeyHeader("Authorization", key, "")` exactly. An optional `account_id` config
value is sent as the `X-Access-Token` header for account scoping; when unset, the header is
omitted entirely (not sent empty) — the same conditional-header pattern as Stripe's
`Stripe-Account`/`account_id`.

Documented parity deviation: legacy stores `account_id` in `cfg.Secrets["account_id"]` rather than
`cfg.Config`. This bundle declares `account_id` as a plain (non-`x-secret`) `spec.json` config
property instead, because `streams.json`'s `base.headers` only tolerates an absent value for a
`config.*`-templated header that is declared-but-optional (`read.go`'s `resolveHeaders`); a
`secrets.*`-templated header is *always* a hard error when the secret is unset (F4,
`docs/migration/conventions.md` §3), which would break the common case of running without an
account id at all. Reclassifying the field never changes the emitted `X-Access-Token` header value
for any configured account id, and never changes any emitted record data — only which
`RuntimeConfig` map callers populate it in.

## Streams notes

The original 4 legacy-parity streams (`projects`/`pipelines`/`agents`/`contexts`) share the base
page-number pagination shape (`pagination.type: page_number`, `page_param: page`, `size_param:
limit`, `start_page: 1`) — a page shorter than `page_size` stops pagination, matching legacy's own
`harvest`'s "short page means the listing is exhausted" rule exactly (`codefresh.go:159-162`).

- **`projects`** (`GET /projects`, records at `projects`): identity is the API's own top-level
  `id` field; `computed_fields` renames `projectName`/`pipelinesNumber`/`updatedAt` to
  `project_name`/`pipelines_number`/`updated_at`, and copies `favorite`/`pipelines_number` as bare
  `{{ record.<path> }}` references so the engine's typed extraction preserves their real
  boolean/integer wire types (no stringify workaround).
- **`pipelines`** (`GET /pipelines`, records at `docs`): pipeline documents are Mongo-style records
  keyed by `_id`, with `name`/`project`/`is_public`/`created_at`/`updated_at` nested under
  `metadata` — matching legacy's `metadataField` reads exactly (`streams.go:143-152`).
  `computed_fields` maps `record._id` to `id` and reaches into `record.metadata.*` for the rest.
- **`agents`** (`GET /agents`, records at the response root — a bare JSON array, `records.path:
  ""`): agent objects are also Mongo-style, keyed by `_id`, with `name`/`version`/`status`/
  `created_at` at the top level.
- **`contexts`** (`GET /contexts`, records at the response root): context objects have no top-level
  `id`/`_id` at all — identity is `metadata.name`, matching legacy's `recordID` fallback chain
  landing on `metadataField(item, "name")` for this resource (contexts are the one stream where
  legacy's `id`/`_id` branches never match). `owner` is read from the top-level field, matching
  legacy's own `codefreshContextRecord` (`streams.go:164-175`).

Documented parity deviation: legacy's `recordID` is a 4-branch runtime fallback (`id`, then `_id`,
then `metadata.name`, then `projectName`) tried uniformly across all 4 resource types. The engine's
`computed_fields` dialect has no conditional/fallback reference syntax (a template is a single
fixed reference or filter chain, never an "A-or-B" expression — same limitation breezy-hr's `_id`
mapping documents). This bundle instead hard-codes, per stream, the ONE branch that resource's real
Codefresh API response actually populates (`id` for projects, `_id` for pipelines and agents,
`metadata.name` for contexts) — this never changes emitted data for any real Codefresh API
response; it only differs from legacy for a hypothetical cross-shaped response (e.g. a project
object missing `id` but carrying `_id`) legacy's fallback chain defends against but the real API
has never been observed to send for that resource type.

**New Pass B streams** (verified against the live `/openapi.json` response schemas/examples, not
guessed):

- **`builds`** (`GET /workflow`, records at `workflows.docs`): Codefresh's actual build-history
  resource — legacy never modeled this at all. Reuses the base `page_number` pagination (the
  response's own `pagination.page`/`pagination.pageSize` match `page`/`limit` exactly); the
  response also carries a `pagination.nextPage` boolean, but the engine's `page_number` paginator
  only implements the short-page stop signal (same class of harmless-extra-request deviation as
  `docs/migration/conventions.md` §5 ledger item 13/jamf-pro). `GET /workflow` supports a
  `startDate`+`endDate` range filter (both required together per its own parameter docs), which
  does not fit the engine's lower-bound-only `incremental.request_param` shape (there's no
  declarative way to pair a computed lower bound with a second, closing upper-bound param) — this
  stream is full-refresh only, not incremental; see Known limits.
- **`images`** (`GET /images`, records at `docs`): the endpoint's own OpenAPI `parameters` list
  documents `limit`/`offset`, and its recorded example response wraps the array in `{docs, total,
  limit, offset}` (`pagination.type: offset_limit`) even though the same spec's `schema.type` is
  declared as a bare array — the worked example is the actual recorded wire shape and is what this
  bundle follows (`docs/migration/conventions.md` §4's "recorded-real-shape" rule). Docker image
  records are keyed by Mongo `_id`; `computed_fields` renames it to `id`.
- **`registries`** (`GET /registries`, records at the response root, not paginated): every
  registry-provider variant (ECR/GCR/Docker Hub/etc — a `oneOf`-discriminated union in the OpenAPI
  spec) shares one common base shape (`_id`/`name`/`provider`/`kind`/`domain`/`primary`/`default`/
  `internal`/`behindFirewall`); this bundle projects only that shared base, never a
  provider-specific credential field (those are inherently sensitive and provider-varying, not a
  syncable common schema).
- **`triggers`** (`GET /hermes/triggers`, records at the response root, not paginated): pipeline
  trigger bindings; primary key is the `(event, pipeline)` pair (a trigger has no single-field id
  of its own). `filters.tag` and the nested `event-data.*` sub-object are surfaced via
  `computed_fields`.
- **`trigger_events`** (`GET /hermes/events`, records at the response root, not paginated):
  available trigger-event integrations (e.g. "Docker Hub push"); primary key is `uri`. The raw
  API's own `secret` field (a webhook-URL secret token) is deliberately NOT in this stream's schema
  — schema-mode projection drops it, so it never reaches an emitted record.
- **`annotations`** (`GET /annotations`, records at the response root, not paginated): free-form
  key/value tags attached to Codefresh entities (projects, pipelines, builds, etc); `entity_id`/
  `entity_type` identify what the annotation is attached to.

None of the original 4 legacy-parity streams declares an `incremental` block: legacy's `harvest`
has no incremental filtering at all (every read is a full page-number sweep), so this bundle
matches that shape exactly rather than inventing an incremental cursor legacy never had. None of
the 6 new Pass B streams declares one either — `images`/`registries`/`triggers`/`trigger_events`/
`annotations` have no documented server-side lower-bound date filter, and `builds`'s `startDate`
requires a paired `endDate` the engine's incremental dialect cannot express (see above).

## Write actions & risks

10 write actions, all requiring approval (`risk` set on every action; `confirm: "destructive"` on
every delete):

- **`create_project`** (`POST /projects`) — creates a new Codefresh project (`projectName`
  required).
- **`delete_project`** (`DELETE /projects/{{ record.id }}`) — **destructive/irreversible**; treats
  404 as already-deleted (`missing_ok_status: [404]`).
- **`create_pipeline`** (`POST /pipelines`) — creates a new pipeline from a `{metadata: {name},
  spec}` body.
- **`update_pipeline`** (`PUT /pipelines/{{ record.name }}`) — replaces an existing pipeline's
  full spec; `path_fields: ["name"]` excludes only the flat top-level `name` field used for the
  path from the body, so the body's own nested `metadata.name`/`spec` still round-trip untouched
  (Codefresh pipeline names are `org/repo`-shaped and contain `/`; the engine's default
  per-path-segment `urlencode` percent-encodes it to `%2F` on the wire, which Go's `net/http`
  server — and Codefresh's real API, which accepts the same org/repo-shaped name in this position
  — decodes back to the literal name; verified this round-trips correctly against
  `httptest.Server`, see `write_request_shape:update_pipeline`/`delete_pipeline`/`run_pipeline`).
- **`delete_pipeline`** (`DELETE /pipelines/{{ record.name }}`) — **destructive/irreversible**.
- **`run_pipeline`** (`POST /pipelines/run/{{ record.name }}`) — triggers a REAL pipeline
  execution; consumes actual build minutes/resources on the connected Codefresh account, not a
  reversible action even though it isn't a delete.
- **`create_context`** (`POST /contexts`) — creates a new shared configuration context; contexts
  can hold arbitrary key/value configuration (potentially sensitive, e.g. registry credentials),
  so this write is treated as approval-required like every other mutation here.
- **`delete_context`** (`DELETE /contexts/{{ record.name }}`) — **destructive/irreversible**.
- **`create_agent`** (`POST /agents`) — registers a new Codefresh runner agent record (`name`
  required).
- **`delete_agent`** (`DELETE /agent/{{ record.id }}`) — **destructive/irreversible**
  deregistration.

## Known limits

- Full Codefresh API surface remains out of scope beyond the 10 streams/10 writes above; see
  `api_surface.json`'s `excluded` entries (419 method+path operations reviewed against the live
  OpenAPI spec). The largest excluded areas are: raw Kubernetes/Helm cluster-proxy passthrough
  (`requires_elevated_scope`/`out_of_scope` — arbitrary cluster API forwarding, not Codefresh's own
  data), GitOps/Argo CD environment-v2 integration (`out_of_scope` — a distinct product area), and
  account/customer/IDP/ABAC/service-user administration (`requires_elevated_scope` — needs
  account-owner/reseller-admin privilege, not ordinary API-key access).
- **`team`, `clusters`, `step-types`, `runtime-environments`, and `audit` GET list endpoints are
  excluded, not merely deferred**: the live OpenAPI spec (`https://g.codefresh.io/api/openapi.json`)
  documents each of their response bodies as a bare `{"type": "object"}` placeholder with no
  property list at all — no field-accurate schema or fixture can be authored for these honestly
  from the published spec alone, and no alternative authoritative source (the API's own docs page
  is a JS-rendered Swagger UI with no separate field reference) was found. Implementing them would
  require either reverse-engineering field names from a live authenticated call (out of scope for
  a fixture-based migration) or accepting a schema that is a guess, which the `docs/migration/
  conventions.md` §2 "Schema-as-projection" rule forbids.
- **`builds`'s `startDate`/`endDate` range filter is not modeled as `incremental`.** See Streams
  notes — the engine's `incremental.request_param` sends exactly one lower-bound value; `GET
  /workflow`'s own parameter docs require `startDate` and `endDate` together, and there is no
  declarative mechanism to pair a computed lower bound with a second, closing upper-bound
  parameter. `builds` is full-refresh only.
- **`page_size` is fixed at `50` (matching legacy's default) for the 4 legacy-parity streams and
  `builds`, and neither `page_size` nor `max_pages` is runtime-configurable.** Legacy exposes both
  as config overrides (`codefreshPageSize`/`codefreshMaxPages`, `codefresh.go:268-296`, clamped
  1-100 / `0`/`all`/`unlimited`). The engine's `page_number` paginator's `PageSize` is a static
  bundle-authored int (not template-resolvable from `config.*`), and there is no
  `MaxPages`-equivalent config-driven knob for this paginator type either; `max_pages` is unbounded
  (matching legacy's own `max_pages=0`/`all`/`unlimited` default). `page_size` is set to legacy's
  own default (`codefreshDefaultPageSize = 50`, `codefresh.go:31`) rather than a small
  fixture-convenience value — the mandatory 2-page conformance fixture
  (`fixtures/streams/projects/{page_1,page_2}.json`) is authored with 50 records on page 1 (full
  page, continues) and 1 record on page 2 (short page, stops) to honestly exercise the short-page
  stop rule at the real page size. `page_size`/`max_pages` not being runtime-configurable is the
  only remaining config-surface narrowing versus legacy for these streams.
- **Legacy's fixture-mode-only synthetic record shape is not modeled.** Legacy's `readFixture` path
  (only reached when `config.mode == "fixture"`) emits 2 synthetic records per stream with a
  single hard-coded shape shared across all 4 resource types (`codefresh.go:170-203`) — a
  test/conformance-harness affordance, not a real API response shape. This bundle's
  `fixtures/streams/**` instead model each stream's own REAL wire shape (bare array vs. `{docs:
  [...]}` vs. `{projects: [...]}`, Mongo `_id` vs. plain `id`, `metadata`-nested fields), which is
  the correct and more useful fixture-authoring target for `conformance`'s dynamic checks per
  `docs/migration/conventions.md` §4 ("recorded-real-shape" fixtures) — not a parity gap.
- `account_id`'s secret-to-config reclassification (see Auth setup) is a classification-only
  deviation; it never changes the emitted `X-Access-Token` header value or any record data.
