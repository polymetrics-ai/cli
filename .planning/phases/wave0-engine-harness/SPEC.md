# SPEC — wave0-engine-harness

Phase: `wave0-engine-harness` · Milestone: `connector-architecture-v2`
Design of record: `docs/architecture/connector-architecture-v2-design.md` (engine/bundles/hooks/connectorgen/conformance v2)
and `docs/architecture/connector-certification-design.md` (certify harness).
PRD: `docs/plans/universal-programming-loop-prd.md`. Wave context: `docs/migration/orchestration-plan.md` (wave 0 row).

## 1. What wave0 delivers

Wave0 builds the declarative engine and all migration tooling **alongside** the legacy connector
machinery, proves it with three golden migrations (stripe, searxng, postgres) under
engine-vs-legacy parity tests, and leaves the tree green with legacy behavior unchanged.
**Nothing legacy is deleted and no registry behavior changes in wave0.**

### 1.1 `internal/connectors/engine/` — declarative runtime engine (design §B)

New package interpreting defs bundles on top of the UNCHANGED `connsdk` toolkit
(`internal/connectors/connsdk/http.go` `Requester.Do/DoForm/DoJSON`, `paginate.go` four paginators +
`Harvest`, `auth.go` `Bearer/Basic/APIKeyHeader/APIKeyQuery/OAuth2ClientCredentials`,
`extract.go` `RecordsAt/StringAt`, `state.go` `Cursor/WithCursor/MaxCursor`). Files:

| File | Contents |
|---|---|
| `bundle.go` | `Bundle`, `Metadata`, `HTTPBase`, `AuthSpec`, `StreamSpec`, `RecordsSpec`, `IncrementalSpec`, `WriteAction`, `DeleteSpec`, `PaginationSpec`, `RateLimitSpec`, `ErrorRule`, `APISurface` types (design §B.2) + `LoadAll(fsys fs.FS) ([]Bundle, error)` / `Load(fsys fs.FS, name string) (Bundle, error)` loader with structural validation (required files, name regex `^[a-z0-9][a-z0-9-]*$`, dir name == metadata.name) |
| `interpolate.go` | `{{ config.k }}`, `{{ secrets.k }}`, `{{ record.dotted.path }}`, `{{ cursor }}`; filters `urlencode` (default for path-segment insertion), `unix_seconds`, `base64`; `when` conditions: `==`, `in [...]`, truthiness. No loops/arithmetic/user functions (design §B.3, ~150 lines) |
| `schema.go` | Minimal internal draft-07 compiler + validator — **no new deps**. Keywords: `type` (incl. type arrays), `required`, `properties`, `items`, `enum`, `pattern`, `minProperties`, `additionalProperties` (bool form); `format`, `default`, `title`, `description`, `$schema` accepted as annotations only. Extensions: `x-secret` (→ `Schema.SecretKeys()`), `x-primary-key`, `x-cursor-field`. Unknown keywords are a compile ERROR (keeps bundles honest) |
| `errors.go` | Typed `engine.Error{Connector, Stream, Action, Page, RecordIndex, Class, Hint, Err}` wrapping `connsdk.HTTPError`; `error_map` rule application (`status`, `match_body`, `class`, `hint`); message text passes through `safety.RedactErrorText` (`internal/safety/safety.go:50`) |
| `auth.go` | `selectAuth(cfg, specs) (connsdk.Authenticator, error)`: evaluates `when` in declared order, first match wins; modes `none|bearer|basic|api_key_header|api_key_query|oauth2_client_credentials|custom` mapping onto the existing `connsdk/auth.go` constructors; `custom` resolves an `AuthHook` via the hook registry |
| `paginate.go` | `newPaginator(spec PaginationSpec, pageSize int) (connsdk.Paginator, error)` for all **6 types**: `link_header`, `page_number`, `offset_limit`, `cursor`, `next_url`, `none`. `cursor` supports BOTH token sources: `token_path` (body token → `connsdk.CursorPaginator`) and `last_record_field` + optional `stop_path` (stripe's `starting_after`/`has_more` loop, today hand-written at `internal/connectors/stripe/stripe.go:147`); `next_url` reads an absolute next-page URL from a body path (aircall style); `none` = single request. New paginator impls live here and satisfy `connsdk.Paginator` — connsdk itself is not modified |
| `read.go` | `Read(ctx, req, emit)` per design §B.4: resolve stream, build `connsdk.Requester` (base URL, headers — a header whose interpolated value is empty is OMITTED, matching stripe's conditional `Stripe-Account`), initial query = static `query` + incremental lower bound from `req.State["cursor"]` fallback `start_config_key`, formatted per `param_format` (`rfc3339|unix_seconds|date|github_date_range`); paginator loop → `RecordsAt` → `filter` (`field_absent`/`field_equals`) → projection (`schema` default: emit only declared properties; `passthrough` opt-out) → `computed_fields` (dotted extraction from raw record) → `MaxCursor` tracking → emit. `client_filtered` incremental drops records below the cursor. Optional `RateLimitSpec{requests_per_minute}` inter-request wait with injectable sleeper. Generic `StatefulReader.InitialState` |
| `write.go` | `ValidateWrite` (compile-once `record_schema`, errors carry record index), `DryRunWrite` (validation + resolved method/path preview with secrets redacted), `Write` per design §B.5: default body = all record fields minus `path_fields`; `body_type: json|form|none`; `body_fields` allow-list for delete bodies; `kind: delete` honors `delete.missing_ok_status` (404 on idempotent delete counts as written). One request per record via `Requester.Do`/`DoForm` |
| `hooks.go` | Hook interfaces `Hooks/AuthHook/RecordHook/StreamHook/WriteHook/CheckHook` (design §B.7) + process-global registry `RegisterHooks(name, factory)` / `HooksFor(name)` (registry lives in engine to avoid import cycles; `internal/connectors/hooks/<name>/` packages call `engine.RegisterHooks` from init) |
| `connector.go` | `engine.New(b Bundle, h Hooks) *Connector` implementing the **current** `connectors.Connector` interface (`internal/connectors/connectors.go:256` — including `Metadata()` and `Write`), plus `connectors.WriteValidator`, `connectors.DryRunWriter`, `connectors.StatefulReader`, `connectors.ManifestProvider` (manifest synthesized from the bundle so `pm connectors inspect` output shape is preserved), plus the NEW `DefinitionProvider` (§1.8). Also `engine.Base` for Tier-3 natives: serves `Metadata()/Definition()/spec` from a bundle. Derived sync modes per design §B.6 exposed on `Manifest().SyncModes` and `Definition().Streams[i].SyncModes` |
| `schema/` (data) | Meta-schemas for the bundle files themselves: `metadata.schema.json`, `spec.schema.json`, `streams.schema.json`, `writes.schema.json`, `api_surface.schema.json` — written in the same minimal draft-07 dialect and used by `connectorgen validate` (and referenced by the migration executor prompt template in `docs/prompts/universal-programming-loop-prompts.md`) |

### 1.2 `internal/connectors/defs/` — embedded bundle tree

`defs/defs.go` is the only Go file: `package defs` with a single `//go:embed all:*` and
`var FS embed.FS` (single-pattern embed so optional files — `writes.json`, `fixtures/` — never
break compilation; the loader iterates directories and skips non-dirs). Wave0 ships exactly three
bundles: `defs/stripe/`, `defs/searxng/`, `defs/postgres/` (layout per design §A; contracts in
`DATA-MODEL.md`). `streams.json` is OPTIONAL when `metadata.capabilities.dynamic_schema == true`
(postgres discovers streams at runtime).

### 1.3 `internal/connectors/hooks/` + hook registry

Directory scaffold + `hooks/hookset/hookset_gen.go` (generated by `connectorgen gen`; blank imports,
empty in wave0 — the goldens need no hooks: stripe/searxng are pure declarative, postgres is Tier-3).
Hook dispatch is proven by engine unit tests with in-test fake hooks, not by a production hook.

### 1.4 `cmd/connectorgen` — validate | gen | new

- `validate [dir]` (default `internal/connectors/defs`): loads every bundle from `os.DirFS`,
  compiles all schemas, resolves every `{{ }}` against `spec.json` properties, checks `schema` refs
  exist, PK/cursor fields exist in schemas, write `path_fields ⊆ record_schema properties`,
  api_surface rules 1–5 (design §E.1), naming regex, `docs.md` fixed headings, fixture presence for
  the first stream, secret-literal scan of fixtures. Non-zero exit + per-bundle findings (`--json`).
- `gen`: regenerates `hooks/hookset/hookset_gen.go` and `native/nativeset/nativeset_gen.go`
  (blank-import wiring; NOTHING imports these two packages in wave0 — see §2 coexistence).
- `new <name>`: scaffolds a defs bundle from embedded templates.
`cmd/registrygen` is NOT replaced in wave0 (deleted in wave6); its `skip` map
(`cmd/registrygen/main.go:30`) gains entries `defs`, `engine`, `hooks`, `native`, `conformance`,
`certify` so the new packages are never blank-imported as connectors. This is the ONLY legacy-file
edit in the phase and it is orchestrator-owned.

### 1.5 `internal/connectors/conformance/` — conformance v2

New package (legacy `internal/connectors/native_conformance.go` stays untouched until wave6):
- Static checks per bundle: `spec_schema_valid`, `stream_schemas_valid`, `pk_fields_exist`,
  `cursor_fields_exist`, `interpolations_resolve`, `write_schemas_valid`, `surface_complete`,
  `docs_present`, `secret_redaction`, `fixtures_present`.
- Dynamic checks: an `httptest.Server` replays `fixtures/streams/<stream>/page_N.json`
  (request-keyed envelope, see `DATA-MODEL.md` §5) against the REAL engine:
  `check_fixture`, `read_fixture_nonempty` (first stream mandatory), `pagination_terminates`
  (multi-page fixtures consumed exactly once), `records_match_schema`, `cursor_advances`
  (post-read cursor == max fixture cursor; re-read sends `request_param`),
  `write_validate` + `write_request_shape` (dry-run/execution vs `fixtures/writes/<action>.json`
  `expect` block), `delete_semantics` (`missing_ok_status`).
- Entry points: `Run(ctx, bundle, hooks) Report` + a `TestConformance` Go test that iterates
  `defs.FS` with one subtest per connector (`go test ./internal/connectors/conformance -run 'TestConformance/<name>'`).
- Self-tests: a seeded-invalid bundle corpus under `conformance/testdata/invalid/` must FAIL with
  the expected check names (shared with the connectorgen eval, `EVAL-PLAN.md` §3).

### 1.6 `internal/connectors/certify/` — certification harness CORE ONLY

Scope = implementation-order steps 1–2 of `docs/architecture/connector-certification-design.md`:
- `report.go` — `CertificationReport` (`kind: ConnectorCertification`, `schema_version: 1`),
  save/load under `.polymetrics/certifications/<connector>.json` + history dir, matrix rendering.
- `cliharness.go` — in-process driver over `cli.Run(args, stdout, stderr) int`
  (`internal/cli/cli.go:24`): ephemeral `os.MkdirTemp` root, `--root`/`--json` injection,
  stdout/stderr capture, envelope `kind` + exit-code assertion, secret-value scan of captured
  output (exact/base64/URL-encoded forms).
- `certify.go` — minimal `Runner`/`Options` (single connector, serial stages).
- `stages_source.go` — stages 0–11: preflight, fixture_conformance (skip-with-reason when the
  connector has no defs bundle — true for `sample`), manual_json (`pm connectors inspect --json`,
  `internal/cli/cli.go:198`), credentials add/test, catalog, `etl_full_refresh_append` (live-local),
  overwrite + overwrite_deduped + incremental_append_deduped as CAPTURE replays through the
  built-in `file` connector, `etl_incremental_append` + resume (cursor monotonic), query_contract.
  Proven end-to-end against the built-in `sample` connector (`internal/connectors/connectors.go:461`)
  from a Go test — **no CLI wiring** (`pm connectors certify` = step 5, later phase).
- EXCLUDED from wave0 (later phases): write protocol/ledger/sweeper, flow + schedule stages,
  `--all`/batch/creds.yaml, record/replay `httpx` seam, `cmd/certstatus`, the `--credential` flag
  fix on `pm etl check/read` (not needed for sample; documented as wave1 prerequisite).

### 1.7 Migration program artifacts

- `docs/migration/conventions.md` — the single migration recipe (bundle authoring rules, tier
  triggers, naming, fixture rules, documented parity-deviation ledger).
- `docs/migration/result.schema.json`, `docs/migration/review.schema.json` — agent I/O contracts
  per `docs/migration/orchestration-plan.md` §"Per-agent task spec".
- `docs/migration/inventory.json` generated by new tool `cmd/inventorygen`: per connector dir
  `{name, path, loc, runtime_kind, bucket(S/M/L/XL), catalog_slugs, documentation_url,
  stream_count}` sourced from dir scan + `connectors.ConnectorCatalog()` + `ManifestOf`.
- `.golangci.yml` + Makefile targets: `lint`, `connectorgen-validate`; `verify` extended with both
  (see `RUNBOOK.md`; golangci-lint acquisition is an open coordinator question — no go.mod change
  either way).

### 1.8 Interface additions (no breaking changes) — see `API-CONTRACT.md`

Added to `internal/connectors/connectors.go` (or a new `definition.go` in the same package):
`Definition`, `StreamSummary`, `WriteActionInfo`, `RiskSpec` reuse, and
`type DefinitionProvider interface { Definition() Definition }`. The core `Connector` interface,
`Manifest`/`ManifestProvider` (`internal/connectors/manifest.go`), `NewRegistry`/`NewLiveRegistry`,
slug/catalog machinery, and `internal/app/sync_modes.go` are ALL UNCHANGED in wave0.

### 1.9 Golden migrations + parity tests

| Golden | Target | Parity mechanism |
|---|---|---|
| stripe | `defs/stripe/` bundle (5 streams, `cursor`/`starting_after` pagination, `unix_seconds` incremental on `created`, writes `create_customer`/`update_customer` with `body_type: form`) | One `httptest.Server` serving recorded pages; legacy `stripe.Connector{Client: srv.Client()}` with `base_url` config vs `engine.New(bundle, nil)` against the same server: identical record sequences per stream, identical write request method/path/form body, manifest-surface equality (stream names/PKs/cursor fields/write action names vs `connectors.ManifestOf`) |
| searxng | `defs/searxng/` bundle (2 streams `search`/`reddit` over `/search`, `page_number` pagination `pageno` with no size param, `max_pages: 1`, templated `q` incl. `site:reddit.com` scoping, optional bearer) | Same httptest parity as stripe (legacy `searxng.Connector{Client: ...}` + `base_url`); read-only |
| postgres | `internal/connectors/native/postgres/` Tier-3 component split (`connector.go`, `connection.go`, `reader.go`, `cataloger.go`, `cdc.go` stub; no `writer.go` — read-only) embedding `engine.Base` over `defs/postgres/` bundle | No httptest (SQL): parity = fixture-mode Check/Catalog/Read outputs identical to legacy `internal/connectors/postgres`, identical config-validation error table (host/port/sslmode), `Definition()` served from the bundle. Reader/cataloger logic ported from `postgres.go` (pgx already in go.mod — no new dep) |

Legacy packages `internal/connectors/{stripe,searxng,postgres}/` keep compiling and stay
registered exactly as today **until wave6**.

## 2. Coexistence mechanism (DECIDED — the load-bearing scope guardrail)

1. **Registry untouched.** `connectors.NewRegistry()/NewLiveRegistry()`
   (`internal/connectors/connectors.go:273,277`) and `registryset` (via `appRegistry()` at
   `internal/cli/cli.go:983`) are not modified. `cmd/registrygen` keeps generating
   `registryset/registry_gen.go`; its skip map is extended (§1.4) so new packages are never
   imported as connectors.
2. **Engine connectors are constructed in tests only.** Parity/conformance tests build them
   directly: `bundles, _ := engine.LoadAll(defs.FS); c := engine.New(bundleFor("stripe"), nil)`.
   No `RegisterFactory` call is made for engine-backed stripe/searxng, and
   `native/postgres` has **no init() registration** in wave0 (its file carries a
   `// wave6: RegisterFactory("postgres", New)` marker). This avoids the
   `RegisterFactory` overwrite semantics (`connectors.go:60`) silently flipping a golden.
3. **The flip is a wave6 change**: registry rebuild from `engine.LoadAll(defs.FS)` + nativeset,
   legacy deletion, naming clean break — behind the roadmap HUMAN GATE.
4. Rollback consequence: reverting wave0 removes only additive packages/files + the registrygen
   skip-map entries; production behavior is bit-identical before/after (see `RUNBOOK.md`).

## 3. Out of scope for wave0 (explicit)

- Any deletion: `slug.go`, `catalog_data.json`, `manifest.go` types, `native_conformance.go`,
  `native_port.go`, `registryset`, live-registry split — all stay.
- `internal/app` changes (sync-mode validation consults derived modes only from wave6).
- CLI surface changes (no `pm connectors certify` subcommand, no capability filters, no bare-name
  clean break).
- Certify write/flow/schedule stages, batch mode, record/replay, `cmd/certstatus`.
- Pilot/fan-out connector migrations (waves 1–4), api_surface fail-first-run enforcement beyond the
  three goldens.
- New Go module dependencies of any kind (hard rule; a needed dep is a typed `NEEDS_NEW_DEP`
  blocker to the coordinator).

## 4. Acceptance (from `.planning/ROADMAP.md`, restated verbatim as the gate)

1. Engine unit tests green (interpolation, auth selection, pagination matrix, read/write paths,
   error mapping).
2. 3 goldens migrated with engine-vs-legacy parity tests passing: stripe, searxng, postgres.
3. `connectorgen validate` rejects seeded-invalid bundles; accepts the goldens.
4. Conformance v2 passes for the 3 goldens (static + httptest fixture replay).
5. Certify source stages pass against the `sample` connector end-to-end.
6. `go build ./... && go test ./... && golangci-lint run` green (plus existing `make verify`).

Quantitative exit metrics: `EVAL-PLAN.md`.

## 5. Human gates flagged for this phase

- **golangci-lint acquisition** (tool, not module dep): pin via
  `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@<pinned>` in the Makefile (no
  go.mod change, network at gate time) OR require a local binary install. Coordinator decides;
  neither adds a module dependency. golangci-lint is currently NOT on PATH in this environment.
- No dependency additions, schema migrations, production deploys, auth/security changes,
  destructive data actions, or quality-gate reductions are planned. If any task discovers one, it
  stops and escalates (stop conditions in `PLAN.md` §Rules).
