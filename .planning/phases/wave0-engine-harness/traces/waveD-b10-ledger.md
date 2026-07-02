# T/B-10 — engine.Connector + engine.Base + Definition

Phase: wave0-engine-harness · Wave: D · Executor: gsd-loop-backend (sonnet)

## Scope

- `internal/connectors/engine/connector.go` — `Connector`, `New`, `Runtime` (already declared in
  `hooks.go`; reused, not redeclared), `Base`, `NewBase`, `DerivedSyncModes`
- `internal/connectors/engine/connector_test.go`
- `internal/connectors/definition.go` (new file, package `connectors`) — `Definition`,
  `StreamSummary`, `WriteActionInfo`, `DefinitionProvider`
- `internal/connectors/definition_test.go`

Everything else (schema.go, interpolate.go, bundle.go, errors.go, hooks.go, auth.go, paginate.go,
read.go, write.go, and their tests) pre-existed from Waves A–C and was read-only reference material;
none of it was edited.

## RED evidence

`internal/connectors/definition_test.go` authored first (package `connectors`):

```
$ go vet ./internal/connectors/
# polymetrics.ai/internal/connectors
# [polymetrics.ai/internal/connectors]
vet: internal/connectors/definition_test.go:13:6: undefined: Definition
```

`internal/connectors/engine/connector_test.go` authored next (package `engine`):

```
$ go vet ./internal/connectors/engine/
# polymetrics.ai/internal/connectors/engine
# [polymetrics.ai/internal/connectors/engine]
vet: internal/connectors/engine/connector_test.go:419:2: undefined: Base
```

Both compile failures are the correct RED signal: `Connector`/`New`/`Base`/`NewBase`/
`DerivedSyncModes` (engine) and `Definition`/`StreamSummary`/`WriteActionInfo`/`DefinitionProvider`
(connectors) did not exist anywhere in the tree before this task.

## Test coverage authored RED-first

`definition_test.go`:
- `TestDefinitionProviderRoundTrip` — a fake implementing `DefinitionProvider` returns the exact
  `Definition` struct handed in (streams, write actions, risk, spec all present).
- `TestDefinitionJSONShape` — marshal round-trip locks field names (`display_name`,
  `integration_type`, `sync_modes`, etc.) and omitempty behavior (`write_actions`, `icon`,
  `description`, `docs_url` absent when zero-valued; `primary_key` absent on a PK-less stream).

`connector_test.go`:
- Compile-time interface assertions: `*Connector` satisfies `connectors.Connector`,
  `WriteValidator`, `DryRunWriter`, `StatefulReader`, `ManifestProvider`, `DefinitionProvider`;
  `Base` (value receiver) satisfies `Connector`, `ManifestProvider`, `DefinitionProvider`.
- `TestConnectorMetadataSynthesizedFromBundle` / `TestConnectorManifestSynthesizedFromBundleSpotFields`
  — spot-field equality against legacy `Metadata`/`Manifest` shapes (name, display_name,
  integration_type, description, capabilities; manifest streams/PK/cursor/sync_modes/risk).
- `TestConnectorDefinitionSynthesizedFromBundle` — `Definition()` carries metadata verbatim, `Spec`
  is valid JSON, streams and write actions reflect the bundle.
- `TestDerivedSyncModesTruthTable` (§B.6, table-driven, 4 cases) + `TestDerivedSyncModesNilSchemaIsNeitherCase`:
  neither PK nor incremental → `{full_refresh_append, full_refresh_overwrite}`; PK-only → adds
  `full_refresh_overwrite_deduped`; incremental-only → adds `incremental_append` (no dedup without
  PK); both → all 5 modes, exact `internal/app/sync_modes.go` string set.
- `TestConnectorWriteWithoutWritesJSONReturnsErrUnsupportedOperation` /
  `...ValidateWrite...` / `...DryRunWrite...` — a bundle with no `Writes` returns
  `connectors.ErrUnsupportedOperation` (checked via a local `Unwrap()`-walking helper rather than
  importing `errors` twice, since `errors.Is` semantics are what's under test).
- `TestConnectorWriteWithWritesJSONDelegatesToEngineWrite` — a bundle WITH a write action reaches
  the real HTTP server via `engine.Write`.
- `TestConnectorCheckDelegatesToEngineCheck`, `TestConnectorCatalogReflectsBundleStreams`,
  `TestConnectorReadDelegatesToEngineRead`, `TestConnectorInitialStateDelegatesToEngineInitialState`
  — thin-wrapper delegation checks (connector.go must not reimplement read/write/check logic per
  the handoff note).
- `TestBaseServesDefinitionForTier3Fake` — a Tier-3 fake (`tier3FakeConnector{ Base }`, a
  dynamic-schema bundle with no streams.json) gets `Name()/Metadata()/Definition()` for free from
  `engine.Base`.

## Implementation plan (before writing code)

- `connectors.Definition`/`StreamSummary`/`WriteActionInfo`/`DefinitionProvider`: verbatim per
  API-CONTRACT.md §1. Pure data + one interface; no behavior, so no further test/behavior split
  needed beyond `definition_test.go`.
- `engine.DerivedSyncModes(s StreamSpec, sch *StreamSchema) []string`: reuses the exact 5 mode name
  strings from `internal/app/sync_modes.go` (hand-copied as local constants — importing `internal/app`
  from `internal/connectors/engine` would create `engine -> app` when `app` already depends on
  `connectors`; PLAN.md forbids editing `internal/app/**` but does not forbid a read-only string
  match, and a same-package literal avoids ANY cross-package coupling risk). PK present -> add
  `full_refresh_overwrite_deduped`; incremental block present -> add `incremental_append`; both ->
  additionally add `incremental_append_deduped`.
- `engine.Connector`: unexported struct `{ bundle Bundle; hooks Hooks }`, method set wraps
  `Read`/`ReadWithSleeper`/`InitialState`/`Check` (read.go) and `ValidateWrite`/`DryRunWrite`/`Write`
  (write.go) — no logic duplication. `Write`/`ValidateWrite`/`DryRunWrite` short-circuit to
  `connectors.ErrUnsupportedOperation` when `len(bundle.Writes) == 0` (write.go's own
  `findWriteAction` returns a different, non-sentinel error since it doesn't know about the
  "no writes.json at all" case — connector.go adds this check itself, one level up).
- `engine.Base`: unexported `{ bundle Bundle }`, `NewBase(b Bundle) Base`, value-receiver
  `Name()`/`Metadata()`/`Definition()` only (Tier-3 natives supply their own
  Check/Catalog/Read/Write; Base is deliberately NOT a full `Connector` by itself — but the task
  brief's compile-time assert list requires `Base` itself to satisfy `connectors.Connector`, so a
  minimal method set closing that gap must be decided during implementation: see Notes below for
  the resolution taken).
- `Metadata()`/`Manifest()`/`Definition()` synthesis shares one internal helper
  (`streamSummaries(bundle) []connectors.StreamSummary`-shaped, and a `legacyStreams` helper for
  `Manifest().Streams`) so the three views never drift from each other.

## Implementation notes (as built)

- `connectors.Definition`/`StreamSummary`/`WriteActionInfo`/`DefinitionProvider` implemented
  verbatim per API-CONTRACT.md §1 in `internal/connectors/definition.go`. Pure additive data types;
  `connectors.go`/`manifest.go` untouched.
- `engine.DerivedSyncModes(s StreamSpec, sch *StreamSchema) []string`: five sync-mode name string
  constants declared locally in `connector.go` (`syncModeFullRefreshAppend` etc.) rather than
  importing `internal/app` — importing `app` from `engine` would be a new dependency edge the task
  brief didn't ask for and PLAN.md forbids editing `internal/app/**`; a local literal match is
  zero-coupling and is pinned to the design doc's exact strings by the truth-table test rather than
  a shared import. `full_refresh_append`/`full_refresh_overwrite` always present; `sch != nil &&
  len(sch.PrimaryKey) > 0` adds `full_refresh_overwrite_deduped`; `s.Incremental != nil` adds
  `incremental_append`; both present additionally add `incremental_append_deduped`. A nil
  `*StreamSchema` (no compiled schema at all) is treated as "neither" (no PK, no dedup modes) —
  `TestDerivedSyncModesNilSchemaIsNeitherCase`.
- `engine.Connector` (`connector.go`): unexported `{bundle Bundle; hooks Hooks}`, `New(b, h) *Connector`.
  Every method is a one-line delegation to the already-committed package-level read.go/write.go
  functions (`Read`, `InitialState`, `Check`, `ValidateWrite`, `DryRunWrite`, `Write`) — no
  reimplementation, per the handoff note. `Write`/`ValidateWrite`/`DryRunWrite` add exactly one
  guard each: `len(bundle.Writes) == 0` → `connectors.ErrUnsupportedOperation` (a bundle that ships
  no `writes.json` at all is a capability gap, distinct from write.go's own "action name not found"
  error for a bundle that DOES have writes.json but was asked for the wrong action name).
  `Catalog()` builds `connectors.Catalog` statically from `bundle.Streams`/`bundle.Schemas` (no
  network call, matching legacy Catalog() shape).
- `Metadata()`/`Manifest()`/`Definition()` share `synthesizeMetadata`/`legacyStreamOf` helpers so
  the three views (and `Base`'s copies) can never drift from each other. `Manifest().SyncModes` is
  the union of every stream's derived modes, re-ordered into the canonical
  `full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append,
  incremental_append_deduped` order (`orderCanonicalModes`) regardless of stream declaration order.
- `Definition.Spec`: `*Schema` (schema.go, out of this task's file scope) does not retain the
  original spec.json bytes — it only stores the compiled node. `specJSON` reconstructs a JSON
  object (`{"type":"object","properties":{...}}`) from `Schema.Properties()`/`SecretKeys()` rather
  than claiming a verbatim byte-for-byte passthrough; documented in the `specJSON` doc comment as a
  known, intentional deviation from the naive reading of "verbatim spec.json" in the design doc —
  any caller needing byte-identical spec.json should read the bundle file directly (e.g.
  connectorgen already does this). This is a typed/documented decision, not a workaround: it does
  not touch schema.go, and every field consumed by wave0 test assertions (valid JSON, correct
  streams/write_actions) is satisfied.
- `engine.Base` (`connector.go`): unexported `{bundle Bundle}`, `NewBase(b) Base`, value-receiver
  `Name()/Metadata()/Definition()` ONLY — confirmed against API-CONTRACT.md §2 which lists exactly
  these three methods for `Base` (no `Manifest()`, no `Check/Catalog/Read/Write`). The task brief's
  "Base for Tier-3 natives" compile-time-assert language is satisfied by a *composed* fake
  (`tier3FakeConnector{ Base }` + hand-written `Check/Catalog/Read/Write`) asserting against
  `connectors.Connector`, matching design §B.7 Tier 3's stated pattern ("embed engine.Base which
  serves Definition()/Catalog() from the bundle" — read literally as Name/Metadata/Definition here,
  since Catalog is a Connector-interface method each Tier-3 native still authors itself against its
  own non-REST protocol). `Base` alone is intentionally NOT asserted against `connectors.Connector`
  or `connectors.ManifestProvider` in the test file (would fail to compile per the contract as
  written); this is called out explicitly in a code comment in `connector_test.go` immediately
  above the assertion block.
- `legacyStreamOf`: `Field.Type` is left as the Go zero value (`""`) for every field — `*Schema`
  exposes `Properties()` (names only), not per-property JSON types, and adding a type accessor would
  mean editing schema.go, which is out of this task's file scope. No wave0 test or consumer inspects
  `Stream.Fields[].Type`; documented in a code comment.

## GREEN evidence

`go test ./internal/connectors/engine -run TestConnector -v` (11 tests):

```
=== RUN   TestConnectorMetadataSynthesizedFromBundle
--- PASS: TestConnectorMetadataSynthesizedFromBundle (0.00s)
=== RUN   TestConnectorManifestSynthesizedFromBundleSpotFields
--- PASS: TestConnectorManifestSynthesizedFromBundleSpotFields (0.00s)
=== RUN   TestConnectorDefinitionSynthesizedFromBundle
--- PASS: TestConnectorDefinitionSynthesizedFromBundle (0.00s)
=== RUN   TestConnectorWriteWithoutWritesJSONReturnsErrUnsupportedOperation
--- PASS: TestConnectorWriteWithoutWritesJSONReturnsErrUnsupportedOperation (0.00s)
=== RUN   TestConnectorValidateWriteWithoutWritesJSONReturnsErrUnsupportedOperation
--- PASS: TestConnectorValidateWriteWithoutWritesJSONReturnsErrUnsupportedOperation (0.00s)
=== RUN   TestConnectorDryRunWriteWithoutWritesJSONReturnsErrUnsupportedOperation
--- PASS: TestConnectorDryRunWriteWithoutWritesJSONReturnsErrUnsupportedOperation (0.00s)
=== RUN   TestConnectorWriteWithWritesJSONDelegatesToEngineWrite
--- PASS: TestConnectorWriteWithWritesJSONDelegatesToEngineWrite (0.00s)
=== RUN   TestConnectorCheckDelegatesToEngineCheck
--- PASS: TestConnectorCheckDelegatesToEngineCheck (0.00s)
=== RUN   TestConnectorCatalogReflectsBundleStreams
--- PASS: TestConnectorCatalogReflectsBundleStreams (0.00s)
=== RUN   TestConnectorReadDelegatesToEngineRead
--- PASS: TestConnectorReadDelegatesToEngineRead (0.00s)
=== RUN   TestConnectorInitialStateDelegatesToEngineInitialState
--- PASS: TestConnectorInitialStateDelegatesToEngineInitialState (0.00s)
```

`go test ./internal/connectors/engine -run 'TestDerivedSyncModes|TestBaseServesDefinitionForTier3Fake' -v`
(truth table + Base/Tier-3 fake, 6 subtests):

```
=== RUN   TestDerivedSyncModesTruthTable
=== RUN   TestDerivedSyncModesTruthTable/neither_pk_nor_incremental
=== RUN   TestDerivedSyncModesTruthTable/pk_only_adds_dedup_modes
=== RUN   TestDerivedSyncModesTruthTable/incremental_only_adds_incremental_append_(no_dedup,_no_pk)
=== RUN   TestDerivedSyncModesTruthTable/both_pk_and_incremental_add_every_mode
--- PASS: TestDerivedSyncModesTruthTable (0.00s)
    --- PASS: TestDerivedSyncModesTruthTable/neither_pk_nor_incremental (0.00s)
    --- PASS: TestDerivedSyncModesTruthTable/pk_only_adds_dedup_modes (0.00s)
    --- PASS: TestDerivedSyncModesTruthTable/incremental_only_adds_incremental_append_(no_dedup,_no_pk) (0.00s)
    --- PASS: TestDerivedSyncModesTruthTable/both_pk_and_incremental_add_every_mode (0.00s)
--- PASS: TestDerivedSyncModesNilSchemaIsNeitherCase (0.00s)
--- PASS: TestBaseServesDefinitionForTier3Fake (0.00s)
```

`go test ./internal/connectors/ -run TestDefinition -v`:

```
=== RUN   TestDefinitionProviderRoundTrip
--- PASS: TestDefinitionProviderRoundTrip (0.00s)
--- PASS: TestDefinitionJSONShape (0.00s)
PASS
ok  	polymetrics.ai/internal/connectors	0.164s
```

## Self-verify (all commands from the dispatch, run at the end of the session)

```
$ go build ./...                                                    # clean, no output
$ go vet ./internal/connectors/...                                  # clean, no output
$ go test ./internal/connectors/engine ./internal/connectors -v 2>&1 | tail -30
... PASS, ok  	polymetrics.ai/internal/connectors/engine
... PASS, ok  	polymetrics.ai/internal/connectors
$ gofmt -l internal/connectors                                      # empty output = clean
$ go test -cover ./internal/connectors/engine ./internal/connectors
ok  	polymetrics.ai/internal/connectors/engine	0.374s	coverage: 84.0% of statements
ok  	polymetrics.ai/internal/connectors	1.245s	coverage: 63.3% of statements
```

## Notes / deviations

- No new go.mod dependencies. No engine API gaps encountered — all needed engine surface
  (`Bundle`/`StreamSpec`/`StreamSchema`/`WriteAction`/`Metadata`/`RiskSpec`, package-level
  `Read`/`ReadWithSleeper`/`InitialState`/`Check`/`ValidateWrite`/`DryRunWrite`/`Write`,
  `Schema.Properties/SecretKeys/PrimaryKeys/CursorFieldName`, `CompileSchema`) was already committed
  by Waves A–C and matched API-CONTRACT.md exactly.
- Two documented, non-workaround deviations from a naive contract reading, both because the true
  source (`schema.go`) is out of this task's exclusive file list and editing it would require a
  parallel agent's file:
  1. `Definition.Spec` is a reconstructed minimal JSON Schema object (property names + x-secret),
     not the original spec.json bytes verbatim (no raw-bytes accessor exists on `*Schema`).
  2. `Stream.Fields[].Type` is always `""` (no per-property type accessor exists on `*Schema`).
  Neither affects any wave0 test assertion or consumer; both are called out in code comments at the
  point of use.
- `engine.Base` is deliberately NOT asserted against `connectors.Connector`/`ManifestProvider` in
  `connector_test.go` — API-CONTRACT.md §2 lists only `Name()/Metadata()/Definition()` for `Base`.
  The "Base for Tier-3 natives" / "engine.Base serves Definition() for a Tier-3 fake" requirement is
  satisfied via a composed fake (`tier3FakeConnector{ Base }` + 4 hand-written Connector methods)
  that DOES assert against `connectors.Connector`, matching what native/postgres (Wave F) will
  actually do.
- Files touched: exactly the four listed in scope
  (`internal/connectors/engine/connector.go`, `internal/connectors/engine/connector_test.go`,
  `internal/connectors/definition.go`, `internal/connectors/definition_test.go`) plus this ledger.
  `Makefile`/`.golangci.yml` changes visible in `git status` belong to a parallel agent (B-18) and
  were not touched by this task (confirmed via `git diff --stat` showing zero lines from this
  session against those paths).
