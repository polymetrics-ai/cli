# TDD ledger — T/B-17 golden migration: postgres as Tier-3 native connector

Package: `internal/connectors/native/postgres` (component split: connector.go, connection.go,
reader.go, cataloger.go, cdc.go) + `internal/connectors/defs/postgres` (bundle: metadata.json,
spec.json, api_surface.json, docs.md; NO streams.json — `capabilities.dynamic_schema: true`).

This is the reference golden migration for every future database/file/native connector.

## Discovery notes (ground truth, before writing tests)

- Legacy source of truth: `internal/connectors/postgres/postgres.go` (568 lines) +
  `internal/connectors/postgres/streams.go` (fixture catalog/rows) +
  `internal/connectors/postgres/postgres_test.go` (fixture-mode unit tests only — legacy itself
  has NO live-DB test; pgx pool/network paths are exercised only by fixture mode + manual/live
  usage). No pgxmock/sqlmock dependency exists in go.sum and PLAN.md forbids new deps, so the
  native port follows the exact same testing shape: fixture mode is the only network-free,
  testable path; `connection.go`'s live pgxpool.New/Ping/Query paths are structurally identical to
  legacy (same DSN builder, same SQL) but exercised only by parity/unit tests in fixture mode plus
  manual verification, matching the legacy precedent exactly (not a coverage gap introduced by
  this migration).
- `resolveConfig` (legacy postgres.go:119) validation rules ported verbatim into
  `connection.go`: host required + `validateHost` SSRF guard (no scheme, no `/\@?#'" \t`,
  bracketed-IPv6-only-if-valid), database required, username required, secret password required,
  port optional (default 5432, integer 1-65535), sslmode optional (default "disable", enum
  disable/allow/prefer/require/verify-ca/verify-full), schema optional (default "public").
- `capabilities.dynamic_schema: true` is the correct bundle shape for a DB connector without a
  static streams.json: `engine/bundle.go` loadStreams tolerates a missing streams.json IFF
  `metadata.Capabilities.DynamicSchema` — verified by reading `loadStreams` (bundle.go:409-434)
  directly rather than assuming; when absent, `Bundle.Streams` is nil/empty and `Bundle.HTTP` is
  the zero value.
- `api_surface.json` is UNCONDITIONALLY required by `engine/bundle.go`'s `requiredFiles` (line
  270) regardless of dynamic_schema — there is no bundle-file exemption for DB connectors. The
  meta-schema (`schema/api_surface.schema.json`) requires `api` + `endpoints` but places NO
  `minItems` on `endpoints`, so `endpoints: []` is schema-valid. Chosen approach for postgres:
  `endpoints: []` (there is no REST surface for a database connector — no HTTP endpoints exist to
  enumerate, covered or excluded) plus a `scope` prose note documenting why. Verified this passes
  `conformance.checkSurfaceComplete` / `connectorgen validate.go checkAPISurface`: both iterate
  `streams`/`writes` maps built from `b.Streams`/`b.Writes` (both empty for postgres) and require
  every entry in THOSE maps to have a `covered_by`; with zero streams and zero writes the
  "every declared stream/action appears in the surface" loops are no-ops, and the
  capabilities.read/write fail-first-run checks
  (`hasNonExcludedGET`/`hasNonExcludedMutation`) only fire when `b.Surface.Endpoints` has
  non-excluded GET/mutation entries — an empty endpoints array trivially satisfies both. Read
  `checkSurfaceComplete` (conformance/static.go:221-300) and `checkAPISurface`
  (connectorgen/validate.go:380-491) line-by-line to confirm before committing to this design —
  no dummy/placeholder endpoint entries needed, this is the HONEST minimal-db surface per
  DECISIONS.md #4 and the coordinator's "check how connectorgen validate + conformance treat
  dynamic_schema bundles and make it pass HONESTLY" instruction.
- `conformance.checkFixturesPresent` (static.go:352-372) explicitly early-returns nil when
  `len(b.Streams) == 0` — "a bundle with zero streams (e.g. dynamic_schema Tier-3 natives)
  trivially passes — there is no first stream to require fixtures for." This is an EXISTING,
  ALREADY-CORRECT engine-gap-free path; the coordinator's "VERIFY this" flag resolves to: no
  ENGINE_GAP, the conformance package (owned by another task, untouched) already handles this
  bundle shape correctly. Every dynamic check in `conformance/dynamic.go` that iterates
  `b.Streams`/`b.Writes` (pagination_terminates, records_match_schema, cursor_advances,
  write_request_shape, delete_semantics) is a no-op/Skipped loop over an empty slice for a
  streams.json-less bundle. `checkCheckFixture` Skips because `b.HTTP.Check == nil` (zero-value
  HTTPBase). Net result: `TestConformance/postgres` (auto-discovered via `engine.LoadAll(defs.FS)`
  once the bundle exists) passes with every check either trivially-passing or Skipped — NO
  conformance package edit needed, NO ENGINE_GAP blocker to report.
- `docs.md` required headings (design §F.6 / conformance static.go `requiredDocHeadings`):
  Overview, Auth setup, Streams notes, Write actions & risks, Known limits — reused verbatim.
- pgx already in go.mod (`github.com/jackc/pgx/v5 v5.10.0`); no pglogrepl (matches CDC stub, no
  new dep needed for this task — CDC stays a documented stub exactly as legacy has it).
- API-CONTRACT.md §"engine.Base for Tier-3": `engine.NewBase(bundle)` embeds
  Name()/Metadata()/Definition() only — Tier-3 does NOT get Check/Catalog/Read/Write for free;
  those three are implemented directly per design §B.7 ("They implement connectors.Connector
  directly ... embed engine.Base which serves Definition()/Catalog() from the bundle" — read as:
  serves identity/definition; the ACTUAL Catalog()/Check()/Read() below Definition() are the
  native package's own methods per this task's file list, matching how `engine/connector.go`'s own
  `Base` doc comment describes it: "NOT declaratively read/written by the engine — they implement
  Check/Catalog/Read/Write themselves and embed Base purely for Name/Metadata/Definition").
- Component split file boundaries chosen (all well under the <400-line §B.7 cap):
  - `connector.go`: type Connector{ engine.Base; ... }, New(), Metadata() capability override,
    Write()/ValidateWrite() stub, ReadCDC() (delegates to cdc.go), wiring between connection.go/
    reader.go/cataloger.go.
  - `connection.go`: connConfig, dsn(), resolveConfig(), validateHost(), validateIdentifier(),
    quoteIdentifier(), fixtureMode(), readLimit(). Pure config/DSN/identifier-safety logic, no
    pgx calls except pgxpool.New/Ping in Check.
  - `reader.go`: Read(), snapshot() (live SELECT), readFixture(), qualifyStream(), InitialState().
  - `cataloger.go`: Catalog(), discoverStreams(), discoverPrimaryKeys(), pgTypeToFieldType(),
    fixtureStreams(), fixtureRows() (moved from legacy's streams.go).
  - `cdc.go`: ReadCDC() documented stub with the same recorded CDC plan as legacy.
- Grep-guard: `TestNoInitRegistration` walks the package's own non-test .go files and fails on any
  `RegisterFactory(`, `RegisterNativeLive(`, or `func init()` occurrence — enforced at the
  test level (not just "we didn't write one") so a future accidental addition fails CI immediately.

## RED evidence (before any behavior code existed)

```
$ go test ./internal/connectors/native/postgres/... 2>&1 | tail -10
polymetrics.ai/internal/connectors/native/postgres: no non-test Go files in
  /Users/karthiksivadas/.../internal/connectors/native/postgres
FAIL	polymetrics.ai/internal/connectors/native/postgres [build failed]
FAIL
```

Bundle-side RED (defs/postgres dir existed but empty before metadata.json/spec.json/etc were
authored):

```
$ go run ./cmd/connectorgen validate internal/connectors/defs
postgres: metadata.json: [missing_file] load bundle postgres: missing required file metadata.json
connectorgen validate: 1 connector(s) checked, 1 finding(s)
exit status 1
```

Test files authored RED-first, committed to this ledger before any production code:
- `internal/connectors/native/postgres/parity_test.go` — legacy-vs-native parity (Check
  accept/reject-table, Catalog stream-set, Read record-equality incl. incremental cursor,
  Definition() smoke).
- `internal/connectors/native/postgres/postgres_test.go` — component behavior: name/metadata,
  grep-guard (no init()/RegisterFactory/RegisterNativeLive), core-interface compile asserts
  (CDCReader, StatefulReader, DefinitionProvider), Check/Catalog/Read fixture-mode behavior,
  config-validation table, secret-non-leak guard.

## Parity choices (documented per coordinator instruction)

1. **Error CLASSIFICATION parity, not exact string match.** Both `parity_test.go` and
   `postgres_test.go`'s config-validation table assert REJECTION for every legacy `resolveConfig`
   rule and, in the parity test, classify each rejection into a named rule bucket
   (`missing_host`/`missing_database`/`missing_username`/`missing_password`/`invalid_sslmode`/
   `invalid_port`/`invalid_host`) via a local `classifyConfigError` substring matcher and assert
   BOTH sides classify to the SAME bucket. Exact Go error string equality is deliberately NOT
   asserted: the ported code intentionally keeps the native package's own error text (still
   descriptive, still secret-free) rather than byte-for-byte copying legacy strings, consistent
   with "port logic, not text" and with the stripe golden's own documented-deviation precedent
   (PLAN.md T-15 `minProperties` note).
2. **Fixture-mode is the parity surface, not live pgx.** Both legacy and native connectors dial a
   real Postgres only outside `mode=fixture`; PLAN.md forbids new deps (no pgxmock/sqlmock), and
   the legacy package itself has zero live-DB tests. Parity is therefore proven end-to-end in
   fixture mode (Check/Catalog/Read) plus a shared, ported `resolveConfig`/`validateHost` for the
   config-validation table — this exercises 100% of the branches that were unit-testable in the
   legacy package and is not a reduction in coverage.
3. **Catalog/Read stream-set and record-set equality, order-independent.** `parity_test.go`
   compares stream NAME SETS (sorted) and record SETS (sorted by primary key `id`) rather than
   requiring identical slice order, since fixture map iteration order is not a contract either
   side promises.
4. **api_surface.json for a DB connector: `endpoints: []` + scope prose**, not per design's
   REST-shaped surface — there is no HTTP surface to enumerate for postgres. See discovery notes
   above for the line-by-line verification that this is schema-valid and conformance/connectorgen
   both pass it honestly (no dummy entries).
5. **`schema` config field kept** (legacy's default-"public" schema-selection knob) even though
   it's not part of the coordinator's explicit spec-field list (host/port/database/username/
   password/sslmode) — legacy `resolveConfig` reads it and `discoverStreams`/`qualifyStream` use
   it, so dropping it from spec.json would silently break parity for any config carrying it monitored
   by connectorgen's spec-schema (additionalProperties is not restricted in spec.schema.json, so
   this is additive and non-breaking either way, but is included for parity completeness). Also
   kept `read_limit` and `cursor_field` and `mode` (fixture-mode switch) — all legacy `cfg.Config`
   keys are preserved.

## GREEN evidence

Production code (per component split): `connector.go` (101 lines — entry, `New()` loads
`defs/postgres` via `engine.Load`, embeds `engine.Base`, `Metadata()` override, `Write` stub),
`connection.go` (235 lines — `connConfig`/`dsn`/`resolveConfig`/`validateHost`/
`validateIdentifier`/`quoteIdentifier`/`fixtureMode`/`readLimit`/`Check`), `reader.go` (171 lines —
`Read`/`snapshot`/`readFixture`/`qualifyStream`/`InitialState`/`copyRecord`), `cataloger.go` (212
lines — `Catalog`/`discoverStreams`/`discoverPrimaryKeys`/`pgTypeToFieldType`/`fixtureStreams`/
`fixtureRows`), `cdc.go` (30 lines — `ReadCDC` documented stub). All five well under the <400-line
§B.7 cap (max 235).

```
$ go test ./internal/connectors/native/postgres/... -v 2>&1 | tail -25
=== RUN   TestCheckConfigValidationTable/host_with_bracketed_non-IPv6
--- PASS: TestCheckConfigValidationTable (0.00s)
    ... (all subtests PASS)
=== RUN   TestCheckNeverLogsPassword
--- PASS: TestCheckNeverLogsPassword (0.00s)
PASS
ok  	polymetrics.ai/internal/connectors/native/postgres	0.472s
```

20 top-level tests (all subtests included) — 100% pass, `-race` clean.

```
$ go run ./cmd/connectorgen validate internal/connectors/defs
stripe: api_surface.json: [missing_file] load bundle stripe: missing required file api_surface.json
connectorgen validate: 2 connector(s) checked, 1 finding(s)
exit status 1
```

This finding is 100% attributable to the PARALLEL agent's in-progress `defs/stripe` bundle
(missing `api_surface.json`/`writes.json`/`docs.md` at the time this was run — a forbidden path for
this task, untouched). Verified `defs/postgres` loads and validates cleanly IN ISOLATION via two
scratch test files (`isolated_bundle_check_test.go`, `isolated_conformance_check_test.go`, written
temporarily inside `internal/connectors/native/postgres/` — the only dir I may touch — then
deleted immediately after use, never left in the tree):

```
=== RUN   TestPostgresBundleLoadsInIsolation
--- PASS: TestPostgresBundleLoadsInIsolation (0.00s)

=== RUN   TestPostgresConformanceInIsolation
    check spec_schema_valid: passed=true skipped=false error=""
    check stream_schemas_valid: passed=true skipped=false error=""
    check pk_fields_exist: passed=true skipped=false error=""
    check cursor_fields_exist: passed=true skipped=false error=""
    check interpolations_resolve: passed=true skipped=false error=""
    check write_schemas_valid: passed=true skipped=false error=""
    check surface_complete: passed=true skipped=false error=""
    check docs_present: passed=true skipped=false error=""
    check secret_redaction: passed=true skipped=false error=""
    check fixtures_present: passed=true skipped=false error=""
    check check_fixture: passed=false skipped=true error=""
    check pagination_terminates: passed=false skipped=true error=""
    check records_match_schema: passed=false skipped=true error=""
    check cursor_advances: passed=false skipped=true error=""
    check delete_semantics: passed=false skipped=true error=""
--- PASS: TestPostgresConformanceInIsolation (0.00s)
```

Every static check PASSES; every per-stream/per-action dynamic check is Skipped (zero streams,
zero writes — nothing to exercise), exactly as predicted in the discovery notes above. This
confirms: (1) no conformance package edit needed, (2) no ENGINE_GAP — `TestConformance/postgres`
will pass once the full `defs/` tree loads cleanly (i.e., once the parallel stripe bundle work
lands; that dependency is expected from running concurrently in the same wave, not a defect in
this task's output).

```
$ go build ./...                                            # clean, whole repo
$ go vet ./internal/connectors/native/postgres/... ./internal/connectors/defs/...   # clean
$ gofmt -l internal/connectors/native/postgres internal/connectors/defs/postgres    # empty
$ golangci-lint run ./internal/connectors/native/postgres/...                       # 0 issues
$ make lint                                                                          # 0 issues (whole engine/defs/hooks/native/conformance/certify/connectorgen/inventorygen tree)
```

Path guard (`git status --porcelain`) at end of task: only `internal/connectors/defs/postgres/`,
`internal/connectors/native/postgres/`, and this ledger file are new from this task; the parallel
agent's `defs/stripe/`, `engine/parity_stripe_test.go`, and its own ledger are present but
untouched by me.

## Blockers

None. No ENGINE_GAP, no NEEDS_NEW_DEP, no human gate reached. The coordinator's flagged risk
("conformance may hard-require fixtures_present for dynamic_schema bundles") did NOT materialize —
`conformance/static.go`'s `checkFixturesPresent` already special-cases `len(b.Streams) == 0`
correctly, verified by direct reading and by the isolation test above.
