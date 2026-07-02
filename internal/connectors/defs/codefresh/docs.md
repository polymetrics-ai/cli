# Overview

Codefresh reads projects, pipelines, runner agents, and shared configuration contexts through the
Codefresh REST API (`https://g.codefresh.io/api`). This bundle is a legacy-parity migration of
`internal/connectors/codefresh` (the hand-written connector it replaces); the legacy package stays
registered and unchanged until the wave6 registry flip. Codefresh is read-only here — there are no
obvious safe reverse-ETL writes, so `capabilities.write` is `false` and this bundle ships no
`writes.json`.

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

All 4 streams share the same page-number pagination shape (`pagination.type: page_number`,
`page_param: page`, `size_param: limit`, `start_page: 1`) — a page shorter than `page_size` stops
pagination, matching legacy's own `harvest`'s "short page means the listing is exhausted" rule
exactly (`codefresh.go:159-162`).

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

None of the 4 streams declares an `incremental` block: legacy's `harvest` has no incremental
filtering at all (every read is a full page-number sweep), so this bundle matches that shape
exactly rather than inventing an incremental cursor legacy never had.

## Write actions & risks

None. Codefresh is read-only in this connector (`capabilities.write: false`, matching legacy's own
`Write` stub which unconditionally returns `connectors.ErrUnsupportedOperation`); no `writes.json`
is shipped.

## Known limits

- Full Codefresh API surface (pipeline triggers/runs, builds, deployments, user/team management)
  is out of scope for this migration; see `api_surface.json`'s `excluded` entries. Only the 4
  legacy-parity read streams are implemented.
- **`page_size` is fixed at `2` (not legacy's default of `50`), and neither `page_size` nor
  `max_pages` is runtime-configurable.** Legacy exposes both as config overrides
  (`codefreshPageSize`/`codefreshMaxPages`, `codefresh.go:268-296`, clamped 1-100 / `0`/`all`/
  `unlimited`). The engine's `page_number` paginator's `PageSize` is a static bundle-authored int
  (not template-resolvable from `config.*`), and there is no `MaxPages`-equivalent config-driven
  knob for this paginator type either; `max_pages` is unbounded (matching legacy's own
  `max_pages=0`/`all`/`unlimited` default, since `PaginationSpec` here declares no `max_pages`
  cap). `page_size` is set to `2` specifically so the mandatory 2-page conformance fixture
  (`fixtures/streams/projects/{page_1,page_2}.json`) is realistic to author and honestly exercises
  the short-page stop rule (`conformance`'s `pagination_terminates` check requires the replay
  server to serve exactly one request per fixture page — a `page_size` of 50 against a small
  hand-authored fixture would short-circuit after page 1 and never touch page 2 at all), matching
  bamboo-hr's and callrail's identical documented precedent
  (`docs/migration/conventions.md`). This changes the real per-page record count from legacy's 50
  to 2 — a REST-shape difference (more, smaller requests), never a data-emission difference (every
  record is still read exactly once, across more pages).
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
